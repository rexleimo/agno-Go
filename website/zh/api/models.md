# Models API 参考 / Models API Reference

## OpenAI

**创建 / Create:**
```go
func New(modelID string, config Config) (*OpenAI, error)

type Config struct {
    APIKey      string  // 必需 / Required
    BaseURL     string  // 可选 (默认: https://api.openai.com/v1) / Optional (default: https://api.openai.com/v1)
    Temperature float64 // 可选 (默认: 0.7) / Optional (default: 0.7)
    MaxTokens   int     // 可选 / Optional
}
```

**支持的模型 / Supported Models:**
- `gpt-4`
- `gpt-4-turbo`
- `gpt-4o`
- `gpt-4o-mini`
- `gpt-3.5-turbo`

**示例 / Example:**
```go
model, err := openai.New("gpt-4o-mini", openai.Config{
    APIKey:      os.Getenv("OPENAI_API_KEY"),
    Temperature: 0.7,
    MaxTokens:   1000,
})
```

## Anthropic

**创建 / Create:**
```go
func New(modelID string, config Config) (*Anthropic, error)

type Config struct {
    APIKey      string  // 必需 / Required
    Temperature float64 // 可选 / Optional
    MaxTokens   int     // 可选 (默认: 4096) / Optional (default: 4096)
}
```

**支持的模型 / Supported Models:**
- `claude-3-5-sonnet-20241022`
- `claude-3-opus-20240229`
- `claude-3-sonnet-20240229`
- `claude-3-haiku-20240307`

**示例 / Example:**
```go
model, err := anthropic.New("claude-3-5-sonnet-20241022", anthropic.Config{
    APIKey:    os.Getenv("ANTHROPIC_API_KEY"),
    MaxTokens: 2048,
})
```

## Ollama

**创建 / Create:**
```go
func New(modelID string, config Config) (*Ollama, error)

type Config struct {
    BaseURL     string  // 可选 (默认: http://localhost:11434) / Optional (default: http://localhost:11434)
    Temperature float64 // 可选 / Optional
}
```

**支持的模型 / Supported Models:**
- 本地 Ollama 安装中可用的任何模型 / Any model available in local Ollama installation
- 常用: `llama3`, `mistral`, `codellama`, `phi` / Common: `llama3`, `mistral`, `codellama`, `phi`

**示例 / Example:**
```go
model, err := ollama.New("llama3", ollama.Config{
    BaseURL:     "http://localhost:11434",
    Temperature: 0.8,
})
```
