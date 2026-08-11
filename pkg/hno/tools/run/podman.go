package run

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// containerExecutor runs payloads in disposable OCI containers via podman or
// docker. Every Run creates a fresh container with fixed policy flags that
// callers cannot relax.
type containerExecutor struct {
	binPath     string
	image       string
	timeout     time.Duration
	memoryLimit int64
	pidLimit    int64
	outputLimit int64
}

func newContainerExecutor(binary, path string, config Config) *containerExecutor {
	return &containerExecutor{
		binPath:     path,
		image:       config.Image,
		timeout:     DefaultTimeout,
		memoryLimit: DefaultMemoryLimit,
		pidLimit:    DefaultPidLimit,
		outputLimit: DefaultOutputLimit,
	}
}

// Run executes one payload in a fresh, disposable container.
func (e *containerExecutor) Run(ctx context.Context, spec Spec) (Result, error) {
	if err := validateSpec(spec); err != nil {
		return Result{}, err
	}
	spec = e.defaults(spec)

	template, ok := runtimeTemplates[spec.Runtime]
	if !ok {
		return Result{}, fmt.Errorf("unknown sandbox runtime %q", spec.Runtime)
	}

	start := time.Now()
	args, cidfile, err := e.buildArgs(spec, template)
	if err != nil {
		return Result{}, err
	}

	runCtx, cancel := context.WithTimeout(ctx, spec.Timeout)
	defer cancel()

	command := exec.CommandContext(runCtx, e.binPath, args...)
	command.Stdin = strings.NewReader(spec.Code)

	var stdout, stderr limitedBuffer
	stdout.max = int(spec.OutputLimit)
	stderr.max = int(spec.OutputLimit)

	var stdoutErr, stderrErr error
	var writers sync.WaitGroup
	stdoutPipe, err := command.StdoutPipe()
	if err != nil {
		return Result{}, fmt.Errorf("sandbox stdout pipe: %w", err)
	}
	stderrPipe, err := command.StderrPipe()
	if err != nil {
		return Result{}, fmt.Errorf("sandbox stderr pipe: %w", err)
	}
	writers.Add(2)
	go func() {
		defer writers.Done()
		_, stdoutErr = io.Copy(&stdout, stdoutPipe)
	}()
	go func() {
		defer writers.Done()
		_, stderrErr = io.Copy(&stderr, stderrPipe)
	}()

	runErr := command.Run()
	writers.Wait()

	duration := time.Since(start)
	result := Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	if cid := readCidFile(cidfile); cid != "" {
		_ = e.forceRemove(cid)
	}
	_ = os.Remove(cidfile)

	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		result.TimedOut = true
		return result, fmt.Errorf("%w after %s", ErrTimeout, spec.Timeout)
	}
	if runErr != nil {
		result.ExitCode = exitCodeOf(runErr)
		if stdoutErr != nil || stderrErr != nil {
			return result, fmt.Errorf("sandbox run failed: %w (capture: %v)", runErr, errors.Join(stdoutErr, stderrErr))
		}
		return result, fmt.Errorf("sandbox run failed: %w", runErr)
	}
	result.ExitCode = 0
	if stdoutErr != nil || stderrErr != nil {
		return result, fmt.Errorf("sandbox output capture: %w", errors.Join(stdoutErr, stderrErr))
	}
	return result, nil
}

// Close releases provider resources. The container executor holds none.
func (e *containerExecutor) Close() error { return nil }

// buildArgs constructs the podman/docker invocation. Policy flags are fixed
// and never derived from caller input.
func (e *containerExecutor) buildArgs(spec Spec, template runtimeTemplate) ([]string, string, error) {
	dir, err := os.MkdirTemp("", "agno-run-*")
	if err != nil {
		return nil, "", fmt.Errorf("sandbox temp dir: %w", err)
	}
	cidfile := filepath.Join(dir, "cid")

	args := []string{"run", "--rm", "-i"}
	args = append(args,
		"--network", "none",
		"--cap-drop", "ALL",
		"--security-opt", "no-new-privileges",
		"--pids-limit", strconv.FormatInt(spec.PidLimit, 10),
		"--memory", strconv.FormatInt(spec.MemoryLimit, 10),
		"--memory-swap", strconv.FormatInt(spec.MemoryLimit, 10),
		"--read-only",
		"--tmpfs", "/tmp:rw,size=64m",
		"--cidfile", cidfile,
	)
	for _, entry := range spec.Env {
		args = append(args, "--env", entry)
	}
	if spec.Workspace != "" {
		absolute, err := filepath.Abs(spec.Workspace)
		if err != nil {
			return nil, "", fmt.Errorf("sandbox workspace: %w", err)
		}
		info, err := os.Stat(absolute)
		if err != nil {
			return nil, "", fmt.Errorf("sandbox workspace %q: %w", absolute, err)
		}
		if !info.IsDir() {
			return nil, "", fmt.Errorf("sandbox workspace %q is not a directory", absolute)
		}
		args = append(args, "--volume", absolute+":/workspace:rw")
	}
	args = append(args, e.image)
	args = append(args, template.Entrypoint...)
	return args, cidfile, nil
}

// forceRemove deletes a container by ID with its own short timeout so that a
// timed-out run can still be cleaned up.
func (e *containerExecutor) forceRemove(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	remove := exec.CommandContext(ctx, e.binPath, "rm", "-f", id)
	if err := remove.Run(); err != nil {
		return fmt.Errorf("remove sandbox container %q: %w", id, err)
	}
	return nil
}

func readCidFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func exitCodeOf(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

// limitedBuffer captures at most max bytes. Excess bytes are discarded
// silently: output flooding must fail closed without failing the run.
type limitedBuffer struct {
	mu        sync.Mutex
	buf       bytes.Buffer
	max       int
	truncated bool
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	total := len(p)
	remaining := b.max - b.buf.Len()
	if remaining > 0 {
		if len(p) > remaining {
			p = p[:remaining]
			b.truncated = true
		}
		_, _ = b.buf.Write(p)
	} else {
		b.truncated = true
	}
	// Report the full input length so io.Copy never sees a short write;
	// truncation is a policy, not an I/O error.
	return total, nil
}

func (b *limitedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.truncated {
		return b.buf.String()
	}
	return b.buf.String() + "\n...output truncated..."
}

func (b *limitedBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Len()
}
