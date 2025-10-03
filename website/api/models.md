# Models API Reference

## OpenAI

**Create:**
```go
func New(modelID string, config Config) (*OpenAI, error)

type Config struct {
    APIKey      string  // Required
    BaseURL     string  // Optional (default: https://api.openai.com/v1)
    Temperature float64 // Optional (default: 0.7)
    MaxTokens   int     // Optional
}
```

**Supported Models:**
- `gpt-4`
- `gpt-4-turbo`
- `gpt-4o`
- `gpt-4o-mini`
- `gpt-3.5-turbo`

**Example:**
```go
model, err := openai.New("gpt-4o-mini", openai.Config{
    APIKey:      os.Getenv("OPENAI_API_KEY"),
    Temperature: 0.7,
    MaxTokens:   1000,
})
```

## Anthropic

**Create:**
```go
func New(modelID string, config Config) (*Anthropic, error)

type Config struct {
    APIKey      string  // Required
    Temperature float64 // Optional
    MaxTokens   int     // Optional (default: 4096)
}
```

**Supported Models:**
- `claude-3-5-sonnet-20241022`
- `claude-3-opus-20240229`
- `claude-3-sonnet-20240229`
- `claude-3-haiku-20240307`

**Example:**
```go
model, err := anthropic.New("claude-3-5-sonnet-20241022", anthropic.Config{
    APIKey:    os.Getenv("ANTHROPIC_API_KEY"),
    MaxTokens: 2048,
})
```

## Ollama

**Create:**
```go
func New(modelID string, config Config) (*Ollama, error)

type Config struct {
    BaseURL     string  // Optional (default: http://localhost:11434)
    Temperature float64 // Optional
}
```

**Supported Models:**
- Any model available in local Ollama installation
- Common: `llama3`, `mistral`, `codellama`, `phi`

**Example:**
```go
model, err := ollama.New("llama3", ollama.Config{
    BaseURL:     "http://localhost:11434",
    Temperature: 0.8,
})
```
