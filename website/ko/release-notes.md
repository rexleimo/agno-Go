---
title: 릴리스 노트
description: Agno-Go의 버전 히스토리 및 릴리스 노트
outline: deep
---

# 릴리스 노트

## 버전 1.2.1 (2025-10-15)

### 🧭 문서 재구성

- 명확한 분리:
  - `website/` → 구현된 대외 문서 (VitePress 사이트)
  - `docs/` → 설계 초안, 마이그레이션 계획, 태스크, 개발자/내부 문서
- 정책과 진입점을 담은 `docs/README.md` 추가
- 기여자 온보딩을 위한 `CONTRIBUTING.md` 추가

### 🔗 링크 수정

- README, CLAUDE, CHANGELOG, 릴리스 노트 링크를 `website/advanced/*`, `website/guide/*`로 정규화
- `docs/` 하위 중복 구현 문서로의 오래된 링크 제거

### 🌐 사이트 업데이트

- API: AgentOS 페이지에 지식 API 추가 (/api/agentos)
- Workflow History, Performance 페이지를 표준 참조로 통일

### ✅ 동작 변경

- 없음 (문서/구조 조정만 포함)

### ✨ 신규 (이번 릴리스에서 구현됨)

- A2A 스트리밍 이벤트 유형 필터 (SSE)
  - `POST /api/v1/agents/:id/run/stream?types=token,complete`
  - 요청한 이벤트만 출력; 표준 SSE 형식; Context 취소 지원
- AgentOS 컨텐츠 추출 미들웨어
  - JSON/Form의 `content/metadata/user_id/session_id`를 Context로 주입
  - `MaxRequestSize` 크기 보호 및 경로 스킵 지원
- Google Sheets 도구 (서비스 계정)
  - `read_range`, `write_range`, `append_rows`; JSON/파일 자격 증명 지원
- 최소 지식 적재 엔드포인트
  - `POST /api/v1/knowledge/content` 는 `text/plain` 및 `application/json` 지원

