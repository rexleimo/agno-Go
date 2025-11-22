package cerebras

import (
	"github.com/rexleimo/agno-go/internal/agent"
	"github.com/rexleimo/agno-go/internal/model"
	"github.com/rexleimo/agno-go/pkg/providers/shared"
)

const defaultEndpoint = "https://api.cerebras.ai/v1"

func New(status model.ProviderStatus, endpoint, apiKey string) model.ChatProvider {
	return shared.NewOpenAICompat(agent.ProviderCerebras, shared.Config{
		Endpoint:   endpointOrDefault(endpoint, defaultEndpoint),
		APIKey:     apiKey,
		Status:     status,
		ChatPath:   "/chat/completions",
		EmbedPath:  "/embeddings",
		ParseUsage: true,
	})
}

func NewEmbed(status model.ProviderStatus, endpoint, apiKey string) model.EmbeddingProvider {
	return shared.NewOpenAICompat(agent.ProviderCerebras, shared.Config{
		Endpoint:   endpointOrDefault(endpoint, defaultEndpoint),
		APIKey:     apiKey,
		Status:     status,
		ParseUsage: true,
	})
}

func endpointOrDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
