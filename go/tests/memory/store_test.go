package memory_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rexleimo/agno-go/internal/agent"
	memimpl "github.com/rexleimo/agno-go/internal/memory"
	memstore "github.com/rexleimo/agno-go/pkg/memory"
)

func TestInMemoryStoreLifecycle(t *testing.T) {
	store := memstore.NewInMemoryStore()
	agentID := uuid.New()
	sessionID := uuid.New()

	if err := store.UpsertSession(context.Background(), agentID, sessionID); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	msg := agent.Message{ID: uuid.New(), Role: agent.RoleUser, Content: "hi"}
	if err := store.AppendMessage(context.Background(), agentID, sessionID, msg); err != nil {
		t.Fatalf("append message: %v", err)
	}
	result := agent.ToolCallResult{ToolCallID: "tc1", Status: agent.ToolStatusSuccess, Output: "ok"}
	if err := store.AppendToolResult(context.Background(), agentID, sessionID, result); err != nil {
		t.Fatalf("append tool result: %v", err)
	}
	history, err := store.LoadHistory(context.Background(), agentID, sessionID, agent.HistoryOptions{TokenWindow: 10})
	if err != nil {
		t.Fatalf("load history: %v", err)
	}
	if len(history) != 1 || len(history[0].ToolCalls) != 0 {
		t.Fatalf("unexpected history entries: %+v", history)
	}
	if err := store.DeleteSession(context.Background(), agentID, sessionID); err != nil {
		t.Fatalf("delete session: %v", err)
	}
	if _, err := store.LoadHistory(context.Background(), agentID, sessionID, agent.HistoryOptions{}); err == nil {
		t.Fatalf("expected error after delete")
	}
}

func TestBoltStoreTokenWindow(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bolt.db")
	store, err := memimpl.NewBoltStore(path)
	if err != nil {
		t.Fatalf("new bolt: %v", err)
	}
	defer func() { _ = store.Close() }()

	agentID := uuid.New()
	sessionID := uuid.New()
	if err := store.UpsertSession(context.Background(), agentID, sessionID); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	for i := 0; i < 5; i++ {
		msg := agent.Message{
			ID:        uuid.New(),
			Role:      agent.RoleUser,
			Content:   "hello",
			Usage:     agent.Usage{PromptTokens: 2, CompletionTokens: 1},
			CreatedAt: time.Now(),
		}
		if err := store.AppendMessage(context.Background(), agentID, sessionID, msg); err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}
	history, err := store.LoadHistory(context.Background(), agentID, sessionID, agent.HistoryOptions{TokenWindow: 6})
	if err != nil {
		t.Fatalf("load history: %v", err)
	}
	if len(history) > 3 {
		t.Fatalf("expected trimmed history, got %d entries", len(history))
	}
}

func TestBadgerStoreCRUD(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "badger")
	store, err := memimpl.NewBadgerStore(dir, 0)
	if err != nil {
		t.Fatalf("new badger: %v", err)
	}
	defer func() { _ = store.Close() }()

	agentID := uuid.New()
	sessionID := uuid.New()
	if err := store.UpsertSession(context.Background(), agentID, sessionID); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	msg := agent.Message{ID: uuid.New(), Role: agent.RoleAssistant, Content: "resp"}
	if err := store.AppendMessage(context.Background(), agentID, sessionID, msg); err != nil {
		t.Fatalf("append: %v", err)
	}
	history, err := store.LoadHistory(context.Background(), agentID, sessionID, agent.HistoryOptions{TokenWindow: 1})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(history) != 1 || history[0].Content != "resp" {
		t.Fatalf("unexpected history: %+v", history)
	}
	if err := store.DeleteSession(context.Background(), agentID, sessionID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := store.LoadHistory(context.Background(), agentID, sessionID, agent.HistoryOptions{}); err == nil {
		t.Fatalf("expected error after delete")
	}
	// ensure dir exists
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("badger dir missing: %v", err)
	}
}

func TestBadgerStoreTTLExpires(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "badger-ttl")
	store, err := memimpl.NewBadgerStore(dir, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("new badger: %v", err)
	}
	defer func() { _ = store.Close() }()

	agentID := uuid.New()
	sessionID := uuid.New()
	if err := store.UpsertSession(context.Background(), agentID, sessionID); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	msg := agent.Message{ID: uuid.New(), Role: agent.RoleAssistant, Content: "ttl"}
	if err := store.AppendMessage(context.Background(), agentID, sessionID, msg); err != nil {
		t.Fatalf("append: %v", err)
	}
	time.Sleep(600 * time.Millisecond)
	if _, err := store.LoadHistory(context.Background(), agentID, sessionID, agent.HistoryOptions{}); err == nil {
		t.Fatalf("expected ttl expiry")
	}
}
