package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/rexleimo/agno-go/pkg/agno/models"
	"github.com/rexleimo/agno-go/pkg/agno/types"
)

const (
	defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"
)

// Gemini wraps the Google Gemini API client
type Gemini struct {
	models.BaseModel
	config     Config
	httpClient *http.Client
}

// Config contains Gemini-specific configuration
type Config struct {
	APIKey          string
	BaseURL         string
	Temperature     float64
	MaxTokens       int
	ThinkingBudget  int
	IncludeThoughts *bool
}

// New creates a new Gemini model instance
func New(modelID string, config Config) (*Gemini, error) {
	if config.APIKey == "" {
		return nil, types.NewInvalidConfigError("Gemini API key is required", nil)
	}

	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 8192
	}

	return &Gemini{
		BaseModel: models.BaseModel{
			ID:       modelID,
			Provider: "gemini",
			Name:     modelID,
		},
		config:     config,
		httpClient: &http.Client{},
	}, nil
}

// SupportsReasoning returns whether the current configuration enables reasoning features
func (g *Gemini) SupportsReasoning() bool {
	if g == nil {
		return false
	}

	modelID := strings.ToLower(g.ID)
	if strings.Contains(modelID, "2.5") ||
		strings.Contains(modelID, "thinking") ||
		strings.Contains(modelID, "reasoning") {
		return true
	}

	if g.config.ThinkingBudget > 0 {
		return true
	}

	if g.config.IncludeThoughts != nil && *g.config.IncludeThoughts {
		return true
	}

	return false
}

// Invoke calls the Gemini API synchronously
func (g *Gemini) Invoke(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
	geminiReq := g.buildGeminiRequest(req)

	reqBody, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, types.NewAPIError("failed to marshal request", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", g.config.BaseURL, g.ID, g.config.APIKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, types.NewAPIError("failed to create HTTP request", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, types.NewAPIError("failed to call Gemini API", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, types.NewAPIError(fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(body)), nil)
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, types.NewAPIError("failed to decode response", err)
	}

	return g.convertResponse(&geminiResp), nil
}

