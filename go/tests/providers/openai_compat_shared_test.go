package providers

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/rexleimo/agno-go/internal/agent"
	"github.com/rexleimo/agno-go/internal/model"
	"github.com/rexleimo/agno-go/pkg/providers/shared"
)

func TestOpenAICompatHeadersAndStatus(t *testing.T) {
	var sawAuth bool
	client := shared.NewOpenAICompat(agent.ProviderOpenAI, shared.Config{
		Endpoint: "http://stub",
		APIKey:   "key",
		Status:   model.ProviderStatus{Provider: agent.ProviderOpenAI, Status: model.ProviderAvailable},
		ExtraHeaders: map[string]string{
			"X-Test": "yes",
		},
		HTTPClient: &http.Client{
			Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
				if req.Header.Get("Authorization") != "Bearer key" || req.Header.Get("X-Test") != "yes" {
					t.Fatalf("missing headers: %+v", req.Header)
				}
				sawAuth = true
				body := `{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(body)),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			}),
		},
	})
	_, err := client.Chat(context.Background(), model.ChatRequest{
		Model:    agent.ModelConfig{ModelID: "gpt-test"},
		Messages: []agent.Message{{Role: agent.RoleUser, Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("chat: %v", err)
	}
	if !sawAuth {
		t.Fatalf("auth header not observed")
	}
	if client.Status().Status != model.ProviderAvailable {
		t.Fatalf("unexpected status: %+v", client.Status())
	}
}

func TestOpenAICompatStreamError(t *testing.T) {
	client := shared.NewOpenAICompat(agent.ProviderOpenAI, shared.Config{
		Endpoint: "http://stub",
		APIKey:   "key",
		Status:   model.ProviderStatus{Provider: agent.ProviderOpenAI, Status: model.ProviderAvailable},
		HTTPClient: &http.Client{
			Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
				body := `{"error":{"message":"blocked"}}`
				return &http.Response{
					StatusCode: http.StatusTooManyRequests,
					Body:       io.NopCloser(strings.NewReader(body)),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			}),
		},
	})

	err := client.Stream(context.Background(), model.ChatRequest{
		Model:    agent.ModelConfig{ModelID: "gpt-test", Stream: true},
		Messages: []agent.Message{{Role: agent.RoleUser, Content: "hello"}},
	}, func(ev model.ChatStreamEvent) error { return nil })
	if err == nil {
		t.Fatalf("expected stream error on 429")
	}
}
