package run

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestValidateSpec(t *testing.T) {
	valid := Spec{Runtime: "python", Code: "print(1)"}

	tests := []struct {
		name    string
		spec    Spec
		wantErr string
	}{
		{name: "valid python", spec: valid},
		{name: "valid node", spec: Spec{Runtime: "node", Code: "console.log(1)"}},
		{name: "valid shell", spec: Spec{Runtime: "shell", Code: "echo hi"}},
		{name: "empty runtime", spec: Spec{Code: "x"}, wantErr: "unknown sandbox runtime"},
		{name: "unknown runtime", spec: Spec{Runtime: "ruby", Code: "x"}, wantErr: "unknown sandbox runtime"},
		{name: "empty code", spec: Spec{Runtime: "python"}, wantErr: "non-empty code"},
		{name: "blank code", spec: Spec{Runtime: "python", Code: "  \n "}, wantErr: "non-empty code"},
		{name: "negative timeout", spec: Spec{Runtime: "python", Code: "x", Timeout: -time.Second}, wantErr: "timeout"},
		{name: "negative memory", spec: Spec{Runtime: "python", Code: "x", MemoryLimit: -1}, wantErr: "memory limit"},
		{name: "negative pids", spec: Spec{Runtime: "python", Code: "x", PidLimit: -1}, wantErr: "pid limit"},
		{name: "negative output", spec: Spec{Runtime: "python", Code: "x", OutputLimit: -1}, wantErr: "output limit"},
		{name: "unsupported network", spec: Spec{Runtime: "python", Code: "x", Network: NetworkPolicy(7)}, wantErr: "network policy"},
		{name: "bad env no equals", spec: Spec{Runtime: "python", Code: "x", Env: []string{"SECRET"}}, wantErr: "KEY=VALUE"},
		{name: "bad env leading equals", spec: Spec{Runtime: "python", Code: "x", Env: []string{"=v"}}, wantErr: "KEY=VALUE"},
		{name: "good env", spec: Spec{Runtime: "python", Code: "x", Env: []string{"A=1", "B=two"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSpec(tt.spec)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("validateSpec() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("validateSpec() error = nil, want %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("validateSpec() error = %q, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestRuntimeTemplates(t *testing.T) {
	tests := []struct {
		runtime  string
		firstArg string
	}{
		{runtime: "python", firstArg: "python3"},
		{runtime: "node", firstArg: "node"},
		{runtime: "shell", firstArg: "bash"},
	}
	for _, tt := range tests {
		t.Run(tt.runtime, func(t *testing.T) {
			template, ok := runtimeTemplates[tt.runtime]
			if !ok {
				t.Fatalf("runtime %q missing from templates", tt.runtime)
			}
			if len(template.Entrypoint) == 0 || template.Entrypoint[0] != tt.firstArg {
				t.Fatalf("runtime %q entrypoint = %v, want first arg %q", tt.runtime, template.Entrypoint, tt.firstArg)
			}
		})
	}
}

func TestNewExecutorProbeOrder(t *testing.T) {
	tests := []struct {
		name      string
		binaries  map[string]string
		backend   string
		wantBin   string
		wantErrIs error
	}{
		{
			name:     "podman preferred",
			binaries: map[string]string{"podman": "/usr/bin/podman", "docker": "/usr/bin/docker"},
			wantBin:  "/usr/bin/podman",
		},
		{
			name:     "docker fallback",
			binaries: map[string]string{"docker": "/usr/bin/docker"},
			wantBin:  "/usr/bin/docker",
		},
		{
			name:      "no provider fails closed",
			binaries:  map[string]string{},
			wantErrIs: ErrNoProvider,
		},
		{
			name:      "forced backend missing",
			binaries:  map[string]string{"docker": "/usr/bin/docker"},
			backend:   "podman",
			wantErrIs: ErrNoProvider,
		},
		{
			name:      "unsupported backend",
			binaries:  map[string]string{"podman": "/usr/bin/podman"},
			backend:   "bwrap",
			wantErrIs: ErrUnsupportedBackend,
		},
		{
			name:     "forced podman",
			binaries: map[string]string{"podman": "/usr/bin/podman", "docker": "/usr/bin/docker"},
			backend:  "podman",
			wantBin:  "/usr/bin/podman",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewExecutor(Config{Binaries: tt.binaries, Backend: tt.backend})
			if tt.wantErrIs != nil {
				if err == nil {
					t.Fatalf("NewExecutor() error = nil, want %v", tt.wantErrIs)
				}
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("NewExecutor() error = %v, want %v", err, tt.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewExecutor() error = %v, want nil", err)
			}
			container, ok := executor.(*containerExecutor)
			if !ok {
				t.Fatalf("NewExecutor() type = %T, want *containerExecutor", executor)
			}
			if container.binPath != tt.wantBin {
				t.Fatalf("binPath = %q, want %q", container.binPath, tt.wantBin)
			}
			if container.image != "agent-runner:1" {
				t.Fatalf("image = %q, want default agent-runner:1", container.image)
			}
		})
	}
}

func TestBuildArgsPolicyIsFixed(t *testing.T) {
	executor := &containerExecutor{binPath: "/usr/bin/podman", image: "agent-runner:1"}
	spec := Spec{
		Runtime:     "python",
		Code:        "print(1)",
		MemoryLimit: 256 << 20,
		PidLimit:    64,
		Env:         []string{"A=1"},
	}
	args, cidfile, err := executor.buildArgs(spec, runtimeTemplates["python"])
	if err != nil {
		t.Fatalf("buildArgs() error = %v", err)
	}
	if cidfile == "" {
		t.Fatal("buildArgs() cidfile is empty")
	}

	joined := strings.Join(args, " ")
	required := []string{
		"run", "--rm", "-i",
		"--network", "none",
		"--cap-drop", "ALL",
		"--security-opt", "no-new-privileges",
		"--pids-limit", "64",
		"--memory", "268435456",
		"--memory-swap", "268435456",
		"--read-only",
		"--tmpfs",
		"--env", "A=1",
		"agent-runner:1",
		"python3", "-",
	}
	for _, fragment := range required {
		if !strings.Contains(joined, fragment) {
			t.Fatalf("buildArgs() missing %q in %q", fragment, joined)
		}
	}
	forbidden := []string{"--privileged", "--device", "--network host", "--network=host"}
	for _, fragment := range forbidden {
		if strings.Contains(joined, fragment) {
			t.Fatalf("buildArgs() must not contain %q in %q", fragment, joined)
		}
	}
}

func TestBuildArgsRejectsMissingWorkspace(t *testing.T) {
	executor := &containerExecutor{binPath: "/usr/bin/podman", image: "agent-runner:1"}
	_, _, err := executor.buildArgs(Spec{Runtime: "python", Code: "x", Workspace: "C:\\definitely\\missing\\workspace"}, runtimeTemplates["python"])
	if err == nil {
		t.Fatal("buildArgs() with missing workspace error = nil, want error")
	}
	if !strings.Contains(err.Error(), "workspace") {
		t.Fatalf("buildArgs() error = %q, want workspace mention", err)
	}
}

func TestLimitedBufferTruncates(t *testing.T) {
	buffer := &limitedBuffer{max: 8}
	written, err := buffer.Write([]byte("0123456789abcdef"))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if written != 16 {
		t.Fatalf("Write() n = %d, want 16 (excess must be swallowed)", written)
	}
	if buffer.String() != "01234567\n...output truncated..." {
		t.Fatalf("String() = %q, want truncated marker", buffer.String())
	}

	small := &limitedBuffer{max: 32}
	if _, err := small.Write([]byte("ok")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if small.String() != "ok" {
		t.Fatalf("String() = %q, want %q", small.String(), "ok")
	}
}

func TestExitCodeOf(t *testing.T) {
	if exitCodeOf(nil) != -1 {
		t.Fatal("exitCodeOf(nil) = non -1")
	}
	if exitCodeOf(errors.New("plain")) != -1 {
		t.Fatal("exitCodeOf(plain error) = non -1")
	}
}
