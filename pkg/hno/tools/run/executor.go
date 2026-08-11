// Package run executes untrusted code and shell commands inside disposable
// sandboxes. It is language-agnostic: a runtime template maps a language name
// to an entrypoint inside a prebuilt image, and the sandbox core never parses
// or interprets the payload.
package run

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Default limits applied when a Spec field is left at its zero value.
const (
	DefaultTimeout     = 30 * time.Second
	DefaultMemoryLimit = 512 << 20 // 512 MiB
	DefaultPidLimit    = 128
	DefaultOutputLimit = 1 << 20 // 1 MiB
)

// NetworkPolicy controls container network access. Only NetworkNone is
// implemented in the initial version; the field exists so future policies
// do not change the Spec shape.
type NetworkPolicy int

const (
	// NetworkNone denies all network access inside the sandbox.
	NetworkNone NetworkPolicy = iota
)

// Runtime templates. The entrypoint receives the payload on stdin, so no
// payload bytes ever appear in the process argument list.
var runtimeTemplates = map[string]runtimeTemplate{
	"python": {Entrypoint: []string{"python3", "-"}},
	"node":   {Entrypoint: []string{"node", "-"}},
	"shell":  {Entrypoint: []string{"bash", "-s"}},
}

// runtimeTemplate describes how one language is launched inside the sandbox
// image.
type runtimeTemplate struct {
	Entrypoint []string
}

// Spec describes one sandboxed execution. Zero-value fields fall back to
// executor defaults.
type Spec struct {
	// Runtime selects the template: python, node, or shell.
	Runtime string
	// Code is the payload sent to the entrypoint on stdin.
	Code string
	// Timeout bounds the whole run. Zero uses DefaultTimeout.
	Timeout time.Duration
	// MemoryLimit bounds the sandbox memory in bytes. Zero uses
	// DefaultMemoryLimit.
	MemoryLimit int64
	// PidLimit bounds processes inside the sandbox. Zero uses
	// DefaultPidLimit.
	PidLimit int64
	// Network policy. Zero is NetworkNone.
	Network NetworkPolicy
	// Workspace is a host directory mounted read-write at /workspace.
	// Empty means no workspace is mounted.
	Workspace string
	// Env is an explicit allowlist of KEY=VALUE pairs passed into the
	// sandbox. Secrets must never be listed here.
	Env []string
	// OutputLimit caps captured stdout/stderr in bytes. Zero uses
	// DefaultOutputLimit.
	OutputLimit int64
}

// Result reports a completed sandboxed run.
type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
	// TimedOut is true when the run was killed for exceeding Spec.Timeout.
	TimedOut bool
}

// Executor runs untrusted payloads inside disposable sandboxes. Implementations
// must be safe to call from concurrent goroutines.
type Executor interface {
	Run(ctx context.Context, spec Spec) (Result, error)
	Close() error
}

// Config configures provider selection for NewExecutor.
type Config struct {
	// Backend forces a provider: "podman", "docker", or "auto" (default).
	Backend string
	// Image overrides the sandbox image. Defaults to "agent-runner:1".
	Image string
	// Binaries is an optional probe override used by tests.
	Binaries map[string]string
}

// ErrNoProvider is returned when no sandbox runtime binary is available.
// Callers must fail closed and never fall back to a bare exec.Command.
var ErrNoProvider = errors.New("no sandbox provider available: install podman or docker, or construct a local executor explicitly for development")

// ErrUnsupportedBackend is returned when Config.Backend names a provider this
// build does not implement.
var ErrUnsupportedBackend = errors.New("unsupported sandbox backend")

// ErrTimeout is returned when a run exceeds its Spec.Timeout.
var ErrTimeout = errors.New("sandbox run timed out")

// NewExecutor selects a container provider by probing for podman, then
// docker. It never silently degrades: when no provider is found it returns
// ErrNoProvider, and callers must treat that as an execution denial.
func NewExecutor(config Config) (Executor, error) {
	if config.Image == "" {
		config.Image = "agent-runner:1"
	}
	lookPath := exec.LookPath
	if config.Binaries != nil {
		lookPath = func(name string) (string, error) {
			if path, ok := config.Binaries[name]; ok && path != "" {
				return path, nil
			}
			return "", exec.ErrNotFound
		}
	}

	backends := []string{"podman", "docker"}
	if config.Backend != "" && config.Backend != "auto" {
		found := false
		for _, name := range backends {
			if config.Backend == name {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("%w %q", ErrUnsupportedBackend, config.Backend)
		}
		backends = []string{config.Backend}
	}

	for _, name := range backends {
		path, err := lookPath(name)
		if err != nil {
			continue
		}
		return newContainerExecutor(name, path, config), nil
	}
	return nil, ErrNoProvider
}

// validateSpec rejects specs that must never reach a provider. Unknown
// runtimes and missing payloads are configuration errors, not runtime
// questions.
func validateSpec(spec Spec) error {
	if _, ok := runtimeTemplates[spec.Runtime]; !ok {
		return fmt.Errorf("unknown sandbox runtime %q (supported: python, node, shell)", spec.Runtime)
	}
	if strings.TrimSpace(spec.Code) == "" {
		return fmt.Errorf("sandbox spec requires non-empty code")
	}
	if spec.Timeout < 0 {
		return fmt.Errorf("sandbox timeout must not be negative")
	}
	if spec.MemoryLimit < 0 {
		return fmt.Errorf("sandbox memory limit must not be negative")
	}
	if spec.PidLimit < 0 {
		return fmt.Errorf("sandbox pid limit must not be negative")
	}
	if spec.OutputLimit < 0 {
		return fmt.Errorf("sandbox output limit must not be negative")
	}
	if spec.Network != NetworkNone {
		return fmt.Errorf("unsupported sandbox network policy %d", spec.Network)
	}
	for _, entry := range spec.Env {
		if !strings.Contains(entry, "=") || strings.HasPrefix(entry, "=") {
			return fmt.Errorf("sandbox env entry %q must be KEY=VALUE", entry)
		}
	}
	return nil
}

// defaults applies executor defaults to zero-valued spec fields.
func (e *containerExecutor) defaults(spec Spec) Spec {
	if spec.Timeout == 0 {
		spec.Timeout = e.timeout
	}
	if spec.MemoryLimit == 0 {
		spec.MemoryLimit = e.memoryLimit
	}
	if spec.PidLimit == 0 {
		spec.PidLimit = e.pidLimit
	}
	if spec.OutputLimit == 0 {
		spec.OutputLimit = e.outputLimit
	}
	return spec
}
