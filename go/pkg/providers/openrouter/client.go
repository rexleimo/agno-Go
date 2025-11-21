package openrouter

import (
	"github.com/rexleimo/agno-go/internal/agent"
	"github.com/rexleimo/agno-go/internal/model"
	"github.com/rexleimo/agno-go/pkg/providers/shared"
)

const defaultEndpoint = "https://openrouter.ai/api/v1"

func New(status model.ProviderStatus, endpoint, apiKey string, headers map[string]string) model.ChatProvider {
	return shared.NewOpenAICompat(agent.ProviderOpenRouter, shared.Config{
		Endpoint:            endpointOrDefault(endpoint, defaultEndpoint),
		APIKey:              apiKey,
		Status:              status,
		ExtraHeaders:        headers,
		ParseUsage:          true,
		ErrorParser:         openRouterErrorParser,
		UsageExtractorChat:  usageFromOpenRouter,
		UsageExtractorEmbed: usageFromOpenRouterEmbed,
	})
}

func NewEmbed(status model.ProviderStatus, endpoint, apiKey string, headers map[string]string) model.EmbeddingProvider {
	return shared.NewOpenAICompat(agent.ProviderOpenRouter, shared.Config{
		Endpoint:            endpointOrDefault(endpoint, defaultEndpoint),
		APIKey:              apiKey,
		Status:              status,
		ExtraHeaders:        headers,
		ParseUsage:          true,
		ErrorParser:         openRouterErrorParser,
		UsageExtractorChat:  usageFromOpenRouter,
		UsageExtractorEmbed: usageFromOpenRouterEmbed,
	})
}

func endpointOrDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func openRouterErrorParser(body map[string]any) string {
	if errObj, ok := body["error"].(map[string]any); ok {
		if msg, ok := errObj["message"].(string); ok && msg != "" {
			return msg
		}
		if desc, ok := errObj["description"].(string); ok && desc != "" {
			return desc
		}
	}
	return ""
}

func usageFromOpenRouter(resp shared.OACompatChatResp) agent.Usage {
	return agent.Usage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		LatencyMs:        0,
	}
}

func usageFromOpenRouterEmbed(resp shared.OACompatEmbedResp) agent.Usage {
	return agent.Usage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		LatencyMs:        0,
	}
}
