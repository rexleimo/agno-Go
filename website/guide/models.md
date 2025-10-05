# Models - LLM Providers

Agno-Go supports multiple LLM providers with a unified interface.

---

## Supported Models

### OpenAI
- GPT-4o, GPT-4o-mini, GPT-4 Turbo, GPT-3.5 Turbo
- Full streaming support
- Function calling

### Anthropic Claude
- Claude 3.5 Sonnet, Claude 3 Opus, Claude 3 Sonnet, Claude 3 Haiku
- Streaming support
- Tool use

### GLM (智谱AI) ⭐ NEW in v1.0.2
- GLM-4, GLM-4V (vision), GLM-3-Turbo
- Chinese language optimized
- Custom JWT authentication
- Function calling support

### Ollama
- Run models locally (Llama, Mistral, etc.)
- Privacy-first
- No API costs

---

## OpenAI

### Setup

```go
import "github.com/rexleimo/agno-go/pkg/agno/models/openai"

model, err := openai.New("gpt-4o-mini", openai.Config{
    APIKey:      os.Getenv("OPENAI_API_KEY"),
    Temperature: 0.7,
    MaxTokens:   1000,
})
```

### Configuration

```go
type Config struct {
    APIKey      string  // Required: Your OpenAI API key
    BaseURL     string  // Optional: Custom endpoint (default: https://api.openai.com/v1)
    Temperature float64 // Optional: 0.0-2.0 (default: 0.7)
    MaxTokens   int     // Optional: Max response tokens
}
```

### Supported Models

| Model | Context | Best For |
|-------|---------|----------|
| `gpt-4o` | 128K | Most capable, multimodal |
| `gpt-4o-mini` | 128K | Fast, cost-effective |
| `gpt-4-turbo` | 128K | Advanced reasoning |
| `gpt-3.5-turbo` | 16K | Simple tasks, fast |

### Example

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

### Setup

```go
import "github.com/rexleimo/agno-go/pkg/agno/models/anthropic"

model, err := anthropic.New("claude-3-5-sonnet-20241022", anthropic.Config{
    APIKey:    os.Getenv("ANTHROPIC_API_KEY"),
    MaxTokens: 2048,
})
```

### Configuration

```go
type Config struct {
    APIKey      string  // Required: Your Anthropic API key
    Temperature float64 // Optional: 0.0-1.0
    MaxTokens   int     // Optional: Max response tokens (default: 4096)
}
```

### Supported Models

| Model | Context | Best For |
|-------|---------|----------|
| `claude-3-5-sonnet-20241022` | 200K | Most intelligent, coding |
| `claude-3-opus-20240229` | 200K | Complex tasks |
| `claude-3-sonnet-20240229` | 200K | Balanced performance |
| `claude-3-haiku-20240307` | 200K | Fast responses |

### Example

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

### Setup

```go
import "github.com/rexleimo/agno-go/pkg/agno/models/glm"

model, err := glm.New("glm-4", glm.Config{
    APIKey:      os.Getenv("ZHIPUAI_API_KEY"),  // Format: {key_id}.{key_secret}
    Temperature: 0.7,
    MaxTokens:   1024,
})
```

### Configuration

```go
type Config struct {
    APIKey      string  // Required: API key in format {key_id}.{key_secret}
    BaseURL     string  // Optional: Custom endpoint (default: https://open.bigmodel.cn/api/paas/v4)
    Temperature float64 // Optional: 0.0-1.0
    MaxTokens   int     // Optional: Max response tokens
    TopP        float64 // Optional: Top-p sampling parameter
    DoSample    bool    // Optional: Whether to use sampling
}
```

### Supported Models

| Model | Context | Best For |
|-------|---------|----------|
| `glm-4` | 128K | General conversation, Chinese language |
| `glm-4v` | 128K | Vision tasks, multimodal |
| `glm-3-turbo` | 128K | Fast responses, cost-effective |

### API Key Format

GLM uses a special API key format with two parts:

```
{key_id}.{key_secret}
```

Get your API key at: https://open.bigmodel.cn/

### Example

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
        Name:         "GLM Assistant",
        Model:        model,
        Toolkits:     []toolkit.Toolkit{calculator.New()},
        Instructions: "You are a helpful AI assistant.",
    })

    // Chinese language support
    output, _ := agent.Run(context.Background(), "你好！请计算 123 * 456")
    fmt.Println(output.Content)
}
```

### Authentication

GLM uses JWT (JSON Web Token) authentication:

1. API key is parsed into `key_id` and `key_secret`
2. JWT token is generated with HMAC-SHA256 signing
3. Token is valid for 7 days
4. Automatically regenerated for each request

This is handled automatically by the SDK.

---

## Ollama (Local Models)

### Setup

1. Install Ollama: https://ollama.ai
2. Pull a model: `ollama pull llama2`
3. Use in Agno-Go:

```go
import "github.com/rexleimo/agno-go/pkg/agno/models/ollama"

