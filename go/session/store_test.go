package session

import (
	"context"
	"testing"
	"time"

	"github.com/agno-agi/agno-go/go/workflow"
)

func TestMemoryStoreCRUD(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	now := time.Now()
	rec := &Session{
		ID:       ID("session-1"),
		Workflow: "wf-1",
		Context: UserContext{
			UserID:    "user-1",
			Payload:   map[string]any{"foo": "bar"},
			StartedAt: now,
		},
		Status: StatusRunning,
		History: []HistoryEntry{
			{Timestamp: now, Source: "agent", Message: "hello", Metadata: map[string]any{"k": "v"}},
		},
		Result: &Result{
			Success: true,
			Data:    map[string]any{"result": 1},
		},
		UpdatedAt: now,
	}

	if err := store.Save(ctx, rec); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := store.Get(ctx, rec.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if got.ID != rec.ID || got.Workflow != rec.Workflow {
		t.Fatalf("unexpected record %#v", got)
	}

	// mutate retrieved record and ensure store value is unaffected
	got.Context.Payload["foo"] = "baz"
	got.History[0].Metadata["k"] = "changed"

	gotAgain, err := store.Get(ctx, rec.ID)
	if err != nil {
		t.Fatalf("get again: %v", err)
	}

	if gotAgain.Context.Payload["foo"] != "bar" {
		t.Fatalf("expected payload to remain unchanged, got %v", gotAgain.Context.Payload)
	}
	if gotAgain.History[0].Metadata["k"] != "v" {
		t.Fatalf("expected history metadata to remain unchanged")
	}

	if err := store.Delete(ctx, rec.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if _, err := store.Get(ctx, rec.ID); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestMemoryStoreSearch(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	now := time.Now()
	sessions := []*Session{
		{
			ID:        ID("s1"),
			Workflow:  workflow.ID("wf-1"),
			Context:   UserContext{UserID: "alice"},
			Status:    StatusCompleted,
			UpdatedAt: now,
		},
		{
			ID:        ID("s2"),
			Workflow:  workflow.ID("wf-2"),
			Context:   UserContext{UserID: "bob"},
			Status:    StatusFailed,
			UpdatedAt: now.Add(-2 * time.Hour),
		},
		{
			ID:        ID("s3"),
			Workflow:  workflow.ID("wf-1"),
			Context:   UserContext{UserID: "alice"},
			Status:    StatusRunning,
			UpdatedAt: now.Add(-1 * time.Hour),
		},
	}

	for _, s := range sessions {
		if err := store.Save(ctx, s); err != nil {
			t.Fatalf("save session %s: %v", s.ID, err)
		}
	}

	results, err := store.Search(ctx, SearchFilter{
		UserID:   "alice",
		Workflow: workflow.ID("wf-1"),
		Statuses: []Status{StatusCompleted},
	})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 1 || results[0].ID != ID("s1") {
		t.Fatalf("unexpected results %+v", results)
	}

	results, err = store.Search(ctx, SearchFilter{
		Since: now.Add(-90 * time.Minute),
		Limit: 1,
	})
	if err != nil {
		t.Fatalf("search w limit: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result with limit 1, got %d", len(results))
	}
}

func TestNewStoreDriverRegistration(t *testing.T) {
	_, err := NewStore(DriverMemory, Options{})
	if err != nil {
		t.Fatalf("memory driver should be registered: %v", err)
	}

	if _, err := NewStore(DriverSQLite, Options{}); err == nil {
		t.Fatalf("expected error for sqlite without factory")
	}

	driverName := Driver("sqlite-temp")
	RegisterDriver(driverName, func(Options) (Store, error) {
		return NewMemoryStore(), nil
	})
	if _, err := NewStore(driverName, Options{}); err != nil {
		t.Fatalf("custom driver should succeed: %v", err)
	}
}
