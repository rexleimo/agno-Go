package agent

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// ErrSessionNotFound is returned when session lookups fail.
var ErrSessionNotFound = errors.New("session not found")

// HistoryOptions controls how much history to return, including token window limits.
type HistoryOptions struct {
	// TokenWindow trims history to approximately this many tokens (prompt+completion).
	// A value <=0 returns full history.
	TokenWindow int
}

// Store defines the contract for persisting session state.
// Implementations may be in-memory or embedded KV (e.g., Bolt/Badger) but must remain pure Go.
type Store interface {
	// UpsertSession initializes or ensures the session namespace exists.
	UpsertSession(ctx context.Context, agentID, sessionID uuid.UUID) error
	// AppendMessage appends a chat message to session history.
	AppendMessage(ctx context.Context, agentID, sessionID uuid.UUID, msg Message) error
	// AppendToolResult records the outcome of a tool call and associates it with history if present.
	AppendToolResult(ctx context.Context, agentID, sessionID uuid.UUID, result ToolCallResult) error
	// LoadHistory returns session messages respecting the token window.
	LoadHistory(ctx context.Context, agentID, sessionID uuid.UUID, opts HistoryOptions) ([]Message, error)
	// DeleteSession removes all state for the given session.
	DeleteSession(ctx context.Context, agentID, sessionID uuid.UUID) error
}
