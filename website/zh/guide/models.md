# Models - LLM 提供商

Agno-Go 通过统一接口支持多个 LLM 提供商。

---

## 支持的模型

### OpenAI
- GPT-4o、GPT-4o-mini、GPT-4 Turbo、GPT-3.5 Turbo
- 完整流式传输支持
- 函数调用

### Anthropic Claude
- Claude 3.5 Sonnet、Claude 3 Opus、Claude 3 Sonnet、Claude 3 Haiku
- 流式传输支持
- 工具使用

### GLM (智谱AI) ⭐ v1.0.2 新增
- GLM-4、GLM-4V（视觉）、GLM-3-Turbo
- 中文语言优化
- 自定义 JWT 认证
- 函数调用支持

### Ollama
- 本地运行模型 (Llama、Mistral 等)
- 隐私优先
- 无 API 费用

---

## OpenAI

### 设置

```go
import "github.com/rexleimo/agno-go/pkg/agno/models/openai"

model, err := openai.New("gpt-4o-mini", openai.Config{
    APIKey:      os.Getenv("OPENAI_API_KEY"),
    Temperature: 0.7,
    MaxTokens:   1000,
})
```

### 配置

```go
type Config struct {
    APIKey      string  // 必需: 您的 OpenAI API 密钥
    BaseURL     string  // 可选: 自定义端点 (默认: https://api.openai.com/v1)
    Temperature float64 // 可选: 0.0-2.0 (默认: 0.7)
    MaxTokens   int     // 可选: 最大响应 Token 数
}
```

### 支持的模型

| 模型 | 上下文 | 最适合 |
|-------|---------|----------|
| `gpt-4o` | 128K | 最强大,多模态 |
| `gpt-4o-mini` | 128K | 快速,经济实惠 |
| `gpt-4-turbo` | 128K | 高级推理 |
| `gpt-3.5-turbo` | 16K | 简单任务,快速 |

### 示例

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

### 设置

```go
import "github.com/rexleimo/agno-go/pkg/agno/models/anthropic"

model, err := anthropic.New("claude-3-5-sonnet-20241022", anthropic.Config{
    APIKey:    os.Getenv("ANTHROPIC_API_KEY"),
    MaxTokens: 2048,
})
```

### 配置

```go
type Config struct {
    APIKey      string  // 必需: 您的 Anthropic API 密钥
    Temperature float64 // 可选: 0.0-1.0
    MaxTokens   int     // 可选: 最大响应 Token 数 (默认: 4096)
}
```

### 支持的模型

| 模型 | 上下文 | 最适合 |
|-------|---------|----------|
| `claude-3-5-sonnet-20241022` | 200K | 最智能,编程 |
| `claude-3-opus-20240229` | 200K | 复杂任务 |
| `claude-3-sonnet-20240229` | 200K | 平衡性能 |
| `claude-3-haiku-20240307` | 200K | 快速响应 |

### 示例

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

## GLM (智谱AI)

### 设置

```go
import "github.com/rexleimo/agno-go/pkg/agno/models/glm"

model, err := glm.New("glm-4", glm.Config{
    APIKey:      os.Getenv("ZHIPUAI_API_KEY"),  // 格式: {key_id}.{key_secret}
    Temperature: 0.7,
    MaxTokens:   1024,
})
```

### 配置

```go
type Config struct {
    APIKey      string  // 必需: API 密钥，格式为 {key_id}.{key_secret}
    BaseURL     string  // 可选: 自定义端点 (默认: https://open.bigmodel.cn/api/paas/v4)
    Temperature float64 // 可选: 0.0-1.0
    MaxTokens   int     // 可选: 最大响应 Token 数
    TopP        float64 // 可选: Top-p 采样参数
    DoSample    bool    // 可选: 是否使用采样
}
```

### 支持的模型

| 模型 | 上下文 | 最适合 |
|-------|---------|----------|
| `glm-4` | 128K | 通用对话、中文语言 |
| `glm-4v` | 128K | 视觉任务、多模态 |
| `glm-3-turbo` | 128K | 快速响应、成本优化 |

### API 密钥格式

GLM 使用特殊的 API 密钥格式，包含两部分：

```
{key_id}.{key_secret}
```

在此获取您的 API 密钥: https://open.bigmodel.cn/

### 示例

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models/glm"
    "github.com/rexleimo/agno-go/pkg/agno/tools/calculator"
    "github.com/rexleimo/agno-go/pkg/agno/tools/toolkit"
)

func main() {
    model, err := glm.New("glm-4", glm.Config{
        APIKey:      os.Getenv("ZHIPUAI_API_KEY"),
        Temperature: 0.7,
        MaxTokens:   1024,
    })
    if err != nil {
        log.Fatal(err)
    }

    agent, _ := agent.New(agent.Config{
        Name:         "GLM 助手",
        Model:        model,
        Toolkits:     []toolkit.Toolkit{calculator.New()},
        Instructions: "你是一个有用的 AI 助手。",
    })

    // 中文语言支持
    output, _ := agent.Run(context.Background(), "你好！请计算 123 * 456")
    fmt.Println(output.Content)
}
```

### 认证

GLM 使用 JWT (JSON Web Token) 认证：

1. API 密钥被解析为 `key_id` 和 `key_secret`
2. 使用 HMAC-SHA256 签名生成 JWT token
3. Token 有效期为 7 天
4. 每次请求自动重新生成

这些都由 SDK 自动处理。

---

## Ollama (本地模型)

### 设置

1. 安装 Ollama: https://ollama.ai
2. 拉取模型: `ollama pull llama2`
3. 在 Agno-Go 中使用:

```go
import "github.com/rexleimo/agno-go/pkg/agno/models/ollama"

