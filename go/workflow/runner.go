package workflow

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrAgentNotRegistered indicates that a workflow references an agent that was
// not supplied in the run request.
var ErrAgentNotRegistered = errors.New("workflow: agent not registered")

// ErrNoAgents indicates that the run request did not include any agents.
var ErrNoAgents = errors.New("workflow: no agents provided")

// StepExecutor executes a workflow step and produces structured results.
type StepExecutor interface {
	Execute(ctx context.Context, req StepExecutionRequest) (StepExecutionResult, error)
}

// StepExecutionRequest bundles the runtime context passed to the executor.
type StepExecutionRequest struct {
	SessionID string
	Step      Step
	Input     map[string]any
	Metadata  map[string]string
}

// StepExecutionResult captures normalized step output, reasoning metadata and
// metrics. It intentionally mirrors StepRun for parity diffing.
type StepExecutionResult struct {
	Output    StepIODigest
	Reasoning []ReasoningStep
	Metrics   map[string]float64
}

// StepInput describes the external payload associated with a workflow step.
type StepInput struct {
	StepID  StepID
	Payload map[string]any
}

// RunRequest drives a WorkflowRun execution.
type RunRequest struct {
	RunID     string
	SessionID string
	AgentIDs  []string
	Inputs    []StepInput
	Metadata  map[string]string
}

// Runner executes a Workflow using the injected StepExecutor.
type Runner struct {
	workflow Workflow
	executor StepExecutor
	now      func() time.Time
}

// NewRunner builds a Runner for the provided workflow. The executor must not
// be nil.
func NewRunner(wf Workflow, executor StepExecutor) (*Runner, error) {
	if executor == nil {
		return nil, errors.New("workflow: executor is required")
	}
	return &Runner{
		workflow: wf,
		executor: executor,
		now:      time.Now,
	}, nil
}

// Run executes the workflow sequentially while preserving parity metadata for
// telemetry and contract tests. Parallel and coordinator-worker patterns are
// represented via BackgroundTasks metadata until the full concurrency model is
// implemented.
func (r *Runner) Run(ctx context.Context, req RunRequest) (*WorkflowRun, error) {
	if req.SessionID == "" {
		return nil, errors.New("workflow: sessionID is required")
	}
	if len(req.AgentIDs) == 0 {
		return nil, ErrNoAgents
	}
	agentSet := make(map[string]struct{}, len(req.AgentIDs))
	for _, id := range req.AgentIDs {
		agentSet[id] = struct{}{}
	}

	inputMap := make(map[StepID]map[string]any, len(req.Inputs))
	for _, input := range req.Inputs {
		inputMap[input.StepID] = clonePayload(input.Payload)
	}

	run := r.bootstrapRun(req)

	for i, step := range r.workflow.Steps {
		if _, ok := agentSet[step.AgentID]; !ok {
			return nil, fmt.Errorf("%w: %s", ErrAgentNotRegistered, step.AgentID)
		}
		stepRun := StepRun{
			ID:      step.ID,
			AgentID: step.AgentID,
			Name:    step.Name,
			Status:  StepStatusPending,
			Input: StepIODigest{
				Payload: clonePayload(inputMap[step.ID]),
			},
			Metrics: map[string]float64{},
		}
		run.Steps = append(run.Steps, stepRun)

		updatedRun, err := r.executeStep(ctx, run, i, inputMap)
		if updatedRun != nil {
			run = updatedRun
		}
		if err != nil {
			return run, err
		}
	}

	run.Status = RunStatusCompleted
	run.ResourcesUsed.ToolCalls = len(run.Steps)
	r.appendBackgroundTasks(run)
	if run.ResourcesUsed.TotalTokens == 0 {
		run.ResourcesUsed.TotalTokens = run.ResourcesUsed.PromptTokens + run.ResourcesUsed.CompletionTokens
	}
	return run, nil
}

func (r *Runner) bootstrapRun(req RunRequest) *WorkflowRun {
	now := r.now()
	runID := req.RunID
	if strings.TrimSpace(runID) == "" {
		runID = fmt.Sprintf("%s-%d", r.workflow.ID, now.UnixNano())
	}
	return &WorkflowRun{
		ID:                   runID,
		SessionID:            req.SessionID,
		WorkflowRef:          r.workflow.ID,
		PatternType:          r.workflow.PatternType,
		Status:               RunStatusRunning,
		Metadata:             cloneStringMap(req.Metadata),
		StartedAt:            now,
		UpdatedAt:            now,
		TerminationCondition: r.workflow.TerminationCondition,
		ResourcesUsed: ResourceUsage{
			Additional: map[string]float64{},
		},
	}
}

