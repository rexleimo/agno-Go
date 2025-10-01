# Agno-Go 架构设计

## 核心理念
**简单、高效、可扩展**

---

## 整体架构

```
┌─────────────────────────────────────────┐
│          Application Layer              │
│  (CLI Tools, Web API, Custom Apps)      │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│         Core Abstractions               │
│  ┌─────────┐  ┌──────┐  ┌──────────┐   │
│  │  Agent  │  │ Team │  │ Workflow │   │
│  └─────────┘  └──────┘  └──────────┘   │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│        Foundation Layer                  │
│  ┌────────┐ ┌───────┐ ┌──────┐         │
│  │ Models │ │ Tools │ │Memory│ ...     │
│  └────────┘ └───────┘ └──────┘         │
└─────────────────────────────────────────┘
```

---

## 核心接口设计

### 1. Model Interface (模型)
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

### 2. Agent (智能体)
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

### 3. Toolkit (工具包)
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

### 4. VectorDB (向量数据库)
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

## 数据流

### 单次对话流程
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

### 流式响应
```go
// 使用 channel 实现流式输出
ch := agent.RunStream(ctx, "hello")
for event := range ch {
    switch event.Type {
    case "content":
        fmt.Print(event.Data)
    case "tool_call":
        // 处理工具调用
    case "completed":
        // 完成
    }
}
```

---

## 并发模型

### Goroutine 使用原则
1. **每个 Agent.Run 独立 goroutine** (用户控制)
2. **Tool 执行可并行** (框架控制)
3. **使用 context.Context 控制生命周期**

```go
// 示例: 并发运行多个 agent
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

## 错误处理

### 分层错误设计
```go
// 基础错误类型
type AgnoError struct {
    Code    string
    Message string
    Cause   error
}

// 特定错误
var (
    ErrModelTimeout    = &AgnoError{Code: "MODEL_TIMEOUT", ...}
    ErrToolExecution   = &AgnoError{Code: "TOOL_ERROR", ...}
    ErrInvalidInput    = &AgnoError{Code: "INVALID_INPUT", ...}
)

// 使用 errors.Is/As
if errors.Is(err, ErrModelTimeout) {
    // 重试逻辑
}
```

---

## 配置管理

### 简单配置文件
```yaml
# agno.yaml
agent:
  model:
    provider: openai
    id: gpt-4
    api_key: ${OPENAI_API_KEY}

  tools:
    - http_client
    - file_ops

  memory:
    type: in_memory
    max_messages: 100
```

```go
// 加载配置
config, err := LoadConfig("agno.yaml")
agent := NewAgent(config)
```

---

## 扩展机制

### 1. 自定义模型
```go
type MyModel struct {
    // 实现 Model interface
}

func (m *MyModel) Invoke(ctx context.Context, req *InvokeRequest) (*ModelResponse, error) {
    // 自定义逻辑
}
```

### 2. 自定义工具
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
                // 工具逻辑
                return "result", nil
            },
        },
    }
}
```

### 3. 中间件模式
```go
type Middleware func(next Handler) Handler
type Handler func(context.Context, *Request) (*Response, error)

// 示例: 日志中间件
func LoggingMiddleware(next Handler) Handler {
    return func(ctx context.Context, req *Request) (*Response, error) {
        log.Printf("Request: %+v", req)
        resp, err := next(ctx, req)
        log.Printf("Response: %+v", resp)
        return resp, err
    }
}
```

---

## 性能优化策略

### 1. 对象池
```go
var messagePool = sync.Pool{
    New: func() interface{} {
        return &Message{}
    },
}

// 使用
msg := messagePool.Get().(*Message)
defer messagePool.Put(msg)
```

### 2. 减少内存分配
- 使用 `strings.Builder` 代替字符串拼接
- 预分配 slice 容量: `make([]T, 0, expectedSize)`
- 避免不必要的拷贝

### 3. 并发优化
- 限制 goroutine 数量 (worker pool)
- 使用 buffered channel
- 合理使用 `sync.WaitGroup` 和 `errgroup`

---

## 测试架构

### 1. Mock 接口
```go
type MockModel struct {
    InvokeFunc func(context.Context, *InvokeRequest) (*ModelResponse, error)
}

func (m *MockModel) Invoke(ctx context.Context, req *InvokeRequest) (*ModelResponse, error) {
    if m.InvokeFunc != nil {
        return m.InvokeFunc(ctx, req)
    }
    return &ModelResponse{Content: "mock response"}, nil
}
```

### 2. 测试分层
- **单元测试**: 每个函数/方法
- **集成测试**: Agent + Model + Tool
- **端到端测试**: 完整场景
- **性能测试**: benchmark

---

## 部署架构

### 1. 单机部署
```
┌─────────────┐
│  agno-go    │
│  (binary)   │
└─────────────┘
```

### 2. API 服务部署
```
        ┌──────────┐
        │ Load     │
        │ Balancer │
        └────┬─────┘
             │
    ┌────────┴────────┐
    ▼                 ▼
┌────────┐       ┌────────┐
│agno-go │       │agno-go │
│  API   │       │  API   │
└───┬────┘       └───┬────┘
    │                │
    └────────┬───────┘
             ▼
      ┌──────────┐
      │ Database │
      │ & Vector │
      │   DB     │
      └──────────┘
```

---

## 设计原则总结

1. **接口优于实现** - 依赖抽象
2. **组合优于继承** - 使用 struct embedding
3. **显式优于隐式** - 不过度抽象
4. **标准库优先** - 减少依赖
5. **失败快速** - 早期错误检查
6. **并发安全** - 默认线程安全设计

---

**保持简单,持续迭代**
