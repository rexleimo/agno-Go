package runtime_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rexleimo/agno-go/internal/agent"
	"github.com/rexleimo/agno-go/internal/model"
	"github.com/rexleimo/agno-go/internal/runtime"
	rtmiddleware "github.com/rexleimo/agno-go/internal/runtime/middleware"
	"github.com/rexleimo/agno-go/pkg/memory"
)

func TestStreamBackpressureBlocksNewRequestsButKeepsInFlightStreaming(t *testing.T) {
	limiter := rtmiddleware.NewConcurrencyLimiter(1, 0, 200*time.Millisecond)
	router := model.NewRouter()
	router.RegisterChatProvider(&slowStreamProvider{delay: 150 * time.Millisecond})
	svc := runtime.NewService(memory.NewInMemoryStore(), router)
	server := runtime.NewServer(router.Statuses, "test", svc, runtime.WithConcurrencyLimiter(limiter))

	agentID := createAgent(t, server.Router)
	sessionID := createSession(t, server.Router, agentID)

	payload := `{"messages":[{"role":"user","content":"load"}]}`

	streamDone := make(chan struct{})
	var resp1 *http.Response
	go func() {
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/agents/%s/sessions/%s/messages?stream=true", agentID, sessionID), strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		rr := newStreamingRecorder()
		server.Router.ServeHTTP(rr, req)
		resp1 = rr.Result()
		close(streamDone)
	}()

	time.Sleep(25 * time.Millisecond) // allow the first stream to enter the limiter

	req2, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/agents/%s/sessions/%s/messages?stream=true", agentID, sessionID), strings.NewReader(payload))
	req2.Header.Set("Content-Type", "application/json")
	rr2 := newStreamingRecorder()
	server.Router.ServeHTTP(rr2, req2)
	resp2 := rr2.Result()
	if resp2.StatusCode != http.StatusTooManyRequests {
		body, _ := io.ReadAll(resp2.Body)
		t.Fatalf("expected 429 backpressure, got %d body=%s", resp2.StatusCode, string(body))
	}
	if reqID := resp2.Header.Get("X-Request-ID"); reqID == "" {
		t.Fatalf("expected request id on backpressure response")
	}
	var bp rtmiddleware.BackpressureResponse
	if err := json.NewDecoder(resp2.Body).Decode(&bp); err != nil {
		t.Fatalf("decode backpressure body: %v", err)
	}
	if bp.Error != "backpressure" || bp.Limit != 1 {
		t.Fatalf("unexpected backpressure payload: %+v", bp)
	}

	select {
	case <-streamDone:
	case <-time.After(2 * time.Second):
		t.Fatalf("stream did not complete in time")
	}
	defer func() { _ = resp1.Body.Close() }()
	if resp1.StatusCode != http.StatusMultiStatus {
		t.Fatalf("expected 207 for stream, got %d", resp1.StatusCode)
	}
	data, _ := io.ReadAll(resp1.Body)
	if !bytes.Contains(data, []byte("chunk-1")) || !bytes.Contains(data, []byte("chunk-2")) {
		t.Fatalf("stream body missing expected chunks: %s", string(data))
	}
}

type slowStreamProvider struct {
	delay time.Duration
}

func (p *slowStreamProvider) Name() agent.Provider {
	return agent.ProviderOpenAI
}

func (p *slowStreamProvider) Status() model.ProviderStatus {
	return model.ProviderStatus{
		Provider:     agent.ProviderOpenAI,
		Status:       model.ProviderAvailable,
		Capabilities: []model.Capability{model.CapabilityChat, model.CapabilityStreaming},
	}
}

func (p *slowStreamProvider) Chat(ctx context.Context, req model.ChatRequest) (*model.ChatResponse, error) {
	return &model.ChatResponse{
		Message: agent.Message{
			Role:    agent.RoleAssistant,
			Content: "ok",
		},
	}, nil
}

func (p *slowStreamProvider) Stream(ctx context.Context, req model.ChatRequest, fn model.StreamHandler) error {
	if err := fn(model.ChatStreamEvent{Type: "token", Delta: "chunk-1"}); err != nil {
		return err
	}
	timer := time.NewTimer(p.delay)
	select {
	case <-ctx.Done():
		timer.Stop()
		return ctx.Err()
	case <-timer.C:
	}
	if err := fn(model.ChatStreamEvent{Type: "token", Delta: "chunk-2"}); err != nil {
		return err
	}
	return fn(model.ChatStreamEvent{Type: "end", Done: true})
}

// newStreamingRecorder ensures interfaces needed by SSE handlers are available.
func newStreamingRecorder() *responseRecorderWithFlush {
	return &responseRecorderWithFlush{ResponseRecorder: httptest.NewRecorder()}
}

type responseRecorderWithFlush struct {
	*httptest.ResponseRecorder
}

func (r *responseRecorderWithFlush) Flush() {}
