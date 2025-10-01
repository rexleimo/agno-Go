package anthropic

import (
	"context"
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
