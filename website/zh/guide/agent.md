# Agent

**Agent** 是一个可以使用工具、维护对话上下文并独立执行任务的自主 AI 实体。

## 概述

```go
import "github.com/rexleimo/agno-Go/pkg/agno/agent"

agent, err := agent.New(agent.Config{
    Name:         "My Agent",
    Model:        model,
    Toolkits:     []toolkit.Toolkit{calculator.New()},
    Instructions: "You are a helpful assistant",
    MaxLoops:     10,
})

output, err := agent.Run(context.Background(), "What is 2+2?")
```

## 配置

### Config 结构

```go
type Config struct {
    Name         string            // Agent 名称
    Model        models.Model      // LLM 模型
    Toolkits     []toolkit.Toolkit // 可用工具
    Memory       memory.Memory     // 对话记忆
    Instructions string            // 系统指令
    MaxLoops     int               // 最大工具调用循环次数 (默认: 10)
    PreHooks     []hooks.Hook      // 执行前钩子
    PostHooks    []hooks.Hook      // 执行后钩子
}
```

### 参数

- **Name** (必需): 人类可读的 Agent 标识符
- **Model** (必需): LLM 模型实例 (OpenAI、Claude 等)
- **Toolkits** (可选): 可用工具列表
- **Memory** (可选): 默认为内存存储,最多保留 100 条消息
- **Instructions** (可选): 系统提示词/角色设定
- **MaxLoops** (可选): 防止无限工具调用循环 (默认: 10)
- **PreHooks** (可选): 执行前验证钩子
- **PostHooks** (可选): 执行后验证钩子

## 基本用法

### 简单 Agent

```go
package main

import (
    "context"
    "fmt"
    "github.com/rexleimo/agno-Go/pkg/agno/agent"
    "github.com/rexleimo/agno-Go/pkg/agno/models/openai"
)

func main() {
    model, _ := openai.New("gpt-4o-mini", openai.Config{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })

    ag, _ := agent.New(agent.Config{
        Name:         "Assistant",
        Model:        model,
        Instructions: "You are a helpful assistant",
    })

    output, _ := ag.Run(context.Background(), "Hello!")
    fmt.Println(output.Content)
}
```

### 带工具的 Agent

```go
import (
    "github.com/rexleimo/agno-Go/pkg/agno/tools/calculator"
    "github.com/rexleimo/agno-Go/pkg/agno/tools/http"
)

ag, _ := agent.New(agent.Config{
    Name:  "Smart Assistant",
    Model: model,
    Toolkits: []toolkit.Toolkit{
        calculator.New(),
        http.New(),
    },
    Instructions: "You can do math and make HTTP requests",
})

output, _ := ag.Run(ctx, "Calculate 15 * 23 and fetch https://api.github.com")
```

## 高级特性

### 自定义记忆

```go
import "github.com/rexleimo/agno-Go/pkg/agno/memory"

// Create memory with custom limit
mem := memory.NewInMemory(50) // Keep last 50 messages

ag, _ := agent.New(agent.Config{
    Memory: mem,
    // ... other config
})
```

### 钩子与防护栏

使用钩子验证输入和输出:

```go
import "github.com/rexleimo/agno-Go/pkg/agno/guardrails"

// Built-in prompt injection guard
promptGuard := guardrails.NewPromptInjectionGuardrail()

// Custom validation hook
customHook := func(ctx context.Context, input *hooks.HookInput) error {
    if len(input.Input) > 1000 {
        return fmt.Errorf("input too long")
    }
    return nil
}

ag, _ := agent.New(agent.Config{
    PreHooks:  []hooks.Hook{customHook, promptGuard},
    PostHooks: []hooks.Hook{outputValidator},
    // ... other config
})
```

### Context 和超时

```go
import "time"

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

output, err := ag.Run(ctx, "Complex task...")
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        fmt.Println("Timeout!")
    }
}
```

## Run 输出

`Run` 方法返回 `*RunOutput`:

```go
type RunOutput struct {
    Content  string                 // Agent 的响应
    Messages []types.Message        // 完整消息历史
    Metadata map[string]interface{} // 附加数据
}
```

示例:

```go
output, err := ag.Run(ctx, "Tell me a joke")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Response:", output.Content)
fmt.Println("Messages:", len(output.Messages))
fmt.Println("Metadata:", output.Metadata)
```

## 记忆管理

### 清除记忆

```go
// Clear all conversation history
ag.ClearMemory()
```

### 访问记忆

```go
// Get current messages
messages := ag.GetMemory().GetMessages()
fmt.Println("History:", len(messages))
```

## 错误处理

```go
output, err := ag.Run(ctx, input)
if err != nil {
    switch {
    case errors.Is(err, types.ErrInvalidInput):
        // Handle invalid input
    case errors.Is(err, types.ErrRateLimit):
        // Handle rate limit
    case errors.Is(err, context.DeadlineExceeded):
        // Handle timeout
    default:
        // Handle other errors
    }
}
```

## 最佳实践

### 1. 始终使用 Context

```go
ctx := context.Background()
output, err := ag.Run(ctx, input)
```

### 2. 设置适当的 MaxLoops

```go
// For simple tasks
MaxLoops: 5

// For complex reasoning
MaxLoops: 15
```

### 3. 提供清晰的指令

```go
Instructions: `You are a customer support agent.
- Be polite and professional
- Use tools to look up information
- If unsure, ask for clarification`
```

### 4. 使用类型安全的工具配置

```go
calc := calculator.New()
httpClient := http.New(http.Config{
    Timeout: 10 * time.Second,
})

ag, _ := agent.New(agent.Config{
    Toolkits: []toolkit.Toolkit{calc, httpClient},
})
```

## 性能考虑

- **Agent 创建**: 平均约 180ns
- **内存占用**: 每个 Agent 约 1.2KB
- **并发 Agent**: 完全线程安全,可自由使用 goroutine

```go
// Concurrent agents
for i := 0; i < 100; i++ {
    go func(id int) {
        ag, _ := agent.New(config)
        output, _ := ag.Run(ctx, input)
        fmt.Printf("Agent %d: %s\n", id, output.Content)
    }(i)
}
```

## 示例

查看工作示例:

- [Simple Agent](https://github.com/rexleimo/agno-Go/tree/main/cmd/examples/simple_agent)
- [Claude Agent](https://github.com/rexleimo/agno-Go/tree/main/cmd/examples/claude_agent)
- [Ollama Agent](https://github.com/rexleimo/agno-Go/tree/main/cmd/examples/ollama_agent)
- [Agent with Guardrails](https://github.com/rexleimo/agno-Go/tree/main/cmd/examples/agent_with_guardrails)

## API 参考

完整的 API 文档,请参阅 [Agent API Reference](/api/agent)。

## 下一步

- [Team](/guide/team) - 多 Agent 协作
- [Workflow](/guide/workflow) - 基于步骤的编排
- [Tools](/guide/tools) - 内置和自定义工具
- [Models](/guide/models) - LLM 提供商配置
