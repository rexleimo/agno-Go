package workflow

import (
	"context"
	"fmt"

	"github.com/yourusername/agno-go/pkg/agno/agent"
)

// Step represents a basic workflow step that executes an agent
type Step struct {
	ID          string
	Name        string
	Agent       *agent.Agent
	Description string
}

// StepConfig contains step configuration
type StepConfig struct {
	ID          string
	Name        string
	Agent       *agent.Agent
	Description string
}

// NewStep creates a new step
func NewStep(config StepConfig) (*Step, error) {
	if config.Agent == nil {
		return nil, fmt.Errorf("agent is required for step")
	}

	if config.ID == "" {
		config.ID = fmt.Sprintf("step-%s", config.Name)
	}

	if config.Name == "" {
		config.Name = config.ID
	}

	return &Step{
		ID:          config.ID,
		Name:        config.Name,
		Agent:       config.Agent,
		Description: config.Description,
	}, nil
}

// Execute runs the step
func (s *Step) Execute(ctx context.Context, execCtx *ExecutionContext) (*ExecutionContext, error) {
	// Use the current output as input, or initial input if no output yet
	input := execCtx.Output
	if input == "" {
		input = execCtx.Input
	}

	// Run the agent
	output, err := s.Agent.Run(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("step %s execution failed: %w", s.ID, err)
	}

	// Update execution context
	execCtx.Output = output.Content
	execCtx.Set(fmt.Sprintf("step_%s_output", s.ID), output.Content)

	return execCtx, nil
}

// GetID returns the step ID
func (s *Step) GetID() string {
	return s.ID
}

// GetType returns the node type
func (s *Step) GetType() NodeType {
	return NodeTypeStep
}
