package session

import (
	"time"

	"github.com/agno-agi/agno-go/go/workflow"
)

// ID uniquely identifies a session or task instance.
type ID string

// Status indicates the lifecycle state of a session.
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

// UserContext captures initial user input and metadata relevant to a session.
type UserContext struct {
	UserID    string
	Channel   string
	Locale    string
	Payload   map[string]any
	StartedAt time.Time
}

// HistoryEntry records a single interaction or event within the session.
type HistoryEntry struct {
	Timestamp time.Time
	Source    string
	Message   string
	Metadata  map[string]any
}

// Result captures the final outcome of a session in a generic form.
type Result struct {
	Success bool
	Reason  string
	Data    map[string]any
}

// Session represents a single run of a workflow for a given user context.
type Session struct {
	ID        ID
	Workflow  workflow.ID
	Context   UserContext
	History   []HistoryEntry
	Status    Status
	Result    *Result
	TraceID   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
