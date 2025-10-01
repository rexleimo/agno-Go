package openai

import (
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
			modelID: "gpt-4o-mini",
			config: Config{
				APIKey: "test-key",
			},
			wantErr: false,
		},
		{
			name:    "missing API key",
			modelID: "gpt-4",
			config:  Config{},
			wantErr: true,
		},
		{
			name:    "with custom base URL",
			modelID: "gpt-3.5-turbo",
			config: Config{
				APIKey:  "test-key",
				BaseURL: "https://custom.openai.com",
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
				if model == nil {
					t.Error("New() returned nil model")
					return
				}
				if model.GetID() != tt.modelID {
					t.Errorf("GetID() = %v, want %v", model.GetID(), tt.modelID)
				}
				if model.GetProvider() != "openai" {
					t.Errorf("GetProvider() = %v, want openai", model.GetProvider())
				}
			}
		})
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
				APIKey: "sk-test",
			},
			wantErr: false,
		},
		{
			name:    "empty API key",
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

func TestOpenAI_buildChatRequest(t *testing.T) {
	model, err := New("gpt-4o-mini", Config{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	tests := []struct {
		name string
		req  *models.InvokeRequest
		want string // modelID
	}{
		{
			name: "basic request",
			req: &models.InvokeRequest{
				Messages: []*types.Message{
					types.NewUserMessage("Hello"),
				},
			},
			want: "gpt-4o-mini",
		},
		{
			name: "with temperature",
			req: &models.InvokeRequest{
				Messages: []*types.Message{
					types.NewUserMessage("Hello"),
				},
				Temperature: 0.8,
			},
			want: "gpt-4o-mini",
		},
		{
			name: "with max tokens",
			req: &models.InvokeRequest{
				Messages: []*types.Message{
					types.NewUserMessage("Hello"),
				},
				MaxTokens: 100,
			},
			want: "gpt-4o-mini",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatReq := model.buildChatRequest(tt.req)
			if chatReq.Model != tt.want {
				t.Errorf("buildChatRequest() model = %v, want %v", chatReq.Model, tt.want)
			}
			if len(chatReq.Messages) != len(tt.req.Messages) {
				t.Errorf("buildChatRequest() messages count = %v, want %v", len(chatReq.Messages), len(tt.req.Messages))
			}
			if tt.req.Temperature > 0 && chatReq.Temperature != float32(tt.req.Temperature) {
				t.Errorf("buildChatRequest() temperature = %v, want %v", chatReq.Temperature, tt.req.Temperature)
			}
			if tt.req.MaxTokens > 0 && chatReq.MaxTokens != tt.req.MaxTokens {
				t.Errorf("buildChatRequest() max_tokens = %v, want %v", chatReq.MaxTokens, tt.req.MaxTokens)
			}
		})
	}
}

func TestOpenAI_buildChatRequest_WithTools(t *testing.T) {
	model, err := New("gpt-4o-mini", Config{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	req := &models.InvokeRequest{
		Messages: []*types.Message{
			types.NewUserMessage("Calculate something"),
		},
		Tools: []models.ToolDefinition{
			{
				Type: "function",
				Function: models.FunctionSchema{
					Name:        "add",
					Description: "Add two numbers",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"a": map[string]interface{}{"type": "number"},
							"b": map[string]interface{}{"type": "number"},
						},
					},
				},
			},
		},
	}

	chatReq := model.buildChatRequest(req)

	if len(chatReq.Tools) != 1 {
		t.Errorf("buildChatRequest() tools count = %v, want 1", len(chatReq.Tools))
	}

	if chatReq.Tools[0].Function.Name != "add" {
		t.Errorf("buildChatRequest() tool name = %v, want add", chatReq.Tools[0].Function.Name)
	}
}

func TestOpenAI_buildChatRequest_WithToolCalls(t *testing.T) {
	model, err := New("gpt-4o-mini", Config{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	req := &models.InvokeRequest{
		Messages: []*types.Message{
			{
				Role: types.RoleAssistant,
				ToolCalls: []types.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Function: types.ToolCallFunction{
							Name:      "add",
							Arguments: `{"a": 1, "b": 2}`,
						},
					},
				},
			},
			{
				Role:       types.RoleTool,
				Content:    "3",
				ToolCallID: "call_1",
			},
		},
	}

	chatReq := model.buildChatRequest(req)

	if len(chatReq.Messages) != 2 {
		t.Fatalf("buildChatRequest() messages count = %v, want 2", len(chatReq.Messages))
	}

	// Check assistant message with tool calls
	if len(chatReq.Messages[0].ToolCalls) != 1 {
		t.Errorf("buildChatRequest() assistant tool calls = %v, want 1", len(chatReq.Messages[0].ToolCalls))
	}

	// Check tool message
	if chatReq.Messages[1].ToolCallID != "call_1" {
		t.Errorf("buildChatRequest() tool call ID = %v, want call_1", chatReq.Messages[1].ToolCallID)
	}
}

func TestOpenAI_GetProvider(t *testing.T) {
	model, err := New("gpt-4o-mini", Config{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	if model.GetProvider() != "openai" {
		t.Errorf("GetProvider() = %v, want openai", model.GetProvider())
	}
}

func TestOpenAI_GetID(t *testing.T) {
	modelID := "gpt-4o-mini"
	model, err := New(modelID, Config{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	if model.GetID() != modelID {
		t.Errorf("GetID() = %v, want %v", model.GetID(), modelID)
	}
}
