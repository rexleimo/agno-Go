package workflow

import (
	"testing"
	"time"
)

func TestStepRunDuration(t *testing.T) {
	now := time.Now()
	step := StepRun{
		StartedAt:   now.Add(-2 * time.Second),
		CompletedAt: now,
	}
	if d := step.Duration(); d != 2*time.Second {
		t.Fatalf("expected 2s duration, got %v", d)
	}

	step = StepRun{
		StartedAt: now,
	}
	if d := step.Duration(); d != 0 {
		t.Fatalf("expected zero duration for incomplete step, got %v", d)
	}
}

func TestRunStatusIsTerminal(t *testing.T) {
	cases := []struct {
		status   RunStatus
		expected bool
	}{
		{RunStatusPending, false},
		{RunStatusRunning, false},
		{RunStatusCompleted, true},
		{RunStatusFailed, true},
		{RunStatusPaused, true},
	}

	for _, tc := range cases {
		if tc.status.IsTerminal() != tc.expected {
			t.Fatalf("status %s expected terminal=%v", tc.status, tc.expected)
		}
	}
}
