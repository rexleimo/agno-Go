package telemetry

import (
	"testing"
	"time"
)

func TestNormalizeSetsDefaults(t *testing.T) {
	event := Event{
		Type:    EventType("custom"),
		Payload: map[string]any{},
	}
	normalized := Normalize(event)
	if normalized.Runtime != RuntimeGo {
		t.Fatalf("expected runtime=go, got %s", normalized.Runtime)
	}
	if normalized.Type != EventUnknown {
		t.Fatalf("expected unknown event type, got %s", normalized.Type)
	}
	if normalized.Payload["original_event_type"] != EventType("custom") {
		t.Fatalf("missing original event type in payload: %+v", normalized.Payload)
	}
}

func TestValidateRuntime(t *testing.T) {
	event := Event{
		ID:        "1",
		Timestamp: time.Now(),
		Runtime:   RuntimeGo,
		Type:      EventRunStarted,
	}
	if err := Validate(event); err != nil {
		t.Fatalf("expected valid event: %v", err)
	}
	event.Runtime = "invalid"
	if err := Validate(event); err == nil {
		t.Fatalf("expected validation error for runtime")
	}
}
