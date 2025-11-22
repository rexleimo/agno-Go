package runtime_test

import (
	"context"
	"testing"

	"github.com/rexleimo/agno-go/internal/agent"
	"github.com/rexleimo/agno-go/internal/model"
	"github.com/rexleimo/agno-go/internal/runtime"
	"github.com/rexleimo/agno-go/pkg/memory"
	"github.com/rexleimo/agno-go/pkg/providers/stub"
)

func TestToggleToolRegistersAndDisables(t *testing.T) {
	router := model.NewRouter()
	router.RegisterChatProvider(stub.New(agent.ProviderOpenAI, model.ProviderAvailable, nil))
	svc := runtime.NewService(memory.NewInMemoryStore(), router)

	agentID, err := svc.CreateAgent(context.Background(), agent.Agent{
		Name: "toggle-agent",
		Model: agent.ModelConfig{
			Provider: agent.ProviderOpenAI,
			ModelID:  "m1",
			Stream:   true,
		},
	})
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}

	cfg, err := svc.ToggleTool(context.Background(), agentID, "time", true)
	if err != nil {
		t.Fatalf("enable tool: %v", err)
	}
	if !contains(cfg.Enabled, "time") || !contains(cfg.Registered, "time") {
		t.Fatalf("tool not enabled/registered: %+v", cfg)
	}

	cfg, err = svc.ToggleTool(context.Background(), agentID, "time", false)
	if err != nil {
		t.Fatalf("disable tool: %v", err)
	}
	if contains(cfg.Enabled, "time") {
		t.Fatalf("tool still enabled after disable")
	}
}

func contains(list []string, v string) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}
	return false
}
