package telemetry

import (
	"testing"
	"time"
)

func TestInMemoryRecorderRecordAndFilter(t *testing.T) {
	recorder := NewInMemoryRecorder()
	recorder.Record(Event{ID: "1", Timestamp: time.Now(), Type: EventRunStarted, SessionID: "s1"})
	recorder.Record(Event{ID: "2", Timestamp: time.Now(), Type: EventToolCall, SessionID: "s2", Payload: map[string]any{"prompt_tokens": 10, "latency_ms": 5}})

	events := recorder.List(Filter{SessionID: "s2"})
	if len(events) != 1 {
		t.Fatalf("expected one event, got %d", len(events))
	}
	if events[0].Type != EventToolCall {
		t.Fatalf("unexpected event type: %s", events[0].Type)
	}

	stats := recorder.Stats()
	if stats.EventCount != 2 {
		t.Fatalf("unexpected stats count: %+v", stats)
	}
	if stats.PromptTokens != 10 || stats.LatencyMillis != 5 {
		t.Fatalf("unexpected stats aggregations: %+v", stats)
	}
}

func TestInMemoryRecorderLimit(t *testing.T) {
	recorder := NewInMemoryRecorder()
	for i := 0; i < 5; i++ {
		recorder.Record(Event{ID: string(rune('a' + i)), Timestamp: time.Now(), Type: EventRunStarted})
	}
	if got := len(recorder.List(Filter{Limit: 2})); got != 2 {
		t.Fatalf("expected limit=2, got %d", got)
	}
}
