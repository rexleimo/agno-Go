package agent

import "testing"

func TestRunUS1Example(t *testing.T) {
	input := US1Input{
		Query: "test query",
	}

	out, err := RunUS1Example(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Query != input.Query {
		t.Fatalf("expected query %q, got %q", input.Query, out.Query)
	}
	if out.Result == nil {
		t.Fatalf("expected non-nil result")
	}
	if out.Result["workflow_id"] == nil {
		t.Fatalf("expected workflow_id in result")
	}
	if out.Result["status"] == nil {
		t.Fatalf("expected status in result")
	}
}
