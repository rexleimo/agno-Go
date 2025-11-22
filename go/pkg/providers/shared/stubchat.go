package shared

import (
	"context"
	"strings"

	"github.com/rexleimo/agno-go/internal/agent"
	"github.com/rexleimo/agno-go/internal/model"
)

// StdChat implements minimal chat/stream/embed behavior for providers not yet fully implemented.
type StdChat struct {
	NameVal     agent.Provider
	StatusVal   model.ProviderStatus
	Placeholder string
	EmbedVector []float64
}

func (s StdChat) Name() agent.Provider { return s.NameVal }

func (s StdChat) Status() model.ProviderStatus { return s.StatusVal }

func (s StdChat) Chat(ctx context.Context, req model.ChatRequest) (*model.ChatResponse, error) {
	if s.StatusVal.Status != model.ProviderAvailable {
		return nil, model.ErrProviderUnavailable
	}
	content := s.replyText(req.Messages)
	msg := agent.Message{Role: agent.RoleAssistant, Content: content}
	return &model.ChatResponse{Message: msg}, nil
}

func (s StdChat) Stream(ctx context.Context, req model.ChatRequest, fn model.StreamHandler) error {
	if s.StatusVal.Status != model.ProviderAvailable {
		return model.ErrProviderUnavailable
	}
	parts := strings.Fields(s.replyText(req.Messages))
	if len(parts) == 0 {
		parts = []string{s.Placeholder}
	}
	for _, token := range parts {
		if err := fn(model.ChatStreamEvent{Type: "token", Delta: token + " "}); err != nil {
			return err
		}
	}
	return fn(model.ChatStreamEvent{Type: "end", Done: true})
}

func (s StdChat) Embed(ctx context.Context, req model.EmbeddingRequest) (*model.EmbeddingResponse, error) {
	if s.StatusVal.Status != model.ProviderAvailable {
		return nil, model.ErrProviderUnavailable
	}
	vec := s.EmbedVector
	if len(vec) == 0 {
		vec = []float64{0.1, 0.2, 0.3}
	}
	vectors := make([][]float64, len(req.Input))
	for i := range req.Input {
		vectors[i] = vec
	}
	return &model.EmbeddingResponse{Vectors: vectors}, nil
}

func (s StdChat) replyText(msgs []agent.Message) string {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == agent.RoleUser {
			return s.Placeholder + ": " + msgs[i].Content
		}
	}
	return s.Placeholder
}
