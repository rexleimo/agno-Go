package telemetry

import (
	"sync"
)

// Filter constrains the events returned by the recorder.
type Filter struct {
	SessionID     string
	WorkflowRunID string
	AgentID       string
	EventType     EventType
	Runtime       string
	Limit         int
}

// Stats aggregates numeric metrics derived from telemetry payloads.
type Stats struct {
	EventCount       int
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	LatencyMillis    int
}

// Store represents a Recorder that can also list stored events and expose
// aggregate stats.
type Store interface {
	Recorder
	List(Filter) []Event
	Stats() Stats
}

// InMemoryRecorder stores telemetry events for inspection and HTTP export.
type InMemoryRecorder struct {
	mu     sync.RWMutex
	events []Event
	stats  Stats
}

// NewInMemoryRecorder constructs a recorder backed by in-memory storage.
func NewInMemoryRecorder() *InMemoryRecorder {
	return &InMemoryRecorder{
		events: make([]Event, 0, 64),
	}
}

// Record stores the event after enforcing runtime and event-type constraints.
func (r *InMemoryRecorder) Record(event Event) {
	normalized := Normalize(event)
	if normalized.Runtime != RuntimeGo {
		normalized.Runtime = RuntimeGo
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, normalized)
	r.updateStatsLocked(normalized)
}

// List returns events filtered by the provided constraints.
func (r *InMemoryRecorder) List(filter Filter) []Event {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Event, 0, len(r.events))
	runtimeFilter := filter.Runtime
	if runtimeFilter == "" {
		runtimeFilter = RuntimeGo
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}
	for _, evt := range r.events {
		if runtimeFilter != "" && evt.Runtime != runtimeFilter {
			continue
		}
		if filter.SessionID != "" && evt.SessionID != filter.SessionID {
			continue
		}
		if filter.WorkflowRunID != "" && evt.WorkflowRunID != filter.WorkflowRunID {
			continue
		}
		if filter.AgentID != "" && evt.AgentID != filter.AgentID {
			continue
		}
		if filter.EventType != "" && evt.Type != filter.EventType {
			continue
		}
		result = append(result, evt)
		if len(result) >= limit {
			break
		}
	}
	return result
}

// Stats returns a snapshot of the aggregated metrics.
func (r *InMemoryRecorder) Stats() Stats {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.stats
}

func (r *InMemoryRecorder) updateStatsLocked(event Event) {
	r.stats.EventCount++
	if v, ok := metricFromPayload(event.Payload, "prompt_tokens"); ok {
		r.stats.PromptTokens += v
	}
	if v, ok := metricFromPayload(event.Payload, "completion_tokens"); ok {
		r.stats.CompletionTokens += v
	}
	if v, ok := metricFromPayload(event.Payload, "total_tokens"); ok {
		r.stats.TotalTokens += v
	}
	if v, ok := metricFromPayload(event.Payload, "latency_ms"); ok {
		r.stats.LatencyMillis += v
	}
}

func metricFromPayload(payload map[string]any, key string) (int, bool) {
	if len(payload) == 0 {
		return 0, false
	}
	val, ok := payload[key]
	if !ok {
		return 0, false
	}
	switch v := val.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case float32:
		return int(v), true
	default:
		return 0, false
	}
}
