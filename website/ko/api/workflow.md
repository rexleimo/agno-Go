# Workflow API 레퍼런스

## workflow.New

새로운 워크플로우를 생성합니다.

**함수 시그니처:**
```go
func New(config Config) (*Workflow, error)
```

**매개변수:**
```go
type Config struct {
    Name  string      // 워크플로우 이름
    Steps []Primitive // 워크플로우 단계
}
```

**예제:**
```go
wf, err := workflow.New(workflow.Config{
    Name: "Data Processing",
    Steps: []workflow.Primitive{
        workflow.NewStep("fetch", fetchAgent),
        workflow.NewStep("process", processAgent),
        workflow.NewStep("output", outputAgent),
    },
})
```

## Workflow 원시 연산

### 1. Step

에이전트 또는 함수를 실행합니다.

**함수 시그니처:**
```go
func NewStep(name string, target interface{}) *Step
```

**타겟 타입:**
- `*agent.Agent`: 에이전트 실행
- `func(ctx *ExecutionContext) (*RunOutput, error)`: 커스텀 함수

**예제:**
```go
step := workflow.NewStep("analyze", analyzerAgent)

// 또는 커스텀 함수
step := workflow.NewStep("transform", func(ctx *workflow.ExecutionContext) (*workflow.RunOutput, error) {
    input := ctx.Input
    // 입력 변환
    return &workflow.RunOutput{Content: transformed}, nil
})
```

### 2. Condition

조건부 분기입니다.

**함수 시그니처:**
```go
func NewCondition(name string, condition func(*ExecutionContext) bool,
                   thenStep, elseStep Primitive) *Condition
```

**예제:**
```go
cond := workflow.NewCondition("check_sentiment",
    func(ctx *workflow.ExecutionContext) bool {
        result := ctx.GetResult("classify")
        return strings.Contains(result.Content, "positive")
    },
    workflow.NewStep("positive_handler", positiveAgent),
    workflow.NewStep("negative_handler", negativeAgent),
)
```

### 3. Loop

반복 루프입니다.

**함수 시그니처:**
```go
func NewLoop(name string, condition func(*ExecutionContext) bool,
             body Primitive, maxIterations int) *Loop
```

**예제:**
```go
loop := workflow.NewLoop("retry",
    func(ctx *workflow.ExecutionContext) bool {
        result := ctx.GetResult("attempt")
        return result == nil || strings.Contains(result.Content, "error")
    },
    workflow.NewStep("attempt", retryAgent),
    5, // 최대 5회 반복
)
```

### 4. Parallel

병렬 실행입니다.

**함수 시그니처:**
```go
func NewParallel(name string, steps []Primitive) *Parallel
```

**예제:**
```go
parallel := workflow.NewParallel("gather",
    []workflow.Primitive{
        workflow.NewStep("source1", agent1),
        workflow.NewStep("source2", agent2),
        workflow.NewStep("source3", agent3),
    },
)
```

### 5. Router

동적 라우팅입니다.

**함수 시그니처:**
```go
func NewRouter(name string, routes map[string]Primitive,
               selector func(*ExecutionContext) string) *Router
```

**예제:**
```go
router := workflow.NewRouter("route_by_type",
    map[string]workflow.Primitive{
        "email":  workflow.NewStep("email", emailAgent),
        "chat":   workflow.NewStep("chat", chatAgent),
        "phone":  workflow.NewStep("phone", phoneAgent),
    },
    func(ctx *workflow.ExecutionContext) string {
        // 입력에 따라 라우트 결정
        if strings.Contains(ctx.Input, "@") {
            return "email"
        }
        return "chat"
    },
)
```

## ExecutionContext

워크플로우 컨텍스트에 접근합니다.

**메서드:**
```go
func (ctx *ExecutionContext) GetResult(stepName string) *RunOutput
func (ctx *ExecutionContext) SetData(key string, value interface{})
func (ctx *ExecutionContext) GetData(key string) interface{}
```

**예제:**
```go
step := workflow.NewStep("use_context", func(ctx *workflow.ExecutionContext) (*workflow.RunOutput, error) {
    previous := ctx.GetResult("previous_step")
    userData := ctx.GetData("user_data")

    // 이전 결과 사용
    return &workflow.RunOutput{Content: result}, nil
})
```
