package agentos

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/agno-go/pkg/agno/session"
)

func TestNewServer(t *testing.T) {
	server, err := NewServer(nil)
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if server == nil {
		t.Fatal("Expected non-nil server")
	}

	if server.config.Address != ":8080" {
		t.Errorf("Address = %v, want ':8080'", server.config.Address)
	}
}

func TestNewServer_WithConfig(t *testing.T) {
	storage := session.NewMemoryStorage()

	server, err := NewServer(&Config{
		Address:        ":9090",
		SessionStorage: storage,
		Debug:          true,
	})
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	if server.config.Address != ":9090" {
		t.Errorf("Address = %v, want ':9090'", server.config.Address)
	}

	if !server.config.Debug {
		t.Error("Debug should be true")
	}
}

func TestHealthEndpoint(t *testing.T) {
	server, _ := NewServer(nil)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %v, want %v", w.Code, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Status = %v, want 'healthy'", response["status"])
	}
}

func TestCreateSession(t *testing.T) {
	server, _ := NewServer(nil)

	reqBody := CreateSessionRequest{
		AgentID: "test-agent",
		UserID:  "test-user",
		Name:    "Test Session",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/sessions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Status = %v, want %v", w.Code, http.StatusCreated)
	}

	var response SessionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.SessionID == "" {
		t.Error("SessionID should not be empty")
	}

	if response.AgentID != "test-agent" {
		t.Errorf("AgentID = %v, want 'test-agent'", response.AgentID)
	}

	if response.UserID != "test-user" {
		t.Errorf("UserID = %v, want 'test-user'", response.UserID)
	}
}

func TestCreateSession_MissingAgentID(t *testing.T) {
	server, _ := NewServer(nil)

	reqBody := CreateSessionRequest{
		UserID: "test-user",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/sessions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestGetSession(t *testing.T) {
	server, _ := NewServer(nil)

	// First create a session
	sess := session.NewSession("test-session-123", "test-agent")
	server.sessionStorage.Create(context.Background(), sess)

	// Then get it
	req, _ := http.NewRequest("GET", "/api/v1/sessions/test-session-123", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %v, want %v", w.Code, http.StatusOK)
	}

	var response SessionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.SessionID != "test-session-123" {
		t.Errorf("SessionID = %v, want 'test-session-123'", response.SessionID)
	}
}

func TestGetSession_NotFound(t *testing.T) {
	server, _ := NewServer(nil)

	req, _ := http.NewRequest("GET", "/api/v1/sessions/non-existent", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %v, want %v", w.Code, http.StatusNotFound)
	}
}

func TestUpdateSession(t *testing.T) {
	server, _ := NewServer(nil)

	// Create a session first
	sess := session.NewSession("test-session-123", "test-agent")
	sess.Name = "Original Name"
	server.sessionStorage.Create(context.Background(), sess)

	// Update it
	updateReq := UpdateSessionRequest{
		Name: "Updated Name",
		Metadata: map[string]interface{}{
			"key": "value",
		},
	}

	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/api/v1/sessions/test-session-123", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %v, want %v", w.Code, http.StatusOK)
	}

	var response SessionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Name != "Updated Name" {
		t.Errorf("Name = %v, want 'Updated Name'", response.Name)
	}

	if response.Metadata["key"] != "value" {
		t.Error("Metadata not updated correctly")
	}
}

func TestDeleteSession(t *testing.T) {
	server, _ := NewServer(nil)

	// Create a session first
	sess := session.NewSession("test-session-123", "test-agent")
	server.sessionStorage.Create(context.Background(), sess)

	// Delete it
	req, _ := http.NewRequest("DELETE", "/api/v1/sessions/test-session-123", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %v, want %v", w.Code, http.StatusOK)
	}

	// Verify it's deleted
	_, err := server.sessionStorage.Get(context.Background(), "test-session-123")
	if err != session.ErrSessionNotFound {
		t.Error("Session should be deleted")
	}
}

func TestListSessions(t *testing.T) {
	server, _ := NewServer(nil)

	// Create multiple sessions
	sess1 := session.NewSession("sess-1", "agent-1")
	sess1.UserID = "user-1"
	server.sessionStorage.Create(context.Background(), sess1)

	sess2 := session.NewSession("sess-2", "agent-1")
	sess2.UserID = "user-2"
	server.sessionStorage.Create(context.Background(), sess2)

	sess3 := session.NewSession("sess-3", "agent-2")
	sess3.UserID = "user-1"
	server.sessionStorage.Create(context.Background(), sess3)

	// List all sessions
	req, _ := http.NewRequest("GET", "/api/v1/sessions", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %v, want %v", w.Code, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	count := int(response["count"].(float64))
	if count != 3 {
		t.Errorf("Count = %v, want 3", count)
	}
}

func TestListSessions_WithFilter(t *testing.T) {
	server, _ := NewServer(nil)

	// Create multiple sessions
	sess1 := session.NewSession("sess-1", "agent-1")
	sess1.UserID = "user-1"
	server.sessionStorage.Create(context.Background(), sess1)

	sess2 := session.NewSession("sess-2", "agent-1")
	sess2.UserID = "user-2"
	server.sessionStorage.Create(context.Background(), sess2)

	sess3 := session.NewSession("sess-3", "agent-2")
	sess3.UserID = "user-1"
	server.sessionStorage.Create(context.Background(), sess3)

	// Filter by agent_id
	req, _ := http.NewRequest("GET", "/api/v1/sessions?agent_id=agent-1", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %v, want %v", w.Code, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	count := int(response["count"].(float64))
	if count != 2 {
		t.Errorf("Count = %v, want 2", count)
	}
}

func TestAgentRun(t *testing.T) {
	server, _ := NewServer(nil)

	// Agent needs to be registered first (will return 404 if not registered)
	runReq := AgentRunRequest{
		Input: "Hello, agent!",
	}

	body, _ := json.Marshal(runReq)
	req, _ := http.NewRequest("POST", "/api/v1/agents/test-agent/run", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	// Expect 404 since agent is not registered
	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %v, want %v (agent not registered)", w.Code, http.StatusNotFound)
	}
}

func TestAgentRun_WithSession(t *testing.T) {
	server, _ := NewServer(nil)

	// Create a session first
	sess := session.NewSession("test-session", "test-agent")
	server.sessionStorage.Create(context.Background(), sess)

	runReq := AgentRunRequest{
		Input:     "Hello, agent!",
		SessionID: "test-session",
	}

	body, _ := json.Marshal(runReq)
	req, _ := http.NewRequest("POST", "/api/v1/agents/test-agent/run", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	// Expect 404 since agent is not registered
	if w.Code != http.StatusNotFound {
		t.Errorf("Status = %v, want %v (agent not registered)", w.Code, http.StatusNotFound)
	}
}
