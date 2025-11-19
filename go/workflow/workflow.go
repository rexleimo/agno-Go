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
	ID       StepID
	AgentID  string
	Name     string
	Metadata map[string]string
}

// RoutingRule describes how control flows from one step to another.
type RoutingRule struct {
	From StepID
	To   StepID
	// Condition is an opaque expression evaluated against the step result or
	// session state; interpretation is left to higher layers.
	Condition string
}

// TerminationCondition describes when a workflow run should be considered
// complete.
type TerminationCondition struct {
	// MaxIterations is used for loop-like workflows. Zero means no explicit
	// iteration limit.
	MaxIterations int
	// OnError indicates how errors should influence termination semantics
	// (for example, "fail-fast" vs "continue-on-error").
	OnError string
}

// Workflow describes a collaboration pattern between multiple agents.
type Workflow struct {
	ID                   ID
	Name                 string
	PatternType          PatternType
	Steps                []Step
	EntryPoints          []StepID
	TerminationCondition TerminationCondition
	RoutingRules         []RoutingRule
}