func (r *Runner) executeStep(ctx context.Context, run *WorkflowRun, idx int, inputs map[StepID]map[string]any) (*WorkflowRun, error) {
	if idx >= len(run.Steps) {
		return run, nil
	}
	stepRun := &run.Steps[idx]
	step := r.workflow.Steps[idx]
	start := r.now()
	stepRun.Status = StepStatusRunning
	stepRun.StartedAt = start
	run.UpdatedAt = start

	if err := ctx.Err(); err != nil {
		stepRun.Status = StepStatusFailed
		stepRun.Error = err.Error()
		stepRun.CompletedAt = r.now()
		run.Status = RunStatusFailed
		return run, err
	}

	result, err := r.executor.Execute(ctx, StepExecutionRequest{
		SessionID: run.SessionID,
		Step:      step,
		Input:     inputs[step.ID],
		Metadata:  run.Metadata,
	})
	completed := r.now()
	stepRun.CompletedAt = completed
	run.UpdatedAt = completed
	if err != nil {
		stepRun.Status = StepStatusFailed
		stepRun.Error = err.Error()
		run.Status = RunStatusFailed
		return run, err
	}

	stepRun.Status = StepStatusCompleted
	stepRun.Output = result.Output
	if len(result.Reasoning) > 0 {
		stepRun.Reasoning = append([]ReasoningStep(nil), result.Reasoning...)
		run.ReasoningTrace = append(run.ReasoningTrace, result.Reasoning...)
	}
	if len(result.Metrics) > 0 {
		stepRun.Metrics = cloneMetrics(result.Metrics)
		r.mergeMetrics(run, result.Metrics)
	}
	r.recordRouting(step.ID, completed, run)
	return run, nil
}

func (r *Runner) mergeMetrics(run *WorkflowRun, metrics map[string]float64) {
	for key, value := range metrics {
		switch strings.ToLower(key) {
		case "prompt_tokens":
			run.ResourcesUsed.PromptTokens += int(value)
		case "completion_tokens":
			run.ResourcesUsed.CompletionTokens += int(value)
		case "total_tokens":
			run.ResourcesUsed.TotalTokens += int(value)
		case "latency_ms":
			run.ResourcesUsed.LatencyMillis += int(value)
		case "tool_calls":
			run.ResourcesUsed.ToolCalls += int(value)
		default:
			if run.ResourcesUsed.Additional == nil {
				run.ResourcesUsed.Additional = map[string]float64{}
			}
			run.ResourcesUsed.Additional[key] = run.ResourcesUsed.Additional[key] + value
		}
	}
}

func (r *Runner) recordRouting(stepID StepID, ts time.Time, run *WorkflowRun) {
	for _, rule := range r.workflow.RoutingRules {
		if rule.From != stepID {
			continue
		}
		app := RoutingRuleApplication{
			RuleID:    fmt.Sprintf("%s->%s", rule.From, rule.To),
			From:      rule.From,
			To:        rule.To,
			Condition: rule.Condition,
			Result:    "applied",
			Timestamp: ts,
		}
		run.RoutingRulesApplied = append(run.RoutingRulesApplied, app)
	}
}

func (r *Runner) appendBackgroundTasks(run *WorkflowRun) {
	if r.workflow.PatternType != PatternParallel && r.workflow.PatternType != PatternCoordinatorWorker {
		return
	}
	for _, step := range run.Steps {
		bt := BackgroundTaskState{
			TaskID:    string(step.ID),
			Name:      step.Name,
			Type:      string(r.workflow.PatternType),
			Status:    string(step.Status),
			Metadata:  map[string]any{"agent_id": step.AgentID},
			StartedAt: step.StartedAt,
			EndedAt:   step.CompletedAt,
		}
		run.BackgroundTasks = append(run.BackgroundTasks, bt)
	}
}

func clonePayload(src map[string]any) map[string]any {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneMetrics(src map[string]float64) map[string]float64 {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]float64, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
