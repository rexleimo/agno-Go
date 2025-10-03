# Workflow API Reference

## workflow.New

Create a new workflow.

**Signature:**
```go
func New(config Config) (*Workflow, error)
```

**Parameters:**
```go
type Config struct {
    Name  string      // Workflow name
    Steps []Primitive // Workflow steps
}
```

**Example:**
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

## Workflow Primitives

### 1. Step

Execute an agent or function.

**Signature:**
```go
func NewStep(name string, target interface{}) *Step
```

**Target types:**
- `*agent.Agent`: Run agent
- `func(ctx *ExecutionContext) (*RunOutput, error)`: Custom function

**Example:**
```go
step := workflow.NewStep("analyze", analyzerAgent)

// Or custom function
step := workflow.NewStep("transform", func(ctx *workflow.ExecutionContext) (*workflow.RunOutput, error) {
    input := ctx.Input
    // Transform input
    return &workflow.RunOutput{Content: transformed}, nil
})
```

### 2. Condition

Conditional branching.

**Signature:**
```go
func NewCondition(name string, condition func(*ExecutionContext) bool,
                   thenStep, elseStep Primitive) *Condition
```

**Example:**
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

Iterative loops.

**Signature:**
```go
func NewLoop(name string, condition func(*ExecutionContext) bool,
             body Primitive, maxIterations int) *Loop
```

**Example:**
```go
loop := workflow.NewLoop("retry",
    func(ctx *workflow.ExecutionContext) bool {
        result := ctx.GetResult("attempt")
        return result == nil || strings.Contains(result.Content, "error")
    },
    workflow.NewStep("attempt", retryAgent),
    5, // Max 5 iterations
)
```

### 4. Parallel

Parallel execution.

**Signature:**
```go
func NewParallel(name string, steps []Primitive) *Parallel
```

**Example:**
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

Dynamic routing.

**Signature:**
```go
func NewRouter(name string, routes map[string]Primitive,
               selector func(*ExecutionContext) string) *Router
```

**Example:**
```go
router := workflow.NewRouter("route_by_type",
    map[string]workflow.Primitive{
        "email":  workflow.NewStep("email", emailAgent),
        "chat":   workflow.NewStep("chat", chatAgent),
        "phone":  workflow.NewStep("phone", phoneAgent),
    },
    func(ctx *workflow.ExecutionContext) string {
        // Determine route based on input
        if strings.Contains(ctx.Input, "@") {
            return "email"
        }
        return "chat"
    },
)
```

## ExecutionContext

Access workflow context.

**Methods:**
```go
func (ctx *ExecutionContext) GetResult(stepName string) *RunOutput
func (ctx *ExecutionContext) SetData(key string, value interface{})
func (ctx *ExecutionContext) GetData(key string) interface{}
```

**Example:**
```go
step := workflow.NewStep("use_context", func(ctx *workflow.ExecutionContext) (*workflow.RunOutput, error) {
    previous := ctx.GetResult("previous_step")
    userData := ctx.GetData("user_data")

    // Use previous results
    return &workflow.RunOutput{Content: result}, nil
})
```
