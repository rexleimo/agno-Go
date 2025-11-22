package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/rexleimo/agno-go/internal/model"
)

func TestProviderEnvGating(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-test")
	t.Setenv("GEMINI_API_KEY", "sk-gemini")
	t.Setenv("GLM4_API_KEY", "sk-glm4")
	root := repoRoot(t)
	cfg, err := LoadWithEnv(filepath.Join(root, "config", "default.yaml"), "")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	statuses := cfg.ProviderStatuses()
	var openaiStatus, glm4Status model.Availability
	for _, st := range statuses {
		switch st.Provider {
		case "openai":
			openaiStatus = st.Status
		case "glm4":
			glm4Status = st.Status
		}
	}
	if openaiStatus != model.ProviderAvailable {
		t.Fatalf("expected openai available, got %s", openaiStatus)
	}
	if glm4Status != model.ProviderAvailable {
		t.Fatalf("expected glm4 available, got %s", glm4Status)
	}
}

func TestResolveEnvOverrides(t *testing.T) {
	t.Setenv("GOMEMLIMIT", "256MiB")
	t.Setenv("GOGC", "80")
	cfg := &Config{}
	cfg.applyRuntimeEnv()
	if os.Getenv("GOMEMLIMIT") != "256MiB" {
		t.Fatalf("GOMEMLIMIT not applied")
	}
	if os.Getenv("GOGC") != "80" {
		t.Fatalf("GOGC not applied")
	}
}

func repoRoot(tb testing.TB) string {
	tb.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		tb.Fatalf("cannot resolve caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", ".."))
}