// InvokeStream calls the Gemini API with streaming response
func (g *Gemini) InvokeStream(ctx context.Context, req *models.InvokeRequest) (<-chan types.ResponseChunk, error) {
	geminiReq := g.buildGeminiRequest(req)

	reqBody, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, types.NewAPIError("failed to marshal request", err)
	}

	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?key=%s&alt=sse", g.config.BaseURL, g.ID, g.config.APIKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, types.NewAPIError("failed to create HTTP request", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, types.NewAPIError("failed to call Gemini API", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, types.NewAPIError(fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(body)), nil)
	}

	chunks := make(chan types.ResponseChunk)

	go func() {
		defer close(chunks)
		defer resp.Body.Close()

		decoder := NewSSEDecoder(resp.Body)
		for {
			data, err := decoder.Next()
			if err != nil {
				if err != io.EOF {
					chunks <- types.ResponseChunk{
						Done:  true,
						Error: err,
					}
				} else {
					chunks <- types.ResponseChunk{Done: true}
				}
				return
			}

			var geminiResp GeminiResponse
			if err := json.Unmarshal(data, &geminiResp); err != nil {
				continue
			}

			chunk := g.convertToChunk(&geminiResp)
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

// buildGeminiRequest converts InvokeRequest to Gemini API request
func (g *Gemini) buildGeminiRequest(req *models.InvokeRequest) *GeminiRequest {
	geminiReq := &GeminiRequest{
		Contents:         make([]Content, 0),
		GenerationConfig: &GenerationConfig{},
	}

	// Set generation config
	if req.Temperature > 0 {
		geminiReq.GenerationConfig.Temperature = req.Temperature
	} else if g.config.Temperature > 0 {
		geminiReq.GenerationConfig.Temperature = g.config.Temperature
	}

	if req.MaxTokens > 0 {
		geminiReq.GenerationConfig.MaxOutputTokens = req.MaxTokens
	} else if g.config.MaxTokens > 0 {
		geminiReq.GenerationConfig.MaxOutputTokens = g.config.MaxTokens
	}

	var thinkingBudget int
	var includeThoughtsPtr *bool

	if g.config.ThinkingBudget > 0 {
		thinkingBudget = g.config.ThinkingBudget
	}
	if g.config.IncludeThoughts != nil {
		includeThoughtsPtr = g.config.IncludeThoughts
	}

	if req.Extra != nil {
		if val, ok := req.Extra["thinking_budget"]; ok {
			if budget, ok := toInt(val); ok {
				thinkingBudget = budget
			}
		}
		if val, ok := req.Extra["include_thoughts"]; ok {
			if include, ok := toBool(val); ok {
				includeThoughtsPtr = &include
			}
		}
		if val, ok := req.Extra["includeThoughts"]; ok {
			if include, ok := toBool(val); ok {
				includeThoughtsPtr = &include
			}
		}
		if cfgRaw, ok := req.Extra["thinking_config"]; ok {
			if cfg, ok := cfgRaw.(map[string]interface{}); ok {
				if v, ok := cfg["budget_tokens"]; ok {
					if budget, ok := toInt(v); ok {
						thinkingBudget = budget
					}
				}
				if v, ok := cfg["budgetTokens"]; ok {
					if budget, ok := toInt(v); ok {
						thinkingBudget = budget
					}
				}
				if v, ok := cfg["include_thoughts"]; ok {
					if include, ok := toBool(v); ok {
						includeThoughtsPtr = &include
					}
				}
				if v, ok := cfg["includeThoughts"]; ok {
					if include, ok := toBool(v); ok {
						includeThoughtsPtr = &include
					}
				}
			}
		}
	}

	if thinkingBudget > 0 || includeThoughtsPtr != nil {
		tc := &ThinkingConfig{}
		if thinkingBudget > 0 {
			tc.BudgetTokens = thinkingBudget
		}
		if includeThoughtsPtr != nil {
			tc.IncludeThoughts = includeThoughtsPtr
		}
		geminiReq.ThinkingConfig = tc
	}

	// Convert messages to Gemini format
	var systemInstruction string
	for _, msg := range req.Messages {
		switch msg.Role {
		case types.RoleSystem:
			systemInstruction = msg.Content
		case types.RoleUser:
			geminiReq.Contents = append(geminiReq.Contents, Content{
				Role: "user",
				Parts: []Part{
					{Text: msg.Content},
				},
			})
		case types.RoleAssistant:
			content := Content{
				Role:  "model",
				Parts: make([]Part, 0),
			}

			// Add text content
			if msg.Content != "" {
				content.Parts = append(content.Parts, Part{Text: msg.Content})
			}

			// Add tool calls if present
			if len(msg.ToolCalls) > 0 {
				for _, tc := range msg.ToolCalls {
					var args map[string]interface{}
					json.Unmarshal([]byte(tc.Function.Arguments), &args)
					content.Parts = append(content.Parts, Part{
						FunctionCall: &FunctionCall{
							Name: tc.Function.Name,
							Args: args,
						},
					})
				}
			}

			geminiReq.Contents = append(geminiReq.Contents, content)
		case types.RoleTool:
			// Tool results are sent as function responses
			geminiReq.Contents = append(geminiReq.Contents, Content{
				Role: "function",
				Parts: []Part{
					{
						FunctionResponse: &FunctionResponse{
							Name: msg.Name,
							Response: map[string]interface{}{
								"result": msg.Content,
							},
						},
					},
				},
			})
		}
	}

	// Add system instruction if present
	if systemInstruction != "" {
		geminiReq.SystemInstruction = &Content{
			Parts: []Part{
				{Text: systemInstruction},
			},
		}
	}

	// Convert tools
	if len(req.Tools) > 0 {
		geminiReq.Tools = make([]Tool, 1)
		geminiReq.Tools[0].FunctionDeclarations = make([]FunctionDeclaration, len(req.Tools))

		for i, tool := range req.Tools {
			geminiReq.Tools[0].FunctionDeclarations[i] = FunctionDeclaration{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			}
		}
	}

	return geminiReq
}

// convertResponse converts Gemini response to ModelResponse
func (g *Gemini) convertResponse(resp *GeminiResponse) *types.ModelResponse {
	if resp == nil {
		return &types.ModelResponse{
			Model: g.ID,
		}
	}

	modelResp := &types.ModelResponse{
		Model: g.ID,
		Usage: types.Usage{
			PromptTokens:     resp.UsageMetadata.PromptTokenCount,
			CompletionTokens: resp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      resp.UsageMetadata.TotalTokenCount,
		},
	}

	if len(resp.Candidates) == 0 {
		return modelResp
	}

	candidate := resp.Candidates[0]
	modelResp.Metadata.FinishReason = candidate.FinishReason
	if resp.UsageMetadata.ThoughtsTokenCount > 0 {
		if modelResp.Metadata.Extra == nil {
			modelResp.Metadata.Extra = make(map[string]interface{})
		}
		modelResp.Metadata.Extra["thoughts_token_count"] = resp.UsageMetadata.ThoughtsTokenCount
	}

	var contentBuilder strings.Builder
	var reasoningBuilder strings.Builder

	// Extract content, reasoning and tool calls
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			if part.Thought {
				if reasoningBuilder.Len() > 0 {
					reasoningBuilder.WriteString("\n")
				}
				reasoningBuilder.WriteString(part.Text)
			} else {
				contentBuilder.WriteString(part.Text)
			}
		}

		if part.FunctionCall != nil {
			argsJSON, _ := json.Marshal(part.FunctionCall.Args)
			modelResp.ToolCalls = append(modelResp.ToolCalls, types.ToolCall{
				ID:   generateToolCallID(),
				Type: "function",
				Function: types.ToolCallFunction{
					Name:      part.FunctionCall.Name,
					Arguments: string(argsJSON),
				},
			})
		}
	}

	content := strings.TrimSpace(contentBuilder.String())
	if content != "" {
		modelResp.Content = content
	}

	if reasoningBuilder.Len() > 0 {
		reasoning := strings.TrimSpace(reasoningBuilder.String())
		if reasoning != "" {
			rc := types.NewReasoningContent(reasoning)
			if resp.UsageMetadata.ThoughtsTokenCount > 0 {
				rc = rc.WithTokenCount(resp.UsageMetadata.ThoughtsTokenCount)
			}
			modelResp.ReasoningContent = rc
		}
	}

	return modelResp
}

