package culture

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestManager_AddAndListKnowledge(t *testing.T) {
	store := NewInMemoryStore()
	now := time.Date(2024, 10, 1, 12, 0, 0, 0, time.UTC)

	manager, err := NewManager(store, WithClock(func() time.Time { return now }))
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ctx := context.Background()
	entry, err := manager.AddKnowledge(ctx, &Entry{
		Content: "Use friendly tone with customers",
		Tags:    []string{"support", "voice"},
	})
	if err != nil {
		t.Fatalf("AddKnowledge failed: %v", err)
	}

	if entry.ID == "" {
		t.Fatal("expected generated ID")
	}
	if entry.CreatedAt != now {
		t.Fatalf("expected CreatedAt %v, got %v", now, entry.CreatedAt)
	}
	if len(entry.Tags) != 2 || entry.Tags[0] != "support" || entry.Tags[1] != "voice" {
		t.Fatalf("unexpected tags order: %#v", entry.Tags)
	}

	results, err := manager.ListKnowledge(ctx, Filter{Tags: []string{"voice"}})
	if err != nil {
		t.Fatalf("ListKnowledge failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != entry.ID {
		t.Fatalf("expected entry %s, got %s", entry.ID, results[0].ID)
	}
}

func TestManager_UpdateKnowledge(t *testing.T) {
	store := NewInMemoryStore()
	manager, err := NewManager(store)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ctx := context.Background()
	entry, _ := manager.AddKnowledge(ctx, &Entry{Content: "Use concise greetings"})

	updated, err := manager.UpdateKnowledge(ctx, &Entry{
		ID:      entry.ID,
		Content: "Use personalised greetings",
		Tags:    []string{"greeting"},
	})
	if err != nil {
		t.Fatalf("UpdateKnowledge failed: %v", err)
	}

	if updated.Content != "Use personalised greetings" {
		t.Fatalf("unexpected content: %s", updated.Content)
	}
	if len(updated.Tags) != 1 || updated.Tags[0] != "greeting" {
		t.Fatalf("unexpected tags: %#v", updated.Tags)
	}

	stored, err := manager.GetKnowledge(ctx, entry.ID)
	if err != nil {
		t.Fatalf("GetKnowledge failed: %v", err)
	}
	if stored.Content != "Use personalised greetings" {
		t.Fatalf("store not updated: %s", stored.Content)
	}
}

func TestManager_AddKnowledgeAsync(t *testing.T) {
	store := NewInMemoryStore()
	manager, err := NewManager(store)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ctx := context.Background()
	ch := manager.AddKnowledgeAsync(ctx, &Entry{Content: "Respond in markdown"})

	select {
	case res := <-ch:
		if res.Err != nil {
			t.Fatalf("async result error: %v", res.Err)
		}
		if res.Entry == nil || res.Entry.Content != "Respond in markdown" {
			t.Fatalf("unexpected entry: %#v", res.Entry)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for async result")
	}
}

func TestManager_RemoveAndClear(t *testing.T) {
	store := NewInMemoryStore()
	manager, err := NewManager(store)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	ctx := context.Background()
	entry, _ := manager.AddKnowledge(ctx, &Entry{Content: "Log hand-offs"})

	if err := manager.RemoveKnowledge(ctx, entry.ID); err != nil {
		t.Fatalf("RemoveKnowledge failed: %v", err)
	}
	if _, err := manager.GetKnowledge(ctx, entry.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	_, _ = manager.AddKnowledge(ctx, &Entry{Content: "Keep escalation notes"})
	if err := manager.Clear(ctx); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}
	results, _ := manager.ListKnowledge(ctx, Filter{})
	if len(results) != 0 {
		t.Fatalf("expected empty store, got %d", len(results))
	}
}
