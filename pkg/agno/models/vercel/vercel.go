package vercel

import (
    "context"
    "net/http"
    "time"

    "github.com/rexleimo/agno-go/pkg/agno/models"
    "github.com/rexleimo/agno-go/pkg/agno/types"
    "github.com/sashabaranov/go-openai"
)

// Vercel provider (OpenAI-compatible endpoint configurable)
type Vercel struct {
    models.BaseModel
    client *openai.Client
    config Config
}

type Config struct {
    APIKey      string
    BaseURL     string // required to point to Vercel inference endpoint
    Temperature float64
    MaxTokens   int
    Timeout     time.Duration
}

func New(modelID string, config Config) (*Vercel, error) {
    if config.APIKey == "" {
        return nil, types.NewInvalidConfigError("Vercel API key is required", nil)
    }
    baseURL := config.BaseURL
    if baseURL == "" {
        // leave OpenAI base as fallback; users should set BaseURL for Vercel
        baseURL = openai.DefaultConfig(config.APIKey).BaseURL
    }
    clientConfig := openai.DefaultConfig(config.APIKey)
    clientConfig.BaseURL = baseURL
    to := config.Timeout
    if to == 0 { to = 60 * time.Second }
    clientConfig.HTTPClient = &http.Client{Timeout: to}

    return &Vercel{
        BaseModel: models.BaseModel{ID: modelID, Provider: "vercel"},
        client:    openai.NewClientWithConfig(clientConfig),
        config: Config{
            APIKey:      config.APIKey,
            BaseURL:     baseURL,
            Temperature: config.Temperature,
            MaxTokens:   config.MaxTokens,
            Timeout:     to,
        },
    }, nil
}

func (v *Vercel) Invoke(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
    chatReq := v.buildChatRequest(req)
    resp, err := v.client.CreateChatCompletion(ctx, chatReq)
    if err != nil { return nil, types.NewAPIError("failed to call Vercel API", err) }
    if len(resp.Choices) == 0 { return nil, types.NewAPIError("no response from Vercel", nil) }
    ch := resp.Choices[0]
    mr := &types.ModelResponse{
        ID: resp.ID, Content: ch.Message.Content, Model: resp.Model,
        Usage: types.Usage{PromptTokens: resp.Usage.PromptTokens, CompletionTokens: resp.Usage.CompletionTokens, TotalTokens: resp.Usage.TotalTokens},
        Metadata: types.Metadata{FinishReason: string(ch.FinishReason)},
    }
    if len(ch.Message.ToolCalls) > 0 {
        mr.ToolCalls = make([]types.ToolCall, len(ch.Message.ToolCalls))
        for i, tc := range ch.Message.ToolCalls {
            mr.ToolCalls[i] = types.ToolCall{ID: tc.ID, Type: string(tc.Type), Function: types.ToolCallFunction{Name: tc.Function.Name, Arguments: tc.Function.Arguments}}
        }
    }
    return mr, nil
}

func (v *Vercel) InvokeStream(ctx context.Context, req *models.InvokeRequest) (<-chan types.ResponseChunk, error) {
    chatReq := v.buildChatRequest(req); chatReq.Stream = true
    stream, err := v.client.CreateChatCompletionStream(ctx, chatReq)
    if err != nil { ch := make(chan types.ResponseChunk,1); ch<-types.ResponseChunk{Error:err,Done:true}; close(ch); return ch,nil }
    ch := make(chan types.ResponseChunk)
    go func(){
        defer close(ch); defer stream.Close()
        for { resp, err := stream.Recv(); if err != nil { ch<-types.ResponseChunk{Error:err,Done:true}; return }
            for _, choice := range resp.Choices {
                if delta := choice.Delta; delta.Content != "" { ch<-types.ResponseChunk{Content: delta.Content} }
            }
        }
    }()
    return ch, nil
}

func (v *Vercel) buildChatRequest(req *models.InvokeRequest) openai.ChatCompletionRequest {
    chatReq := openai.ChatCompletionRequest{ Model: v.ID, Messages: make([]openai.ChatCompletionMessage, len(req.Messages)) }
    for i, m := range req.Messages {
        msg := openai.ChatCompletionMessage{ Role: string(m.Role), Content: m.Content, Name: m.Name }
        if m.ToolCallID != "" { msg.ToolCallID = m.ToolCallID }
        if len(m.ToolCalls) > 0 { calls := make([]openai.ToolCall, len(m.ToolCalls)); for j, tc := range m.ToolCalls { calls[j] = openai.ToolCall{ ID: tc.ID, Type: openai.ToolType(tc.Type), Function: openai.FunctionCall{Name: tc.Function.Name, Arguments: tc.Function.Arguments} } }; msg.ToolCalls = calls }
        chatReq.Messages[i] = msg
    }
    if len(req.Tools) > 0 { tools := make([]openai.Tool, len(req.Tools)); for i, td := range req.Tools { tools[i] = openai.Tool{ Type: openai.ToolType(td.Type), Function: &openai.FunctionDefinition{Name: td.Function.Name, Description: td.Function.Description, Parameters: td.Function.Parameters} } }; chatReq.Tools = tools }
    temp, max := models.MergeConfig(req.Temperature, v.config.Temperature, req.MaxTokens, v.config.MaxTokens)
    if temp > 0 { chatReq.Temperature = float32(temp) }
    if max > 0 { chatReq.MaxTokens = max }
    return chatReq
}

