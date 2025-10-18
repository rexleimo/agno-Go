package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rexleimo/agno-go/pkg/agno/models"
	"github.com/rexleimo/agno-go/pkg/agno/types"
)

const (
	defaultBaseURL = "https://api.anthropic.com/v1"
	apiVersion     = "2023-06-01"
)

var nonThinkingModels = map[string]struct{}{
	"claude-3-haiku-20240307":   {},
	"claude-3-5-haiku-20241022": {},
	"claude-3-5-haiku-latest":   {},
}

// Anthropic wraps the Anthropic Claude API client
type Anthropic struct {
	models.BaseModel
	config     Config
	httpClient *http.Client
}

// Config contains Anthropic-specific configuration
// Config 包含Anthropic特定配置
type Config struct {
	APIKey      string
	BaseURL     string
	Temperature float64
	MaxTokens   int
	Timeout     time.Duration // Request timeout / 请求超时时间
	Thinking    *ThinkingConfig
}

// ThinkingConfig represents Anthropic extended thinking configuration
type ThinkingConfig struct {
	Type            string `json:"type"`
	BudgetTokens    int    `json:"budget_tokens"`
	MaxOutputTokens int    `json:"max_output_tokens,omitempty"`
}

// New creates a new Anthropic Claude model instance
func New(modelID string, config Config) (*Anthropic, error) {
	if config.APIKey == "" {
		return nil, types.NewInvalidConfigError("Anthropic API key is required", nil)
	}

	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}

	if config.MaxTokens == 0 {
		config.MaxTokens = 4096
	}

	// Set default timeout if not specified
	// 如果未指定则设置默认超时时间
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second // Default 60 seconds / 默认60秒
	}

	return &Anthropic{
		BaseModel: models.BaseModel{
			ID:       modelID,
			Provider: "anthropic",
			Name:     modelID,
		},
		config: config,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// SupportsReasoning indicates whether the model is configured for extended thinking
func (a *Anthropic) SupportsReasoning() bool {
	if a == nil || a.config.Thinking == nil {
		return false
	}

	cfg := a.config.Thinking
	if !strings.EqualFold(cfg.Type, "enabled") {
		return false
	}

	if cfg.BudgetTokens <= 0 {
		return false
	}

	modelID := strings.ToLower(a.ID)
	if _, exists := nonThinkingModels[modelID]; exists {
		return false
	}

	return true
}

// Invoke calls the Anthropic API synchronously
func (a *Anthropic) Invoke(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
	claudeReq := a.buildClaudeRequest(req)

	reqBody, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, types.NewAPIError("failed to marshal request", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.config.BaseURL+"/messages", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, types.NewAPIError("failed to create HTTP request", err)
	}

	a.setHeaders(httpReq)

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, types.NewAPIError("failed to call Anthropic API", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, types.NewAPIError(fmt.Sprintf("API error: %s", string(body)), nil)
	}

	var claudeResp ClaudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return nil, types.NewAPIError("failed to decode response", err)
	}

	return a.convertResponse(&claudeResp), nil
}

