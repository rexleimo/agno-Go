package anthropic

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/yourusername/agno-go/pkg/agno/models"
	"github.com/yourusername/agno-go/pkg/agno/types"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		modelID string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid config",
			modelID: "claude-3-opus-20240229",
			config: Config{
				APIKey: "test-key",
			},
			wantErr: false,
		},
		{
			name:    "missing API key",
			modelID: "claude-3-opus-20240229",
			config:  Config{},
			wantErr: true,
		},
		{
			name:    "custom base URL",
			modelID: "claude-3-sonnet-20240229",
			config: Config{
				APIKey:  "test-key",
				BaseURL: "https://custom.api.com",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := New(tt.modelID, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if model.GetID() != tt.modelID {
					t.Errorf("GetID() = %v, want %v", model.GetID(), tt.modelID)
				}
				if model.GetProvider() != "anthropic" {
					t.Errorf("GetProvider() = %v, want anthropic", model.GetProvider())
				}
			}
		})
	}
}

func TestBuildClaudeRequest(t *testing.T) {
	model, _ := New("claude-3-opus-20240229", Config{
		APIKey:      "test-key",
		Temperature: 0.7,
		MaxTokens:   1000,
	})

	tests := []struct {
		name string
		req  *models.InvokeRequest
		want int // expected number of messages
	}{
		{
			name: "basic messages",
			req: &models.InvokeRequest{
				Messages: []*types.Message{
					{Role: types.RoleSystem, Content: "You are helpful"},
					{Role: types.RoleUser, Content: "Hello"},
				},
			},
			want: 1, // system message becomes separate field
		},
		{
			name: "with tools",
			req: &models.InvokeRequest{
				Messages: []*types.Message{
					{Role: types.RoleUser, Content: "Calculate 2+2"},
				},
				Tools: []models.ToolDefinition{
					{
						Type: "function",
						Function: models.FunctionSchema{
							Name:        "calculator",
							Description: "Perform calculations",
							Parameters: map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"operation": map[string]interface{}{
										"type": "string",
									},
								},
							},
						},
					},
				},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claudeReq := model.buildClaudeRequest(tt.req)
			if len(claudeReq.Messages) != tt.want {
				t.Errorf("buildClaudeRequest() messages = %v, want %v", len(claudeReq.Messages), tt.want)
			}
		})
	}
}

func TestConvertResponse(t *testing.T) {
	model, _ := New("claude-3-opus-20240229", Config{APIKey: "test-key"})

	claudeResp := &ClaudeResponse{
		ID:    "msg_123",
		Model: "claude-3-opus-20240229",
		Content: []ContentBlock{
			{
				Type: "text",
				Text: "Hello, world!",
			},
		},
		StopReason: "end_turn",
		Usage: ClaudeUsage{
			InputTokens:  10,
			OutputTokens: 5,
		},
	}

	modelResp := model.convertResponse(claudeResp)

	if modelResp.Content != "Hello, world!" {
		t.Errorf("Content = %v, want 'Hello, world!'", modelResp.Content)
	}
	if modelResp.Usage.TotalTokens != 15 {
		t.Errorf("TotalTokens = %v, want 15", modelResp.Usage.TotalTokens)
	}
}

func TestConvertResponseWithToolCalls(t *testing.T) {
	model, _ := New("claude-3-opus-20240229", Config{APIKey: "test-key"})

	claudeResp := &ClaudeResponse{
		ID:    "msg_123",
		Model: "claude-3-opus-20240229",
		Content: []ContentBlock{
			{
				Type: "text",
				Text: "I'll calculate that for you.",
			},
			{
				Type: "tool_use",
				ID:   "call_123",
				Name: "calculator",
				Input: map[string]interface{}{
					"operation": "add",
					"a":         2,
					"b":         2,
				},
			},
		},
		StopReason: "tool_use",
		Usage: ClaudeUsage{
			InputTokens:  15,
			OutputTokens: 20,
		},
	}

	modelResp := model.convertResponse(claudeResp)

	if len(modelResp.ToolCalls) != 1 {
		t.Errorf("ToolCalls length = %v, want 1", len(modelResp.ToolCalls))
	}
	if modelResp.ToolCalls[0].Function.Name != "calculator" {
		t.Errorf("ToolCall name = %v, want 'calculator'", modelResp.ToolCalls[0].Function.Name)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				APIKey: "test-key",
			},
			wantErr: false,
		},
		{
			name:    "missing API key",
			config:  Config{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetHeaders(t *testing.T) {
	model, _ := New("claude-3-opus-20240229", Config{
		APIKey: "test-api-key",
	})

	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil)
	model.setHeaders(req)

	if req.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %v, want application/json", req.Header.Get("Content-Type"))
	}
	if req.Header.Get("x-api-key") != "test-api-key" {
		t.Errorf("x-api-key = %v, want test-api-key", req.Header.Get("x-api-key"))
	}
	if req.Header.Get("anthropic-version") != apiVersion {
		t.Errorf("anthropic-version = %v, want %v", req.Header.Get("anthropic-version"), apiVersion)
	}
}

