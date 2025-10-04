# Workflow APIリファレンス

## workflow.New

新しいワークフローを作成します。

**シグネチャ:**
```go
func New(config Config) (*Workflow, error)
```

**パラメータ:**
```go
type Config struct {
    Name  string      // ワークフロー名
    Steps []Primitive // ワークフローステップ
}
```

**例:**
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

## Workflowプリミティブ

### 1. Step

エージェントまたは関数を実行します。

**シグネチャ:**
```go
func NewStep(name string, target interface{}) *Step
```

**ターゲットの型:**
- `*agent.Agent`: エージェントを実行
- `func(ctx *ExecutionContext) (*RunOutput, error)`: カスタム関数

**例:**
```go
step := workflow.NewStep("analyze", analyzerAgent)

// またはカスタム関数
step := workflow.NewStep("transform", func(ctx *workflow.ExecutionContext) (*workflow.RunOutput, error) {
    input := ctx.Input
    // 入力を変換
    return &workflow.RunOutput{Content: transformed}, nil
})
```

### 2. Condition

条件分岐を実行します。

**シグネチャ:**
```go
func NewCondition(name string, condition func(*ExecutionContext) bool,
                   thenStep, elseStep Primitive) *Condition
```

**例:**
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

反復ループを実行します。

**シグネチャ:**
```go
func NewLoop(name string, condition func(*ExecutionContext) bool,
             body Primitive, maxIterations int) *Loop
```

**例:**
```go
loop := workflow.NewLoop("retry",
    func(ctx *workflow.ExecutionContext) bool {
        result := ctx.GetResult("attempt")
        return result == nil || strings.Contains(result.Content, "error")
    },
    workflow.NewStep("attempt", retryAgent),
    5, // 最大5回の反復
)
```

### 4. Parallel

並列実行を行います。

**シグネチャ:**
```go
func NewParallel(name string, steps []Primitive) *Parallel
```

**例:**
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

動的ルーティングを実行します。

**シグネチャ:**
```go
func NewRouter(name string, routes map[string]Primitive,
               selector func(*ExecutionContext) string) *Router
```

**例:**
```go
router := workflow.NewRouter("route_by_type",
    map[string]workflow.Primitive{
        "email":  workflow.NewStep("email", emailAgent),
        "chat":   workflow.NewStep("chat", chatAgent),
        "phone":  workflow.NewStep("phone", phoneAgent),
    },
    func(ctx *workflow.ExecutionContext) string {
        // 入力に基づいてルートを決定
        if strings.Contains(ctx.Input, "@") {
            return "email"
        }
        return "chat"
    },
)
```

## ExecutionContext

ワークフローコンテキストにアクセスします。

**メソッド:**
```go
func (ctx *ExecutionContext) GetResult(stepName string) *RunOutput
func (ctx *ExecutionContext) SetData(key string, value interface{})
func (ctx *ExecutionContext) GetData(key string) interface{}
```

**例:**
```go
step := workflow.NewStep("use_context", func(ctx *workflow.ExecutionContext) (*workflow.RunOutput, error) {
    previous := ctx.GetResult("previous_step")
    userData := ctx.GetData("user_data")

    // 前の結果を使用
    return &workflow.RunOutput{Content: result}, nil
})
```
