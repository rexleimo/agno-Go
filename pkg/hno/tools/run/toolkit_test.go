package run

import (
	"context"
	"errors"
	"testing"
	"time"
)

type recordingExecutor struct {
	spec     Spec
	result   Result
	err      error
	closeErr error
}

func (e *recordingExecutor) Run(_ context.Context, spec Spec) (Result, error) {
	e.spec = spec
	return e.result, e.err
}

func (e *recordingExecutor) Close() error { return e.closeErr }

func TestCodeExecutionToolkitRunCode(t *testing.T) {
	executor := &recordingExecutor{result: Result{ExitCode: 0, Stdout: "ok"}}
	toolkit := NewToolkit(executor)

	result, err := toolkit.Execute(context.Background(), "run_code", map[string]interface{}{
		"runtime":         "python",
		"code":            "print('ok')",
		"timeout_seconds": 2.5,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := result.(map[string]interface{})
	if output["stdout"] != "ok" || output["exit_code"] != 0 {
		t.Fatalf("Execute() output = %#v", output)
	}
	if executor.spec.Runtime != "python" || executor.spec.Code != "print('ok')" || executor.spec.Timeout != 2500*time.Millisecond {
		t.Fatalf("executor spec = %#v", executor.spec)
	}
}

func TestCodeExecutionToolkitReturnsExecutionOutputOnFailure(t *testing.T) {
	executor := &recordingExecutor{
		result: Result{ExitCode: 4, Stdout: "partial", Stderr: "failed"},
		err:    errors.New("process exited with status 4"),
	}
	toolkit := NewToolkit(executor)

	result, err := toolkit.Execute(context.Background(), "run_code", map[string]interface{}{
		"runtime": "python",
		"code":    "raise SystemExit(4)",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v, want structured tool result", err)
	}
	output := result.(map[string]interface{})
	if output["exit_code"] != 4 || output["stdout"] != "partial" || output["error"] == "" {
		t.Fatalf("Execute() output = %#v", output)
	}
}

func TestCodeExecutionToolkitFailsClosedWithNilExecutor(t *testing.T) {
	toolkit := NewToolkit(nil)
	_, err := toolkit.Execute(context.Background(), "run_code", map[string]interface{}{
		"runtime": "python",
		"code":    "print(1)",
	})
	if err == nil {
		t.Fatal("Execute() error = nil, want configuration error")
	}
}

func TestCodeExecutionToolkitRejectsInvalidTimeout(t *testing.T) {
	toolkit := NewToolkit(&recordingExecutor{})
	_, err := toolkit.Execute(context.Background(), "run_code", map[string]interface{}{
		"runtime":         "python",
		"code":            "print(1)",
		"timeout_seconds": 0.0,
	})
	if err == nil {
		t.Fatal("Execute() error = nil, want timeout validation error")
	}
}

func TestCodeExecutionToolkitClose(t *testing.T) {
	want := errors.New("close failed")
	toolkit := NewToolkit(&recordingExecutor{closeErr: want})
	if err := toolkit.Close(); !errors.Is(err, want) {
		t.Fatalf("Close() error = %v, want %v", err, want)
	}
}