model, err := ollama.New("llama2", ollama.Config{
    BaseURL: "http://localhost:11434",  // Ollama server
})
```

### Configuration

```go
type Config struct {
    BaseURL     string  // Optional: Ollama server URL (default: http://localhost:11434)
    Temperature float64 // Optional: 0.0-1.0
}
```

### Supported Models

Any model available in Ollama:
- `llama2`, `llama3`, `llama3.1`
- `mistral`, `mixtral`
- `codellama`, `deepseek-coder`
- `qwen2`, `gemma2`

### Example

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

## Model Comparison

### Performance

| Provider | Speed | Cost | Privacy | Context |
|----------|-------|------|---------|---------|
| OpenAI GPT-4o-mini | ⚡⚡⚡ | 💰 | ☁️ Cloud | 128K |
| OpenAI GPT-4o | ⚡⚡ | 💰💰💰 | ☁️ Cloud | 128K |
| Anthropic Claude | ⚡⚡ | 💰💰 | ☁️ Cloud | 200K |
| GLM-4 | ⚡⚡⚡ | 💰 | ☁️ Cloud | 128K |
| Ollama | ⚡ | 🆓 Free | 🏠 Local | Varies |

### When to Use Each

**OpenAI GPT-4o-mini**
- Development and testing
- High-volume applications
- Cost-sensitive use cases

**OpenAI GPT-4o**
- Complex reasoning tasks
- Multimodal applications
- Production systems

**Anthropic Claude**
- Long context needs (200K tokens)
- Coding assistance
- Complex analysis

**GLM-4**
- Chinese language applications
- Domestic deployment requirements
- Fast responses with good quality
- Cost-effective for Chinese users

**Ollama**
- Privacy requirements
- No internet connectivity
- Zero API costs
- Development/testing

---

## Switching Models

Switching between models is easy:

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

// Use the same agent code
agent, _ := agent.New(agent.Config{
    Model: openaiModel,  // or claudeModel, glmModel, or ollamaModel
})
```

---

## Advanced Configuration

### Temperature

Controls randomness (0.0 = deterministic, 1.0+ = creative):

```go
model, _ := openai.New("gpt-4o-mini", openai.Config{
    Temperature: 0.0,  // Consistent responses
})

model, _ := openai.New("gpt-4o-mini", openai.Config{
    Temperature: 1.5,  // Creative responses
})
```

### Max Tokens

Limit response length:

```go
model, _ := openai.New("gpt-4o-mini", openai.Config{
    MaxTokens: 500,  // Short responses
})
```

### Custom Endpoints

Use compatible APIs:

```go
model, _ := openai.New("gpt-4o-mini", openai.Config{
    BaseURL: "https://your-proxy.com/v1",  // Custom endpoint
    APIKey:  "your-key",
})
```

---

## Best Practices

### 1. Environment Variables

Store API keys securely:

```go
// Good ✅
APIKey: os.Getenv("OPENAI_API_KEY")

// Bad ❌
APIKey: "sk-proj-..." // Hardcoded
```

### 2. Error Handling

Always check for errors:

```go
model, err := openai.New("gpt-4o-mini", openai.Config{
    APIKey: os.Getenv("OPENAI_API_KEY"),
})
if err != nil {
    log.Fatalf("Failed to create model: %v", err)
}
```

### 3. Model Selection

Choose based on your needs:

```go
// Development: Fast and cheap
devModel, _ := openai.New("gpt-4o-mini", config)

// Production: More capable
prodModel, _ := openai.New("gpt-4o", config)
```

### 4. Context Management

Be mindful of context limits:

```go
// For long conversations, clear memory periodically
if messageCount > 50 {
    agent.ClearMemory()
}
```

---

## Environment Setup

Create a `.env` file:

```bash
# OpenAI
OPENAI_API_KEY=sk-proj-...

# Anthropic
ANTHROPIC_API_KEY=sk-ant-...

# GLM (智谱AI) - Format: {key_id}.{key_secret}
ZHIPUAI_API_KEY=your-key-id.your-key-secret

# Ollama (optional, defaults to localhost)
OLLAMA_BASE_URL=http://localhost:11434
```

Load in your code:

```go
import "github.com/joho/godotenv"

func init() {
    godotenv.Load()
}
```

---

## Next Steps

- Add [Tools](/guide/tools) to enhance model capabilities
- Learn about [Memory](/guide/memory) for conversation history
- Build [Teams](/guide/team) with mixed models
- Explore [Examples](/examples/) for real-world usage

---

## Related Examples

- [Simple Agent](/examples/simple-agent) - OpenAI example
- [Claude Agent](/examples/claude-agent) - Anthropic example
- [GLM Agent](/examples/glm-agent) - GLM (智谱AI) example
- [Ollama Agent](/examples/ollama-agent) - Local model example
