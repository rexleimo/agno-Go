package agent

import (
	"context"
	"testing"

	"github.com/yourusername/agno-go/pkg/agno/memory"
	"github.com/yourusername/agno-go/pkg/agno/models"
	"github.com/yourusername/agno-go/pkg/agno/tools/calculator"
	"github.com/yourusername/agno-go/pkg/agno/tools/toolkit"
	"github.com/yourusername/agno-go/pkg/agno/types"
)

// MockModel is a simple mock for testing
type MockModel struct {
	models.BaseModel
	InvokeFunc       func(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error)
	InvokeStreamFunc func(ctx context.Context, req *models.InvokeRequest) (<-chan types.ResponseChunk, error)
}

func (m *MockModel) Invoke(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
	if m.InvokeFunc != nil {
		return m.InvokeFunc(ctx, req)
	}
	return &types.ModelResponse{
		ID:      "test-response",
		Content: "Mock response",
		Model:   "mock-model",
	}, nil
}

func (m *MockModel) InvokeStream(ctx context.Context, req *models.InvokeRequest) (<-chan types.ResponseChunk, error) {
	if m.InvokeStreamFunc != nil {
		return m.InvokeStreamFunc(ctx, req)
	}
	ch := make(chan types.ResponseChunk)
	close(ch)
	return ch, nil
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				Name:  "TestAgent",
				Model: &MockModel{BaseModel: models.BaseModel{ID: "test", Provider: "mock"}},
			},
			wantErr: false,
		},
		{
			name: "missing model",
			config: Config{
				Name: "TestAgent",
			},
			wantErr: true,
			errMsg:  "model is required",
		},
		{
			name: "with default values",
			config: Config{
				Model: &MockModel{BaseModel: models.BaseModel{ID: "test", Provider: "mock"}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if agent == nil {
					t.Error("New() returned nil agent")
					return
				}
				// Check defaults
				if agent.MaxLoops <= 0 {
					t.Error("MaxLoops should have default value > 0")
				}
				if agent.Memory == nil {
					t.Error("Memory should be initialized")
				}
			}
		})
	}
}

func TestAgent_Run_SimpleResponse(t *testing.T) {
	mockModel := &MockModel{
		BaseModel: models.BaseModel{ID: "test", Provider: "mock"},
		InvokeFunc: func(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
			return &types.ModelResponse{
				ID:      "test-1",
				Content: "Hello, this is a test response",
				Model:   "test",
			}, nil
		},
	}

	agent, err := New(Config{
		Name:  "TestAgent",
		Model: mockModel,
	})
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	output, err := agent.Run(context.Background(), "Hello")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if output.Content != "Hello, this is a test response" {
		t.Errorf("Run() content = %v, want %v", output.Content, "Hello, this is a test response")
	}

	if len(output.Messages) < 2 {
		t.Errorf("Run() should have at least 2 messages (user + assistant)")
	}
}

func TestAgent_Run_EmptyInput(t *testing.T) {
	mockModel := &MockModel{
		BaseModel: models.BaseModel{ID: "test", Provider: "mock"},
	}

	agent, err := New(Config{
		Name:  "TestAgent",
		Model: mockModel,
	})
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	_, err = agent.Run(context.Background(), "")
	if err == nil {
		t.Error("Run() should return error for empty input")
	}
}

func TestAgent_Run_WithToolCalls(t *testing.T) {
	callCount := 0
	mockModel := &MockModel{
		BaseModel: models.BaseModel{ID: "test", Provider: "mock"},
		InvokeFunc: func(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
			callCount++

			// First call: return tool call
			if callCount == 1 {
				return &types.ModelResponse{
					ID:    "test-1",
					Model: "test",
					ToolCalls: []types.ToolCall{
						{
							ID:   "call_1",
							Type: "function",
							Function: types.ToolCallFunction{
								Name:      "add",
								Arguments: `{"a": 5, "b": 3}`,
							},
						},
					},
				}, nil
			}

			// Second call: return final answer
			return &types.ModelResponse{
				ID:      "test-2",
				Content: "The result is 8",
				Model:   "test",
			}, nil
		},
	}

	agent, err := New(Config{
		Name:     "TestAgent",
		Model:    mockModel,
		Toolkits: []toolkit.Toolkit{calculator.New()},
	})
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	output, err := agent.Run(context.Background(), "What is 5 + 3?")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 model calls (tool call + final), got %d", callCount)
	}

	if output.Content != "The result is 8" {
		t.Errorf("Run() content = %v, want %v", output.Content, "The result is 8")
	}

	// Check metadata
	loops, ok := output.Metadata["loops"].(int)
	if !ok || loops != 2 {
		t.Errorf("Run() loops = %v, want 2", loops)
	}
}

func TestAgent_Run_MaxLoops(t *testing.T) {
	mockModel := &MockModel{
		BaseModel: models.BaseModel{ID: "test", Provider: "mock"},
		InvokeFunc: func(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
			// Always return tool calls to trigger max loops
			return &types.ModelResponse{
				ID:    "test",
				Model: "test",
				ToolCalls: []types.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						Function: types.ToolCallFunction{
							Name:      "add",
							Arguments: `{"a": 1, "b": 1}`,
						},
					},
				},
			}, nil
		},
	}

	agent, err := New(Config{
		Name:     "TestAgent",
		Model:    mockModel,
		Toolkits: []toolkit.Toolkit{calculator.New()},
		MaxLoops: 3,
	})
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	_, err = agent.Run(context.Background(), "Test")
	if err == nil {
		t.Error("Run() should return error when max loops reached")
	}
}

func TestAgent_ClearMemory(t *testing.T) {
	mockModel := &MockModel{
		BaseModel: models.BaseModel{ID: "test", Provider: "mock"},
	}

	agent, err := New(Config{
		Name:         "TestAgent",
		Model:        mockModel,
		Instructions: "You are a helpful assistant",
	})
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Add some messages
	agent.Memory.Add(types.NewUserMessage("Hello"))
	agent.Memory.Add(types.NewAssistantMessage("Hi there"))

	if len(agent.Memory.GetMessages()) < 3 { // system + user + assistant
		t.Error("Should have at least 3 messages")
	}

	// Clear memory
	agent.ClearMemory()

	messages := agent.Memory.GetMessages()
	if len(messages) != 1 {
		t.Errorf("After clear, should have 1 message (system), got %d", len(messages))
	}

	if messages[0].Role != types.RoleSystem {
		t.Error("First message after clear should be system message")
	}
}

func TestAgent_WithCustomMemory(t *testing.T) {
	mockModel := &MockModel{
		BaseModel: models.BaseModel{ID: "test", Provider: "mock"},
	}

	customMemory := memory.NewInMemory(5)
	customMemory.Add(types.NewUserMessage("Previous message"))

	agent, err := New(Config{
		Name:   "TestAgent",
		Model:  mockModel,
		Memory: customMemory,
	})
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	messages := agent.Memory.GetMessages()
	if len(messages) != 1 {
		t.Errorf("Should preserve custom memory, got %d messages", len(messages))
	}
}
