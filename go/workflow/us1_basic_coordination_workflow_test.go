package workflow

import "testing"

func TestUS1BasicCoordinationWorkflow(t *testing.T) {
	wf := US1BasicCoordinationWorkflow()

	if wf.ID == "" {
		t.Fatal("workflow ID must not be empty")
	}
	if wf.PatternType != PatternSequential {
		t.Fatalf("expected PatternSequential, got %q", wf.PatternType)
	}
	if len(wf.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(wf.Steps))
	}

	if wf.Steps[0].AgentID != "hn_researcher" {
		t.Errorf("first step should use hn_researcher, got %q", wf.Steps[0].AgentID)
	}
	if wf.Steps[1].AgentID != "article_reader" {
		t.Errorf("second step should use article_reader, got %q", wf.Steps[1].AgentID)
	}
	if len(wf.EntryPoints) != 1 || wf.EntryPoints[0] != wf.Steps[0].ID {
		t.Errorf("expected single entrypoint equal to first step ID")
	}
}