// convertToChunk converts Gemini response to ResponseChunk for streaming
func (g *Gemini) convertToChunk(resp *GeminiResponse) types.ResponseChunk {
	chunk := types.ResponseChunk{}

	if len(resp.Candidates) == 0 {
		chunk.Done = true
		return chunk
	}

	candidate := resp.Candidates[0]

	// Check if done
	if candidate.FinishReason != "" && candidate.FinishReason != "STOP" {
		chunk.Done = true
		return chunk
	}

	// Extract content
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			chunk.Content += part.Text
		}

		if part.FunctionCall != nil {
			argsJSON, _ := json.Marshal(part.FunctionCall.Args)
			chunk.ToolCalls = append(chunk.ToolCalls, types.ToolCall{
				ID:   generateToolCallID(),
				Type: "function",
				Function: types.ToolCallFunction{
					Name:      part.FunctionCall.Name,
					Arguments: string(argsJSON),
				},
			})
		}
	}

	return chunk
}

// GeminiRequest represents the Gemini API request
type GeminiRequest struct {
	Contents          []Content         `json:"contents"`
	SystemInstruction *Content          `json:"systemInstruction,omitempty"`
	Tools             []Tool            `json:"tools,omitempty"`
	GenerationConfig  *GenerationConfig `json:"generationConfig,omitempty"`
	ThinkingConfig    *ThinkingConfig   `json:"thinkingConfig,omitempty"`
}

// Content represents content in the request/response
type Content struct {
	Role  string `json:"role,omitempty"`
	Parts []Part `json:"parts"`
}

