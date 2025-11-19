package workflow

// ID uniquely identifies a workflow or collaboration pattern.
type ID string

// PatternType captures the high-level collaboration style used by a workflow.
type PatternType string

const (
	PatternSequential        PatternType = "sequential"
	PatternParallel          PatternType = "parallel"
	PatternCoordinatorWorker PatternType = "coordinator-worker"
	PatternLoop              PatternType = "loop"
)

// StepID identifies a single workflow step.
type StepID string

// Step represents a single unit of work within a workflow. It typically maps
// to an Agent invocation, a decision point, or a branch.
type Step struct {
	ID       StepID            `json:"id"`
	AgentID  string            `json:"agentId"`
	Name     string            `json:"name,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// RoutingRule describes how control flows from one step to another.
type RoutingRule struct {
	From StepID `json:"from"`
	To   StepID `json:"to"`
	// Condition is an opaque expression evaluated against the step result or
	// session state; interpretation is left to higher layers.
	Condition string `json:"condition,omitempty"`
}

// TerminationCondition describes when a workflow run should be considered
// complete.
type TerminationCondition struct {
	// MaxIterations is used for loop-like workflows. Zero means no explicit
	// iteration limit.
	MaxIterations int `json:"maxIterations,omitempty"`
	// OnError indicates how errors should influence termination semantics
	// (for example, "fail-fast" vs "continue-on-error").
	OnError string `json:"onError,omitempty"`
}

// Workflow describes a collaboration pattern between multiple agents.
type Workflow struct {
	ID                   ID                   `json:"id"`
	Name                 string               `json:"name,omitempty"`
	PatternType          PatternType          `json:"patternType"`
	Steps                []Step               `json:"steps,omitempty"`
	EntryPoints          []StepID             `json:"entryPoints,omitempty"`
	TerminationCondition TerminationCondition `json:"terminationCondition"`
	RoutingRules         []RoutingRule        `json:"routingRules,omitempty"`
}
