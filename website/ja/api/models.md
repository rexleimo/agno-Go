# Models APIリファレンス

## OpenAI

**作成:**
```go
func New(modelID string, config Config) (*OpenAI, error)

type Config struct {
    APIKey      string  // 必須
    BaseURL     string  // オプション (デフォルト: https://api.openai.com/v1)
    Temperature float64 // オプション (デフォルト: 0.7)
    MaxTokens   int     // オプション
}
```

**サポートされているモデル:**
- `gpt-4`
- `gpt-4-turbo`
- `gpt-4o`
- `gpt-4o-mini`
- `gpt-3.5-turbo`

**例:**
```go
model, err := openai.New("gpt-4o-mini", openai.Config{
    APIKey:      os.Getenv("OPENAI_API_KEY"),
    Temperature: 0.7,
    MaxTokens:   1000,
})
```

## Anthropic

**作成:**
```go
func New(modelID string, config Config) (*Anthropic, error)

type Config struct {
    APIKey      string  // 必須
    Temperature float64 // オプション
    MaxTokens   int     // オプション (デフォルト: 4096)
}
```

**サポートされているモデル:**
- `claude-3-5-sonnet-20241022`
- `claude-3-opus-20240229`
- `claude-3-sonnet-20240229`
- `claude-3-haiku-20240307`

**例:**
```go
model, err := anthropic.New("claude-3-5-sonnet-20241022", anthropic.Config{
    APIKey:    os.Getenv("ANTHROPIC_API_KEY"),
    MaxTokens: 2048,
})
```

## Ollama

**作成:**
```go
func New(modelID string, config Config) (*Ollama, error)

type Config struct {
    BaseURL     string  // オプション (デフォルト: http://localhost:11434)
    Temperature float64 // オプション
}
```

**サポートされているモデル:**
- ローカルOllamaインストールで利用可能な任意のモデル
- 一般的なモデル: `llama3`, `mistral`, `codellama`, `phi`

**例:**
```go
model, err := ollama.New("llama3", ollama.Config{
    BaseURL:     "http://localhost:11434",
    Temperature: 0.8,
})
```
