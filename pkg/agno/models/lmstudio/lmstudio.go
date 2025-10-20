package lmstudio

import (
    "context"
    "net/http"
    "time"

    "github.com/rexleimo/agno-go/pkg/agno/models"
    "github.com/rexleimo/agno-go/pkg/agno/types"
    "github.com/sashabaranov/go-openai"
)

const (
    // DefaultBaseURL for LM Studio local server
    DefaultBaseURL = "http://localhost:1234/v1"
)

// LMStudio wraps OpenAI-compatible local server
type LMStudio struct {
    models.BaseModel
    client *openai.Client
    config Config
}

// Config for LM Studio
type Config struct {
    APIKey      string        // optional; LM Studio typically does not require
    BaseURL     string
    Temperature float64
    MaxTokens   int
    Timeout     time.Duration
}

// New creates a new LMStudio instance
func New(modelID string, config Config) (*LMStudio, error) {
    baseURL := config.BaseURL
    if baseURL == "" {
        baseURL = DefaultBaseURL
    }
    // Use placeholder key if not provided
    apiKey := config.APIKey
    if apiKey == "" {
        apiKey = "lm-studio"
    }

    clientConfig := openai.DefaultConfig(apiKey)
    clientConfig.BaseURL = baseURL
    timeout := config.Timeout
    if timeout == 0 {
        timeout = 60 * time.Second
    }
    clientConfig.HTTPClient = &http.Client{Timeout: timeout}

    return &LMStudio{
        BaseModel: models.BaseModel{ID: modelID, Provider: "lmstudio"},
        client:    openai.NewClientWithConfig(clientConfig),
        config: Config{
            APIKey:      apiKey,
            BaseURL:     baseURL,
            Temperature: config.Temperature,
            MaxTokens:   config.MaxTokens,
            Timeout:     timeout,
        },
    }, nil
}

// Invoke calls the local server synchronously
func (l *LMStudio) Invoke(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
    chatReq := l.buildChatRequest(req)
    resp, err := l.client.CreateChatCompletion(ctx, chatReq)
    if err != nil {
        return nil, types.NewAPIError("failed to call LM Studio API", err)
    }
    if len(resp.Choices) == 0 {
        return nil, types.NewAPIError("no response from LM Studio", nil)
    }
    choice := resp.Choices[0]
    return &types.ModelResponse{
        ID:      resp.ID,
        Content: choice.Message.Content,
        Model:   resp.Model,
        Usage: types.Usage{
            PromptTokens:     resp.Usage.PromptTokens,
            CompletionTokens: resp.Usage.CompletionTokens,
            TotalTokens:      resp.Usage.TotalTokens,
        },
        Metadata: types.Metadata{FinishReason: string(choice.FinishReason)},
    }, nil
}

// InvokeStream streams responses
func (l *LMStudio) InvokeStream(ctx context.Context, req *models.InvokeRequest) (<-chan types.ResponseChunk, error) {
    chatReq := l.buildChatRequest(req)
    chatReq.Stream = true
    stream, err := l.client.CreateChatCompletionStream(ctx, chatReq)
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
        }
    }()
    return ch, nil
}

// buildChatRequest converts InvokeRequest to OpenAI ChatCompletionRequest
func (l *LMStudio) buildChatRequest(req *models.InvokeRequest) openai.ChatCompletionRequest {
    chatReq := openai.ChatCompletionRequest{
        Model:    l.ID,
        Messages: make([]openai.ChatCompletionMessage, len(req.Messages)),
    }
    for i, m := range req.Messages {
        msg := openai.ChatCompletionMessage{
            Role:    string(m.Role),
            Content: m.Content,
            Name:    m.Name,
        }
        if m.ToolCallID != "" {
            msg.ToolCallID = m.ToolCallID
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
            msg.ToolCalls = calls
        }
        chatReq.Messages[i] = msg
    }
    if len(req.Tools) > 0 {
        tools := make([]openai.Tool, len(req.Tools))
        for i, tdef := range req.Tools {
            tools[i] = openai.Tool{
                Type: openai.ToolType(tdef.Type),
                Function: &openai.FunctionDefinition{
                    Name:        tdef.Function.Name,
                    Description: tdef.Function.Description,
                    Parameters:  tdef.Function.Parameters,
                },
            }
        }
        chatReq.Tools = tools
    }
    temperature, maxTokens := models.MergeConfig(req.Temperature, l.config.Temperature, req.MaxTokens, l.config.MaxTokens)
    if temperature > 0 {
        chatReq.Temperature = float32(temperature)
    }
    if maxTokens > 0 {
        chatReq.MaxTokens = maxTokens
    }
    return chatReq
}

