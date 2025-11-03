package dto

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// SessionType defines the supported session categories.
type SessionType string

const (
	SessionTypeAgent    SessionType = "agent"
	SessionTypeTeam     SessionType = "team"
	SessionTypeWorkflow SessionType = "workflow"
)

var (
	// ErrInvalidSessionType is returned when a provided session type is not supported.
	ErrInvalidSessionType = errors.New("invalid session type")
)

// ParseSessionType converts the provided string into a SessionType value.
// The input is matched in a case-insensitive manner. When the value is empty
// the function defaults to the agent session type to preserve backwards
// compatibility with the Python service which defaults to agent sessions.
func ParseSessionType(value string) (SessionType, error) {
	if value == "" {
		return SessionTypeAgent, nil
	}
	switch strings.ToLower(value) {
	case string(SessionTypeAgent):
		return SessionTypeAgent, nil
	case string(SessionTypeTeam):
		return SessionTypeTeam, nil
	case string(SessionTypeWorkflow):
		return SessionTypeWorkflow, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrInvalidSessionType, value)
	}
}

// Validate ensures the session type represents one of the supported values.
func (t SessionType) Validate() error {
	switch t {
	case SessionTypeAgent, SessionTypeTeam, SessionTypeWorkflow:
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrInvalidSessionType, string(t))
	}
}

// ComponentColumn returns the database column name that should be used when a
// component_id filter is supplied. The Python implementation maps the
// component identifier based on the session type and the Go implementation
// mirrors that behaviour.
func (t SessionType) ComponentColumn() (string, error) {
	switch t {
	case SessionTypeAgent:
		return "agent_id", nil
	case SessionTypeTeam:
		return "team_id", nil
	case SessionTypeWorkflow:
		return "workflow_id", nil
	default:
		return "", fmt.Errorf("%w: %s", ErrInvalidSessionType, string(t))
	}
}

// SessionRecord models the raw session representation stored in Postgres. It
// mirrors the Python AgentOS schema and is used across the store and service
// layers. JSON fields are decoded into generic maps so the HTTP layer can
// serialise the responses without losing optional fields.
type SessionRecord struct {
	SessionID    string
	SessionType  SessionType
	AgentID      *string
	TeamID       *string
	WorkflowID   *string
	UserID       *string
	SessionData  map[string]any
	AgentData    map[string]any
	TeamData     map[string]any
	WorkflowData map[string]any
	Metadata     map[string]any
	Runs         []map[string]any
	Summary      map[string]any
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Clone performs a deep copy of the record so mutations in upper layers do not
// affect cached values within the store.
func (r *SessionRecord) Clone() (*SessionRecord, error) {
	if r == nil {
		return nil, nil
	}
	copyRecord := *r
	var err error
	copyRecord.SessionData, err = cloneMap(r.SessionData)
	if err != nil {
		return nil, err
	}
	copyRecord.AgentData, err = cloneMap(r.AgentData)
	if err != nil {
		return nil, err
	}
	copyRecord.TeamData, err = cloneMap(r.TeamData)
	if err != nil {
		return nil, err
	}
	copyRecord.WorkflowData, err = cloneMap(r.WorkflowData)
	if err != nil {
		return nil, err
	}
	copyRecord.Metadata, err = cloneMap(r.Metadata)
	if err != nil {
		return nil, err
	}
	copyRecord.Summary, err = cloneMap(r.Summary)
	if err != nil {
		return nil, err
	}
	if r.Runs != nil {
		copyRecord.Runs = make([]map[string]any, len(r.Runs))
		for i, run := range r.Runs {
			copyRecord.Runs[i], err = cloneMap(run)
			if err != nil {
				return nil, err
			}
		}
	}
	return &copyRecord, nil
}

// SessionName returns the human friendly name stored within the session data.
func (r *SessionRecord) SessionName() string {
	if r == nil || r.SessionData == nil {
		return ""
	}
	if value, ok := r.SessionData["session_name"].(string); ok {
		return value
	}
	return ""
}

// SessionState extracts the session_state map from the session data.
func (r *SessionRecord) SessionState() map[string]any {
	if r == nil || r.SessionData == nil {
		return nil
	}
	if value, ok := r.SessionData["session_state"].(map[string]any); ok {
		return value
	}
	return nil
}

func cloneMap(src map[string]any) (map[string]any, error) {
	if src == nil {
		return nil, nil
	}
	buffer, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	var dst map[string]any
	if err := json.Unmarshal(buffer, &dst); err != nil {
		return nil, err
	}
	return dst, nil
}
