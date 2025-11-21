package model

import (
	"context"
	"errors"
	"fmt"

	"github.com/rexleimo/agno-go/internal/agent"
)

// Capability captures provider abilities.
type Capability string

const (
	CapabilityChat      Capability = "chat"
	CapabilityEmbedding Capability = "embedding"
	CapabilityStreaming Capability = "streaming"
)

// Availability describes provider status derived from configuration.
type Availability string

const (
	ProviderAvailable     Availability = "available"
	ProviderNotConfigured Availability = "not-configured"
	ProviderDisabled      Availability = "disabled"
)

// ProviderStatus reports configuration and capability state for a provider.
type ProviderStatus struct {
	Provider     agent.Provider `json:"provider"`
	Status       Availability   `json:"status"`
	Capabilities []Capability   `json:"capabilities,omitempty"`
	MissingEnv   []string       `json:"missingEnv,omitempty"`
	Reason       string         `json:"reason,omitempty"`
}

// ChatRequest models a provider-agnostic chat request.
type ChatRequest struct {
	Model    agent.ModelConfig `json:"model"`
	Messages []agent.Message   `json:"messages"`
	Tools    []agent.ToolCall  `json:"tools,omitempty"`
	Metadata map[string]any    `json:"metadata,omitempty"`
	Stream   bool              `json:"stream,omitempty"`
}

// ChatResponse wraps a single assistant turn and usage data.
type ChatResponse struct {
	Message      agent.Message `json:"message"`
	Usage        agent.Usage   `json:"usage,omitempty"`
	FinishReason string        `json:"finishReason,omitempty"`
}

// ChatStreamEvent represents a streaming delta (text or tool call).
type ChatStreamEvent struct {
	Type         string          `json:"type"` // token|tool_call|end
	Delta        string          `json:"delta,omitempty"`
	ToolCall     *agent.ToolCall `json:"toolCall,omitempty"`
	Usage        agent.Usage     `json:"usage,omitempty"`
	FinishReason string          `json:"finishReason,omitempty"`
	Done         bool            `json:"done,omitempty"`
}

// StreamHandler consumes streaming chat events.
type StreamHandler func(event ChatStreamEvent) error

// EmbeddingRequest models an embedding call.
type EmbeddingRequest struct {
	Model agent.ModelConfig `json:"model"`
	Input []string          `json:"input"`
}

// EmbeddingResponse contains embedding vectors.
type EmbeddingResponse struct {
	Vectors [][]float64 `json:"vectors"`
	Usage   agent.Usage `json:"usage,omitempty"`
}

// ChatProvider defines chat completion capabilities.
type ChatProvider interface {
	Name() agent.Provider
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	Stream(ctx context.Context, req ChatRequest, fn StreamHandler) error
	Status() ProviderStatus
}

// EmbeddingProvider defines embedding capabilities.
type EmbeddingProvider interface {
	Name() agent.Provider
	Embed(ctx context.Context, req EmbeddingRequest) (*EmbeddingResponse, error)
	Status() ProviderStatus
}

// Router dispatches chat/embedding requests to registered providers.
type Router struct {
	chatProviders      map[agent.Provider]ChatProvider
	embeddingProviders map[agent.Provider]EmbeddingProvider
}

// Errors returned by the router when routing fails.
var (
	ErrProviderNotRegistered = errors.New("provider not registered")
	ErrCapabilityUnsupported = errors.New("capability unsupported")
	ErrProviderUnavailable   = errors.New("provider not available")
)

// NewRouter constructs an empty provider router.
func NewRouter() *Router {
	return &Router{
		chatProviders:      make(map[agent.Provider]ChatProvider),
		embeddingProviders: make(map[agent.Provider]EmbeddingProvider),
	}
}

// RegisterChatProvider adds or replaces a chat provider.
func (r *Router) RegisterChatProvider(p ChatProvider) {
	r.chatProviders[p.Name()] = p
}

// RegisterEmbeddingProvider adds or replaces an embedding provider.
func (r *Router) RegisterEmbeddingProvider(p EmbeddingProvider) {
	r.embeddingProviders[p.Name()] = p
}

// Chat routes a completion request to the configured provider.
func (r *Router) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	p, ok := r.chatProviders[req.Model.Provider]
	if !ok {
		return nil, ErrProviderNotRegistered
	}
	status := p.Status()
	if status.Status != ProviderAvailable {
		return nil, fmt.Errorf("%w: %s", ErrProviderUnavailable, status.Status)
	}
	if req.Stream {
		return nil, errors.New("stream requested without stream handler")
	}
	return p.Chat(ctx, req)
}

// Stream routes a streaming completion request.
func (r *Router) Stream(ctx context.Context, req ChatRequest, fn StreamHandler) error {
	p, ok := r.chatProviders[req.Model.Provider]
	if !ok {
		return ErrProviderNotRegistered
	}
	status := p.Status()
	if status.Status != ProviderAvailable {
		return fmt.Errorf("%w: %s", ErrProviderUnavailable, status.Status)
	}
	return p.Stream(ctx, req, fn)
}

// Embed routes an embedding request.
func (r *Router) Embed(ctx context.Context, req EmbeddingRequest) (*EmbeddingResponse, error) {
	p, ok := r.embeddingProviders[req.Model.Provider]
	if !ok {
		return nil, ErrProviderNotRegistered
	}
	status := p.Status()
	if status.Status != ProviderAvailable {
		return nil, fmt.Errorf("%w: %s", ErrProviderUnavailable, status.Status)
	}
	return p.Embed(ctx, req)
}

// Statuses returns provider statuses for health checks.
func (r *Router) Statuses() []ProviderStatus {
	result := make([]ProviderStatus, 0, len(r.chatProviders)+len(r.embeddingProviders))
	seen := make(map[agent.Provider]bool)

	for _, p := range r.chatProviders {
		st := p.Status()
		result = append(result, st)
		seen[p.Name()] = true
	}
	for _, p := range r.embeddingProviders {
		if seen[p.Name()] {
			continue
		}
		result = append(result, p.Status())
	}
	return result
}
