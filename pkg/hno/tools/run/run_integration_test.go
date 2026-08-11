package run

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// Integration tests require a real container runtime and a prepared image.
// Enable with AGNO_RUN_INTEGRATION=1 and point AGNO_RUN_IMAGE at an image
// containing python3, bash, and a network probe binary (python3 alone is
// enough; the network test uses python sockets).
func integrationExecutor(t *testing.T) Executor {
	t.Helper()
	if os.Getenv("AGNO_RUN_INTEGRATION") == "" {
		t.Skip("set AGNO_RUN_INTEGRATION=1 to run sandbox integration tests")
	}

	var backend string
	for _, name := range []string{"podman", "docker"} {
		if _, err := exec.LookPath(name); err == nil {
			backend = name
			break
		}
	}
	if backend == "" {
		t.Skip("no container runtime (podman/docker) available")
	}

	image := os.Getenv("AGNO_RUN_IMAGE")
	if image == "" {
		t.Skip("set AGNO_RUN_IMAGE to the prepared sandbox image")
	}

	executor, err := NewExecutor(Config{Backend: backend, Image: image})
	if err != nil {
		t.Fatalf("NewExecutor(%s) error = %v", backend, err)
	}
	t.Cleanup(func() { _ = executor.Close() })
	return executor
}

func TestIntegrationBasicRun(t *testing.T) {
	executor := integrationExecutor(t)

	result, err := executor.Run(context.Background(), Spec{
		Runtime: "python",
		Code:    "print('sandboxed-ok')",
		Timeout: 30 * time.Second,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0 (stderr: %s)", result.ExitCode, result.Stderr)
	}
	if !strings.Contains(result.Stdout, "sandboxed-ok") {
		t.Fatalf("Stdout = %q, want sandboxed-ok", result.Stdout)
	}
}

func TestIntegrationExitCode(t *testing.T) {
	executor := integrationExecutor(t)

	result, err := executor.Run(context.Background(), Spec{
		Runtime: "python",
		Code:    "import sys; sys.exit(3)",
		Timeout: 30 * time.Second,
	})
	if err == nil {
		t.Fatal("Run() error = nil, want exit error")
	}
	if result.ExitCode != 3 {
		t.Fatalf("ExitCode = %d, want 3", result.ExitCode)
	}
}

func TestIntegrationTimeoutKillsRun(t *testing.T) {
	executor := integrationExecutor(t)

	result, err := executor.Run(context.Background(), Spec{
		Runtime: "shell",
		Code:    "sleep 300",
		Timeout: 2 * time.Second,
	})
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("Run() error = %v, want ErrTimeout", err)
	}
	if !result.TimedOut {
		t.Fatal("Result.TimedOut = false, want true")
	}
}

func TestIntegrationNoNetwork(t *testing.T) {
	executor := integrationExecutor(t)

	// A socket connect must fail inside the sandbox. The payload bounds its
	// own attempt so a misconfigured (networked) sandbox still fails fast.
	result, err := executor.Run(context.Background(), Spec{
		Runtime: "python",
		Code: `
import socket
s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
s.settimeout(3)
try:
    s.connect(("1.1.1.1", 80))
    print("NETWORK-REACHABLE")
except OSError:
    print("NETWORK-BLOCKED")
finally:
    s.close()
`,
		Timeout: 30 * time.Second,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if strings.Contains(result.Stdout, "NETWORK-REACHABLE") {
		t.Fatalf("sandbox reached the network; stdout = %q", result.Stdout)
	}
}

func TestIntegrationOutputTruncated(t *testing.T) {
	executor := integrationExecutor(t)

	result, err := executor.Run(context.Background(), Spec{
		Runtime: "python",
		Code:    "print('x' * 10000)",
		Timeout: 30 * time.Second,
		// tiny cap forces truncation
		OutputLimit: 64,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !strings.Contains(result.Stdout, "output truncated") {
		t.Fatalf("Stdout = %q, want truncation marker", result.Stdout)
	}
}

func TestIntegrationUnknownRuntimeFailsClosed(t *testing.T) {
	executor := integrationExecutor(t)

	_, err := executor.Run(context.Background(), Spec{
		Runtime: "ruby",
		Code:    "puts 1",
		Timeout: 5 * time.Second,
	})
	if err == nil {
		t.Fatal("Run() with unknown runtime error = nil, want error")
	}
}
