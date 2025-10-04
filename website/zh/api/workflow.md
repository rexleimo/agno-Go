# Workflow API 参考 / Workflow API Reference

## workflow.New

创建一个新的工作流。/ Create a new workflow.

**签名 / Signature:**
```go
func New(config Config) (*Workflow, error)
```

**参数 / Parameters:**
```go
type Config struct {
    Name  string      // 工作流名称 / Workflow name
    Steps []Primitive // 工作流步骤 / Workflow steps
}
```

**示例 / Example:**
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

## 工作流原语 / Workflow Primitives

### 1. Step

执行智能体或函数。/ Execute an agent or function.

**签名 / Signature:**
```go
func NewStep(name string, target interface{}) *Step
```

**目标类型 / Target types:**
- `*agent.Agent`: 运行智能体 / Run agent
- `func(ctx *ExecutionContext) (*RunOutput, error)`: 自定义函数 / Custom function

**示例 / Example:**
```go
step := workflow.NewStep("analyze", analyzerAgent)

// 或自定义函数 / Or custom function
step := workflow.NewStep("transform", func(ctx *workflow.ExecutionContext) (*workflow.RunOutput, error) {
    input := ctx.Input
    // 转换输入 / Transform input
    return &workflow.RunOutput{Content: transformed}, nil
})
```

### 2. Condition

条件分支。/ Conditional branching.

**签名 / Signature:**
```go
func NewCondition(name string, condition func(*ExecutionContext) bool,
                   thenStep, elseStep Primitive) *Condition
```

**示例 / Example:**
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

迭代循环。/ Iterative loops.

**签名 / Signature:**
```go
func NewLoop(name string, condition func(*ExecutionContext) bool,
             body Primitive, maxIterations int) *Loop
```

**示例 / Example:**
```go
loop := workflow.NewLoop("retry",
    func(ctx *workflow.ExecutionContext) bool {
        result := ctx.GetResult("attempt")
        return result == nil || strings.Contains(result.Content, "error")
    },
    workflow.NewStep("attempt", retryAgent),
    5, // 最多 5 次迭代 / Max 5 iterations
)
```

### 4. Parallel

并行执行。/ Parallel execution.

**签名 / Signature:**
```go
func NewParallel(name string, steps []Primitive) *Parallel
```

**示例 / Example:**
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

动态路由。/ Dynamic routing.

**签名 / Signature:**
```go
func NewRouter(name string, routes map[string]Primitive,
               selector func(*ExecutionContext) string) *Router
```

**示例 / Example:**
```go
router := workflow.NewRouter("route_by_type",
    map[string]workflow.Primitive{
        "email":  workflow.NewStep("email", emailAgent),
        "chat":   workflow.NewStep("chat", chatAgent),
        "phone":  workflow.NewStep("phone", phoneAgent),
    },
    func(ctx *workflow.ExecutionContext) string {
        // 根据输入确定路由 / Determine route based on input
        if strings.Contains(ctx.Input, "@") {
            return "email"
        }
        return "chat"
    },
)
```

## ExecutionContext

访问工作流上下文。/ Access workflow context.

**方法 / Methods:**
```go
func (ctx *ExecutionContext) GetResult(stepName string) *RunOutput
func (ctx *ExecutionContext) SetData(key string, value interface{})
func (ctx *ExecutionContext) GetData(key string) interface{}
```

**示例 / Example:**
```go
step := workflow.NewStep("use_context", func(ctx *workflow.ExecutionContext) (*workflow.RunOutput, error) {
    previous := ctx.GetResult("previous_step")
    userData := ctx.GetData("user_data")

    // 使用之前的结果 / Use previous results
    return &workflow.RunOutput{Content: result}, nil
})
```
