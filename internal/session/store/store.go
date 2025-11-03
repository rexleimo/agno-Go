package store

import (
	"context"
	"errors"

	"github.com/rexleimo/agno-go/internal/session/dto"
)

// ErrNotFound indicates the store could not locate the requested session.
var ErrNotFound = errors.New("session not found")

// ListSessionsOptions is used to filter and paginate the sessions query.
type ListSessionsOptions struct {
	SessionType dto.SessionType
	UserID      string
	ComponentID string
	SessionName string
	SortBy      string
	SortOrder   string
	Limit       int
	Page        int
}

// Store describes the persistence operations required by the session service.
type Store interface {
	UpsertSession(ctx context.Context, record *dto.SessionRecord, preserveCreated bool) (*dto.SessionRecord, error)
	ListSessions(ctx context.Context, opts ListSessionsOptions) ([]*dto.SessionRecord, int, error)
	GetSession(ctx context.Context, sessionID string, sessionType dto.SessionType) (*dto.SessionRecord, error)
	DeleteSession(ctx context.Context, sessionID string, sessionType dto.SessionType) error
	RenameSession(ctx context.Context, sessionID string, sessionType dto.SessionType, sessionName string) (*dto.SessionRecord, error)
}
