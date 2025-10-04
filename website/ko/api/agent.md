# Agent API 레퍼런스

## agent.New

새로운 에이전트 인스턴스를 생성합니다.

**함수 시그니처:**
```go
func New(config Config) (*Agent, error)
```

**매개변수:**

```go
type Config struct {
    // 필수
    Model models.Model // 사용할 LLM 모델

    // 선택 사항
    Name         string            // 에이전트 이름 (기본값: "Agent")
    Toolkits     []toolkit.Toolkit // 사용 가능한 도구
    Memory       memory.Memory     // 대화 메모리
    Instructions string            // 시스템 지침
    MaxLoops     int               // 최대 도구 호출 루프 (기본값: 10)
}
```

**반환값:**
- `*Agent`: 생성된 에이전트 인스턴스
- `error`: 모델이 nil이거나 설정이 유효하지 않은 경우 에러

**예제:**
```go
model, _ := openai.New("gpt-4", openai.Config{APIKey: apiKey})

ag, err := agent.New(agent.Config{
    Name:         "Assistant",
    Model:        model,
    Toolkits:     []toolkit.Toolkit{calculator.New()},
    Instructions: "You are a helpful assistant.",
    MaxLoops:     15,
})
```

## Agent.Run

입력으로 에이전트를 실행합니다.

**함수 시그니처:**
```go
func (a *Agent) Run(ctx context.Context, input string) (*RunOutput, error)
```

**매개변수:**
- `ctx`: 취소/타임아웃을 위한 Context
- `input`: 사용자 입력 문자열

**반환값:**
```go
type RunOutput struct {
    Content  string                 // 에이전트의 응답
    Metadata map[string]interface{} // 추가 메타데이터
}
```

**에러:**
- `InvalidInputError`: 입력이 비어있음
- `ModelTimeoutError`: LLM 요청 타임아웃
- `ToolExecutionError`: 도구 실행 실패
- `APIError`: LLM API 에러

**예제:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

output, err := ag.Run(ctx, "What is 2+2?")
if err != nil {
    log.Fatal(err)
}
fmt.Println(output.Content)
```

## Agent.ClearMemory

대화 메모리를 초기화합니다.

**함수 시그니처:**
```go
func (a *Agent) ClearMemory()
```

**예제:**
```go
ag.ClearMemory() // 새로운 대화 시작
```