// InvokeStream calls the Anthropic API with streaming response
func (a *Anthropic) InvokeStream(ctx context.Context, req *models.InvokeRequest) (<-chan types.ResponseChunk, error) {
	claudeReq := a.buildClaudeRequest(req)
	claudeReq.Stream = true

	reqBody, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, types.NewAPIError("failed to marshal request", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.config.BaseURL+"/messages", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, types.NewAPIError("failed to create HTTP request", err)
	}

	a.setHeaders(httpReq)

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, types.NewAPIError("failed to call Anthropic API", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, types.NewAPIError(fmt.Sprintf("API error: %s", string(body)), nil)
	}

	chunks := make(chan types.ResponseChunk)

	go func() {
		defer close(chunks)
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		for {
			var event StreamEvent
			if err := decoder.Decode(&event); err != nil {
				if err != io.EOF {
					chunks <- types.ResponseChunk{
						Done:  true,
						Error: err,
					}
				}
				return
			}

			chunk := a.convertStreamEvent(&event)
			select {
			case chunks <- chunk:
				if chunk.Done {
					return
				}
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

// buildClaudeRequest converts InvokeRequest to Claude API request
func (a *Anthropic) buildClaudeRequest(req *models.InvokeRequest) *ClaudeRequest {
	claudeReq := &ClaudeRequest{
		Model:     a.ID,
		Messages:  make([]ClaudeMessage, 0),
		MaxTokens: req.MaxTokens,
		Stream:    false,
	}

	// Set max tokens
	if claudeReq.MaxTokens == 0 {
		claudeReq.MaxTokens = a.config.MaxTokens
	}

	// Set temperature
	if req.Temperature > 0 {
		claudeReq.Temperature = req.Temperature
	} else if a.config.Temperature > 0 {
		claudeReq.Temperature = a.config.Temperature
	}

	var thinkingCfg *ThinkingConfig
	if a.config.Thinking != nil {
		cfgCopy := *a.config.Thinking
		thinkingCfg = &cfgCopy
	}

	if req.Extra != nil {
		if raw, ok := req.Extra["thinking"]; ok {
			if cfgMap, ok := raw.(map[string]interface{}); ok {
				if thinkingCfg == nil {
					thinkingCfg = &ThinkingConfig{}
				}
				if v, ok := cfgMap["type"].(string); ok && v != "" {
					thinkingCfg.Type = v
				}
				if v, ok := cfgMap["Type"].(string); ok && v != "" {
					thinkingCfg.Type = v
				}
				if budget, ok := valueToInt(cfgMap["budget_tokens"]); ok {
					thinkingCfg.BudgetTokens = budget
				}
				if budget, ok := valueToInt(cfgMap["budgetTokens"]); ok {
					thinkingCfg.BudgetTokens = budget
				}
				if maxTokens, ok := valueToInt(cfgMap["max_output_tokens"]); ok {
					thinkingCfg.MaxOutputTokens = maxTokens
				}
				if maxTokens, ok := valueToInt(cfgMap["maxOutputTokens"]); ok {
					thinkingCfg.MaxOutputTokens = maxTokens
				}
			}
		}
		if budget, ok := valueToInt(req.Extra["thinking_budget"]); ok {
			if thinkingCfg == nil {
				thinkingCfg = &ThinkingConfig{Type: "enabled"}
			}
			thinkingCfg.BudgetTokens = budget
			if thinkingCfg.Type == "" {
				thinkingCfg.Type = "enabled"
			}
		}
		if budget, ok := valueToInt(req.Extra["thinkingBudget"]); ok {
			if thinkingCfg == nil {
				thinkingCfg = &ThinkingConfig{Type: "enabled"}
			}
			thinkingCfg.BudgetTokens = budget
			if thinkingCfg.Type == "" {
				thinkingCfg.Type = "enabled"
			}
		}
		if t, ok := req.Extra["thinking_type"].(string); ok && t != "" {
			if thinkingCfg == nil {
				thinkingCfg = &ThinkingConfig{}
			}
			thinkingCfg.Type = t
		}
	}

	if thinkingCfg != nil {
		if thinkingCfg.Type == "" {
			thinkingCfg.Type = "enabled"
		}
		if strings.EqualFold(thinkingCfg.Type, "disabled") || thinkingCfg.BudgetTokens > 0 {
			claudeReq.Thinking = thinkingCfg
		}
	}

	// Convert messages
	var systemPrompt string
	for _, msg := range req.Messages {
		switch msg.Role {
		case types.RoleSystem:
			systemPrompt = msg.Content
		case types.RoleUser, types.RoleAssistant:
			claudeMsg := ClaudeMessage{
				Role:    string(msg.Role),
				Content: msg.Content,
			}
			claudeReq.Messages = append(claudeReq.Messages, claudeMsg)
		case types.RoleTool:
			// Handle tool results
			claudeMsg := ClaudeMessage{
				Role:    "user",
				Content: fmt.Sprintf("Tool result: %s", msg.Content),
			}
			claudeReq.Messages = append(claudeReq.Messages, claudeMsg)
		}
	}

	if systemPrompt != "" {
		claudeReq.System = systemPrompt
	}

	// Convert tools
	if len(req.Tools) > 0 {
		claudeReq.Tools = make([]ClaudeTool, len(req.Tools))
		for i, tool := range req.Tools {
			claudeReq.Tools[i] = ClaudeTool{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				InputSchema: tool.Function.Parameters,
			}
		}
	}

	return claudeReq
}

// convertResponse converts Claude response to ModelResponse
func (a *Anthropic) convertResponse(resp *ClaudeResponse) *types.ModelResponse {
	modelResp := &types.ModelResponse{
		ID:    resp.ID,
		Model: resp.Model,
		Usage: types.Usage{
			PromptTokens:     resp.Usage.InputTokens,
			CompletionTokens: resp.Usage.OutputTokens,
			TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
		Metadata: types.Metadata{
			FinishReason: resp.StopReason,
		},
	}

	var contentBuilder strings.Builder
	var reasoningBuilder strings.Builder
	var redactedReasoning string
	var reasoningSignature string

	// Extract text, reasoning content and tool calls
	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			if block.Text != "" {
				contentBuilder.WriteString(block.Text)
			}
		case "tool_use":
			modelResp.ToolCalls = append(modelResp.ToolCalls, types.ToolCall{
				ID:   block.ID,
				Type: "function",
				Function: types.ToolCallFunction{
					Name:      block.Name,
					Arguments: jsonToString(block.Input),
				},
			})
		case "thinking":
			if block.Thinking != "" {
				if reasoningBuilder.Len() > 0 {
					reasoningBuilder.WriteString("\n")
				}
				reasoningBuilder.WriteString(block.Thinking)
			}
			if block.Signature != "" {
				reasoningSignature = block.Signature
			}
		case "redacted_thinking":
			if block.Data != "" {
				redactedReasoning = block.Data
			}
		}
	}

	if contentBuilder.Len() > 0 {
		modelResp.Content = strings.TrimSpace(contentBuilder.String())
	}

	if reasoningBuilder.Len() > 0 {
		reasoningText := strings.TrimSpace(reasoningBuilder.String())
		if reasoningText != "" {
			reasoning := types.NewReasoningContent(reasoningText)
			if redactedReasoning != "" {
				reasoning = reasoning.WithRedacted(redactedReasoning)
			}
			if resp.Usage.ThinkingTokens > 0 {
				reasoning = reasoning.WithTokenCount(resp.Usage.ThinkingTokens)
			}
			modelResp.ReasoningContent = reasoning
		}
	}

	if reasoningSignature != "" {
		if modelResp.Metadata.Extra == nil {
			modelResp.Metadata.Extra = make(map[string]interface{})
		}
		modelResp.Metadata.Extra["reasoning_signature"] = reasoningSignature
	}

	if resp.Usage.ThinkingTokens > 0 {
		if modelResp.Metadata.Extra == nil {
			modelResp.Metadata.Extra = make(map[string]interface{})
		}
		modelResp.Metadata.Extra["thinking_tokens"] = resp.Usage.ThinkingTokens
	}

	return modelResp
}

// convertStreamEvent converts stream event to ResponseChunk
func (a *Anthropic) convertStreamEvent(event *StreamEvent) types.ResponseChunk {
	chunk := types.ResponseChunk{}

	switch event.Type {
	case "content_block_delta":
		if event.Delta.Type == "text_delta" {
			chunk.Content = event.Delta.Text
		}
	case "message_stop":
		chunk.Done = true
	case "error":
		chunk.Done = true
		chunk.Error = fmt.Errorf("stream error: %s", event.Error.Message)
	}

	return chunk
}

// setHeaders sets required headers for Anthropic API
func (a *Anthropic) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.config.APIKey)
	req.Header.Set("anthropic-version", apiVersion)
}

// ClaudeRequest represents the Anthropic API request
type ClaudeRequest struct {
	Model       string          `json:"model"`
	Messages    []ClaudeMessage `json:"messages"`
	System      string          `json:"system,omitempty"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature,omitempty"`
	Tools       []ClaudeTool    `json:"tools,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
	Thinking    *ThinkingConfig `json:"thinking,omitempty"`
}

// ClaudeMessage represents a message in the conversation
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeTool represents a tool definition
type ClaudeTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// ClaudeResponse represents the Anthropic API response
type ClaudeResponse struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Role       string         `json:"role"`
	Content    []ContentBlock `json:"content"`
	Model      string         `json:"model"`
	StopReason string         `json:"stop_reason"`
	Usage      ClaudeUsage    `json:"usage"`
}

// ContentBlock represents a content block in the response
type ContentBlock struct {
	Type      string                 `json:"type"`
	Text      string                 `json:"text,omitempty"`
	ID        string                 `json:"id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Input     map[string]interface{} `json:"input,omitempty"`
	Thinking  string                 `json:"thinking,omitempty"`
	Signature string                 `json:"signature,omitempty"`
	Data      string                 `json:"data,omitempty"`
}

// ClaudeUsage represents token usage
type ClaudeUsage struct {
	InputTokens    int `json:"input_tokens"`
	OutputTokens   int `json:"output_tokens"`
	ThinkingTokens int `json:"thinking_tokens,omitempty"`
}

// StreamEvent represents a streaming event
type StreamEvent struct {
	Type  string      `json:"type"`
	Delta StreamDelta `json:"delta,omitempty"`
	Error StreamError `json:"error,omitempty"`
}

// StreamDelta represents delta content in streaming
type StreamDelta struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Thinking string `json:"thinking,omitempty"`
}

// StreamError represents an error in streaming
type StreamError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Helper function to convert map to JSON string
func jsonToString(data map[string]interface{}) string {
	if data == nil {
		return "{}"
	}
	jsonBytes, _ := json.Marshal(data)
	return string(jsonBytes)
}

func valueToInt(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int8:
		return int(v), true
	case int16:
		return int(v), true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case uint:
		return int(v), true
	case uint8:
		return int(v), true
	case uint16:
		return int(v), true
	case uint32:
		return int(v), true
	case uint64:
		return int(v), true
	case float32:
		return int(v), true
	case float64:
		return int(v), true
	case string:
		if v == "" {
			return 0, false
		}
		if i, err := strconv.Atoi(v); err == nil {
			return i, true
		}
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return int(i), true
		}
	}
	return 0, false
}

// ValidateConfig validates the Anthropic configuration
func ValidateConfig(config Config) error {
	if config.APIKey == "" {
		return fmt.Errorf("API key is required")
	}
	return nil
}
