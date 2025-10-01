package agent

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/yourusername/agno-go/pkg/agno/memory"
	"github.com/yourusername/agno-go/pkg/agno/models"
	"github.com/yourusername/agno-go/pkg/agno/tools/toolkit"
	"github.com/yourusername/agno-go/pkg/agno/types"
)

// Agent represents an AI agent
type Agent struct {
	ID           string
	Name         string
	Model        models.Model
	Toolkits     []toolkit.Toolkit
	Memory       memory.Memory
	Instructions string
	MaxLoops     int // Maximum tool calling loops
	logger       *slog.Logger
}

// Config contains agent configuration
type Config struct {
	ID           string
	Name         string
	Model        models.Model
	Toolkits     []toolkit.Toolkit
	Memory       memory.Memory
	Instructions string
	MaxLoops     int
	Logger       *slog.Logger
}

// New creates a new agent
func New(config Config) (*Agent, error) {
	if config.Model == nil {
		return nil, types.NewInvalidConfigError("model is required", nil)
	}

	if config.ID == "" {
		config.ID = fmt.Sprintf("agent-%s", config.Model.GetID())
	}

	if config.Name == "" {
		config.Name = config.ID
	}

	if config.Memory == nil {
		config.Memory = memory.NewInMemory(100)
	}

	if config.MaxLoops <= 0 {
		config.MaxLoops = 10
	}

	if config.Logger == nil {
		config.Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	agent := &Agent{
		ID:           config.ID,
		Name:         config.Name,
		Model:        config.Model,
		Toolkits:     config.Toolkits,
		Memory:       config.Memory,
		Instructions: config.Instructions,
		MaxLoops:     config.MaxLoops,
		logger:       config.Logger,
	}

	// Add system message if instructions provided
	if config.Instructions != "" {
		agent.Memory.Add(types.NewSystemMessage(config.Instructions))
	}

	return agent, nil
}

// RunOutput contains the result of agent execution
type RunOutput struct {
	Content  string                 `json:"content"`
	Messages []*types.Message       `json:"messages"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Run executes the agent with the given input
func (a *Agent) Run(ctx context.Context, input string) (*RunOutput, error) {
	if input == "" {
		return nil, types.NewInvalidInputError("input cannot be empty", nil)
	}

	a.logger.Info("agent run started", "agent_id", a.ID, "input", input)

	// Add user message
	userMsg := types.NewUserMessage(input)
	a.Memory.Add(userMsg)

	var finalResponse *types.ModelResponse
	loopCount := 0

	// Tool calling loop
	for loopCount < a.MaxLoops {
		loopCount++

		// Prepare request
		req := &models.InvokeRequest{
			Messages: a.Memory.GetMessages(),
		}

		// Add tools if available
		if len(a.Toolkits) > 0 {
			req.Tools = toolkit.ToModelToolDefinitions(a.Toolkits)
		}

		// Call model
		resp, err := a.Model.Invoke(ctx, req)
		if err != nil {
			a.logger.Error("model invocation failed", "error", err)
			return nil, types.NewAPIError("model invocation failed", err)
		}

		// Store assistant response
		assistantMsg := &types.Message{
			Role:      types.RoleAssistant,
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		}
		a.Memory.Add(assistantMsg)

		// Check if there are tool calls
		if !resp.HasToolCalls() {
			finalResponse = resp
			break
		}

		// Execute tool calls
		a.logger.Info("executing tool calls", "count", len(resp.ToolCalls))
		if err := a.executeToolCalls(ctx, resp.ToolCalls); err != nil {
			a.logger.Error("tool execution failed", "error", err)
			return nil, types.NewToolExecutionError("tool execution failed", err)
		}

		// Continue loop to process tool results
	}

	if loopCount >= a.MaxLoops {
		a.logger.Warn("max loops reached", "max_loops", a.MaxLoops)
		return nil, types.NewError(types.ErrCodeUnknown, "max tool calling loops reached", nil)
	}

	if finalResponse == nil {
		return nil, types.NewError(types.ErrCodeUnknown, "no response from model", nil)
	}

	a.logger.Info("agent run completed", "agent_id", a.ID)

	return &RunOutput{
		Content:  finalResponse.Content,
		Messages: a.Memory.GetMessages(),
		Metadata: map[string]interface{}{
			"loops": loopCount,
			"usage": finalResponse.Usage,
		},
	}, nil
}

// executeToolCalls executes all tool calls and adds results to memory
func (a *Agent) executeToolCalls(ctx context.Context, toolCalls []types.ToolCall) error {
	for _, tc := range toolCalls {
		// Find the toolkit that has this function
		var targetToolkit toolkit.Toolkit
		for _, tk := range a.Toolkits {
			if _, exists := tk.Functions()[tc.Function.Name]; exists {
				targetToolkit = tk
				break
			}
		}

		if targetToolkit == nil {
			errMsg := fmt.Sprintf("function %s not found in any toolkit", tc.Function.Name)
			a.logger.Warn("tool not found", "function", tc.Function.Name)
			a.Memory.Add(types.NewToolMessage(tc.ID, errMsg))
			continue
		}

		// Parse arguments
		args, err := toolkit.ParseArguments(tc.Function.Arguments)
		if err != nil {
			errMsg := fmt.Sprintf("failed to parse arguments: %v", err)
			a.logger.Error("argument parsing failed", "error", err)
			a.Memory.Add(types.NewToolMessage(tc.ID, errMsg))
			continue
		}

		// Execute tool
		a.logger.Info("executing tool", "function", tc.Function.Name, "args", args)

		// Get the function and execute it directly
		fn := targetToolkit.Functions()[tc.Function.Name]
		if fn == nil {
			errMsg := fmt.Sprintf("function %s not found", tc.Function.Name)
			a.logger.Error("function not found", "function", tc.Function.Name)
			a.Memory.Add(types.NewToolMessage(tc.ID, errMsg))
			continue
		}

		result, err := fn.Handler(ctx, args)
		if err != nil {
			errMsg := fmt.Sprintf("tool execution error: %v", err)
			a.logger.Error("tool execution failed", "function", tc.Function.Name, "error", err)
			a.Memory.Add(types.NewToolMessage(tc.ID, errMsg))
			continue
		}

		// Format and store result
		resultStr, err := toolkit.FormatResult(result)
		if err != nil {
			resultStr = fmt.Sprintf("%v", result)
		}

		a.logger.Info("tool executed successfully", "function", tc.Function.Name)
		a.Memory.Add(types.NewToolMessage(tc.ID, resultStr))
	}

	return nil
}

// ClearMemory clears the agent's conversation history
func (a *Agent) ClearMemory() {
	a.Memory.Clear()
	// Re-add system message
	if a.Instructions != "" {
		a.Memory.Add(types.NewSystemMessage(a.Instructions))
	}
}
