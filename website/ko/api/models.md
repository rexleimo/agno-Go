# Models API 레퍼런스

## OpenAI

**생성:**
```go
func New(modelID string, config Config) (*OpenAI, error)

type Config struct {
    APIKey      string  // 필수
    BaseURL     string  // 선택 사항 (기본값: https://api.openai.com/v1)
    Temperature float64 // 선택 사항 (기본값: 0.7)
    MaxTokens   int     // 선택 사항
}
```

**지원 모델:**
- `gpt-4`
- `gpt-4-turbo`
- `gpt-4o`
- `gpt-4o-mini`
- `gpt-3.5-turbo`

**예제:**
```go
model, err := openai.New("gpt-4o-mini", openai.Config{
    APIKey:      os.Getenv("OPENAI_API_KEY"),
    Temperature: 0.7,
    MaxTokens:   1000,
})
```

## Anthropic

**생성:**
```go
func New(modelID string, config Config) (*Anthropic, error)

type Config struct {
    APIKey      string  // 필수
    Temperature float64 // 선택 사항
    MaxTokens   int     // 선택 사항 (기본값: 4096)
}
```

**지원 모델:**
- `claude-3-5-sonnet-20241022`
- `claude-3-opus-20240229`
- `claude-3-sonnet-20240229`
- `claude-3-haiku-20240307`

**예제:**
```go
model, err := anthropic.New("claude-3-5-sonnet-20241022", anthropic.Config{
    APIKey:    os.Getenv("ANTHROPIC_API_KEY"),
    MaxTokens: 2048,
})
```

## Ollama

**생성:**
```go
func New(modelID string, config Config) (*Ollama, error)

type Config struct {
    BaseURL     string  // 선택 사항 (기본값: http://localhost:11434)
    Temperature float64 // 선택 사항
}
```

**지원 모델:**
- 로컬 Ollama 설치에서 사용 가능한 모든 모델
- 일반적인 모델: `llama3`, `mistral`, `codellama`, `phi`

**예제:**
```go
model, err := ollama.New("llama3", ollama.Config{
    BaseURL:     "http://localhost:11434",
    Temperature: 0.8,
})
```