func TestConvertStreamEvent(t *testing.T) {
	model, _ := New("claude-3-opus-20240229", Config{APIKey: "test-key"})

	tests := []struct {
		name      string
		event     *StreamEvent
		wantDone  bool
		wantError bool
		wantText  string
	}{
		{
			name: "text delta",
			event: &StreamEvent{
				Type: "content_block_delta",
				Delta: StreamDelta{
					Type: "text_delta",
					Text: "Hello",
				},
			},
			wantDone: false,
			wantText: "Hello",
		},
		{
			name: "message stop",
			event: &StreamEvent{
				Type: "message_stop",
			},
			wantDone: true,
		},
		{
			name: "error event",
			event: &StreamEvent{
				Type: "error",
				Error: StreamError{
					Type:    "api_error",
					Message: "Rate limit exceeded",
				},
			},
			wantDone:  true,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := model.convertStreamEvent(tt.event)
			if chunk.Done != tt.wantDone {
				t.Errorf("Done = %v, want %v", chunk.Done, tt.wantDone)
			}
			if (chunk.Error != nil) != tt.wantError {
				t.Errorf("Error = %v, wantError %v", chunk.Error, tt.wantError)
			}
			if chunk.Content != tt.wantText {
				t.Errorf("Content = %v, want %v", chunk.Content, tt.wantText)
			}
		})
	}
}

func TestJsonToString(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]interface{}
		want  string
	}{
		{
			name:  "nil map",
			input: nil,
			want:  "{}",
		},
		{
			name: "simple map",
			input: map[string]interface{}{
				"key": "value",
			},
			want: `{"key":"value"}`,
		},
		{
			name: "nested map",
			input: map[string]interface{}{
				"a": 1,
				"b": map[string]interface{}{
					"c": 2,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := jsonToString(tt.input)
			if tt.name == "nil map" && got != tt.want {
				t.Errorf("jsonToString() = %v, want %v", got, tt.want)
			}
			if tt.name == "simple map" && got != tt.want {
				t.Errorf("jsonToString() = %v, want %v", got, tt.want)
			}
			// For nested map, just check it's valid JSON
			if tt.name == "nested map" {
				var parsed map[string]interface{}
				if err := json.Unmarshal([]byte(got), &parsed); err != nil {
					t.Errorf("jsonToString() produced invalid JSON: %v", err)
				}
			}
		})
	}
}

func TestBuildClaudeRequest_WithTemperature(t *testing.T) {
	model, _ := New("claude-3-opus-20240229", Config{
		APIKey:      "test-key",
		Temperature: 0.5,
	})

	tests := []struct {
		name        string
		req         *models.InvokeRequest
		wantTemp    float64
		wantTokens  int
	}{
		{
			name: "use request temperature",
			req: &models.InvokeRequest{
				Messages: []*types.Message{
					{Role: types.RoleUser, Content: "Hello"},
				},
				Temperature: 0.9,
			},
			wantTemp: 0.9,
		},
		{
			name: "use config temperature",
			req: &models.InvokeRequest{
				Messages: []*types.Message{
					{Role: types.RoleUser, Content: "Hello"},
				},
			},
			wantTemp: 0.5,
		},
		{
			name: "use request max tokens",
			req: &models.InvokeRequest{
				Messages: []*types.Message{
					{Role: types.RoleUser, Content: "Hello"},
				},
				MaxTokens: 500,
			},
			wantTokens: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claudeReq := model.buildClaudeRequest(tt.req)
			if tt.wantTemp > 0 && claudeReq.Temperature != tt.wantTemp {
				t.Errorf("Temperature = %v, want %v", claudeReq.Temperature, tt.wantTemp)
			}
			if tt.wantTokens > 0 && claudeReq.MaxTokens != tt.wantTokens {
				t.Errorf("MaxTokens = %v, want %v", claudeReq.MaxTokens, tt.wantTokens)
			}
		})
	}
}

