package telemetry

import "time"

// EventType represents the kind of telemetry event being recorded.
type EventType string

const (
	EventRequestStarted         EventType = "request_started"
	EventRequestFinished        EventType = "request_finished"
	EventProviderError          EventType = "provider_error"
	EventWorkflowStepTransition EventType = "workflow_step_transition"
)

// Event is a structured telemetry record for a single notable occurrence
// during a session or workflow run.
type Event struct {
	ID        string
	Timestamp time.Time
	// SessionID is a string representation of the session identifier. Concrete
	// types (such as session.ID) are converted by the caller to avoid import
	// cycles between packages.
	SessionID string
	// AgentID is optional and may be left empty when the event is not tied to a
	// specific agent.
	AgentID string
	// ProviderID is optional and may be left empty when the event is not tied
	// to a specific provider.
	ProviderID string
	Type       EventType
	// Payload holds non-sensitive, structured data used for diagnostics and
	// parity testing. Sensitive data (such as credentials or personal
	// information) must be stripped by the caller.
	Payload map[string]any
}

// Recorder is a minimal interface for emitting telemetry events. Concrete
// implementations can forward events to logs, metrics backends or tracing
// systems.
type Recorder interface {
	Record(Event)
}

// NoopRecorder is a Recorder that discards all events. It is useful in tests
// and in early wiring stages where a full telemetry pipeline is not yet
// required.
type NoopRecorder struct{}

func (NoopRecorder) Record(Event) {}