model, err := ollama.New("llama2", ollama.Config{
    BaseURL: "http://localhost:11434",  // Ollama server
})
```

### 配置

```go
type Config struct {
    BaseURL     string  // 可选: Ollama 服务器 URL (默认: http://localhost:11434)
    Temperature float64 // 可选: 0.0-1.0
}
```

### 支持的模型

Ollama 中可用的任何模型:
- `llama2`, `llama3`, `llama3.1`
- `mistral`, `mixtral`
- `codellama`, `deepseek-coder`
- `qwen2`, `gemma2`

### 示例

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
    // Make sure Ollama is running and model is pulled
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

## 模型比较

### 性能

| 提供商 | 速度 | 成本 | 隐私 | 上下文 |
|----------|-------|------|---------|---------|
| OpenAI GPT-4o-mini | ⚡⚡⚡ | 💰 | ☁️ 云端 | 128K |
| OpenAI GPT-4o | ⚡⚡ | 💰💰💰 | ☁️ 云端 | 128K |
| Anthropic Claude | ⚡⚡ | 💰💰 | ☁️ 云端 | 200K |
| GLM-4 | ⚡⚡⚡ | 💰 | ☁️ 云端 | 128K |
| Ollama | ⚡ | 🆓 免费 | 🏠 本地 | 可变 |

### 何时使用每种

**OpenAI GPT-4o-mini**
- 开发和测试
- 高容量应用
- 成本敏感的使用场景

**OpenAI GPT-4o**
- 复杂推理任务
- 多模态应用
- 生产系统

**Anthropic Claude**
- 长上下文需求 (200K Token)
- 编程辅助
- 复杂分析

**GLM-4**
- 中文语言应用
- 国内部署要求
- 快速响应且质量好
- 对中国用户成本优化

**Ollama**
- 隐私要求
- 无互联网连接
- 零 API 成本
- 开发/测试

---

## 切换模型

在模型间切换很简单:

```go
// OpenAI
openaiModel, _ := openai.New("gpt-4o-mini", openai.Config{
    APIKey: os.Getenv("OPENAI_API_KEY"),
})

// Claude
claudeModel, _ := anthropic.New("claude-3-5-sonnet-20241022", anthropic.Config{
    APIKey: os.Getenv("ANTHROPIC_API_KEY"),
})

// GLM
glmModel, _ := glm.New("glm-4", glm.Config{
    APIKey: os.Getenv("ZHIPUAI_API_KEY"),
})

// Ollama
ollamaModel, _ := ollama.New("llama2", ollama.Config{})

// 使用相同的 Agent 代码
agent, _ := agent.New(agent.Config{
    Model: openaiModel,  // 或 claudeModel、glmModel、ollamaModel
})
```

---

## 高级配置

### Temperature

控制随机性 (0.0 = 确定性, 1.0+ = 创造性):

```go
model, _ := openai.New("gpt-4o-mini", openai.Config{
    Temperature: 0.0,  // Consistent responses
})

model, _ := openai.New("gpt-4o-mini", openai.Config{
    Temperature: 1.5,  // Creative responses
})
```

### Max Tokens

限制响应长度:

```go
model, _ := openai.New("gpt-4o-mini", openai.Config{
    MaxTokens: 500,  // Short responses
})
```

### 自定义端点

使用兼容的 API:

```go
model, _ := openai.New("gpt-4o-mini", openai.Config{
    BaseURL: "https://your-proxy.com/v1",  // Custom endpoint
    APIKey:  "your-key",
})
```

---

## 最佳实践

### 1. 环境变量

安全地存储 API 密钥:

```go
// Good ✅
APIKey: os.Getenv("OPENAI_API_KEY")

// Bad ❌
APIKey: "sk-proj-..." // Hardcoded
```

### 2. 错误处理

始终检查错误:

```go
model, err := openai.New("gpt-4o-mini", openai.Config{
    APIKey: os.Getenv("OPENAI_API_KEY"),
})
if err != nil {
    log.Fatalf("Failed to create model: %v", err)
}
```

### 3. 模型选择

根据需求选择:

```go
// Development: Fast and cheap
devModel, _ := openai.New("gpt-4o-mini", config)

// Production: More capable
prodModel, _ := openai.New("gpt-4o", config)
```

### 4. Context 管理

注意上下文限制:

```go
// For long conversations, clear memory periodically
if messageCount > 50 {
    agent.ClearMemory()
}
```

---

## 环境设置

创建 `.env` 文件:

```bash
# OpenAI
OPENAI_API_KEY=sk-proj-...

# Anthropic
ANTHROPIC_API_KEY=sk-ant-...

# GLM (智谱AI) - 格式: {key_id}.{key_secret}
ZHIPUAI_API_KEY=your-key-id.your-key-secret

# Ollama (可选, 默认为 localhost)
OLLAMA_BASE_URL=http://localhost:11434
```

在代码中加载:

```go
import "github.com/joho/godotenv"

func init() {
    godotenv.Load()
}
```

---

## 下一步

- 添加 [Tools](/guide/tools) 增强模型能力
- 了解 [Memory](/guide/memory) 的对话历史
- 使用混合模型构建 [Teams](/guide/team)
- 探索 [Examples](/examples/) 的实际用法

---

## 相关示例

- [Simple Agent](/zh/examples/simple-agent) - OpenAI 示例
- [Claude Agent](/zh/examples/claude-agent) - Anthropic 示例
- [GLM Agent](/zh/examples/glm-agent) - GLM (智谱AI) 示例
- [Ollama Agent](/zh/examples/ollama-agent) - 本地模型示例
