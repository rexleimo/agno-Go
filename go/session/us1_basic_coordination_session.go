package session

import (
	"time"

	"github.com/agno-agi/agno-go/go/internal/telemetry"
	"github.com/agno-agi/agno-go/go/workflow"
)

// RunUS1Session constructs a minimal Session instance for the US1 workflow.
// At this stage it does not execute the full agent workflow; instead it
// records the input query and marks the session as completed with a
// placeholder result. The execution semantics will be refined in later
// iterations.
func RunUS1Session(query string, wf workflow.Workflow) Session {
	now := time.Now()

	s := Session{
		ID:       ID("us1-" + ID(now.Format("20060102150405"))),
		Workflow: wf.ID,
		Context: UserContext{
			UserID:    "",
			Channel:   "us1-basic-coordination",
			Locale:    "",
			Payload:   map[string]any{"query": query},
			StartedAt: now,
		},
		Status: StatusCompleted,
		Result: &Result{
			Success: true,
			Reason:  "placeholder",
			Data: map[string]any{
				"query": query,
			},
		},
		TraceID:   "",
		CreatedAt: now,
		UpdatedAt: now,
	}

	entry := HistoryEntry{
		Timestamp: now,
		Source:    "system",
		Message:   "US1 session placeholder run",
		Metadata: map[string]any{
			"step":  "initialized",
			"query": query,
		},
	}

	s.History = append(s.History, entry)

	// Emit a minimal telemetry event for the session run. In real workflows
	// this would be wired to a concrete Recorder implementation.
	ev := telemetry.Event{
		ID:        "us1-session-initialized",
		Timestamp: now,
		SessionID: string(s.ID),
		Type:      telemetry.EventRequestStarted,
		Payload: map[string]any{
			"workflow_id": string(wf.ID),
		},
	}
	var rec telemetry.Recorder = telemetry.NoopRecorder{}
	rec.Record(ev)

	return s
}
