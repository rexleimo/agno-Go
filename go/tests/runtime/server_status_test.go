package runtime_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rexleimo/agno-go/internal/agent"
	"github.com/rexleimo/agno-go/internal/model"
	"github.com/rexleimo/agno-go/internal/runtime"
	"github.com/rexleimo/agno-go/pkg/memory"
	"github.com/rexleimo/agno-go/pkg/providers/stub"
)

func TestPostMessageConflictReturns409(t *testing.T) {
	router := model.NewRouter()
	router.RegisterChatProvider(stub.New(agent.ProviderOpenAI, model.ProviderAvailable, nil))
	svc := runtime.NewService(memory.NewInMemoryStore(), router)
	server := runtime.NewServer(router.Statuses, "test", svc)

	agentID, _ := svc.CreateAgent(context.Background(), agent.Agent{
		Name: "conflict",
		Model: agent.ModelConfig{
			Provider: agent.ProviderOpenAI,
			ModelID:  "stub",
			Stream:   true,
		},
	})
	session, _ := svc.CreateSession(context.Background(), agentID, "", nil)

	// Force state to streaming to trigger conflict
	svc.SetSessionStateForTest(agentID, session.ID, agent.SessionStreaming)

	body := `{"messages":[{"role":"user","content":"hi"}]}`
	req := httptest.NewRequest(http.MethodPost, "/agents/"+agentID.String()+"/sessions/"+session.ID.String()+"/messages", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	server.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestStreamEndpointSetsSSEHeaders(t *testing.T) {
	router := model.NewRouter()
	router.RegisterChatProvider(stub.New(agent.ProviderOpenAI, model.ProviderAvailable, nil))
	svc := runtime.NewService(memory.NewInMemoryStore(), router)
	server := runtime.NewServer(router.Statuses, "test", svc)

	agentID, _ := svc.CreateAgent(context.Background(), agent.Agent{
		Name: "stream",
		Model: agent.ModelConfig{
			Provider: agent.ProviderOpenAI,
			ModelID:  "stub",
			Stream:   true,
		},
	})
	session, _ := svc.CreateSession(context.Background(), agentID, "", nil)

	payload := map[string]any{
		"messages": []map[string]any{
			{"role": "user", "content": "stream please"},
		},
	}
	buf, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/agents/"+agentID.String()+"/sessions/"+session.ID.String()+"/messages?stream=true", bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	server.Router.ServeHTTP(rr, req)

	if rr.Code != http.StatusMultiStatus {
		t.Fatalf("expected 207 for stream, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct == "" || !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("expected SSE content type, got %s", ct)
	}
}
