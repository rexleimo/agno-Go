package modelscope

import (
	"github.com/rexleimo/agno-go/internal/agent"
	"github.com/rexleimo/agno-go/internal/model"
	"github.com/rexleimo/agno-go/pkg/providers/shared"
)

const defaultEndpoint = "https://api.modelscope.cn/v1"

func New(status model.ProviderStatus, endpoint, apiKey string) model.ChatProvider {
	return shared.NewOpenAICompat(agent.ProviderModelScope, shared.Config{
		Endpoint:           endpointOrDefault(endpoint, defaultEndpoint),
		APIKey:             apiKey,
		Status:             status,
		ParseUsage:         true,
		UsageExtractorChat: usageFromCompat,
	})
}

func NewEmbed(status model.ProviderStatus, endpoint, apiKey string) model.EmbeddingProvider {
	return shared.NewOpenAICompat(agent.ProviderModelScope, shared.Config{
		Endpoint:            endpointOrDefault(endpoint, defaultEndpoint),
		APIKey:              apiKey,
		Status:              status,
		ParseUsage:          true,
		UsageExtractorEmbed: usageFromCompatEmbed,
	})
}

func endpointOrDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func usageFromCompat(resp shared.OACompatChatResp) agent.Usage {
	return agent.Usage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		LatencyMs:        0,
	}
}

func usageFromCompatEmbed(resp shared.OACompatEmbedResp) agent.Usage {
	return agent.Usage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		LatencyMs:        0,
	}
}
