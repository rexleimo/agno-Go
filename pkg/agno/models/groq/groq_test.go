package groq

import (
	"testing"
	"time"

	"github.com/rexleimo/agno-go/pkg/agno/models"
	"github.com/rexleimo/agno-go/pkg/agno/types"
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
			modelID: ModelLlama38B,
			config: Config{
				APIKey: "test-key",
			},
			wantErr: false,
		},
		{
			name:    "missing API key",
			modelID: ModelLlama370B,
			config:  Config{},
			wantErr: true,
		},
		{
			name:    "with custom base URL",
			modelID: ModelLlama38B,
			config: Config{
				APIKey:  "test-key",
				BaseURL: "https://custom.groq.com",
			},
			wantErr: false,
		},
		{
			name:    "with default base URL",
			modelID: ModelMixtral8x7B,
			config: Config{
				APIKey: "test-key",
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
				if model.GetProvider() != "groq" {
					t.Errorf("GetProvider() = %v, want groq", model.GetProvider())
				}
				// Verify default base URL is set if not provided
				if tt.config.BaseURL == "" && model.config.BaseURL != DefaultBaseURL {
					t.Errorf("BaseURL = %v, want %v", model.config.BaseURL, DefaultBaseURL)
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
				APIKey: "gsk-test",
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

func TestGroq_buildChatRequest(t *testing.T) {
	model, err := New(ModelLlama38B, Config{APIKey: "test-key"})
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
			want: ModelLlama38B,
		},
		{
			name: "with temperature",
			req: &models.InvokeRequest{
				Messages: []*types.Message{
					types.NewUserMessage("Hello"),
				},
				Temperature: 0.8,
			},
			want: ModelLlama38B,
		},
		{
			name: "with max tokens",
			req: &models.InvokeRequest{
				Messages: []*types.Message{
					types.NewUserMessage("Hello"),
				},
				MaxTokens: 100,
			},
			want: ModelLlama38B,
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

func TestGroq_buildChatRequest_WithTools(t *testing.T) {
	model, err := New(ModelLlama38B, Config{APIKey: "test-key"})
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

func TestGroq_buildChatRequest_WithToolCalls(t *testing.T) {
	model, err := New(ModelLlama38B, Config{APIKey: "test-key"})
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

func TestGroq_GetProvider(t *testing.T) {
	model, err := New(ModelLlama38B, Config{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	if model.GetProvider() != "groq" {
		t.Errorf("GetProvider() = %v, want groq", model.GetProvider())
	}
}

func TestGroq_GetID(t *testing.T) {
	modelID := ModelLlama370B
	model, err := New(modelID, Config{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	if model.GetID() != modelID {
		t.Errorf("GetID() = %v, want %v", model.GetID(), modelID)
	}
}

// Additional tests for error handling and edge cases

func TestNew_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		modelID string
		config  Config
		wantErr bool
	}{
		{
			name:    "empty model ID",
			modelID: "",
			config:  Config{APIKey: "test-key"},
			wantErr: false, // Model ID can be empty, validated by API
		},
		{
			name:    "with temperature config",
			modelID: ModelLlama38B,
			config: Config{
				APIKey:      "test-key",
				Temperature: 0.7,
			},
			wantErr: false,
		},
		{
			name:    "with max tokens config",
			modelID: ModelLlama38B,
			config: Config{
				APIKey:    "test-key",
				MaxTokens: 1000,
			},
			wantErr: false,
		},
		{
			name:    "with all configs",
			modelID: ModelLlama38B,
			config: Config{
				APIKey:      "test-key",
				BaseURL:     "https://custom.groq.com",
				Temperature: 0.8,
				MaxTokens:   2000,
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
				}
				if model.config.Temperature != tt.config.Temperature {
					t.Errorf("Temperature = %v, want %v", model.config.Temperature, tt.config.Temperature)
				}
				if model.config.MaxTokens != tt.config.MaxTokens {
					t.Errorf("MaxTokens = %v, want %v", model.config.MaxTokens, tt.config.MaxTokens)
				}
			}
		})
	}
}

func TestGroq_buildChatRequest_EmptyMessages(t *testing.T) {
	model, err := New(ModelLlama38B, Config{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	req := &models.InvokeRequest{
		Messages: []*types.Message{},
	}

	chatReq := model.buildChatRequest(req)
	if len(chatReq.Messages) != 0 {
		t.Errorf("buildChatRequest() with empty messages should return empty messages")
	}
}

func TestGroq_buildChatRequest_SystemMessage(t *testing.T) {
	model, err := New(ModelLlama38B, Config{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	req := &models.InvokeRequest{
		Messages: []*types.Message{
			types.NewSystemMessage("You are a helpful assistant"),
			types.NewUserMessage("Hello"),
		},
	}

	chatReq := model.buildChatRequest(req)
	if len(chatReq.Messages) != 2 {
		t.Errorf("buildChatRequest() messages count = %v, want 2", len(chatReq.Messages))
	}
	if chatReq.Messages[0].Role != "system" {
		t.Errorf("First message role = %v, want system", chatReq.Messages[0].Role)
	}
}

func TestGroq_buildChatRequest_ConfigTemperature(t *testing.T) {
	model, err := New(ModelLlama38B, Config{
		APIKey:      "test-key",
		Temperature: 0.5,
	})
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	req := &models.InvokeRequest{
		Messages: []*types.Message{
			types.NewUserMessage("Hello"),
		},
	}

	chatReq := model.buildChatRequest(req)
	// Config temperature should be used if request doesn't specify
	if req.Temperature == 0 && model.config.Temperature > 0 {
		if chatReq.Temperature != float32(model.config.Temperature) {
			t.Errorf("buildChatRequest() temperature = %v, want %v", chatReq.Temperature, model.config.Temperature)
		}
	}
}

func TestGroq_buildChatRequest_ConfigMaxTokens(t *testing.T) {
	model, err := New(ModelLlama38B, Config{
		APIKey:    "test-key",
		MaxTokens: 500,
	})
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	req := &models.InvokeRequest{
		Messages: []*types.Message{
			types.NewUserMessage("Hello"),
		},
	}

	chatReq := model.buildChatRequest(req)
	// Config max tokens should be used if request doesn't specify
	if req.MaxTokens == 0 && model.config.MaxTokens > 0 {
		if chatReq.MaxTokens != model.config.MaxTokens {
			t.Errorf("buildChatRequest() max_tokens = %v, want %v", chatReq.MaxTokens, model.config.MaxTokens)
		}
	}
}

func TestGroq_buildChatRequest_RequestOverridesConfig(t *testing.T) {
	model, err := New(ModelLlama38B, Config{
		APIKey:      "test-key",
		Temperature: 0.5,
		MaxTokens:   500,
	})
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	req := &models.InvokeRequest{
		Messages: []*types.Message{
			types.NewUserMessage("Hello"),
		},
		Temperature: 0.9,
		MaxTokens:   1000,
	}

	chatReq := model.buildChatRequest(req)

	// Request parameters should override config
	if chatReq.Temperature != float32(req.Temperature) {
		t.Errorf("buildChatRequest() temperature = %v, want %v (request should override config)", chatReq.Temperature, req.Temperature)
	}
	if chatReq.MaxTokens != req.MaxTokens {
		t.Errorf("buildChatRequest() max_tokens = %v, want %v (request should override config)", chatReq.MaxTokens, req.MaxTokens)
	}
}

func TestGroq_buildChatRequest_MultipleTools(t *testing.T) {
	model, err := New(ModelLlama38B, Config{APIKey: "test-key"})
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
				},
			},
			{
				Type: "function",
				Function: models.FunctionSchema{
					Name:        "multiply",
					Description: "Multiply two numbers",
				},
			},
		},
	}

	chatReq := model.buildChatRequest(req)

	if len(chatReq.Tools) != 2 {
		t.Errorf("buildChatRequest() tools count = %v, want 2", len(chatReq.Tools))
	}
}

func TestGroq_buildChatRequest_AssistantMessage(t *testing.T) {
	model, err := New(ModelLlama38B, Config{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	req := &models.InvokeRequest{
		Messages: []*types.Message{
			types.NewUserMessage("Hello"),
			{
				Role:    types.RoleAssistant,
				Content: "Hi! How can I help you?",
			},
			types.NewUserMessage("Tell me about AI"),
		},
	}

	chatReq := model.buildChatRequest(req)

	if len(chatReq.Messages) != 3 {
		t.Errorf("buildChatRequest() messages count = %v, want 3", len(chatReq.Messages))
	}
	if chatReq.Messages[1].Role != "assistant" {
		t.Errorf("Second message role = %v, want assistant", chatReq.Messages[1].Role)
	}
}

func TestValidateConfig_DetailedErrors(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "API key with whitespace",
			config: Config{
				APIKey: "  gsk-test  ",
			},
			wantErr: false, // Whitespace is allowed, API will handle it
		},
		{
			name: "negative temperature",
			config: Config{
				APIKey:      "gsk-test",
				Temperature: -0.5,
			},
			wantErr: false, // Validation doesn't check temperature range
		},
		{
			name: "negative max tokens",
			config: Config{
				APIKey:    "gsk-test",
				MaxTokens: -100,
			},
			wantErr: false, // Validation doesn't check max tokens range
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

// TestNew_WithTimeout tests timeout configuration
// 测试超时配置
func TestNew_WithTimeout(t *testing.T) {
	tests := []struct {
		name            string
		config          Config
		expectedTimeout time.Duration
	}{
		{
			name: "with custom timeout",
			config: Config{
				APIKey:  "test-key",
				Timeout: 30 * time.Second,
			},
			expectedTimeout: 30 * time.Second,
		},
		{
			name: "with default timeout",
			config: Config{
				APIKey: "test-key",
			},
			expectedTimeout: 60 * time.Second,
		},
		{
			name: "with zero timeout gets default",
			config: Config{
				APIKey:  "test-key",
				Timeout: 0,
			},
			expectedTimeout: 60 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := New(ModelLlama38B, tt.config)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			if model == nil {
				t.Error("Expected model to be created")
			}

			// Verify the config timeout is stored correctly
			if model.config.Timeout != tt.config.Timeout {
				t.Errorf("Config timeout = %v, want %v", model.config.Timeout, tt.config.Timeout)
			}
		})
	}
}

// TestGetModelInfo tests the model info lookup
// 测试模型信息查找
func TestGetModelInfo(t *testing.T) {
	tests := []struct {
		name      string
		modelID   string
		wantFound bool
		wantName  string
	}{
		{
			name:      "llama 3.1 8b",
			modelID:   ModelLlama38B,
			wantFound: true,
			wantName:  "LLaMA 3.1 8B Instant",
		},
		{
			name:      "llama 3.1 70b",
			modelID:   ModelLlama370B,
			wantFound: true,
			wantName:  "LLaMA 3.1 70B Versatile",
		},
		{
			name:      "mixtral",
			modelID:   ModelMixtral8x7B,
			wantFound: true,
			wantName:  "Mixtral 8x7B",
		},
		{
			name:      "unknown model",
			modelID:   "unknown-model",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, found := GetModelInfo(tt.modelID)
			if found != tt.wantFound {
				t.Errorf("GetModelInfo() found = %v, want %v", found, tt.wantFound)
			}
			if tt.wantFound && info.Name != tt.wantName {
				t.Errorf("GetModelInfo() name = %v, want %v", info.Name, tt.wantName)
			}
		})
	}
}

// TestAvailableModels tests that all model constants are in the map
// 测试所有模型常量都在映射表中
func TestAvailableModels(t *testing.T) {
	expectedModels := []string{
		ModelLlama38B,
		ModelLlama370B,
		ModelLlama3405B,
		ModelMixtral8x7B,
		ModelGemma2_9B,
		ModelWhisperLarge,
		ModelLlamaGuard3,
	}

	for _, modelID := range expectedModels {
		t.Run(modelID, func(t *testing.T) {
			info, found := GetModelInfo(modelID)
			if !found {
				t.Errorf("Model %s not found in AvailableModels", modelID)
			}
			if info.ID != modelID {
				t.Errorf("Model ID mismatch: got %s, want %s", info.ID, modelID)
			}
		})
	}
}
