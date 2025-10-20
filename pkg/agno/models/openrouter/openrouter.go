package openrouter

import (
    "context"
    "net/http"
    "time"

    "github.com/rexleimo/agno-go/pkg/agno/models"
    "github.com/rexleimo/agno-go/pkg/agno/types"
    "github.com/sashabaranov/go-openai"
)

const (
    // DefaultBaseURL for OpenRouter
    DefaultBaseURL = "https://openrouter.ai/api/v1"
)

// OpenRouter wraps an OpenAI-compatible client
type OpenRouter struct {
    models.BaseModel
    client *openai.Client
    config Config
}

// Config for OpenRouter
type Config struct {
    APIKey      string
    BaseURL     string
    Temperature float64
    MaxTokens   int
    Timeout     time.Duration

    // Optional recommended headers per OpenRouter docs
    Referer string
    Title   string
}

// headerRoundTripper injects custom headers
type headerRoundTripper struct {
    base    http.RoundTripper
    referer string
    title   string
}

func (rt headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
    if rt.referer != "" {
        req.Header.Set("HTTP-Referer", rt.referer)
    }
    if rt.title != "" {
        req.Header.Set("X-Title", rt.title)
    }
    return rt.base.RoundTrip(req)
}

// New creates a new OpenRouter model instance
func New(modelID string, config Config) (*OpenRouter, error) {
    if config.APIKey == "" {
        return nil, types.NewInvalidConfigError("OpenRouter API key is required", nil)
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
    // custom transport for extra headers
    transport := http.DefaultTransport
    clientConfig.HTTPClient = &http.Client{
        Timeout:   timeout,
        Transport: headerRoundTripper{base: transport, referer: config.Referer, title: config.Title},
    }

    return &OpenRouter{
        BaseModel: models.BaseModel{ID: modelID, Provider: "openrouter"},
        client:    openai.NewClientWithConfig(clientConfig),
        config: Config{
            APIKey:      config.APIKey,
            BaseURL:     baseURL,
            Temperature: config.Temperature,
            MaxTokens:   config.MaxTokens,
            Timeout:     timeout,
            Referer:     config.Referer,
            Title:       config.Title,
        },
    }, nil
}

// Invoke calls OpenRouter synchronously
func (o *OpenRouter) Invoke(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
    chatReq := o.buildChatRequest(req)
    resp, err := o.client.CreateChatCompletion(ctx, chatReq)
    if err != nil {
        return nil, types.NewAPIError("failed to call OpenRouter API", err)
    }
    if len(resp.Choices) == 0 {
        return nil, types.NewAPIError("no response from OpenRouter", nil)
    }
    choice := resp.Choices[0]
    mr := &types.ModelResponse{
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
        mr.ToolCalls = make([]types.ToolCall, len(choice.Message.ToolCalls))
        for i, tc := range choice.Message.ToolCalls {
            mr.ToolCalls[i] = types.ToolCall{
                ID:   tc.ID,
                Type: string(tc.Type),
                Function: types.ToolCallFunction{
                    Name:      tc.Function.Name,
                    Arguments: tc.Function.Arguments,
                },
            }
        }
    }

    return mr, nil
}

// InvokeStream streams responses from OpenRouter
func (o *OpenRouter) InvokeStream(ctx context.Context, req *models.InvokeRequest) (<-chan types.ResponseChunk, error) {
    chatReq := o.buildChatRequest(req)
    chatReq.Stream = true

    stream, err := o.client.CreateChatCompletionStream(ctx, chatReq)
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
                delta := choice.Delta
                if delta.Content != "" {
                    ch <- types.ResponseChunk{Content: delta.Content}
                }
                if len(delta.ToolCalls) > 0 {
                    calls := make([]types.ToolCall, len(delta.ToolCalls))
                    for i, tc := range delta.ToolCalls {
                        calls[i] = types.ToolCall{
                            ID:   tc.ID,
                            Type: string(tc.Type),
                            Function: types.ToolCallFunction{
                                Name:      tc.Function.Name,
                                Arguments: tc.Function.Arguments,
                            },
                        }
                    }
                    ch <- types.ResponseChunk{ToolCalls: calls}
                }
            }
        }
    }()
    return ch, nil
}

// buildChatRequest converts InvokeRequest to OpenAI ChatCompletionRequest
func (o *OpenRouter) buildChatRequest(req *models.InvokeRequest) openai.ChatCompletionRequest {
    chatReq := openai.ChatCompletionRequest{
        Model:    o.ID,
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
            tc := make([]openai.ToolCall, len(m.ToolCalls))
            for j, call := range m.ToolCalls {
                tc[j] = openai.ToolCall{
                    ID:   call.ID,
                    Type: openai.ToolType(call.Type),
                    Function: openai.FunctionCall{
                        Name:      call.Function.Name,
                        Arguments: call.Function.Arguments,
                    },
                }
            }
            msg.ToolCalls = tc
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

    temperature, maxTokens := models.MergeConfig(req.Temperature, o.config.Temperature, req.MaxTokens, o.config.MaxTokens)
    if temperature > 0 {
        chatReq.Temperature = float32(temperature)
    }
    if maxTokens > 0 {
        chatReq.MaxTokens = maxTokens
    }

    return chatReq
}

