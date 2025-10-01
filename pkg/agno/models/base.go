package models

import (
	"context"

	"github.com/yourusername/agno-go/pkg/agno/types"
)

// InvokeRequest contains parameters for model invocation
type InvokeRequest struct {
	Messages    []*types.Message
	Tools       []ToolDefinition
	Temperature float64
	MaxTokens   int
	Stream      bool
	Extra       map[string]interface{}
}

// ToolDefinition defines a tool that can be called by the model
type ToolDefinition struct {
	Type     string         `json:"type"` // "function"
	Function FunctionSchema `json:"function"`
}

// FunctionSchema defines the schema of a callable function
type FunctionSchema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// Model represents a language model interface
type Model interface {
	// Invoke calls the model synchronously
	Invoke(ctx context.Context, req *InvokeRequest) (*types.ModelResponse, error)

	// InvokeStream calls the model with streaming response
	InvokeStream(ctx context.Context, req *InvokeRequest) (<-chan types.ResponseChunk, error)

	// GetProvider returns the model provider name
	GetProvider() string

	// GetID returns the model identifier
	GetID() string

	// GetName returns the model name
	GetName() string
}

// BaseModel provides common functionality for model implementations
type BaseModel struct {
	ID       string
	Name     string
	Provider string
}

// GetProvider returns the model provider
func (m *BaseModel) GetProvider() string {
	return m.Provider
}

// GetID returns the model ID
func (m *BaseModel) GetID() string {
	return m.ID
}

// GetName returns the model name
func (m *BaseModel) GetName() string {
	if m.Name != "" {
		return m.Name
	}
	return m.ID
}
