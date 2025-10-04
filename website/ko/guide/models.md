# Models - LLM 제공업체

Agno-Go는 통합 인터페이스로 여러 LLM 제공업체를 지원합니다.

---

## 지원 모델

### OpenAI
- GPT-4o, GPT-4o-mini, GPT-4 Turbo, GPT-3.5 Turbo
- 완전한 스트리밍 지원
- 함수 호출

### Anthropic Claude
- Claude 3.5 Sonnet, Claude 3 Opus, Claude 3 Sonnet, Claude 3 Haiku
- 스트리밍 지원
- 도구 사용

### Ollama
- 로컬 모델 실행 (Llama, Mistral 등)
- 프라이버시 우선
- API 비용 없음

---

## OpenAI

### 설정

```go
import "github.com/rexleimo/agno-go/pkg/agno/models/openai"

model, err := openai.New("gpt-4o-mini", openai.Config{
    APIKey:      os.Getenv("OPENAI_API_KEY"),
    Temperature: 0.7,
    MaxTokens:   1000,
})
```

### 구성

```go
type Config struct {
    APIKey      string  // 필수: OpenAI API 키
    BaseURL     string  // 선택: 커스텀 엔드포인트 (기본값: https://api.openai.com/v1)
    Temperature float64 // 선택: 0.0-2.0 (기본값: 0.7)
    MaxTokens   int     // 선택: 최대 응답 토큰
}
```

### 지원 모델

| 모델 | 컨텍스트 | 최적 용도 |
|-------|---------|----------|
| `gpt-4o` | 128K | 가장 강력한, 멀티모달 |
| `gpt-4o-mini` | 128K | 빠르고, 비용 효율적 |
| `gpt-4-turbo` | 128K | 고급 추론 |
| `gpt-3.5-turbo` | 16K | 간단한 작업, 빠름 |

### 예제

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models/openai"
)

func main() {
    model, err := openai.New("gpt-4o-mini", openai.Config{
        APIKey:      os.Getenv("OPENAI_API_KEY"),
        Temperature: 0.7,
    })
    if err != nil {
        log.Fatal(err)
    }

    agent, _ := agent.New(agent.Config{
        Name:  "Assistant",
        Model: model,
    })

    output, _ := agent.Run(context.Background(), "Hello!")
    fmt.Println(output.Content)
}
```

---

## Anthropic Claude

### 설정

```go
import "github.com/rexleimo/agno-go/pkg/agno/models/anthropic"

model, err := anthropic.New("claude-3-5-sonnet-20241022", anthropic.Config{
    APIKey:    os.Getenv("ANTHROPIC_API_KEY"),
    MaxTokens: 2048,
})
```

### 구성

```go
type Config struct {
    APIKey      string  // 필수: Anthropic API 키
    Temperature float64 // 선택: 0.0-1.0
    MaxTokens   int     // 선택: 최대 응답 토큰 (기본값: 4096)
}
```

### 지원 모델

| 모델 | 컨텍스트 | 최적 용도 |
|-------|---------|----------|
| `claude-3-5-sonnet-20241022` | 200K | 가장 지능적, 코딩 |
| `claude-3-opus-20240229` | 200K | 복잡한 작업 |
| `claude-3-sonnet-20240229` | 200K | 균형잡힌 성능 |
| `claude-3-haiku-20240307` | 200K | 빠른 응답 |

### 예제

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models/anthropic"
)

func main() {
    model, err := anthropic.New("claude-3-5-sonnet-20241022", anthropic.Config{
        APIKey:    os.Getenv("ANTHROPIC_API_KEY"),
        MaxTokens: 2048,
    })
    if err != nil {
        log.Fatal(err)
    }

    agent, _ := agent.New(agent.Config{
        Name:         "Claude",
        Model:        model,
        Instructions: "You are a helpful assistant.",
    })

    output, _ := agent.Run(context.Background(), "Explain quantum computing")
    fmt.Println(output.Content)
}
```

---

## Ollama (로컬 모델)

### 설정

1. Ollama 설치: https://ollama.ai
2. 모델 다운로드: `ollama pull llama2`
3. Agno-Go에서 사용:

```go
import "github.com/rexleimo/agno-go/pkg/agno/models/ollama"

model, err := ollama.New("llama2", ollama.Config{
    BaseURL: "http://localhost:11434",  // Ollama 서버
})
```

### 구성

```go
type Config struct {
    BaseURL     string  // 선택: Ollama 서버 URL (기본값: http://localhost:11434)
    Temperature float64 // 선택: 0.0-1.0
}
```

### 지원 모델

Ollama에서 사용 가능한 모든 모델:
- `llama2`, `llama3`, `llama3.1`
- `mistral`, `mixtral`
- `codellama`, `deepseek-coder`
- `qwen2`, `gemma2`

### 예제

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models/ollama"
)

