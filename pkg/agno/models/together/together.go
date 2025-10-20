package together

import (
    "context"
    "net/http"
    "time"

    "github.com/rexleimo/agno-go/pkg/agno/models"
    "github.com/rexleimo/agno-go/pkg/agno/types"
    "github.com/sashabaranov/go-openai"
)

const (
    // DefaultBaseURL Together AI OpenAI-compatible endpoint
    DefaultBaseURL = "https://api.together.xyz/v1"
)

// Together wraps an OpenAI-compatible client for Together AI
type Together struct {
    models.BaseModel
    client *openai.Client
    config Config
}

// Config holds Together-specific settings
type Config struct {
    APIKey      string
    BaseURL     string
    Temperature float64
    MaxTokens   int
    Timeout     time.Duration
}

// New creates a new Together model instance
func New(modelID string, config Config) (*Together, error) {
    if config.APIKey == "" {
        return nil, types.NewInvalidConfigError("Together API key is required", nil)
    }

    baseURL := config.BaseURL
    if baseURL == "" {
        baseURL = DefaultBaseURL
    }

    clientConfig := openai.DefaultConfig(config.APIKey)
    clientConfig.BaseURL = baseURL

    timeout := config.Timeout
    if timeout == 0 {
        timeout = 60 * time.Second
    }
    clientConfig.HTTPClient = &http.Client{Timeout: timeout}

    return &Together{
        BaseModel: models.BaseModel{ID: modelID, Provider: "together"},
        client:    openai.NewClientWithConfig(clientConfig),
        config: Config{
            APIKey:      config.APIKey,
            BaseURL:     baseURL,
            Temperature: config.Temperature,
            MaxTokens:   config.MaxTokens,
            Timeout:     timeout,
        },
    }, nil
}

// Invoke calls Together AI synchronously
func (t *Together) Invoke(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
    chatReq := t.buildChatRequest(req)
    resp, err := t.client.CreateChatCompletion(ctx, chatReq)
    if err != nil {
        return nil, types.NewAPIError("failed to call Together API", err)
    }
    if len(resp.Choices) == 0 {
        return nil, types.NewAPIError("no response from Together", nil)
    }

    choice := resp.Choices[0]
    modelResp := &types.ModelResponse{
        ID:      resp.ID,
        Content: choice.Message.Content,
        Model:   resp.Model,
        Usage: types.Usage{
            PromptTokens:     resp.Usage.PromptTokens,
            CompletionTokens: resp.Usage.CompletionTokens,
            TotalTokens:      resp.Usage.TotalTokens,
        },
        Metadata: types.Metadata{FinishReason: string(choice.FinishReason)},
    }

    if len(choice.Message.ToolCalls) > 0 {
        modelResp.ToolCalls = make([]types.ToolCall, len(choice.Message.ToolCalls))
        for i, tc := range choice.Message.ToolCalls {
            modelResp.ToolCalls[i] = types.ToolCall{
                ID:   tc.ID,
                Type: string(tc.Type),
                Function: types.ToolCallFunction{
                    Name:      tc.Function.Name,
                    Arguments: tc.Function.Arguments,
                },
            }
        }
    }

    return modelResp, nil
}

// InvokeStream calls Together AI with streaming response
func (t *Together) InvokeStream(ctx context.Context, req *models.InvokeRequest) (<-chan types.ResponseChunk, error) {
    chatReq := t.buildChatRequest(req)
    chatReq.Stream = true

    stream, err := t.client.CreateChatCompletionStream(ctx, chatReq)
    if err != nil {
        ch := make(chan types.ResponseChunk, 1)
        ch <- types.ResponseChunk{Error: err, Done: true}
        close(ch)
        return ch, nil
    }

    ch := make(chan types.ResponseChunk)
    go func() {
        defer close(ch)
        defer stream.Close()
        for {
            resp, err := stream.Recv()
            if err != nil {
                ch <- types.ResponseChunk{Error: err, Done: true}
                return
            }
            for _, choice := range resp.Choices {
                if delta := choice.Delta; delta.Content != "" {
                    ch <- types.ResponseChunk{Content: delta.Content}
                }
            }
            // No explicit done signal until stream closes
        }
    }()
    return ch, nil
}

// buildChatRequest converts InvokeRequest to OpenAI-compatible ChatCompletionRequest
func (t *Together) buildChatRequest(req *models.InvokeRequest) openai.ChatCompletionRequest {
    chatReq := openai.ChatCompletionRequest{
        Model:    t.ID,
        Messages: make([]openai.ChatCompletionMessage, len(req.Messages)),
    }

    for i, m := range req.Messages {
        chatMsg := openai.ChatCompletionMessage{
            Role:    string(m.Role),
            Content: m.Content,
            Name:    m.Name,
        }

        if m.ToolCallID != "" {
            chatMsg.ToolCallID = m.ToolCallID
        }

        if len(m.ToolCalls) > 0 {
            calls := make([]openai.ToolCall, len(m.ToolCalls))
            for j, tc := range m.ToolCalls {
                calls[j] = openai.ToolCall{
                    ID:   tc.ID,
                    Type: openai.ToolType(tc.Type),
                    Function: openai.FunctionCall{
                        Name:      tc.Function.Name,
                        Arguments: tc.Function.Arguments,
                    },
                }
            }
            chatMsg.ToolCalls = calls
        }

        chatReq.Messages[i] = chatMsg
    }

    // Tools
    if len(req.Tools) > 0 {
        tools := make([]openai.Tool, len(req.Tools))
        for i, tool := range req.Tools {
            tools[i] = openai.Tool{
                Type: openai.ToolType(tool.Type),
                Function: &openai.FunctionDefinition{
                    Name:        tool.Function.Name,
                    Description: tool.Function.Description,
                    Parameters:  tool.Function.Parameters,
                },
            }
        }
        chatReq.Tools = tools
    }

    // Temperature and tokens
    temperature, maxTokens := models.MergeConfig(req.Temperature, t.config.Temperature, req.MaxTokens, t.config.MaxTokens)
    if temperature > 0 {
        chatReq.Temperature = float32(temperature)
    }
    if maxTokens > 0 {
        chatReq.MaxTokens = maxTokens
    }

    return chatReq
}

// ValidateConfig validates Together config
func ValidateConfig(config Config) error {
    if config.APIKey == "" {
        return types.NewInvalidConfigError("Together API key is required", nil)
    }
    return nil
}
