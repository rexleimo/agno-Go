package workflow

import (
	"context"
	"errors"
	"testing"
)

func TestRunnerSequentialSuccess(t *testing.T) {
	wf := Workflow{
		ID:          ID("wf-demo"),
		PatternType: PatternSequential,
		Steps: []Step{
			{ID: StepID("s1"), AgentID: "agent-a", Name: "step-1"},
			{ID: StepID("s2"), AgentID: "agent-b", Name: "step-2"},
		},
		RoutingRules: []RoutingRule{{From: StepID("s1"), To: StepID("s2"), Condition: "always"}},
	}
	exec := &stubExecutor{
		results: map[StepID]StepExecutionResult{
			StepID("s1"): {
				Output:  StepIODigest{Summary: "first", Payload: map[string]any{"ok": true}},
				Metrics: map[string]float64{"prompt_tokens": 10, "latency_ms": 5},
			},
			StepID("s2"): {
				Output:    StepIODigest{Summary: "second"},
				Reasoning: []ReasoningStep{{Order: 1, Type: "thought", Text: "done"}},
				Metrics:   map[string]float64{"completion_tokens": 20},
			},
		},
	}
	runner, err := NewRunner(wf, exec)
	if err != nil {
		t.Fatalf("NewRunner: %v", err)
	}
	ctx := context.Background()
	result, err := runner.Run(ctx, RunRequest{
		SessionID: "session-1",
		AgentIDs:  []string{"agent-a", "agent-b"},
		Inputs: []StepInput{{
			StepID:  StepID("s1"),
			Payload: map[string]any{"topic": "go"},
		}},
		Metadata: map[string]string{"run": "demo"},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Status != RunStatusCompleted {
		t.Fatalf("expected completed status, got %s", result.Status)
	}
	if len(result.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(result.Steps))
	}
	if result.Steps[0].Status != StepStatusCompleted || result.Steps[1].Status != StepStatusCompleted {
		t.Fatalf("expected both steps completed, got %+v", result.Steps)
	}
	if len(result.ReasoningTrace) != 1 {
		t.Fatalf("expected reasoning trace captured")
	}
	if result.ResourcesUsed.PromptTokens != 10 || result.ResourcesUsed.CompletionTokens != 20 {
		t.Fatalf("unexpected aggregated metrics: %+v", result.ResourcesUsed)
	}
	if len(result.RoutingRulesApplied) != 1 {
		t.Fatalf("expected routing rule application to be recorded")
	}
}

func TestRunnerMissingAgent(t *testing.T) {
	wf := Workflow{
		ID:          ID("wf"),
		PatternType: PatternSequential,
		Steps: []Step{
			{ID: StepID("s1"), AgentID: "agent-a"},
		},
	}
	runner, err := NewRunner(wf, &stubExecutor{})
	if err != nil {
		t.Fatalf("NewRunner: %v", err)
	}
	_, err = runner.Run(context.Background(), RunRequest{
		SessionID: "s",
		AgentIDs:  []string{"agent-b"},
	})
	if err == nil || !errors.Is(err, ErrAgentNotRegistered) {
		t.Fatalf("expected ErrAgentNotRegistered, got %v", err)
	}
}

func TestRunnerStepErrorStopsRun(t *testing.T) {
	wf := Workflow{
		ID:          ID("wf"),
		PatternType: PatternSequential,
		Steps:       []Step{{ID: StepID("s1"), AgentID: "agent-a"}, {ID: StepID("s2"), AgentID: "agent-b"}},
	}
	exec := &stubExecutor{
		errors: map[StepID]error{
			StepID("s2"): errors.New("boom"),
		},
	}
	runner, err := NewRunner(wf, exec)
	if err != nil {
		t.Fatalf("NewRunner: %v", err)
	}
	result, err := runner.Run(context.Background(), RunRequest{
		SessionID: "s",
		AgentIDs:  []string{"agent-a", "agent-b"},
	})
	if err == nil {
		t.Fatalf("expected error for failing step")
	}
	if result.Status != RunStatusFailed {
		t.Fatalf("expected failed run, got %s", result.Status)
	}
	if result.Steps[1].Status != StepStatusFailed {
		t.Fatalf("expected second step to fail")
	}
}

type stubExecutor struct {
	results map[StepID]StepExecutionResult
	errors  map[StepID]error
}

func (s *stubExecutor) Execute(_ context.Context, req StepExecutionRequest) (StepExecutionResult, error) {
	if err, ok := s.errors[req.Step.ID]; ok {
		return StepExecutionResult{}, err
	}
	if res, ok := s.results[req.Step.ID]; ok {
		return res, nil
	}
	return StepExecutionResult{Output: StepIODigest{Summary: string(req.Step.ID)}}, nil
}
