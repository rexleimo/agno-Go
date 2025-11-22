package providers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rexleimo/agno-go/internal/model"
	"github.com/rexleimo/agno-go/internal/runtime/config"
)

func TestProvidersMissingEnvAreNotConfigured(t *testing.T) {
	// Ensure env keys are cleared for this test to assert gating logic.
	keys := []string{
		"OPENAI_API_KEY", "GEMINI_API_KEY", "GLM4_API_KEY", "OPENROUTER_API_KEY",
		"SILICONFLOW_API_KEY", "CEREBRAS_API_KEY", "MODELSCOPE_API_KEY", "GROQ_API_KEY",
	}
	restore := unsetEnv(keys)
	defer restore()

	cfgPath := filepath.Join("..", "..", "..", "config", "default.yaml")
	cfg, err := config.LoadWithEnv(cfgPath, "")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	statuses := cfg.ProviderStatuses()
	if len(statuses) == 0 {
		t.Fatalf("expected provider statuses, got none")
	}

	for _, st := range statuses {
		if len(st.MissingEnv) == 0 {
			// Providers without required env (e.g., Ollama) may remain available.
			continue
		}
		if st.Status == model.ProviderAvailable {
			t.Fatalf("expected %s to be gated as not-configured when keys are missing", st.Provider)
		}
		if len(st.MissingEnv) == 0 {
			t.Fatalf("expected missing env reasons for %s", st.Provider)
		}
	}
}

func unsetEnv(keys []string) func() {
	originals := make(map[string]string, len(keys))
	for _, k := range keys {
		originals[k] = os.Getenv(k)
		_ = os.Unsetenv(k)
	}
	return func() {
		for k, v := range originals {
			_ = os.Setenv(k, v)
		}
	}
}
