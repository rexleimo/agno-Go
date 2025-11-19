package telemetry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestTelemetryParityFromPythonLog(t *testing.T) {
	logPath := filepath.Join("..", "..", "..", "specs", "001-migrate-agno-core", "fixtures", "assets", "telemetry", "python_log.json")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read python log: %v", err)
	}
	var pythonEvents []Event
	if err := json.Unmarshal(data, &pythonEvents); err != nil {
		t.Fatalf("parse python log: %v", err)
	}
	recorder := NewInMemoryRecorder()
	for _, evt := range pythonEvents {
		recorder.Record(evt)
	}
	unknown := Event{ID: "unknown", Type: EventType("python_only"), Payload: map[string]any{}}
	recorder.Record(unknown)

	stored := recorder.List(Filter{Limit: len(pythonEvents) + 1})
	if len(stored) != len(pythonEvents)+1 {
		t.Fatalf("unexpected stored events length: %d", len(stored))
	}
	if stored[0].Payload["prompt_tokens"].(float64) != 120 {
		t.Fatalf("expected prompt_tokens preserved, got %+v", stored[0].Payload)
	}
	last := stored[len(stored)-1]
	if last.Type != EventUnknown {
		t.Fatalf("expected unknown event type fallback, got %s", last.Type)
	}
	if last.Payload["original_event_type"] != EventType("python_only") {
		t.Fatalf("expected original event type recorded: %+v", last.Payload)
	}

	stats := recorder.Stats()
	if stats.PromptTokens != 120 {
		t.Fatalf("unexpected prompt token stats: %+v", stats)
	}
	if stats.TotalTokens != 200 {
		t.Fatalf("unexpected total token stats: %+v", stats)
	}
}
