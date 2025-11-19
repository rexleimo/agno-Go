package agent

import (
	"fmt"
	"strings"
)

// ToolkitRef identifies a toolkit/knowledge/memory bundle that can be injected
// into an AgentRuntime during execution.
type ToolkitRef string

// TimeoutPolicy describes upper bounds (in milliseconds) for runtime execution.
type TimeoutPolicy struct {
	// RunMilliseconds caps the entire Agent run duration.
	RunMilliseconds int `json:"runMs,omitempty"`
	// ToolCallMilliseconds caps individual tool/function invocations.
	ToolCallMilliseconds int `json:"toolCallMs,omitempty"`
}

// SessionPolicy mirrors the python session management semantics while exposing
// migration-friendly flags.
type SessionPolicy struct {
	// SessionID allows callers to pin the runtime to a particular session. It
	// can be empty when the runtime is created before a run is scheduled.
	SessionID string `json:"sessionId,omitempty"`
	// OverwriteDBSessionState indicates that existing session_state blobs can be
	// overwritten. Mutually exclusive with EnableAgenticState.
	OverwriteDBSessionState bool `json:"overwriteDbSessionState,omitempty"`
	// EnableAgenticState preserves agent-generated state blobs; cannot be
	// combined with OverwriteDBSessionState.
	EnableAgenticState bool `json:"enableAgenticState,omitempty"`
	// CacheSession toggles in-memory caching for repeated lookups. This does not
	// alter persistence guarantees but influences performance.
	CacheSession bool `json:"cacheSession,omitempty"`
}

// AgentRuntime represents the full declarative configuration required to
// materialize an agent in the Go runtime.
type AgentRuntime struct {
	ID            ID                `json:"id"`
	Name          string            `json:"name,omitempty"`
	ModelRef      string            `json:"modelRef"`
	Toolkits      []ToolkitRef      `json:"toolkits,omitempty"`
	MemoryPolicy  MemoryPolicy      `json:"memoryPolicy"`
	SessionPolicy SessionPolicy     `json:"sessionPolicy"`
	Hooks         map[string]string `json:"hooks,omitempty"`
	Timeouts      TimeoutPolicy     `json:"timeouts,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// AgentRuntimeValidationError keeps track of a single invalid field.
type AgentRuntimeValidationError struct {
	Field   string
	Message string
}

func (e AgentRuntimeValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// AgentRuntimeValidationErrors aggregates multiple validation errors.
type AgentRuntimeValidationErrors []AgentRuntimeValidationError

func (errs AgentRuntimeValidationErrors) Error() string {
	if len(errs) == 0 {
		return "agent runtime validation failed"
	}

	var b strings.Builder
	b.WriteString("agent runtime validation failed:")
	for _, err := range errs {
		b.WriteString(" ")
		b.WriteString(err.Error())
		b.WriteString(";")
	}
	return strings.TrimSuffix(b.String(), ";")
}

// NewAgentRuntime validates the provided configuration and returns a sanitized
// runtime instance.
func NewAgentRuntime(cfg AgentRuntime) (*AgentRuntime, error) {
	rt := cfg
	rt.applyDefaults()
	if err := rt.Validate(); err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *AgentRuntime) applyDefaults() {
	if r.Hooks == nil {
		r.Hooks = map[string]string{}
	}
	if r.Metadata == nil {
		r.Metadata = map[string]string{}
	}
}

// Validate enforces the rules captured in data-model.md for AgentRuntime.
func (r AgentRuntime) Validate() error {
	var errs AgentRuntimeValidationErrors

	if strings.TrimSpace(string(r.ID)) == "" {
		errs = append(errs, AgentRuntimeValidationError{
			Field:   "id",
			Message: "must not be empty",
		})
	}

	if strings.TrimSpace(r.ModelRef) == "" {
		errs = append(errs, AgentRuntimeValidationError{
			Field:   "modelRef",
			Message: "must not be empty",
		})
	}

	if r.MemoryPolicy.WindowSize < 0 {
		errs = append(errs, AgentRuntimeValidationError{
			Field:   "memoryPolicy.windowSize",
			Message: "must be greater than or equal to zero",
		})
	}

	if r.SessionPolicy.OverwriteDBSessionState && r.SessionPolicy.EnableAgenticState {
		errs = append(errs, AgentRuntimeValidationError{
			Field:   "sessionPolicy",
			Message: "overwriteDbSessionState and enableAgenticState cannot both be true",
		})
	}

	if r.Timeouts.RunMilliseconds < 0 {
		errs = append(errs, AgentRuntimeValidationError{
			Field:   "timeouts.runMilliseconds",
			Message: "must be greater than or equal to zero",
		})
	}

	if r.Timeouts.ToolCallMilliseconds < 0 {
		errs = append(errs, AgentRuntimeValidationError{
			Field:   "timeouts.toolCallMilliseconds",
			Message: "must be greater than or equal to zero",
		})
	}

	toolkitSet := map[ToolkitRef]struct{}{}
	for i, toolkit := range r.Toolkits {
		if strings.TrimSpace(string(toolkit)) == "" {
			errs = append(errs, AgentRuntimeValidationError{
				Field:   fmt.Sprintf("toolkits[%d]", i),
				Message: "must not be empty",
			})
			continue
		}
		if _, exists := toolkitSet[toolkit]; exists {
			errs = append(errs, AgentRuntimeValidationError{
				Field:   fmt.Sprintf("toolkits[%d]", i),
				Message: "duplicate toolkit reference",
			})
			continue
		}
		toolkitSet[toolkit] = struct{}{}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
