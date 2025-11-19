package agent

import (
	"context"
	"testing"
	"time"

	"github.com/agno-agi/agno-go/go/session"
)

func TestRuntimeServiceRegisterAndLookup(t *testing.T) {
	store := session.NewMemoryStore()
	svc := NewRuntimeService(store, WithClock(func() time.Time { return time.Unix(1700000000, 0) }))
	ctx := context.Background()

	runtime := AgentRuntime{
		ID:       ID("agent-1"),
		Name:     "Example",
		ModelRef: "openai:gpt-5-mini",
		SessionPolicy: SessionPolicy{
			CacheSession: true,
		},
	}

	registered, err := svc.RegisterAgentRuntime(ctx, runtime)
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if registered == nil || registered.ID != runtime.ID {
		t.Fatalf("unexpected runtime returned: %#v", registered)
	}

	if registered.Metadata == nil {
		registered.Metadata = map[string]string{}
	}
	registered.Metadata["mutated"] = "true"
	stored, ok := svc.AgentRuntime(runtime.ID)
	if !ok {
		t.Fatalf("runtime should be retrievable after registration")
	}
	if stored.Metadata != nil {
		t.Fatalf("expected metadata mutation to not leak into registry")
	}

	all := svc.AgentRuntimes()
	if len(all) != 1 {
		t.Fatalf("expected 1 runtime, got %d", len(all))
	}
	if all[0].ID != runtime.ID {
		t.Fatalf("unexpected runtime ID in list: %s", all[0].ID)
	}
}

func TestRuntimeServiceSessionPolicyPersistsSession(t *testing.T) {
	store := session.NewMemoryStore()
	svc := NewRuntimeService(store)
	ctx := context.Background()

	_, err := svc.RegisterAgentRuntime(ctx, AgentRuntime{
		ID:       ID("agent-sp"),
		ModelRef: "openai:gpt-5-mini",
		SessionPolicy: SessionPolicy{
			SessionID:               "session-123",
			OverwriteDBSessionState: true,
			EnableAgenticState:      false,
			CacheSession:            true,
		},
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	rec, err := store.Get(ctx, session.ID("session-123"))
	if err != nil {
		t.Fatalf("session not created: %v", err)
	}
	policy, ok := rec.Context.Payload["session_policy"].(map[string]any)
	if !ok {
		t.Fatalf("expected session_policy payload to be recorded: %+v", rec.Context.Payload)
	}
	if policy["cache_session"] != true {
		t.Fatalf("expected cache_session flag, got %+v", policy)
	}
	if rec.Status != session.StatusPending {
		t.Fatalf("expected session status pending, got %s", rec.Status)
	}
}

func TestRuntimeServiceValidationError(t *testing.T) {
	svc := NewRuntimeService(nil)
	ctx := context.Background()

	_, err := svc.RegisterAgentRuntime(ctx, AgentRuntime{})
	if err == nil {
		t.Fatalf("expected validation error for empty runtime")
	}
}