func main() {
    // Ollama가 실행 중이고 모델이 다운로드되어 있는지 확인
    model, err := ollama.New("llama2", ollama.Config{
        BaseURL: "http://localhost:11434",
    })
    if err != nil {
        log.Fatal(err)
    }

    agent, _ := agent.New(agent.Config{
        Name:  "Local Assistant",
        Model: model,
    })

    output, _ := agent.Run(context.Background(), "What is Go?")
    fmt.Println(output.Content)
}
```

---

## 모델 비교

### 성능

| 제공업체 | 속도 | 비용 | 프라이버시 | 컨텍스트 |
|----------|-------|------|---------|---------|
| OpenAI GPT-4o-mini | ⚡⚡⚡ | 💰 | ☁️ 클라우드 | 128K |
| OpenAI GPT-4o | ⚡⚡ | 💰💰💰 | ☁️ 클라우드 | 128K |
| Anthropic Claude | ⚡⚡ | 💰💰 | ☁️ 클라우드 | 200K |
| Ollama | ⚡ | 🆓 무료 | 🏠 로컬 | 다양 |

### 각 모델 사용 시기

**OpenAI GPT-4o-mini**
- 개발 및 테스트
- 대용량 애플리케이션
- 비용에 민감한 사용 사례

**OpenAI GPT-4o**
- 복잡한 추론 작업
- 멀티모달 애플리케이션
- 프로덕션 시스템

**Anthropic Claude**
- 긴 컨텍스트 필요 (200K 토큰)
- 코딩 지원
- 복잡한 분석

**Ollama**
- 프라이버시 요구사항
- 인터넷 연결 없음
- API 비용 제로
- 개발/테스트

---

## 모델 전환

모델 전환은 쉽습니다:

```go
// OpenAI
openaiModel, _ := openai.New("gpt-4o-mini", openai.Config{
    APIKey: os.Getenv("OPENAI_API_KEY"),
})

// Claude
claudeModel, _ := anthropic.New("claude-3-5-sonnet-20241022", anthropic.Config{
    APIKey: os.Getenv("ANTHROPIC_API_KEY"),
})

// Ollama
ollamaModel, _ := ollama.New("llama2", ollama.Config{})

// 동일한 에이전트 코드 사용
agent, _ := agent.New(agent.Config{
    Model: openaiModel,  // 또는 claudeModel, 또는 ollamaModel
})
```

---

## 고급 구성

### Temperature

무작위성 제어 (0.0 = 결정론적, 1.0+ = 창의적):

```go
model, _ := openai.New("gpt-4o-mini", openai.Config{
    Temperature: 0.0,  // 일관된 응답
})

model, _ := openai.New("gpt-4o-mini", openai.Config{
    Temperature: 1.5,  // 창의적인 응답
})
```

### Max Tokens

응답 길이 제한:

```go
model, _ := openai.New("gpt-4o-mini", openai.Config{
    MaxTokens: 500,  // 짧은 응답
})
```

### 커스텀 엔드포인트

호환 가능한 API 사용:

```go
model, _ := openai.New("gpt-4o-mini", openai.Config{
    BaseURL: "https://your-proxy.com/v1",  // 커스텀 엔드포인트
    APIKey:  "your-key",
})
```

---

## 모범 사례

### 1. 환경 변수

API 키를 안전하게 저장:

```go
// 좋음 ✅
APIKey: os.Getenv("OPENAI_API_KEY")

// 나쁨 ❌
APIKey: "sk-proj-..." // 하드코딩
```

### 2. 오류 처리

항상 오류 확인:

```go
model, err := openai.New("gpt-4o-mini", openai.Config{
    APIKey: os.Getenv("OPENAI_API_KEY"),
})
if err != nil {
    log.Fatalf("Failed to create model: %v", err)
}
```

### 3. 모델 선택

필요에 따라 선택:

```go
// 개발: 빠르고 저렴
devModel, _ := openai.New("gpt-4o-mini", config)

// 프로덕션: 더 강력함
prodModel, _ := openai.New("gpt-4o", config)
```

### 4. 컨텍스트 관리

컨텍스트 제한에 주의:

```go
// 긴 대화의 경우, 주기적으로 메모리 지우기
if messageCount > 50 {
    agent.ClearMemory()
}
```

---

## 환경 설정

`.env` 파일 생성:

```bash
# OpenAI
OPENAI_API_KEY=sk-proj-...

# Anthropic
ANTHROPIC_API_KEY=sk-ant-...

# Ollama (선택, 기본값은 localhost)
OLLAMA_BASE_URL=http://localhost:11434
```

코드에서 로드:

```go
import "github.com/joho/godotenv"

func init() {
    godotenv.Load()
}
```

---

## 다음 단계

- 모델 능력 향상을 위한 [Tools](/guide/tools) 추가
- 대화 히스토리를 위한 [Memory](/guide/memory) 배우기
- 혼합 모델로 [Teams](/guide/team) 구축
- 실제 사용을 위한 [Examples](/examples/) 탐색

---

## 관련 예제

- [Simple Agent](/examples/simple-agent) - OpenAI 예제
- [Claude Agent](/examples/claude-agent) - Anthropic 예제
- [Ollama Agent](/examples/ollama-agent) - 로컬 모델 예제
