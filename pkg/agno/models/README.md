# Models Package

This package provides common utilities for implementing model providers in Agno-Go.

## Common Utilities (common.go)

### HTTPClient

A shared HTTP client for making API requests:

```go
client := models.NewHTTPClient()

headers := map[string]string{
    "Authorization": "Bearer " + apiKey,
    "Custom-Header": "value",
}

body := map[string]interface{}{
    "prompt": "Hello",
    "max_tokens": 100,
}

resp, err := client.PostJSON(ctx, "https://api.example.com/chat", headers, body)
```

### Response Handling

```go
// For successful responses
var result MyResponseStruct
err := models.ReadJSONResponse(resp, &result)

// For error responses
err := models.ReadErrorResponse(resp)
```

### Helper Functions

**ConvertMessages**: Convert types.Message to generic format
```go
messages := []*types.Message{
    {Role: types.RoleUser, Content: "Hello"},
}
genericMessages := models.ConvertMessages(messages)
```

**MergeConfig**: Merge request-level and model-level configuration
```go
temperature, maxTokens := models.MergeConfig(
    req.Temperature,    // request level
    model.Temperature,  // model level
    req.MaxTokens,      // request level
    model.MaxTokens,    // model level
)
```

**BuildToolDefinitions**: Convert tool definitions to generic format
```go
tools := []models.ToolDefinition{...}
genericTools := models.BuildToolDefinitions(tools)
```

## Usage in Model Providers

### Example: Using HTTPClient

```go
type MyModel struct {
    models.BaseModel
    httpClient *models.HTTPClient
    apiKey     string
}

func New(apiKey string) *MyModel {
    return &MyModel{
        BaseModel:  models.BaseModel{...},
        httpClient: models.NewHTTPClient(),
        apiKey:     apiKey,
    }
}

func (m *MyModel) Invoke(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
    // Build request
    headers := map[string]string{
        "Authorization": "Bearer " + m.apiKey,
    }

    body := map[string]interface{}{
        "messages": models.ConvertMessages(req.Messages),
        "temperature": req.Temperature,
    }

    // Make API call
    resp, err := m.httpClient.PostJSON(ctx, m.baseURL, headers, body)
    if err != nil {
        return nil, err
    }

    // Parse response
    var apiResp MyAPIResponse
    if err := models.ReadJSONResponse(resp, &apiResp); err != nil {
        return nil, err
    }

    // Convert to ModelResponse
    return m.convertResponse(&apiResp), nil
}
```

## Benefits

1. **Code Reuse**: Reduces duplicate HTTP client code across providers
2. **Consistency**: Standardized error handling and response parsing
3. **Maintainability**: Changes to HTTP logic apply to all providers
4. **Testing**: Easier to mock and test with shared utilities

## Providers

### Current Implementations

- **OpenAI** (`openai/openai.go`): Uses official OpenAI Go SDK
  - Coverage: 44.6%
  - Features: GPT-4, GPT-3.5, function calling, streaming

- **Anthropic** (`anthropic/anthropic.go`): Custom HTTP implementation
  - Coverage: 50.9%
  - Features: Claude 3 Opus, Sonnet, Haiku, streaming
  - Can benefit from HTTPClient refactoring

- **Gemini** (`gemini/gemini.go`): Custom HTTP implementation with SSE streaming
  - Coverage: 77.0%
  - Features: Gemini Pro, Gemini Ultra, function calling, SSE streaming
  - Supports system instructions, tool calls, and function responses
  - Example: `cmd/examples/gemini_agent/`

- **DeepSeek** (`deepseek/deepseek.go`): OpenAI-compatible SDK integration
  - Coverage: 81.6%
  - Features: DeepSeek-V3 (deepseek-chat), DeepSeek-R1 (deepseek-reasoner)
  - Full OpenAI API compatibility, function calling, streaming
  - Cost-effective with context caching
  - Example: `cmd/examples/deepseek_agent/`

- **ModelScope** (`modelscope/modelscope.go`): OpenAI-compatible SDK via DashScope
  - Coverage: 78.9%
  - Features: Qwen models (qwen-plus, qwen-turbo, qwen-max), Chinese-optimized
  - Full OpenAI API compatibility through Alibaba Cloud DashScope
  - Excellent Chinese language support
  - Example: `cmd/examples/modelscope_agent/`

- **Ollama** (`ollama/ollama.go`): Custom HTTP implementation
  - Coverage: 43.8%
  - Features: All Ollama models (Llama 2, Mistral, etc.), streaming
  - Can benefit from HTTPClient refactoring

### Adding a New Provider

1. Create a new directory: `pkg/agno/models/yourprovider/`
2. Implement the `Model` interface from `base.go`
3. Use common utilities from `common.go`
4. Add tests with >70% coverage
5. Update this README

See [CLAUDE.md](../../CLAUDE.md#adding-a-model-provider) for detailed instructions.
