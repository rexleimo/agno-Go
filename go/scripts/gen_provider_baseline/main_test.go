package main

import (
	"os"
	"testing"

	"github.com/rexleimo/agno-go/internal/agent"
	"github.com/rexleimo/agno-go/internal/model"
	runtimeconfig "github.com/rexleimo/agno-go/internal/runtime/config"
)

func TestChosenModelPrefersEnv(t *testing.T) {
	envKey := "OPENAI_MODEL"
	t.Setenv(envKey, "env-model")
	if got := chosenModel(agent.ProviderOpenAI, false); got != "env-model" {
		t.Fatalf("expected env override, got %s", got)
	}
}

func TestChosenModelFallback(t *testing.T) {
	os.Unsetenv("GROQ_MODEL")
	if got := chosenModel(agent.ProviderGroq, false); got != chatModels[agent.ProviderGroq] {
		t.Fatalf("expected fallback, got %s", got)
	}
}

func TestBuildClientsUnknown(t *testing.T) {
	chat, embed := buildClients(agent.Provider("unknown"), providerStatus(agent.Provider("unknown")), runtimeconfig.ProviderConfig{})
	if chat != nil || embed != nil {
		t.Fatalf("expected nil clients for unknown provider")
	}
}

func providerStatus(p agent.Provider) model.ProviderStatus {
	return model.ProviderStatus{Provider: p, Status: model.ProviderNotConfigured}
}
