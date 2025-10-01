package session

import (
	"time"

	"github.com/yourusername/agno-go/pkg/agno/agent"
)

// Session represents a conversation session with an agent
type Session struct {
	// Unique session identifier
	SessionID string `json:"session_id"`

	// Agent/Team/Workflow identifiers
	AgentID    string `json:"agent_id,omitempty"`
	TeamID     string `json:"team_id,omitempty"`
	WorkflowID string `json:"workflow_id,omitempty"`

	// User identifier
	UserID string `json:"user_id,omitempty"`

	// Session metadata
	Name     string                 `json:"name,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Session state
	State map[string]interface{} `json:"state,omitempty"`

	// Agent data (for reference)
	AgentData map[string]interface{} `json:"agent_data,omitempty"`

	// Run history
	Runs []*agent.RunOutput `json:"runs,omitempty"`

	// Summary (for long sessions)
	Summary *SessionSummary `json:"summary,omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SessionSummary contains a summary of a session
type SessionSummary struct {
	// Summary text
	Content string `json:"content"`

	// Number of runs in this session
	RunCount int `json:"run_count"`

	// Total tokens used
	TotalTokens int `json:"total_tokens"`

	// Created at timestamp
	CreatedAt time.Time `json:"created_at"`
}

// NewSession creates a new session
func NewSession(sessionID string, agentID string) *Session {
	now := time.Now()
	return &Session{
		SessionID: sessionID,
		AgentID:   agentID,
		Runs:      make([]*agent.RunOutput, 0),
		Metadata:  make(map[string]interface{}),
		State:     make(map[string]interface{}),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AddRun adds a run output to the session
func (s *Session) AddRun(run *agent.RunOutput) {
	// Simply append the new run (no deduplication since RunOutput doesn't have ID)
	s.Runs = append(s.Runs, run)
	s.UpdatedAt = time.Now()
}

// GetRunCount returns the number of runs in this session
func (s *Session) GetRunCount() int {
	return len(s.Runs)
}

// GetLastRun returns the most recent run
func (s *Session) GetLastRun() *agent.RunOutput {
	if len(s.Runs) == 0 {
		return nil
	}
	return s.Runs[len(s.Runs)-1]
}

// CalculateTotalTokens calculates total tokens used across all runs
// Note: Currently RunOutput doesn't track metrics, so this returns 0
// This is a placeholder for future enhancements
func (s *Session) CalculateTotalTokens() int {
	// TODO: When RunOutput gains Metrics field, calculate actual tokens
	return 0
}

// GenerateSummary creates a summary of the session
func (s *Session) GenerateSummary(content string) {
	s.Summary = &SessionSummary{
		Content:     content,
		RunCount:    len(s.Runs),
		TotalTokens: s.CalculateTotalTokens(),
		CreatedAt:   time.Now(),
	}
	s.UpdatedAt = time.Now()
}
