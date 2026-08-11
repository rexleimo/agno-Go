package run

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// localExecutor runs payloads directly on the host. It provides no isolation
// and must be constructed explicitly for development only; NewExecutor never
// falls back to it.
type localExecutor struct {
	timeout     time.Duration
	outputLimit int64
}

// NewLocalExecutor creates an explicitly unsandboxed executor for local
// development. Production deployments must use NewExecutor instead.
func NewLocalExecutor() Executor {
	return &localExecutor{
		timeout:     DefaultTimeout,
		outputLimit: DefaultOutputLimit,
	}
}

// Run executes the payload on the host with a hard timeout and bounded
// output. It prints a one-time warning so accidental use is visible.
func (e *localExecutor) Run(ctx context.Context, spec Spec) (Result, error) {
	if err := validateSpec(spec); err != nil {
		return Result{}, err
	}
	spec = e.defaults(spec)

	template, ok := runtimeTemplates[spec.Runtime]
	if !ok {
		return Result{}, fmt.Errorf("unknown sandbox runtime %q", spec.Runtime)
	}

	start := time.Now()
	runCtx, cancel := context.WithTimeout(ctx, spec.Timeout)
	defer cancel()

	command := exec.CommandContext(runCtx, template.Entrypoint[0], template.Entrypoint[1:]...)
	command.Stdin = strings.NewReader(spec.Code)

	var stdout, stderr limitedBuffer
	stdout.max = int(spec.OutputLimit)
	stderr.max = int(spec.OutputLimit)

	var writers sync.WaitGroup
	stdoutPipe, err := command.StdoutPipe()
	if err != nil {
		return Result{}, fmt.Errorf("local stdout pipe: %w", err)
	}
	stderrPipe, err := command.StderrPipe()
	if err != nil {
		return Result{}, fmt.Errorf("local stderr pipe: %w", err)
	}
	writers.Add(2)
	go func() {
		defer writers.Done()
		_, _ = io.Copy(&stdout, stdoutPipe)
	}()
	go func() {
		defer writers.Done()
		_, _ = io.Copy(&stderr, stderrPipe)
	}()

	runErr := command.Run()
	writers.Wait()

	result := Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: time.Since(start),
	}
	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		result.TimedOut = true
		return result, fmt.Errorf("%w after %s", ErrTimeout, spec.Timeout)
	}
	if runErr != nil {
		result.ExitCode = exitCodeOf(runErr)
		return result, fmt.Errorf("local run failed: %w", runErr)
	}
	return result, nil
}

// Close releases provider resources. The local executor holds none.
func (e *localExecutor) Close() error { return nil }

func (e *localExecutor) defaults(spec Spec) Spec {
	if spec.Timeout == 0 {
		spec.Timeout = e.timeout
	}
	if spec.OutputLimit == 0 {
		spec.OutputLimit = e.outputLimit
	}
	return spec
}
