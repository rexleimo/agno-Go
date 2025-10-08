package groq

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rexleimo/agno-go/pkg/agno/models"
	"github.com/rexleimo/agno-go/pkg/agno/types"
	"github.com/sashabaranov/go-openai"
)

const (
	// DefaultBaseURL is the default Groq API endpoint
	// DefaultBaseURL 是默认的 Groq API 端点
	DefaultBaseURL = "https://api.groq.com/openai/v1"
)

// Groq wraps the Groq client using OpenAI-compatible API
// Groq 使用 OpenAI 兼容的 API 封装 Groq 客户端
type Groq struct {
	models.BaseModel
	client *openai.Client
	config Config
}

// Config contains Groq-specific configuration
// Config 包含 Groq 特定配置
type Config struct {
	APIKey      string        // Groq API Key / Groq API 密钥
	BaseURL     string        // API Base URL / API 基础 URL
	Temperature float64       // Temperature parameter / 温度参数
	MaxTokens   int           // Max tokens to generate / 生成的最大 token 数
	Timeout     time.Duration // Request timeout / 请求超时时间
}

// New creates a new Groq model instance
// New 创建一个新的 Groq 模型实例
func New(modelID string, config Config) (*Groq, error) {
	if config.APIKey == "" {
		return nil, types.NewInvalidConfigError("Groq API key is required", nil)
	}

	// Set default base URL if not provided
	// 如果未提供,设置默认 base URL
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	// Create OpenAI client config for Groq
	// 为 Groq 创建 OpenAI 客户端配置
	clientConfig := openai.DefaultConfig(config.APIKey)
	clientConfig.BaseURL = baseURL

	// Set timeout on HTTP client
	// 在 HTTP 客户端上设置超时
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second // Default 60 seconds / 默认 60 秒
	}
	clientConfig.HTTPClient = &http.Client{
		Timeout: timeout,
	}

	return &Groq{
		BaseModel: models.BaseModel{
			ID:       modelID,
			Provider: "groq",
		},
		client: openai.NewClientWithConfig(clientConfig),
		config: Config{
			APIKey:      config.APIKey,
			BaseURL:     baseURL,
			Temperature: config.Temperature,
			MaxTokens:   config.MaxTokens,
			Timeout:     config.Timeout,
		},
	}, nil
}

// Invoke calls the Groq API synchronously
// Invoke 同步调用 Groq API
func (g *Groq) Invoke(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
	chatReq := g.buildChatRequest(req)

	resp, err := g.client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, types.NewAPIError("failed to call Groq API", err)
	}

	if len(resp.Choices) == 0 {
		return nil, types.NewAPIError("no response from Groq", nil)
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
	// 如果存在,转换工具调用
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

// InvokeStream calls the Groq API with streaming response
// InvokeStream 使用流式响应调用 Groq API
func (g *Groq) InvokeStream(ctx context.Context, req *models.InvokeRequest) (<-chan types.ResponseChunk, error) {
	chatReq := g.buildChatRequest(req)
	chatReq.Stream = true

	stream, err := g.client.CreateChatCompletionStream(ctx, chatReq)
	if err != nil {
		return nil, types.NewAPIError("failed to create Groq stream", err)
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
			// 处理流中的工具调用
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
// buildChatRequest 将 InvokeRequest 转换为 OpenAI ChatCompletionRequest
func (g *Groq) buildChatRequest(req *models.InvokeRequest) openai.ChatCompletionRequest {
	chatReq := openai.ChatCompletionRequest{
		Model:    g.ID,
		Messages: make([]openai.ChatCompletionMessage, len(req.Messages)),
	}

	// Convert messages
	// 转换消息
	for i, msg := range req.Messages {
		chatMsg := openai.ChatCompletionMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
			Name:    msg.Name,
		}

		// Handle tool call responses
		// 处理工具调用响应
		if msg.ToolCallID != "" {
			chatMsg.ToolCallID = msg.ToolCallID
		}

		// Handle tool calls in message
		// 处理消息中的工具调用
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
	// 转换工具
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
	// 设置温度
	if req.Temperature > 0 {
		chatReq.Temperature = float32(req.Temperature)
	} else if g.config.Temperature > 0 {
		chatReq.Temperature = float32(g.config.Temperature)
	}

	// Set max tokens
	// 设置最大 token 数
	if req.MaxTokens > 0 {
		chatReq.MaxTokens = req.MaxTokens
	} else if g.config.MaxTokens > 0 {
		chatReq.MaxTokens = g.config.MaxTokens
	}

	return chatReq
}

// ValidateConfig validates the Groq configuration
// ValidateConfig 验证 Groq 配置
func ValidateConfig(config Config) error {
	if config.APIKey == "" {
		return fmt.Errorf("API key is required")
	}
	return nil
}
