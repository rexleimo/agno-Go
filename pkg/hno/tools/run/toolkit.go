package run

import (
	"context"
	"fmt"
	"time"

	"github.com/rexleimo/agno-go/pkg/hno/tools/toolkit"
)

// CodeExecutionToolkit exposes a fail-closed run_code tool to agents. It
// receives an already-configured Executor so provider selection and Cloud
// billing remain an application-owned decision.
type CodeExecutionToolkit struct {
	*toolkit.BaseToolkit
	executor    Executor
	executorErr error
}

// NewToolkit creates a code execution toolkit. A nil executor is retained as
// a configuration error so every run_code call fails closed rather than
// falling back to a host process.
func NewToolkit(executor Executor) *CodeExecutionToolkit {
	t := &CodeExecutionToolkit{BaseToolkit: toolkit.NewBaseToolkit("code_execution")}
	if executor == nil {
		t.executorErr = fmt.Errorf("sandbox executor cannot be nil")
	} else {
		t.executor = executor
	}
	t.registerFunctions()
	return t
}

// NewToolkitWithConfig creates a toolkit from explicit provider configuration.
func NewToolkitWithConfig(config Config) (*CodeExecutionToolkit, error) {
	executor, err := NewExecutor(config)
	if err != nil {
		return nil, err
	}
	return NewToolkit(executor), nil
}

// Close closes the configured executor.
func (t *CodeExecutionToolkit) Close() error {
	if t.executor == nil {
		return nil
	}
	return t.executor.Close()
}

func (t *CodeExecutionToolkit) registerFunctions() {
	t.RegisterFunction(&toolkit.Function{
		Name:        "run_code",
		Description: "Run code in a configured disposable sandbox. Supported runtimes are python, node, and shell. The executor fails closed when no sandbox provider is configured.",
		Parameters: map[string]toolkit.Parameter{
			"runtime": {
				Type:        "string",
				Description: "Runtime template to use",
				Required:    true,
				Enum:        []string{"python", "node", "shell"},
			},
			"code": {
				Type:        "string",
				Description: "Code or shell script to execute",
				Required:    true,
			},
			"timeout_seconds": {
				Type:        "number",
				Description: "Hard execution deadline in seconds (default: 30)",
				Required:    false,
			},
		},
		Handler: t.runCode,
	})
}

func (t *CodeExecutionToolkit) runCode(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	if t.executorErr != nil {
		return nil, fmt.Errorf("code execution sandbox configuration: %w", t.executorErr)
	}
	runtime, ok := args["runtime"].(string)
	if !ok {
		return nil, fmt.Errorf("runtime must be a string")
	}
	code, ok := args["code"].(string)
	if !ok {
		return nil, fmt.Errorf("code must be a string")
	}

	spec := Spec{Runtime: runtime, Code: code}
	if value, exists := args["timeout_seconds"]; exists {
		seconds, ok := value.(float64)
		if !ok || seconds <= 0 {
			return nil, fmt.Errorf("timeout_seconds must be a positive number")
		}
		spec.Timeout = time.Duration(seconds * float64(time.Second))
	}

	result, err := t.executor.Run(ctx, spec)
	if err != nil {
		return map[string]interface{}{
			"exit_code": result.ExitCode,
			"stdout":    result.Stdout,
			"stderr":    result.Stderr,
			"timed_out": result.TimedOut,
			"error":     err.Error(),
		}, nil
	}
	return map[string]interface{}{
		"exit_code": result.ExitCode,
		"stdout":    result.Stdout,
		"stderr":    result.Stderr,
		"timed_out": result.TimedOut,
	}, nil
}
