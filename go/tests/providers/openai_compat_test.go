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

func TestOpenAICompatWithCustomTransport(t *testing.T) {
	client := shared.NewOpenAICompat(agent.ProviderOpenAI, shared.Config{
		Endpoint: "http://stub",
		APIKey:   "key",
		Status:   model.ProviderStatus{Provider: agent.ProviderOpenAI, Status: model.ProviderAvailable},
		HTTPClient: &http.Client{
			Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
				body := `{"choices":[{"message":{"role":"assistant","content":"ok"}}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioNopCloser(body),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			}),
		},
	})

	resp, err := client.Chat(context.Background(), model.ChatRequest{
		Model:    agent.ModelConfig{ModelID: "model"},
		Messages: []agent.Message{{Role: agent.RoleUser, Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("chat: %v", err)
	}
	if resp.Message.Content != "ok" {
		t.Fatalf("unexpected message: %+v", resp.Message)
	}

	errClient := shared.NewOpenAICompat(agent.ProviderOpenAI, shared.Config{
		Endpoint: "http://stub",
		APIKey:   "key",
		Status:   model.ProviderStatus{Provider: agent.ProviderOpenAI, Status: model.ProviderAvailable},
		HTTPClient: &http.Client{
			Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
				body := `{"error":{"message":"blocked","code":"429"}}`
				return &http.Response{
					StatusCode: http.StatusTooManyRequests,
					Body:       ioNopCloser(body),
					Header:     http.Header{"Content-Type": []string{"application/json"}},
				}, nil
			}),
		},
	})
	if err := errClient.Stream(context.Background(), model.ChatRequest{
		Model:    agent.ModelConfig{ModelID: "model", Stream: true},
		Messages: []agent.Message{{Role: agent.RoleUser, Content: "hi"}},
	}, func(ev model.ChatStreamEvent) error { return nil }); err == nil {
		t.Fatalf("expected stream error")
	}
}

type closer struct{ *strings.Reader }

func (c closer) Close() error { return nil }

func ioNopCloser(s string) io.ReadCloser {
	return closer{strings.NewReader(s)}
}
