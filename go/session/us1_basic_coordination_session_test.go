package session

import (
	"testing"

	"github.com/agno-agi/agno-go/go/workflow"
)

func TestRunUS1Session(t *testing.T) {
	wf := workflow.US1BasicCoordinationWorkflow()
	s := RunUS1Session("test query", wf)

	if s.Status != StatusCompleted {
		t.Fatalf("expected session status completed, got %q", s.Status)
	}
	if s.Result == nil || !s.Result.Success {
		t.Fatalf("expected successful result, got %+v", s.Result)
	}
	if s.Workflow != wf.ID {
		t.Errorf("session workflow ID mismatch: got %q, want %q", s.Workflow, wf.ID)
	}
	if len(s.History) == 0 {
		t.Errorf("expected at least one history entry")
	}
}
