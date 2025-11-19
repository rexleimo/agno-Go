package providers

import (
	"testing"

	"github.com/agno-agi/agno-go/go/internal/telemetry"
)

type recordingRecorder struct {
	events []telemetry.Event
}

func (r *recordingRecorder) Record(ev telemetry.Event) {
	r.events = append(r.events, ev)
}

func TestNewNotMigratedError(t *testing.T) {
	err := NewNotMigratedError(ID("unknown-provider"))
	if err.Code != "not_migrated" {
		t.Fatalf("expected not_migrated code, got %v", err.Code)
	}
}

func TestRecordNotMigrated(t *testing.T) {
	rec := &recordingRecorder{}
	RecordNotMigrated(rec, ID("unknown-provider"))

	if len(rec.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(rec.events))
	}
	ev := rec.events[0]
	if ev.ProviderID != "unknown-provider" {
		t.Errorf("unexpected ProviderID: %q", ev.ProviderID)
	}
	if ev.Payload["error_code"] != "not_migrated" {
		t.Errorf("unexpected error_code in payload: %v", ev.Payload["error_code"])
	}
}
