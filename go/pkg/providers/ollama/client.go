package ollama

import (
	"github.com/rexleimo/agno-go/internal/agent"
	"github.com/rexleimo/agno-go/internal/model"
	"github.com/rexleimo/agno-go/pkg/providers/shared"
)

const defaultEndpoint = "http://localhost:11434/api"

func New(status model.ProviderStatus, endpoint, apiKey string) model.ChatProvider {
	return shared.NewOpenAICompat(agent.ProviderOllama, shared.Config{
		Endpoint:           endpointOrDefault(endpoint, defaultEndpoint),
		APIKey:             apiKey,
		Status:             status,
		ChatPath:           "/chat",
		EmbedPath:          "/embeddings",
		ParseUsage:         true,
		ParseToolArgs:      true,
		UsageExtractorChat: usageFromCompat,
		ErrorParser:        ollamaErrorParser,
	})
}

func NewEmbed(status model.ProviderStatus, endpoint, apiKey string) model.EmbeddingProvider {
	return shared.NewOpenAICompat(agent.ProviderOllama, shared.Config{
		Endpoint:            endpointOrDefault(endpoint, defaultEndpoint),
		APIKey:              apiKey,
		Status:              status,
		ChatPath:            "/chat",
		EmbedPath:           "/embeddings",
		ParseUsage:          true,
		ParseToolArgs:       true,
		UsageExtractorEmbed: usageFromCompatEmbed,
		ErrorParser:         ollamaErrorParser,
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

func ollamaErrorParser(body map[string]any) string {
	if msg, ok := body["error"].(string); ok && msg != "" {
		return msg
	}
	return ""
}
