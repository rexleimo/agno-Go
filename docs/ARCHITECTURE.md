# Agno-Go Architecture Design | Agno-Go 架构设计

## Core Philosophy | 核心理念

**Simple, Efficient, Extensible | 简单、高效、可扩展**

---

## System Architecture | 整体架构

```
┌──────────────────────────────────────────────┐
│         Application Layer | 应用层           │
│  (AgentOS API, CLI Tools, Custom Apps)       │
└───────────────────┬──────────────────────────┘
                    │
┌───────────────────▼──────────────────────────┐
│      Core Abstractions | 核心抽象            │
│  ┌─────────┐  ┌──────┐  ┌──────────┐        │
│  │  Agent  │  │ Team │  │ Workflow │        │
│  └─────────┘  └──────┘  └──────────┘        │
└───────────────────┬──────────────────────────┘
                    │
┌───────────────────▼──────────────────────────┐
│     Foundation Layer | 基础层                │
│  ┌────────┐ ┌───────┐ ┌────────┐ ┌───────┐ │
│  │ Models │ │ Tools │ │ Memory │ │Storage│ │
│  └────────┘ └───────┘ └────────┘ └───────┘ │
└──────────────────────────────────────────────┘
```

---

## Core Interface Design | 核心接口设计

### 1. Model Interface | Model 接口
```go
type Model interface {
    // 同步调用
    Invoke(ctx context.Context, req *InvokeRequest) (*ModelResponse, error)

    // 流式调用
    InvokeStream(ctx context.Context, req *InvokeRequest) (<-chan ResponseChunk, error)

    // 元信息
    GetProvider() string
    GetID() string
}

type InvokeRequest struct {
    Messages    []Message
    Tools       []ToolDefinition
    Temperature float64
    MaxTokens   int
}

type ModelResponse struct {
    Content   string
    ToolCalls []ToolCall
    Usage     Usage
}
```

### 2. Agent Interface | Agent 接口
```go
type Agent struct {
    ID          string
    Name        string
    Model       Model
    Tools       []Toolkit
    Memory      *Memory
    Instructions string
}

func (a *Agent) Run(ctx context.Context, input string) (*RunOutput, error)
func (a *Agent) RunStream(ctx context.Context, input string) (<-chan RunEvent, error)
```

### 3. Toolkit Interface | 工具包接口
```go
type Toolkit interface {
    Name() string
    Functions() map[string]*Function
}

type Function struct {
    Name        string
    Description string
    Parameters  map[string]Parameter // 简化的参数定义
    Handler     func(context.Context, map[string]interface{}) (interface{}, error)
}
```

### 4. VectorDB Interface | 向量数据库接口
```go
type VectorDB interface {
    Insert(ctx context.Context, docs []Document) error
    Search(ctx context.Context, query string, limit int) ([]Document, error)
    Delete(ctx context.Context, ids []string) error
}

type Document struct {
    ID       string
    Content  string
    Metadata map[string]interface{}
    Vector   []float64
}
```

---

## Data Flow | 数据流

### Single Conversation Flow | 单次对话流程
```
User Input
    ↓
[Agent.Run]
    ↓
Memory.Recall ─→ Context
    ↓
Model.Invoke (with Tools)
    ↓
    ├─→ Text Response → Output
    └─→ Tool Calls
            ↓
        Execute Tools
            ↓
        Model.Invoke (with results)
            ↓
        Final Response
            ↓
Memory.Store
    ↓
Output to User
```

### Streaming Response | 流式响应
```go
// Use channel for streaming output | 使用 channel 实现流式输出
ch := agent.RunStream(ctx, "hello")
for event := range ch {
    switch event.Type {
    case "content":
        fmt.Print(event.Data)
    case "tool_call":
        // Handle tool call | 处理工具调用
    case "completed":
        // Completed | 完成
    }
}
```

---

## Concurrency Model | 并发模型

### Goroutine Usage Principles | Goroutine 使用原则
1. **Each Agent.Run in independent goroutine** (user-controlled) | **每个 Agent.Run 独立 goroutine**(用户控制)
2. **Tool execution can be parallel** (framework-controlled) | **Tool 执行可并行**(框架控制)
3. **Use context.Context to control lifecycle** | **使用 context.Context 控制生命周期**

```go
// Example: Run multiple agents concurrently | 示例: 并发运行多个 agent
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

var wg sync.WaitGroup
results := make(chan *RunOutput, 3)

for i := 0; i < 3; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        output, err := agents[id].Run(ctx, input)
        if err == nil {
            results <- output
        }
    }(i)
}

wg.Wait()
close(results)
```

---

## Error Handling | 错误处理

### Layered Error Design | 分层错误设计
```go
// Base error type | 基础错误类型
type AgnoError struct {
    Code    string
    Message string
    Cause   error
}

// Specific errors | 特定错误
var (
    ErrModelTimeout    = &AgnoError{Code: "MODEL_TIMEOUT", ...}
    ErrToolExecution   = &AgnoError{Code: "TOOL_ERROR", ...}
    ErrInvalidInput    = &AgnoError{Code: "INVALID_INPUT", ...}
)

// Use errors.Is/As | 使用 errors.Is/As
if errors.Is(err, ErrModelTimeout) {
    // Retry logic | 重试逻辑
}
```

---

## Extension Mechanisms | 扩展机制

### Custom Model | 自定义模型
```go
type MyModel struct {
    // Implement Model interface | 实现 Model 接口
}

func (m *MyModel) Invoke(ctx context.Context, req *InvokeRequest) (*ModelResponse, error) {
    // Custom logic | 自定义逻辑
}
```

### Custom Tool | 自定义工具
```go
type MyToolkit struct{}

func (t *MyToolkit) Name() string {
    return "my_toolkit"
}

func (t *MyToolkit) Functions() map[string]*Function {
    return map[string]*Function{
        "custom_tool": {
            Name: "custom_tool",
            Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
                // Tool logic | 工具逻辑
                return "result", nil
            },
        },
    }
}
```

---

## Design Principles | 设计原则

1. **Interface over Implementation | 接口优于实现** - Depend on abstractions | 依赖抽象
2. **Composition over Inheritance | 组合优于继承** - Use struct embedding | 使用结构体嵌入
3. **Explicit over Implicit | 显式优于隐式** - No over-abstraction | 不过度抽象
4. **Standard Library First | 标准库优先** - Minimize dependencies | 减少依赖
5. **Fail Fast | 快速失败** - Early error checking | 早期错误检查
6. **Concurrency Safe | 并发安全** - Thread-safe by default | 默认线程安全

---

## Deployment Architecture | 部署架构

### Production Deployment | 生产部署
```
        ┌──────────┐
        │   Load   │
        │ Balancer │
        └────┬─────┘
             │
    ┌────────┴────────┐
    ▼                 ▼
┌────────┐       ┌────────┐
│AgentOS │       │AgentOS │
│  API   │       │  API   │
└───┬────┘       └───┬────┘
    │                │
    └────────┬───────┘
             ▼
      ┌──────────┐
      │PostgreSQL│
      │  Redis   │
      │ ChromaDB │
      └──────────┘
```

---

**For detailed development guide, see [DEVELOPMENT.md](DEVELOPMENT.md)**

**详细开发指南,请参阅 [DEVELOPMENT.md](DEVELOPMENT.md)**
