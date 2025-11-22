package glm4

import (
	"github.com/rexleimo/agno-go/internal/agent"
	"github.com/rexleimo/agno-go/internal/model"
	"github.com/rexleimo/agno-go/pkg/providers/shared"
)

const defaultEndpoint = "https://open.bigmodel.cn/api/paas/v4"

func New(status model.ProviderStatus, endpoint, apiKey string) model.ChatProvider {
	return shared.NewOpenAICompat(agent.ProviderGLM4, shared.Config{
		Endpoint: endpointOrDefault(endpoint, defaultEndpoint),
		APIKey:   apiKey,
		Status:   status,
	})
}

func NewEmbed(status model.ProviderStatus, endpoint, apiKey string) model.EmbeddingProvider {
	return shared.NewOpenAICompat(agent.ProviderGLM4, shared.Config{
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
