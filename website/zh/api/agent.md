# Agent API 参考 / Agent API Reference

## agent.New

创建一个新的智能体实例。/ Create a new agent instance.

**签名 / Signature:**
```go
func New(config Config) (*Agent, error)
```

**参数 / Parameters:**

```go
type Config struct {
    // 必需 / Required
    Model models.Model // 要使用的 LLM 模型 / LLM model to use

    // 可选 / Optional
    Name         string            // 智能体名称 (默认: "Agent") / Agent name (default: "Agent")
    Toolkits     []toolkit.Toolkit // 可用工具 / Available tools
    Memory       memory.Memory     // 对话记忆 / Conversation memory
    Instructions string            // 系统指令 / System instructions
    MaxLoops     int               // 最大工具调用循环次数 (默认: 10) / Max tool call loops (default: 10)
}
```

**返回值 / Returns:**
- `*Agent`: 创建的智能体实例 / Created agent instance
- `error`: 如果 model 为 nil 或配置无效则返回错误 / Error if model is nil or config is invalid

**示例 / Example:**
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

使用输入执行智能体。/ Execute the agent with input.

**签名 / Signature:**
```go
func (a *Agent) Run(ctx context.Context, input string) (*RunOutput, error)
```

**参数 / Parameters:**
- `ctx`: 用于取消/超时的上下文 / Context for cancellation/timeout
- `input`: 用户输入字符串 / User input string

**返回值 / Returns:**
```go
type RunOutput struct {
    Content  string                 // 智能体的响应 / Agent's response
    Metadata map[string]interface{} // 附加元数据 / Additional metadata
}
```

**错误 / Errors:**
- `InvalidInputError`: 输入为空 / Input is empty
- `ModelTimeoutError`: LLM 请求超时 / LLM request timeout
- `ToolExecutionError`: 工具执行失败 / Tool execution failed
- `APIError`: LLM API 错误 / LLM API error

**示例 / Example:**
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

清除对话记忆。/ Clear conversation memory.

**签名 / Signature:**
```go
func (a *Agent) ClearMemory()
```

**示例 / Example:**
```go
ag.ClearMemory() // 开始全新对话 / Start fresh conversation
```
