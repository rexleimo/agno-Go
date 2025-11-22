package providers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rexleimo/agno-go/internal/agent"
	"github.com/rexleimo/agno-go/internal/model"
	runtimeconfig "github.com/rexleimo/agno-go/internal/runtime/config"
)

// Providers integration smoke with positive + negative branches; writes report for coverage artifacts.
func TestProvidersIntegrationReport(t *testing.T) {
	base := repoRoot(t)
	cfgPath := filepath.Join(base, "config", "default.yaml")
	envPath := filepath.Join(base, ".env")
	cfg, err := runtimeconfig.LoadWithEnv(cfgPath, envPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	statuses := cfg.ProviderStatuses()
	configs := cfg.ProviderConfigs()

	var buf bytes.Buffer
	targets := []agent.Provider{
		agent.ProviderOpenAI,
		agent.ProviderOpenRouter,
		agent.ProviderGroq,
		agent.ProviderGemini,
		agent.ProviderGLM4,
		agent.ProviderSiliconFlow,
		agent.ProviderModelScope,
		agent.ProviderOllama,
		agent.ProviderCerebras,
	}

	var available int
	for _, prov := range targets {
		st := findStatus(statuses, prov)
		if st.Status != model.ProviderAvailable {
			buf.WriteString(fmt.Sprintf("provider=%s status=skipped reason=%s missing=%v\n", prov, st.Status, st.MissingEnv))
			continue
		}
		modelID := providerModels[prov]
		if modelID == "" {
			buf.WriteString(fmt.Sprintf("provider=%s status=skipped reason=no-model\n", prov))
			continue
		}
		client, err := newProviderClient(prov, st, configs[prov].Endpoint, configs[prov].APIKey)
		if err != nil {
			buf.WriteString(fmt.Sprintf("provider=%s status=error err=%v\n", prov, err))
			continue
		}
		available++

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		resp, err := client.Chat(ctx, model.ChatRequest{
			Model: agent.ModelConfig{
				Provider: prov,
				ModelID:  modelID,
				Stream:   false,
			},
			Messages: []agent.Message{
				{Role: agent.RoleUser, Content: "hello from integration"},
			},
		})
		cancel()
		if err != nil {
			if isConnectivityError(err) {
				buf.WriteString(fmt.Sprintf("provider=%s status=skipped reason=unreachable err=%v\n", prov, err))
				continue
			}
			buf.WriteString(fmt.Sprintf("provider=%s status=error err=%v\n", prov, err))
		} else {
			tokens := resp.Usage.PromptTokens + resp.Usage.CompletionTokens
			buf.WriteString(fmt.Sprintf("provider=%s status=ok tokens=%d content_len=%d\n", prov, tokens, len(resp.Message.Content)))
		}

		cancelCtx, cancelFn := context.WithCancel(context.Background())
		cancelFn()
		_, err = client.Chat(cancelCtx, model.ChatRequest{
			Model: agent.ModelConfig{
				Provider: prov,
				ModelID:  modelID,
			},
			Messages: []agent.Message{
				{Role: agent.RoleUser, Content: "cancelled request"},
			},
		})
		if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			buf.WriteString(fmt.Sprintf("provider=%s negative=unexpected err=%v\n", prov, err))
		} else {
			buf.WriteString(fmt.Sprintf("provider=%s negative=ok err=%v\n", prov, err))
		}
	}

	if available == 0 {
		t.Log("no providers configured; report will only contain skipped entries")
	}

	logPath := filepath.Join(base, "specs", "001-go-agno-rewrite", "artifacts", "coverage", "providers.log")
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		t.Fatalf("mkdir providers log: %v", err)
	}
	if err := os.WriteFile(logPath, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write providers log: %v", err)
	}
}

func isConnectivityError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "dial tcp") || strings.Contains(msg, "connect:") || strings.Contains(msg, "connection refused") || strings.Contains(msg, "operation not permitted")
}
