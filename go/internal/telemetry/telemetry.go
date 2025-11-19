package telemetry

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"
)

// EventType represents the kind of telemetry event being recorded.
type EventType string

// Supported event types for parity with the Python runtime. Unknown event
// types are normalized to EventUnknown.
const (
	EventRunStarted      EventType = "run_started"
	EventReasoningStep   EventType = "reasoning_step"
	EventToolCall        EventType = "tool_call"
	EventSessionSummary  EventType = "session_summary"
	EventRunCompleted    EventType = "run_completed"
	EventRequestStarted  EventType = "request_started"
	EventRequestFinished EventType = "request_finished"
	EventProviderError   EventType = "provider_error"
	EventWorkflowStep    EventType = "workflow_step_transition"
	EventUnknown         EventType = "unknown_event"
)

// Runtime identifiers shared across languages.
const (
	RuntimeGo     = "go"
	RuntimePython = "python"
)

var supportedEventTypes = map[EventType]struct{}{
	EventRunStarted:      {},
	EventReasoningStep:   {},
	EventToolCall:        {},
	EventSessionSummary:  {},
	EventRunCompleted:    {},
	EventRequestStarted:  {},
	EventRequestFinished: {},
	EventProviderError:   {},
	EventWorkflowStep:    {},
	EventUnknown:         {},
}

// Attachment represents optional metadata related to an event, such as a
// reference to an external artifact.
type Attachment struct {
	Type      string `json:"type"`
	Reference string `json:"reference"`
}

// Event is a structured telemetry record for a single notable occurrence
// during a session or workflow run.
type Event struct {
	ID             string         `json:"eventId"`
	Timestamp      time.Time      `json:"timestamp"`
	Runtime        string         `json:"runtime"`
	Type           EventType      `json:"eventType"`
	SessionID      string         `json:"sessionId,omitempty"`
	WorkflowRunID  string         `json:"workflowRunId,omitempty"`
	AgentID        string         `json:"agentId,omitempty"`
	ProviderID     string         `json:"providerId,omitempty"`
	Payload        map[string]any `json:"payload,omitempty"`
	Attachments    []Attachment   `json:"attachments,omitempty"`
	CorrelationIDs []string       `json:"correlationIds,omitempty"`
}

// Recorder is a minimal interface for emitting telemetry events. Concrete
// implementations can forward events to logs, metrics backends or tracing
// systems.
type Recorder interface {
	Record(Event)
}

// Validate verifies that the event conforms to the TelemetryEnvelope schema.
func Validate(event Event) error {
	if stringsTrim(event.Runtime) == "" {
		return errors.New("telemetry: runtime must be set")
	}
	if event.Runtime != RuntimeGo && event.Runtime != RuntimePython {
		return fmt.Errorf("telemetry: unsupported runtime %q", event.Runtime)
	}
	if _, ok := supportedEventTypes[event.Type]; !ok {
		return fmt.Errorf("telemetry: unsupported event type %q", event.Type)
	}
	return nil
}

// Normalize enforces runtime defaults and coerces unknown event types to the
// EventUnknown constant.
func Normalize(event Event) Event {
	if stringsTrim(event.Runtime) == "" {
		event.Runtime = RuntimeGo
	}
	if _, ok := supportedEventTypes[event.Type]; !ok {
		if event.Payload == nil {
			event.Payload = map[string]any{}
		}
		event.Payload["original_event_type"] = event.Type
		if _, exists := event.Payload["runtime_version"]; !exists {
			event.Payload["runtime_version"] = runtime.Version()
		}
		event.Type = EventUnknown
	}
	return event
}

func stringsTrim(v string) string {
	return strings.TrimSpace(v)
}

// NoopRecorder is a Recorder that discards all events. It is useful in tests
// and in early wiring stages where a full telemetry pipeline is not yet
// required.
type NoopRecorder struct{}

func (NoopRecorder) Record(Event) {}
