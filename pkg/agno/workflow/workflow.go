package workflow

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/rexleimo/agno-go/pkg/agno/types"
)

// Workflow represents a multi-step process
type Workflow struct {
	ID     string
	Name   string
	Steps  []Node
	logger *slog.Logger
}

// Node represents a node in the workflow graph
type Node interface {
	Execute(ctx context.Context, input *ExecutionContext) (*ExecutionContext, error)
	GetID() string
	GetType() NodeType
}

// NodeType represents the type of workflow node
type NodeType string

const (
	NodeTypeStep      NodeType = "step"
	NodeTypeCondition NodeType = "condition"
	NodeTypeLoop      NodeType = "loop"
	NodeTypeParallel  NodeType = "parallel"
	NodeTypeRouter    NodeType = "router"
)

// ExecutionContext holds the execution state and data
type ExecutionContext struct {
	Input    string                 `json:"input"`
	Output   string                 `json:"output"`
	Data     map[string]interface{} `json:"data"`
	Metadata map[string]interface{} `json:"metadata"`
}

// NewExecutionContext creates a new execution context
func NewExecutionContext(input string) *ExecutionContext {
	return &ExecutionContext{
		Input:    input,
		Data:     make(map[string]interface{}),
		Metadata: make(map[string]interface{}),
	}
}

// Set stores a value in the context
func (ec *ExecutionContext) Set(key string, value interface{}) {
	ec.Data[key] = value
}

// Get retrieves a value from the context
func (ec *ExecutionContext) Get(key string) (interface{}, bool) {
	val, ok := ec.Data[key]
	return val, ok
}

// Config contains workflow configuration
type Config struct {
	ID     string
	Name   string
	Steps  []Node
	Logger *slog.Logger
}

// New creates a new workflow
func New(config Config) (*Workflow, error) {
	if config.ID == "" {
		config.ID = fmt.Sprintf("workflow-%s", config.Name)
	}

	if config.Name == "" {
		config.Name = config.ID
	}

	if config.Logger == nil {
		config.Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	return &Workflow{
		ID:     config.ID,
		Name:   config.Name,
		Steps:  config.Steps,
		logger: config.Logger,
	}, nil
}

// Run executes the workflow
func (w *Workflow) Run(ctx context.Context, input string) (*ExecutionContext, error) {
	if input == "" {
		return nil, types.NewInvalidInputError("input cannot be empty", nil)
	}

	w.logger.Info("workflow started", "workflow_id", w.ID, "steps", len(w.Steps))

	execCtx := NewExecutionContext(input)

	for i, step := range w.Steps {
		w.logger.Info("executing step", "step_id", step.GetID(), "step_type", step.GetType(), "sequence", i+1)

		result, err := step.Execute(ctx, execCtx)
		if err != nil {
			w.logger.Error("step execution failed", "step_id", step.GetID(), "error", err)
			return nil, types.NewError(types.ErrCodeUnknown, fmt.Sprintf("step %s failed", step.GetID()), err)
		}

		execCtx = result
	}

	w.logger.Info("workflow completed", "workflow_id", w.ID)
	return execCtx, nil
}

// AddStep adds a step to the workflow
func (w *Workflow) AddStep(step Node) {
	w.Steps = append(w.Steps, step)
}
