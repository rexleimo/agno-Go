package deepseek

import (
	"context"
	"encoding/json"

	"github.com/rexleimo/agno-go/pkg/agno/models"
	"github.com/rexleimo/agno-go/pkg/agno/types"
	"github.com/sashabaranov/go-openai"
)

const (
	defaultBaseURL = "https://api.deepseek.com/v1"
)

// DeepSeek wraps the DeepSeek API client (OpenAI-compatible)
type DeepSeek struct {
	models.BaseModel
	client *openai.Client
	config Config
}

// Config contains DeepSeek-specific configuration
type Config struct {
	APIKey      string
	BaseURL     string
	Temperature float64
	MaxTokens   int
}

// New creates a new DeepSeek model instance
// DeepSeek API is fully compatible with OpenAI API format
func New(modelID string, config Config) (*DeepSeek, error) {
	if config.APIKey == "" {
		return nil, types.NewInvalidConfigError("DeepSeek API key is required", nil)
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	clientConfig := openai.DefaultConfig(config.APIKey)
	clientConfig.BaseURL = baseURL

	return &DeepSeek{
		BaseModel: models.BaseModel{
			ID:       modelID,
			Provider: "deepseek",
			Name:     modelID,
		},
		client: openai.NewClientWithConfig(clientConfig),
		config: config,
	}, nil
}

// Invoke calls the DeepSeek API synchronously
func (d *DeepSeek) Invoke(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
	chatReq := d.buildChatRequest(req)

	resp, err := d.client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, types.NewAPIError("failed to call DeepSeek API", err)
	}

	if len(resp.Choices) == 0 {
		return nil, types.NewAPIError("no response from DeepSeek", nil)
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
		Metadata: types.Metadata{
			FinishReason: string(choice.FinishReason),
		},
	}

	// Convert tool calls if present
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

// InvokeStream calls the DeepSeek API with streaming response
func (d *DeepSeek) InvokeStream(ctx context.Context, req *models.InvokeRequest) (<-chan types.ResponseChunk, error) {
	chatReq := d.buildChatRequest(req)
	chatReq.Stream = true

	stream, err := d.client.CreateChatCompletionStream(ctx, chatReq)
	if err != nil {
		return nil, types.NewAPIError("failed to create DeepSeek stream", err)
	}

	chunks := make(chan types.ResponseChunk)

	go func() {
		defer close(chunks)
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if err != nil {
				chunks <- types.ResponseChunk{
					Done:  true,
					Error: err,
				}
				return
			}

			if len(response.Choices) == 0 {
				continue
			}

			delta := response.Choices[0].Delta
			chunk := types.ResponseChunk{
				Content: delta.Content,
			}

			// Handle tool calls in stream
			if len(delta.ToolCalls) > 0 {
				chunk.ToolCalls = make([]types.ToolCall, len(delta.ToolCalls))
				for i, tc := range delta.ToolCalls {
					chunk.ToolCalls[i] = types.ToolCall{
						ID:   tc.ID,
						Type: string(tc.Type),
						Function: types.ToolCallFunction{
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						},
					}
				}
			}

			select {
			case chunks <- chunk:
			case <-ctx.Done():
				chunks <- types.ResponseChunk{
					Done:  true,
					Error: ctx.Err(),
				}
				return
			}
		}
	}()

	return chunks, nil
}

// buildChatRequest converts InvokeRequest to OpenAI ChatCompletionRequest
func (d *DeepSeek) buildChatRequest(req *models.InvokeRequest) openai.ChatCompletionRequest {
	chatReq := openai.ChatCompletionRequest{
		Model:    d.ID,
		Messages: make([]openai.ChatCompletionMessage, len(req.Messages)),
	}

	// Convert messages
	for i, msg := range req.Messages {
		chatMsg := openai.ChatCompletionMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
			Name:    msg.Name,
		}

		// Handle tool call responses
		if msg.ToolCallID != "" {
			chatMsg.ToolCallID = msg.ToolCallID
		}

		// Handle tool calls in message
		if len(msg.ToolCalls) > 0 {
			chatMsg.ToolCalls = make([]openai.ToolCall, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				chatMsg.ToolCalls[j] = openai.ToolCall{
					ID:   tc.ID,
					Type: openai.ToolType(tc.Type),
					Function: openai.FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
		}

		chatReq.Messages[i] = chatMsg
	}

	// Convert tools
	if len(req.Tools) > 0 {
		chatReq.Tools = make([]openai.Tool, len(req.Tools))
		for i, tool := range req.Tools {
			paramsJSON, _ := json.Marshal(tool.Function.Parameters)
			var params map[string]interface{}
			json.Unmarshal(paramsJSON, &params)

			chatReq.Tools[i] = openai.Tool{
				Type: openai.ToolType(tool.Type),
				Function: &openai.FunctionDefinition{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  params,
				},
			}
		}
	}

	// Set temperature
	if req.Temperature > 0 {
		chatReq.Temperature = float32(req.Temperature)
	} else if d.config.Temperature > 0 {
		chatReq.Temperature = float32(d.config.Temperature)
	}

	// Set max tokens
	if req.MaxTokens > 0 {
		chatReq.MaxTokens = req.MaxTokens
	} else if d.config.MaxTokens > 0 {
		chatReq.MaxTokens = d.config.MaxTokens
	}

	return chatReq
}

// ValidateConfig validates the DeepSeek configuration
func ValidateConfig(config Config) error {
	if config.APIKey == "" {
		return types.NewInvalidConfigError("API key is required", nil)
	}
	return nil
}
