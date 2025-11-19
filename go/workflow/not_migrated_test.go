package workflow

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

func TestNewNotMigratedPatternError(t *testing.T) {
	err := NewNotMigratedPatternError(ID("wf-unknown"), PatternSequential)
	if err.Code != "not_migrated" {
		t.Fatalf("expected not_migrated code, got %v", err.Code)
	}
}

func TestRecordNotMigratedWorkflow(t *testing.T) {
	rec := &recordingRecorder{}
	RecordNotMigratedWorkflow(rec, ID("wf-unknown"), PatternParallel)

	if len(rec.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(rec.events))
	}
	ev := rec.events[0]
	if ev.Payload["error_code"] != "not_migrated" {
		t.Errorf("unexpected error_code: %v", ev.Payload["error_code"])
	}
	if ev.Payload["workflow_id"] != "wf-unknown" {
		t.Errorf("unexpected workflow_id: %v", ev.Payload["workflow_id"])
	}
}
