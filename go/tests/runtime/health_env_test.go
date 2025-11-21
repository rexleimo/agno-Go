package runtime_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rexleimo/agno-go/internal/agent"
	"github.com/rexleimo/agno-go/internal/model"
	"github.com/rexleimo/agno-go/internal/runtime"
	runtimeconfig "github.com/rexleimo/agno-go/internal/runtime/config"
	"github.com/rexleimo/agno-go/pkg/memory"
	"github.com/rexleimo/agno-go/pkg/providers/stub"
)

func TestHealthReportsMissingProviders(t *testing.T) {
	restore := unsetAllProviderKeys()
	defer restore()

	cfgPath := filepath.Join("..", "..", "..", "config", "default.yaml")
	cfg, err := runtimeconfig.LoadWithEnv(cfgPath, "")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	server := runtime.NewServer(cfg.ProviderStatuses, "dev", nil)
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	server.Router.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	var body struct {
		Status    string                 `json:"status"`
		Version   string                 `json:"version"`
		Providers []model.ProviderStatus `json:"providers"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Status != "ok" {
		t.Fatalf("expected status ok, got %s", body.Status)
	}
	for _, p := range body.Providers {
		if p.Status != model.ProviderNotConfigured {
			t.Fatalf("expected %s to be not-configured, got %s", p.Provider, p.Status)
		}
	}
}

func TestHealthReturnsDisabledProviders(t *testing.T) {
	router := model.NewRouter()
	router.RegisterChatProvider(stub.New(agent.ProviderOpenAI, model.ProviderAvailable, nil))
	router.RegisterChatProvider(stub.New(agent.ProviderGemini, model.ProviderDisabled, []string{"disabled explicitly"}))
	svc := runtime.NewService(memory.NewInMemoryStore(), router)
	server := runtime.NewServer(router.Statuses, "dev", svc)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	server.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var body struct {
		Providers []model.ProviderStatus `json:"providers"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode health: %v", err)
	}
	if len(body.Providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(body.Providers))
	}
	var hasDisabled bool
	for _, st := range body.Providers {
		if st.Status == model.ProviderDisabled {
			hasDisabled = true
		}
	}
	if !hasDisabled {
		t.Fatalf("expected disabled provider reported in health")
	}
}

func TestPostMessageBlockedWhenProviderNotConfigured(t *testing.T) {
	router := model.NewRouter()
	router.RegisterChatProvider(stub.New(agent.ProviderOpenAI, model.ProviderNotConfigured, []string{"OPENAI_API_KEY"}))
	svc := runtime.NewService(memory.NewInMemoryStore(), router)
	server := runtime.NewServer(router.Statuses, "dev", svc)

	agentID := createAgent(t, server.Router)
	sessionID := createSession(t, server.Router, agentID)

	payload := `{"messages":[{"role":"user","content":"hello blocked"}]}`
	url := fmt.Sprintf("/agents/%s/sessions/%s/messages", agentID, sessionID)
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	server.Router.ServeHTTP(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		body, _ := io.ReadAll(rr.Body)
		t.Fatalf("expected 503 for not-configured provider, got %d body=%s", rr.Code, string(body))
	}
	body, _ := io.ReadAll(rr.Body)
	if !strings.Contains(strings.ToLower(string(body)), "not-configured") {
		t.Fatalf("expected not-configured reason in body, got %s", string(body))
	}
}

func unsetAllProviderKeys() func() {
	keys := []string{
		"OPENAI_API_KEY", "GEMINI_API_KEY", "GLM4_API_KEY", "OPENROUTER_API_KEY",
		"SILICONFLOW_API_KEY", "CEREBRAS_API_KEY", "MODELSCOPE_API_KEY", "GROQ_API_KEY",
	}
	orig := make(map[string]string, len(keys))
	for _, k := range keys {
		orig[k] = os.Getenv(k)
		_ = os.Unsetenv(k)
	}
	return func() {
		for k, v := range orig {
			_ = os.Setenv(k, v)
		}
	}
}

func createAgent(t *testing.T, handler http.Handler) uuid.UUID {
	t.Helper()
	payload := `{"name":"runtime-health","model":{"provider":"openai","modelId":"test-model","stream":true}}`
	req := httptest.NewRequest(http.MethodPost, "/agents", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		body, _ := io.ReadAll(rr.Body)
		t.Fatalf("create agent status=%d body=%s", rr.Code, string(body))
	}
	var out struct {
		AgentID string `json:"agentId"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode agent response: %v", err)
	}
	id, err := uuid.Parse(out.AgentID)
	if err != nil {
		t.Fatalf("parse agentId: %v", err)
	}
	return id
}

func createSession(t *testing.T, handler http.Handler, agentID uuid.UUID) uuid.UUID {
	t.Helper()
	payload := `{}`
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/agents/%s/sessions", agentID.String()), strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		body, _ := io.ReadAll(rr.Body)
		t.Fatalf("create session status=%d body=%s", rr.Code, string(body))
	}
	var out struct {
		SessionID string `json:"sessionId"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode session response: %v", err)
	}
	id, err := uuid.Parse(out.SessionID)
	if err != nil {
		t.Fatalf("parse sessionId: %v", err)
	}
	return id
}
