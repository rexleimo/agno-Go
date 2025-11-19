package session

import (
	"testing"
	"time"
)

func TestNewSessionRecordFromSession(t *testing.T) {
	now := time.Now()
	s := &Session{
		ID:       ID("session-1"),
		Workflow: "wf",
		Context: UserContext{
			UserID:  "user-1",
			Payload: map[string]any{"topic": "go"},
		},
		History: []HistoryEntry{{
			Timestamp: now,
			Source:    "user",
			Message:   "hello",
			Metadata:  map[string]any{"turn": 1},
		}},
		Status:    StatusCompleted,
		Result:    &Result{Success: true, Reason: "ok", Data: map[string]any{"tokens": 42}},
		CreatedAt: now.Add(-time.Minute),
		UpdatedAt: now,
	}

	record := NewSessionRecordFromSession(s)
	if record.SessionID != string(s.ID) {
		t.Fatalf("expected session id preserved")
	}
	if len(record.History) != 1 || record.History[0].Role != "user" {
		t.Fatalf("expected history to be converted")
	}
	if record.Metrics["tokens"] != 42 {
		t.Fatalf("expected numeric metrics to be extracted, got %+v", record.Metrics)
	}
	if record.Summary.Text != "ok" {
		t.Fatalf("expected summary to reuse result reason")
	}
}

func TestApplySessionRecord(t *testing.T) {
	record := SessionRecord{
		SessionID: "session-2",
		UserID:    "user-2",
		StateBlob: map[string]any{"foo": "bar"},
		History: []RunMessage{{
			Role:      "assistant",
			Content:   "hi",
			Timestamp: time.Now(),
		}},
		Summary:   SessionSummary{Text: "summary"},
		Metrics:   map[string]float64{"tokens": 10},
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now(),
	}

	s := ApplySessionRecord(record, nil)
	if s.ID != ID("session-2") {
		t.Fatalf("expected session id to match, got %s", s.ID)
	}
	if s.Context.UserID != "user-2" {
		t.Fatalf("expected user id populated")
	}
	if len(s.History) != 1 || s.History[0].Source != "assistant" {
		t.Fatalf("expected history reconstructed")
	}
	if s.Result == nil || s.Result.Reason != "summary" {
		t.Fatalf("expected result summary to be applied")
	}
	metrics, ok := s.Result.Data["metrics"].(map[string]float64)
	if !ok || metrics["tokens"] != 10 {
		t.Fatalf("expected metrics embedded in result data, got %+v", s.Result.Data)
	}
}
