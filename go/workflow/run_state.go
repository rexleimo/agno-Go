package workflow

import (
	"time"
)

// RunStatus represents the lifecycle state of a workflow run.
type RunStatus string

// StepStatus represents the lifecycle state of an individual workflow step.
type StepStatus string

const (
	RunStatusPending   RunStatus = "pending"
	RunStatusRunning   RunStatus = "running"
	RunStatusCompleted RunStatus = "completed"
	RunStatusFailed    RunStatus = "failed"
	RunStatusPaused    RunStatus = "paused"

	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusPaused    StepStatus = "paused"
)

// WorkflowRun captures the serialized representation of a workflow execution.
type WorkflowRun struct {
	ID                   string                   `json:"runId"`
	SessionID            string                   `json:"sessionId"`
	WorkflowRef          ID                       `json:"workflowRef"`
	PatternType          PatternType              `json:"patternType"`
	Status               RunStatus                `json:"status"`
	Steps                []StepRun                `json:"steps,omitempty"`
	RoutingRulesApplied  []RoutingRuleApplication `json:"routingRulesApplied,omitempty"`
	ReasoningTrace       []ReasoningStep          `json:"reasoningTrace,omitempty"`
	BackgroundTasks      []BackgroundTaskState    `json:"backgroundTasks,omitempty"`
	ResourcesUsed        ResourceUsage            `json:"metrics"`
	Metadata             map[string]string        `json:"metadata,omitempty"`
	StartedAt            time.Time                `json:"startedAt"`
	UpdatedAt            time.Time                `json:"updatedAt"`
	TerminationCondition TerminationCondition     `json:"terminationCondition"`
}

// StepRun stores the state of a single workflow step invocation.
type StepRun struct {
	ID          StepID             `json:"stepId"`
	AgentID     string             `json:"agentId"`
	Name        string             `json:"name,omitempty"`
	Status      StepStatus         `json:"status"`
	Input       StepIODigest       `json:"input"`
	Output      StepIODigest       `json:"output"`
	Reasoning   []ReasoningStep    `json:"reasoning,omitempty"`
	StartedAt   time.Time          `json:"startedAt"`
	CompletedAt time.Time          `json:"completedAt"`
	Metrics     map[string]float64 `json:"metrics,omitempty"`
	Error       string             `json:"error,omitempty"`
}

// StepIODigest summarizes input/output payloads for serialization and parity.
type StepIODigest struct {
	Summary string         `json:"summary,omitempty"`
	Payload map[string]any `json:"payload,omitempty"`
}

// Duration returns how long the step took to complete. Zero is returned when
// either timestamp is unset or the step has not finished yet.
func (s StepRun) Duration() time.Duration {
	if s.StartedAt.IsZero() || s.CompletedAt.IsZero() {
		return 0
	}
	return s.CompletedAt.Sub(s.StartedAt)
}

// RoutingRuleApplication records an evaluated routing rule and its outcome.
type RoutingRuleApplication struct {
	RuleID    string    `json:"ruleId"`
	From      StepID    `json:"from"`
	To        StepID    `json:"to"`
	Condition string    `json:"condition"`
	Result    string    `json:"result"`
	Timestamp time.Time `json:"timestamp"`
}

// ReasoningStep mirrors the openapi ReasoningStep payload for telemetry.
type ReasoningStep struct {
	Order    int            `json:"order"`
	Type     string         `json:"type"`
	Text     string         `json:"text"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// BackgroundTaskState captures the lifecycle of async tasks triggered by a run.
type BackgroundTaskState struct {
	TaskID    string         `json:"taskId"`
	Name      string         `json:"name"`
	Type      string         `json:"type"`
	Status    string         `json:"status"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	StartedAt time.Time      `json:"startedAt"`
	EndedAt   time.Time      `json:"endedAt"`
}

// ResourceUsage aggregates performance and billing data for the run.
type ResourceUsage struct {
	PromptTokens     int                `json:"promptTokens,omitempty"`
	CompletionTokens int                `json:"completionTokens,omitempty"`
	TotalTokens      int                `json:"totalTokens,omitempty"`
	LatencyMillis    int                `json:"latencyMs,omitempty"`
	ToolCalls        int                `json:"toolCalls,omitempty"`
	Additional       map[string]float64 `json:"additional,omitempty"`
}

// IsTerminal indicates if the workflow run has reached a terminal state.
func (s RunStatus) IsTerminal() bool {
	return s == RunStatusCompleted || s == RunStatusFailed || s == RunStatusPaused
}
