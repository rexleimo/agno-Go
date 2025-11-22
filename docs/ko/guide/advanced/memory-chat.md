
# 고급 가이드: 메모리 강화 챗봇

이 가이드는 여러 턴과 세션에 걸쳐 메모리를 활용하는 챗 경험을 구축하는 방법을 설명합니다. 특정 스토리지 제품에 의존하지 않고, 기존 HTTP API와 메타데이터 필드를 어떻게 활용할 수 있는지를 중심으로 다룹니다.

## 1. 메모리 유형

메모리는 크게 세 가지 계층으로 생각할 수 있습니다.

- **대화 히스토리**: 현재 세션 내 최근 메시지들  
- **사용자 프로필**: 사용자에 대한 장기 정보(선호도, 프로필 필드 등)  
- **지식 레코드**: 도메인 특화 사실(과거 상호작용, 중요한 이벤트 등)  

Agno-Go는 Session과 Message를 통해 첫 번째 계층(대화 히스토리)을 기본 제공하며, 나머지 계층은 설정과 자체 서비스로 연결할 수 있습니다.

## 2. 메모리 지원 에이전트 생성

Go 애플리케이션에서는 보통 HTTP를 통해 AgentOS 런타임과 통신합니다. Quickstart와
동일한 흐름을 Go 코드로 구현하면 다음과 같습니다.

```go
package main

import (
  "bytes"
  "encoding/json"
  "log"
  "net/http"
  "time"
)

type Agent struct {
  Name        string                 `json:"name"`
  Description string                 `json:"description"`
  Model       map[string]any         `json:"model"`
  Tools       []map[string]any       `json:"tools"`
  Config      map[string]any         `json:"config"`
}

func main() {
  client := &http.Client{Timeout: 10 * time.Second}

  agent := Agent{
    Name:        "memory-chat-agent",
    Description: "A chat agent that uses session history and external memory.",
    Model: map[string]any{
      "provider": "openai",
      "modelId":  "gpt-4o-mini",
      "stream":   true,
    },
    Tools:  nil,
    Config: map[string]any{},
  }

  body, err := json.Marshal(agent)
  if err != nil {
    log.Fatalf("marshal agent: %v", err)
  }

  resp, err := client.Post("http://localhost:8080/agents", "application/json", bytes.NewReader(body))
  if err != nil {
    log.Fatalf("create agent: %v", err)
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusCreated {
    log.Fatalf("unexpected status: %s", resp.Status)
  }

  // 실제 애플리케이션에서는 여기서 응답을 decode 해서 agentId 를 얻은 뒤,
  // 이후 섹션에서 설명하는 것처럼 세션 생성 및 메시지 전송을 수행합니다.
}
```

터미널이나 API 클라이언트에서 HTTP 엔드포인트를 빠르게 시험해 보고 싶다면,
아래와 같이 동등한 `curl` 명령을 사용할 수 있습니다.

```bash
curl -X POST http://localhost:8080/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "memory-chat-agent",
    "description": "A chat agent that uses session history and external memory.",
    "model": {
      "provider": "openai",
      "modelId": "gpt-4o-mini",
      "stream": true
    },
    "tools": [],
    "config": {}
  }'
```

메모리 지원 에이전트인지 여부를 가르는 핵심은, 이후에 설명하는 것처럼 세션을
어떻게 구성하고 어떤 메타데이터를 전달하느냐에 있습니다.

## 3. 세션과 메타데이터 활용

세션을 생성할 때 사용자 식별자와 메타데이터를 함께 보낼 수 있습니다.

```bash
curl -X POST http://localhost:8080/agents/<agent-id>/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-1234",
    "metadata": {
      "source": "advanced-memory-chat",
      "segment": "beta-testers"
    }
  }'
```

애플리케이션은 `userId`와 `metadata`를 사용해 자체 스토리지에서 사용자 프로필을 조회하거나 갱신하고, 이 정보를 후속 메시지에 반영할 수 있습니다.

## 4. 프롬프트에 메모리 통합

메시지를 보낼 때 이미 알고 있는 사실과 컨텍스트를 `content`에 합쳐 보낼 수 있습니다.

```bash
curl -X POST "http://localhost:8080/agents/<agent-id>/sessions/<session-id>/messages" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "user",
    "content": "이전에 저를 위해 독서 계획을 추천해 줬죠. 그때의 제안 내용과 제가 짧은 독서 세션을 선호한다는 점을 반영해 이번 주 계획을 제안해 주세요."
  }'
```

백엔드에서는 다음과 같은 작업을 할 수 있습니다.

- 메모리 스토어에서 과거 상호작용이나 노트를 불러오기  
- 요약이나 핵심 사실을 프롬프트에 포함하기  
- 표준 메시지 엔드포인트를 통해 런타임에 전달하기  

## 5. 구성과 스토리지

메모리에 크게 의존하는 시나리오에서는:

- 어느 메모리 백엔드를 사용할지(인메모리 vs 로컬 영구 스토리지 등)를 “Configuration & Security Practices” 문서를 참고해 결정합니다.  
- 추가 인프라(데이터베이스, 캐시, 큐 등)는 내부 운영 문서에 기록하고, AgentOS 런타임은 HTTP 동작과 계약에 집중하도록 유지합니다.  
- `.env`와 `config/default.yaml` 설정이 공식 문서의 가이드(특히 보존 기간과 데이터 위치)에 맞는지 확인합니다.  

## 6. 테스트와 개선

메모리 강화 챗봇을 검증할 때는:

- 단기/장기 메모리 동작을 모두 포함하는 테스트 케이스를 설계합니다.  
- 메모리 사용량이 증가해도 `/health`와 Quickstart 플로우로 런타임의 안정성을 확인합니다.  
- 레이턴시와 리소스 사용량을 모니터링하면서 요약 빈도나 재생 길이 등 메모리 전략을 실제 지표에 따라 조정합니다.  