// Part represents a part of the content
type Part struct {
	Text             string            `json:"text,omitempty"`
	FunctionCall     *FunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *FunctionResponse `json:"functionResponse,omitempty"`
	Thought          bool              `json:"thought,omitempty"`
}

// FunctionCall represents a function call
type FunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args,omitempty"`
}

// FunctionResponse represents a function response
type FunctionResponse struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

// Tool represents a tool definition
type Tool struct {
	FunctionDeclarations []FunctionDeclaration `json:"functionDeclarations,omitempty"`
}

// FunctionDeclaration represents a function declaration
type FunctionDeclaration struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// GenerationConfig represents generation configuration
type GenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	TopP            float64 `json:"topP,omitempty"`
	TopK            int     `json:"topK,omitempty"`
}

// ThinkingConfig represents reasoning configuration for Gemini thinking models
type ThinkingConfig struct {
	IncludeThoughts *bool `json:"includeThoughts,omitempty"`
	BudgetTokens    int   `json:"budgetTokens,omitempty"`
}

// GeminiResponse represents the Gemini API response
type GeminiResponse struct {
	Candidates    []Candidate   `json:"candidates"`
	UsageMetadata UsageMetadata `json:"usageMetadata"`
}

// Candidate represents a response candidate
type Candidate struct {
	Content      Content `json:"content"`
	FinishReason string  `json:"finishReason,omitempty"`
	Index        int     `json:"index"`
}

// UsageMetadata represents usage information
type UsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
	ThoughtsTokenCount   int `json:"thoughtsTokenCount,omitempty"`
}

// SSEDecoder decodes Server-Sent Events
type SSEDecoder struct {
	reader io.Reader
	buffer []byte
}

// NewSSEDecoder creates a new SSE decoder
func NewSSEDecoder(r io.Reader) *SSEDecoder {
	return &SSEDecoder{
		reader: r,
		buffer: make([]byte, 0),
	}
}

// Next reads the next SSE event
func (d *SSEDecoder) Next() ([]byte, error) {
	buf := make([]byte, 4096)
	for {
		n, err := d.reader.Read(buf)
		if n > 0 {
			d.buffer = append(d.buffer, buf[:n]...)

			// Look for complete SSE message
			for {
				idx := bytes.Index(d.buffer, []byte("\n\n"))
				if idx == -1 {
					break
				}

				message := d.buffer[:idx]
				d.buffer = d.buffer[idx+2:]

				// Extract data from SSE format
				lines := bytes.Split(message, []byte("\n"))
				for _, line := range lines {
					if bytes.HasPrefix(line, []byte("data: ")) {
						data := bytes.TrimPrefix(line, []byte("data: "))
						data = bytes.TrimSpace(data)
						if len(data) > 0 && !bytes.Equal(data, []byte("[DONE]")) {
							return data, nil
						}
					}
				}
			}
		}

		if err != nil {
			if err == io.EOF && len(d.buffer) > 0 {
				// Return remaining buffer
				data := d.buffer
				d.buffer = nil
				return data, io.EOF
			}
			return nil, err
		}
	}
}

// generateToolCallID generates a unique tool call ID
func generateToolCallID() string {
	return fmt.Sprintf("call_%d", len(strings.Repeat("x", 24)))
}

func toInt(value interface{}) (int, bool) {
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

func toBool(value interface{}) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		if v == "" {
			return false, false
		}
		if b, err := strconv.ParseBool(v); err == nil {
			return b, true
		}
	case int:
		return v != 0, true
	case int8:
		return v != 0, true
	case int16:
		return v != 0, true
	case int32:
		return v != 0, true
	case int64:
		return v != 0, true
	case uint:
		return v != 0, true
	case uint8:
		return v != 0, true
	case uint16:
		return v != 0, true
	case uint32:
		return v != 0, true
	case uint64:
		return v != 0, true
	case float32:
		return v != 0, true
	case float64:
		return v != 0, true
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return i != 0, true
		}
	}
	return false, false
}

// ValidateConfig validates the Gemini configuration
func ValidateConfig(config Config) error {
	if config.APIKey == "" {
		return fmt.Errorf("API key is required")
	}
	return nil
}
