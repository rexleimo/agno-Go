package siliconflow

import (
	"github.com/rexleimo/agno-go/internal/agent"
	"github.com/rexleimo/agno-go/internal/model"
	"github.com/rexleimo/agno-go/pkg/providers/shared"
)

const defaultEndpoint = "https://api.siliconflow.cn/v1"

func New(status model.ProviderStatus, endpoint, apiKey string) model.ChatProvider {
	return shared.NewOpenAICompat(agent.ProviderSiliconFlow, shared.Config{
		Endpoint: endpointOrDefault(endpoint, defaultEndpoint),
		APIKey:   apiKey,
		Status:   status,
		// SiliconFlow follows OpenAI paths; no extra headers needed.
	})
}

func NewEmbed(status model.ProviderStatus, endpoint, apiKey string) model.EmbeddingProvider {
	return shared.NewOpenAICompat(agent.ProviderSiliconFlow, shared.Config{
		Endpoint: endpointOrDefault(endpoint, defaultEndpoint),
		APIKey:   apiKey,
		Status:   status,
	})
}

func endpointOrDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
