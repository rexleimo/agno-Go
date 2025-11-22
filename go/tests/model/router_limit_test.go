package model_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rexleimo/agno-go/internal/agent"
	"github.com/rexleimo/agno-go/internal/model"
)

func TestRouterRetriesAfterTimeout(t *testing.T) {
	router := model.NewRouter(
		model.WithTimeout(5*time.Millisecond),
		model.WithRetries(1, 2*time.Millisecond),
	)
	provider := &flakyChatProvider{}
	router.RegisterChatProvider(provider)

	req := model.ChatRequest{
		Model: agent.ModelConfig{
			Provider: agent.ProviderOpenAI,
			ModelID:  "retry-model",
		},
	}
	resp, err := router.Chat(context.Background(), req)
	if err != nil {
		t.Fatalf("chat failed: %v", err)
	}
	if resp.Message.Content != "ok" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if calls := atomic.LoadInt32(&provider.calls); calls != 2 {
		t.Fatalf("expected 2 attempts, got %d", calls)
	}
}

func TestRouterRejectsWhenConcurrencyExceeded(t *testing.T) {
	router := model.NewRouter(model.WithMaxConcurrency(1), model.WithTimeout(time.Second))
	blocker := &blockingChatProvider{
		entered: make(chan struct{}),
		release: make(chan struct{}),
	}
	router.RegisterChatProvider(blocker)

	req := model.ChatRequest{Model: agent.ModelConfig{Provider: agent.ProviderOpenAI, ModelID: "blocker"}}
	errCh := make(chan error, 1)
	go func() {
		_, err := router.Chat(context.Background(), req)
		errCh <- err
	}()

	select {
	case <-blocker.entered:
	case <-time.After(time.Second):
		t.Fatalf("first call did not start")
	}

	if _, err := router.Chat(context.Background(), req); !errors.Is(err, model.ErrProviderUnavailable) {
		t.Fatalf("expected provider unavailable due to concurrency, got %v", err)
	}

	close(blocker.release)
	if err := <-errCh; err != nil {
		t.Fatalf("first call failed: %v", err)
	}
}

type flakyChatProvider struct {
	calls int32
}

func (p *flakyChatProvider) Name() agent.Provider { return agent.ProviderOpenAI }

func (p *flakyChatProvider) Status() model.ProviderStatus {
	return model.ProviderStatus{
		Provider:     agent.ProviderOpenAI,
		Status:       model.ProviderAvailable,
		Capabilities: []model.Capability{model.CapabilityChat},
	}
}

func (p *flakyChatProvider) Chat(ctx context.Context, req model.ChatRequest) (*model.ChatResponse, error) {
	if atomic.AddInt32(&p.calls, 1) == 1 {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return &model.ChatResponse{
		Message: agent.Message{
			Role:    agent.RoleAssistant,
			Content: "ok",
		},
	}, nil
}

func (p *flakyChatProvider) Stream(ctx context.Context, req model.ChatRequest, fn model.StreamHandler) error {
	return errors.New("stream not implemented")
}

type blockingChatProvider struct {
	entered chan struct{}
	release chan struct{}
}

func (p *blockingChatProvider) Name() agent.Provider { return agent.ProviderOpenAI }

func (p *blockingChatProvider) Status() model.ProviderStatus {
	return model.ProviderStatus{
		Provider:     agent.ProviderOpenAI,
		Status:       model.ProviderAvailable,
		Capabilities: []model.Capability{model.CapabilityChat},
	}
}

func (p *blockingChatProvider) Chat(ctx context.Context, req model.ChatRequest) (*model.ChatResponse, error) {
	select {
	case <-p.entered:
	default:
		close(p.entered)
	}
	select {
	case <-p.release:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return &model.ChatResponse{
		Message: agent.Message{
			Role:    agent.RoleAssistant,
			Content: "ok",
		},
	}, nil
}

func (p *blockingChatProvider) Stream(ctx context.Context, req model.ChatRequest, fn model.StreamHandler) error {
	return errors.New("stream not implemented")
}
