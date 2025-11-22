package groq

import (
	"github.com/rexleimo/agno-go/internal/agent"
	"github.com/rexleimo/agno-go/internal/model"
	"github.com/rexleimo/agno-go/pkg/providers/shared"
)

const defaultEndpoint = "https://api.groq.com/openai/v1"

func New(status model.ProviderStatus, endpoint, apiKey string) model.ChatProvider {
	return shared.NewOpenAICompat(agent.ProviderGroq, shared.Config{
		Endpoint:            endpointOrDefault(endpoint, defaultEndpoint),
		APIKey:              apiKey,
		Status:              status,
		EmbedPath:           "/embeddings",
		ParseUsage:          true,
		ParseToolArgs:       true,
		UsageExtractorChat:  usageFromCompat,
		UsageExtractorEmbed: usageFromCompatEmbed,
		ErrorParser:         groqErrorParser,
	})
}

func NewEmbed(status model.ProviderStatus, endpoint, apiKey string) model.EmbeddingProvider {
	return shared.NewOpenAICompat(agent.ProviderGroq, shared.Config{
		Endpoint:            endpointOrDefault(endpoint, defaultEndpoint),
		APIKey:              apiKey,
		Status:              status,
		EmbedPath:           "/embeddings",
		ParseUsage:          true,
		ParseToolArgs:       true,
		UsageExtractorChat:  usageFromCompat,
		UsageExtractorEmbed: usageFromCompatEmbed,
		ErrorParser:         groqErrorParser,
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

func groqErrorParser(body map[string]any) string {
	if msg, ok := body["message"].(string); ok && msg != "" {
		return msg
	}
	if errObj, ok := body["error"].(map[string]any); ok {
		if msg, ok := errObj["message"].(string); ok && msg != "" {
			return msg
		}
	}
	return ""
}