엔터프라이즈 검수 절차: [`docs/ENTERPRISE_MIGRATION_PLAN.md`](https://github.com/rexleimo/agno-Go/blob/main/docs/ENTERPRISE_MIGRATION_PLAN.md) 참고.

## 버전 1.1.0 (2025-10-08)

### 🎉 하이라이트

이번 릴리스는 프로덕션 준비 멀티 에이전트 시스템을 위한 강력한 새 기능을 제공합니다:

- **A2A 인터페이스** - 표준화된 에이전트 간 통신 프로토콜
- **세션 상태 관리** - 워크플로우 단계 간 영구 상태
- **멀티 테넌트 지원** - 단일 에이전트 인스턴스로 여러 사용자에게 서비스 제공
- **모델 타임아웃 구성** - LLM 호출을 위한 세밀한 타임아웃 제어

---

### ✨ 새로운 기능

#### A2A (Agent-to-Agent) 인터페이스

JSON-RPC 2.0 기반의 에이전트 간 상호작용을 위한 표준화된 통신 프로토콜.

**주요 기능:**
- RESTful API 엔드포인트 (`/a2a/message/send`, `/a2a/message/stream`)
- 멀티미디어 지원 (텍스트, 이미지, 파일, JSON 데이터)
- 스트리밍을 위한 Server-Sent Events (SSE)
- Python Agno A2A 구현과 호환

**빠른 예제:**
```go
import "github.com/rexleimo/agno-go/pkg/agentos/a2a"

// A2A 인터페이스 생성
a2a := a2a.New(a2a.Config{
    Agents: []a2a.Entity{myAgent},
    Prefix: "/a2a",
})

// 라우트 등록 (Gin)
router := gin.Default()
a2a.RegisterRoutes(router)
```

📚 **자세히 알아보기:** [A2A 인터페이스 문서](/ko/api/a2a)

---

#### 워크플로우 세션 상태 관리

워크플로우 단계 간 상태를 유지하기 위한 스레드 안전 세션 관리.

**주요 기능:**
- 단계 간 영구 상태 저장소
- `sync.RWMutex`를 사용한 스레드 안전성
- 병렬 브랜치 격리를 위한 딥 카피
- 데이터 손실을 방지하는 스마트 병합 전략
- Python Agno v2.1.2의 경쟁 조건 수정

**빠른 예제:**
```go
// 세션 정보로 컨텍스트 생성
execCtx := workflow.NewExecutionContextWithSession(
    "input",
    "session-123",  // 세션 ID
    "user-a",       // 사용자 ID
)

// 세션 상태에 액세스
execCtx.SetSessionState("key", "value")
value, _ := execCtx.GetSessionState("key")
```

📚 **자세히 알아보기:** [세션 상태 문서](/ko/guide/session-state)

---

#### 멀티 테넌트 지원

완전한 데이터 격리를 보장하면서 단일 Agent 인스턴스로 여러 사용자에게 서비스 제공.

**주요 기능:**
- 사용자별로 격리된 대화 기록
- Memory 인터페이스의 선택적 `userID` 매개변수
- 기존 코드와의 하위 호환성
- 스레드 안전 동시 작업
- 정리를 위한 `ClearAll()` 메서드

**빠른 예제:**
```go
// 멀티 테넌트 에이전트 생성
agent, _ := agent.New(&agent.Config{
    Name:   "customer-service",
    Model:  model,
    Memory: memory.NewInMemory(100),
})

// 사용자 A의 대화
agent.UserID = "user-a"
output, _ := agent.Run(ctx, "My name is Alice")

// 사용자 B의 대화
agent.UserID = "user-b"
output, _ := agent.Run(ctx, "My name is Bob")
```

📚 **자세히 알아보기:** [멀티 테넌트 문서](/ko/advanced/multi-tenant)

---

#### 모델 타임아웃 구성

세밀한 제어로 LLM 호출에 대한 요청 타임아웃 구성.

**주요 기능:**
- 기본값: 60초
- 범위: 1초에서 10분
- 지원 모델: OpenAI, Anthropic Claude
- 컨텍스트 인식 타임아웃 처리

**빠른 예제:**
```go
// 사용자 정의 타임아웃이 있는 OpenAI
model, _ := openai.New("gpt-4", openai.Config{
    APIKey:  apiKey,
    Timeout: 30 * time.Second,
})

// 사용자 정의 타임아웃이 있는 Claude
claude, _ := anthropic.New("claude-3-opus", anthropic.Config{
    APIKey:  apiKey,
    Timeout: 45 * time.Second,
})
```

📚 **자세히 알아보기:** [모델 구성](/ko/guide/models#timeout-configuration)

---

### 🐛 버그 수정

- **워크플로우 경쟁 조건** - 병렬 단계 실행 데이터 경쟁 수정
  - Python Agno v2.1.2에는 공유 `session_state` dict로 인한 덮어쓰기 문제가 있었습니다
  - Go 구현은 브랜치당 독립적인 SessionState 복제본을 사용합니다
  - 스마트 병합 전략으로 동시 실행에서 데이터 손실 방지

---

### 📚 문서

모든 새로운 기능에는 포괄적인 이중 언어 문서 (영어/중문)가 포함되어 있습니다:

- [A2A 인터페이스 가이드](/ko/api/a2a) - 완전한 프로토콜 사양
- [세션 상태 가이드](/ko/guide/session-state) - 워크플로우 상태 관리
- [멀티 테넌트 가이드](/ko/advanced/multi-tenant) - 데이터 격리 패턴
- [모델 구성](/ko/guide/models#timeout-configuration) - 타임아웃 설정

---

### 🧪 테스트

**새로운 테스트 스위트:**
- `session_state_test.go` - 세션 상태 테스트 543줄
- `memory_test.go` - 멀티 테넌트 메모리 테스트 (새 테스트 케이스 4개)
- `agent_test.go` - 멀티 테넌트 에이전트 테스트
- `openai_test.go` - 타임아웃 구성 테스트
- `anthropic_test.go` - 타임아웃 구성 테스트

**테스트 결과:**
- ✅ 모든 테스트가 `-race` 감지기로 통과
- ✅ 워크플로우 커버리지: 79.4%
- ✅ 메모리 커버리지: 93.1%
- ✅ 에이전트 커버리지: 74.7%

---

### 📊 성능

**성능 저하 없음** - 모든 벤치마크가 일관됩니다:
- Agent 인스턴스화: ~180ns/op (Python보다 16배 빠름)
- 메모리 사용량: ~1.2KB/에이전트
- 스레드 안전 동시 작업

---

### ⚠️ 호환성 문제

**없음.** 이번 릴리스는 v1.0.x와 완전히 하위 호환됩니다.

---

### 🔄 마이그레이션 가이드

**마이그레이션 불필요** - 모든 새 기능은 추가 기능이며 하위 호환됩니다.

**선택적 개선 사항:**

1. **멀티 테넌트 지원 활성화:**
   ```go
   // 에이전트 구성에 UserID 추가
   agent := agent.New(agent.Config{
       UserID: "user-123",  // NEW
       Memory: memory.NewInMemory(100),
   })
   ```

2. **워크플로우에서 세션 상태 사용:**
   ```go
   // 세션이 있는 컨텍스트 생성
   ctx := workflow.NewExecutionContextWithSession(
       "input",
       "session-id",
       "user-id",
   )
   ```

3. **모델 타임아웃 구성:**
   ```go
   // 모델 구성에 Timeout 추가
   model, _ := openai.New("gpt-4", openai.Config{
       APIKey:  apiKey,
       Timeout: 30 * time.Second,  // NEW
   })
   ```

---

### 📦 설치

```bash
go get github.com/rexleimo/agno-go@v1.1.0
```

---

### 🔗 링크

- **GitHub 릴리스:** [v1.1.0](https://github.com/rexleimo/agno-go/releases/tag/v1.1.0)
- **전체 변경 로그:** [CHANGELOG.md](https://github.com/rexleimo/agno-go/blob/main/CHANGELOG.md)
- **문서:** [https://agno-go.dev](https://agno-go.dev)

---

## 버전 1.0.3 (2025-10-06)

### 🧪 개선

- **JSON 직렬화 테스트 강화** - utils/serialize 패키지에서 100% 테스트 커버리지 달성
- **성능 벤치마크** - Python Agno 성능 테스트 패턴과 정렬
- **포괄적인 문서** - 이중 언어 패키지 문서 추가

---

## 버전 1.0.2 (2025-10-05)

### ✨ 추가

#### GLM (智谱AI) 프로바이더

- Zhipu AI의 GLM 모델과 완전 통합
- GLM-4, GLM-4V (비전), GLM-3-Turbo 지원
- 사용자 정의 JWT 인증 (HMAC-SHA256)
- 동기 및 스트리밍 API 호출
- 도구/함수 호출 지원

---

## 버전 1.0.0 (2025-10-02)

### 🎉 초기 릴리스

Agno-Go v1.0은 Agno 멀티 에이전트 프레임워크의 고성능 Go 구현입니다.

#### 핵심 기능
- **Agent** - 도구 지원이 있는 단일 자율 에이전트
- **Team** - 4가지 모드를 통한 멀티 에이전트 협력
- **Workflow** - 5가지 프리미티브를 통한 단계 기반 오케스트레이션

#### LLM 프로바이더
- OpenAI (GPT-4, GPT-3.5, GPT-4 Turbo)
- Anthropic (Claude 3.5 Sonnet, Claude 3 Opus/Sonnet/Haiku)
- Ollama (로컬 모델)

---

**최종 업데이트:** 2025-10-08