func TestBuildClaudeRequest_ToolMessages(t *testing.T) {
	model, _ := New("claude-3-opus-20240229", Config{APIKey: "test-key"})

	req := &models.InvokeRequest{
		Messages: []*types.Message{
			{Role: types.RoleUser, Content: "Calculate 2+2"},
			{Role: types.RoleTool, Content: "4", ToolCallID: "call_123"},
		},
	}

	claudeReq := model.buildClaudeRequest(req)

	// Tool messages should be converted to user messages
	if len(claudeReq.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(claudeReq.Messages))
	}

	// Second message should be user role with tool result
	if claudeReq.Messages[1].Role != "user" {
		t.Errorf("Tool message role = %v, want user", claudeReq.Messages[1].Role)
	}
}

func TestGetProvider(t *testing.T) {
	model, _ := New("claude-3-opus-20240229", Config{APIKey: "test-key"})
	if model.GetProvider() != "anthropic" {
		t.Errorf("GetProvider() = %v, want anthropic", model.GetProvider())
	}
}

func TestGetID(t *testing.T) {
	modelID := "claude-3-sonnet-20240229"
	model, _ := New(modelID, Config{APIKey: "test-key"})
	if model.GetID() != modelID {
		t.Errorf("GetID() = %v, want %v", model.GetID(), modelID)
	}
}

func TestConvertResponse_MultipleTextBlocks(t *testing.T) {
	model, _ := New("claude-3-opus-20240229", Config{APIKey: "test-key"})

	claudeResp := &ClaudeResponse{
		ID:    "msg_123",
		Model: "claude-3-opus-20240229",
		Content: []ContentBlock{
			{Type: "text", Text: "First part. "},
			{Type: "text", Text: "Second part."},
		},
		StopReason: "end_turn",
		Usage: ClaudeUsage{
			InputTokens:  10,
			OutputTokens: 15,
		},
	}

	modelResp := model.convertResponse(claudeResp)

	expectedContent := "First part. Second part."
	if modelResp.Content != expectedContent {
		t.Errorf("Content = %v, want %v", modelResp.Content, expectedContent)
	}
}

func TestConvertResponse_EmptyContent(t *testing.T) {
	model, _ := New("claude-3-opus-20240229", Config{APIKey: "test-key"})

	claudeResp := &ClaudeResponse{
		ID:         "msg_123",
		Model:      "claude-3-opus-20240229",
		Content:    []ContentBlock{},
		StopReason: "end_turn",
		Usage: ClaudeUsage{
			InputTokens:  5,
			OutputTokens: 0,
		},
	}

	modelResp := model.convertResponse(claudeResp)

	if modelResp.Content != "" {
		t.Errorf("Content = %v, want empty string", modelResp.Content)
	}
	if len(modelResp.ToolCalls) != 0 {
		t.Errorf("ToolCalls length = %v, want 0", len(modelResp.ToolCalls))
	}
}

// Integration test - only runs if ANTHROPIC_API_KEY is set
func TestInvoke_Integration(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: ANTHROPIC_API_KEY not set")
	}

	model, err := New("claude-3-haiku-20240307", Config{
		APIKey:      apiKey,
		Temperature: 0.7,
		MaxTokens:   100,
	})
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	req := &models.InvokeRequest{
		Messages: []*types.Message{
			{Role: types.RoleUser, Content: "Say hello in one word"},
		},
	}

	resp, err := model.Invoke(context.Background(), req)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if resp.Content == "" {
		t.Error("Response content is empty")
	}
	if resp.Usage.TotalTokens == 0 {
		t.Error("Usage tokens should be > 0")
	}

	t.Logf("Response: %s", resp.Content)
	t.Logf("Tokens: %d", resp.Usage.TotalTokens)
}
