package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/rexleimo/agno-go/internal/session/dto"
	"github.com/rexleimo/agno-go/internal/session/store"
)

var (
	// ErrNoStoresConfigured indicates the service was initialised without any stores.
	ErrNoStoresConfigured = errors.New("no session stores configured")
	// ErrDatabaseRequired indicates a db_id must be supplied when multiple stores are configured.
	ErrDatabaseRequired = errors.New("db_id query parameter is required when multiple databases are configured")
	// ErrDatabaseNotFound signals that the requested db_id does not match a configured store.
	ErrDatabaseNotFound = errors.New("database not found")
)

// Config defines the session service configuration.
type Config struct {
	Stores    map[string]store.Store
	DefaultDB string
}

// Service orchestrates session operations across storage backends.
type Service struct {
	stores        map[string]store.Store
	defaultStore  string
	configuredIDs []string
}

// New constructs a session service instance.
func New(cfg Config) (*Service, error) {
	if len(cfg.Stores) == 0 {
		return nil, ErrNoStoresConfigured
	}

	service := &Service{
		stores: make(map[string]store.Store, len(cfg.Stores)),
	}
	for id, st := range cfg.Stores {
		service.stores[id] = st
		service.configuredIDs = append(service.configuredIDs, id)
	}

	if cfg.DefaultDB != "" {
		if _, ok := service.stores[cfg.DefaultDB]; !ok {
			return nil, fmt.Errorf("default db_id %q not configured", cfg.DefaultDB)
		}
		service.defaultStore = cfg.DefaultDB
	} else if len(service.stores) == 1 {
		for id := range service.stores {
			service.defaultStore = id
		}
	}

	return service, nil
}

// ListSessionsInput describes the filters accepted by ListSessions.
type ListSessionsInput struct {
	SessionType dto.SessionType
	ComponentID string
	UserID      string
	SessionName string
	SortBy      string
	SortOrder   string
	Limit       int
	Page        int
	DatabaseID  string
}

// SessionListItem mirrors the Python SessionSchema response.
type SessionListItem struct {
	SessionID    string         `json:"session_id"`
	SessionName  string         `json:"session_name"`
	SessionState map[string]any `json:"session_state"`
	CreatedAt    string         `json:"created_at"`
	UpdatedAt    string         `json:"updated_at"`
}

// PaginationMeta matches the Python pagination payload.
type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

// ListSessionsOutput wraps the paginated session list response.
type ListSessionsOutput struct {
	Data []SessionListItem `json:"data"`
	Meta PaginationMeta    `json:"meta"`
}

// GetSessionInput captures the parameters for retrieving a single session.
type GetSessionInput struct {
	SessionID   string
	SessionType dto.SessionType
	DatabaseID  string
}

// SessionDetail holds the detailed payload returned for a single session.
type SessionDetail struct {
	Type dto.SessionType
	Data map[string]any
}

// CreateSessionInput defines the fields accepted when creating/upserting a session.
type CreateSessionInput struct {
	SessionID    string
	SessionType  dto.SessionType
	SessionName  string
	SessionState map[string]any
	Metadata     map[string]any
	UserID       string
	AgentID      string
	TeamID       string
	WorkflowID   string
	AgentData    map[string]any
	TeamData     map[string]any
	WorkflowData map[string]any
	Runs         []map[string]any
	Summary      map[string]any
	DatabaseID   string
}

// RenameSessionInput captures the parameters for renaming a session.
type RenameSessionInput struct {
	SessionID   string
	SessionType dto.SessionType
	SessionName string
	DatabaseID  string
}

// DeleteSessionInput captures the parameters for deleting a session.
type DeleteSessionInput struct {
	SessionID   string
	SessionType dto.SessionType
	DatabaseID  string
}

// ListSessions returns paginated session summaries matching the provided filters.
func (s *Service) ListSessions(ctx context.Context, input ListSessionsInput) (*ListSessionsOutput, error) {
	storeImpl, _, err := s.resolveStore(input.DatabaseID)
	if err != nil {
		return nil, err
	}

	sessionType := input.SessionType
	if sessionType == "" {
		sessionType = dto.SessionTypeAgent
	}

	if err := sessionType.Validate(); err != nil {
		return nil, err
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 20
	}
	page := input.Page
	if page <= 0 {
		page = 1
	}

	records, totalCount, err := storeImpl.ListSessions(ctx, store.ListSessionsOptions{
		SessionType: sessionType,
		ComponentID: input.ComponentID,
		UserID:      input.UserID,
		SessionName: input.SessionName,
		SortBy:      input.SortBy,
		SortOrder:   input.SortOrder,
		Limit:       limit,
		Page:        page,
	})
	if err != nil {
		return nil, err
	}

	items := make([]SessionListItem, 0, len(records))
	for _, record := range records {
		item := toListItem(record)
		items = append(items, item)
	}

	output := &ListSessionsOutput{
		Data: items,
		Meta: PaginationMeta{
			Page:       page,
			Limit:      limit,
			TotalCount: totalCount,
			TotalPages: computeTotalPages(totalCount, limit),
		},
	}

	return output, nil
}

// GetSession retrieves the detailed session payload.
func (s *Service) GetSession(ctx context.Context, input GetSessionInput) (*SessionDetail, error) {
	storeImpl, _, err := s.resolveStore(input.DatabaseID)
	if err != nil {
		return nil, err
	}
	sessionType := input.SessionType
	if sessionType == "" {
		sessionType = dto.SessionTypeAgent
	}
	if err := sessionType.Validate(); err != nil {
		return nil, err
	}

	record, err := storeImpl.GetSession(ctx, input.SessionID, sessionType)
	if err != nil {
		return nil, err
	}

	detail := buildSessionDetail(record)
	return &SessionDetail{Type: sessionType, Data: detail}, nil
}

// CreateSession upserts a session record and returns the resulting detail.
func (s *Service) CreateSession(ctx context.Context, input CreateSessionInput) (*SessionDetail, error) {
	storeImpl, _, err := s.resolveStore(input.DatabaseID)
	if err != nil {
		return nil, err
	}

	sessionType := input.SessionType
	if sessionType == "" {
		sessionType = dto.SessionTypeAgent
	}
	if err := sessionType.Validate(); err != nil {
		return nil, err
	}

	sessionID := input.SessionID
	if sessionID == "" {
		sessionID = uuid.NewString()
	}

	record := &dto.SessionRecord{
		SessionID:    sessionID,
		SessionType:  sessionType,
		SessionData:  map[string]any{},
		Metadata:     input.Metadata,
		AgentData:    input.AgentData,
		TeamData:     input.TeamData,
		WorkflowData: input.WorkflowData,
		Runs:         input.Runs,
		Summary:      input.Summary,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if input.AgentID != "" {
		record.AgentID = ptr(input.AgentID)
	}
	if input.TeamID != "" {
		record.TeamID = ptr(input.TeamID)
	}
	if input.WorkflowID != "" {
		record.WorkflowID = ptr(input.WorkflowID)
	}
	if input.UserID != "" {
		record.UserID = ptr(input.UserID)
	}

	if input.SessionName != "" {
		record.SessionData["session_name"] = input.SessionName
	}
	if input.SessionState != nil {
		record.SessionData["session_state"] = input.SessionState
	}

	if record.Metadata == nil {
		record.Metadata = map[string]any{}
	}

	if record.SessionData == nil {
		record.SessionData = map[string]any{}
	}

	createdRecord, err := storeImpl.UpsertSession(ctx, record, false)
	if err != nil {
		return nil, err
	}

	detail := buildSessionDetail(createdRecord)
	return &SessionDetail{Type: sessionType, Data: detail}, nil
}

// RenameSession updates the session name and returns the detailed payload.
func (s *Service) RenameSession(ctx context.Context, input RenameSessionInput) (*SessionDetail, error) {
	storeImpl, _, err := s.resolveStore(input.DatabaseID)
	if err != nil {
		return nil, err
	}
	if err := input.SessionType.Validate(); err != nil {
		return nil, err
	}

	updated, err := storeImpl.RenameSession(ctx, input.SessionID, input.SessionType, input.SessionName)
	if err != nil {
		return nil, err
	}
	detail := buildSessionDetail(updated)
	return &SessionDetail{Type: input.SessionType, Data: detail}, nil
}

// DeleteSession removes the session from the configured store.
func (s *Service) DeleteSession(ctx context.Context, input DeleteSessionInput) error {
	storeImpl, _, err := s.resolveStore(input.DatabaseID)
	if err != nil {
		return err
	}
	if err := input.SessionType.Validate(); err != nil {
		return err
	}

	return storeImpl.DeleteSession(ctx, input.SessionID, input.SessionType)
}

// GetSessionRuns retrieves the run history for a session via the session record.
func (s *Service) GetSessionRuns(ctx context.Context, input GetSessionInput) ([]map[string]any, error) {
	storeImpl, _, err := s.resolveStore(input.DatabaseID)
	if err != nil {
		return nil, err
	}
	sessionType := input.SessionType
	if sessionType == "" {
		sessionType = dto.SessionTypeAgent
	}
	if err := sessionType.Validate(); err != nil {
		return nil, err
	}

	record, err := storeImpl.GetSession(ctx, input.SessionID, sessionType)
	if err != nil {
		return nil, err
	}

	if record.Runs == nil {
		return []map[string]any{}, nil
	}

	cloned := make([]map[string]any, len(record.Runs))
	for i, run := range record.Runs {
		clone := make(map[string]any, len(run))
		for k, v := range run {
			clone[k] = v
		}
		cloned[i] = clone
	}
	return cloned, nil
}

func (s *Service) resolveStore(dbID string) (store.Store, string, error) {
	if dbID != "" {
		st, ok := s.stores[dbID]
		if !ok {
			return nil, "", ErrDatabaseNotFound
		}
		return st, dbID, nil
	}

	if s.defaultStore != "" {
		return s.stores[s.defaultStore], s.defaultStore, nil
	}

	if len(s.stores) > 1 {
		return nil, "", ErrDatabaseRequired
	}

	for id, st := range s.stores {
		return st, id, nil
	}
	return nil, "", ErrNoStoresConfigured
}

func toListItem(record *dto.SessionRecord) SessionListItem {
	return SessionListItem{
		SessionID:    record.SessionID,
		SessionName:  record.SessionName(),
		SessionState: cloneState(record.SessionState()),
		CreatedAt:    record.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    record.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func computeTotalPages(totalCount, limit int) int {
	if limit <= 0 {
		return 0
	}
	return int(math.Ceil(float64(totalCount) / float64(limit)))
}

func buildSessionDetail(record *dto.SessionRecord) map[string]any {
	if record == nil {
		return map[string]any{}
	}

	base := map[string]any{
		"session_id":   record.SessionID,
		"session_name": record.SessionName(),
		"session_state": func() any {
			state := record.SessionState()
			if state == nil {
				return nil
			}
			return cloneState(state)
		}(),
		"created_at": record.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at": record.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if metadata := cloneMap(record.Metadata); metadata != nil {
		base["metadata"] = metadata
	}

	sessionType := record.SessionType
	switch sessionType {
	case dto.SessionTypeAgent:
		if record.UserID != nil {
			base["user_id"] = *record.UserID
		}
		if record.AgentID != nil {
			base["agent_id"] = *record.AgentID
		}
		base["agent_session_id"] = record.SessionID
		if data := cloneMap(record.AgentData); data != nil {
			base["agent_data"] = data
		}
		metrics := extractMetrics(record.SessionData)
		if metrics != nil {
			base["metrics"] = metrics
			if totalTokens, ok := metrics["total_tokens"]; ok {
				base["total_tokens"] = totalTokens
			}
		}
		if chatHistory, ok := extractChatHistory(record.SessionData); ok {
			base["chat_history"] = chatHistory
		}
		if record.Summary != nil {
			base["session_summary"] = cloneMap(record.Summary)
		}
	case dto.SessionTypeTeam:
		if record.TeamID != nil {
			base["team_id"] = *record.TeamID
		}
		if record.UserID != nil {
			base["user_id"] = *record.UserID
		}
		if data := cloneMap(record.TeamData); data != nil {
			base["team_data"] = data
		}
		metrics := extractMetrics(record.SessionData)
		if metrics != nil {
			base["metrics"] = metrics
			if totalTokens, ok := metrics["total_tokens"]; ok {
				base["total_tokens"] = totalTokens
			}
		}
		if chatHistory, ok := extractChatHistory(record.SessionData); ok {
			base["chat_history"] = chatHistory
		}
		if record.Summary != nil {
			base["session_summary"] = cloneMap(record.Summary)
		}
	case dto.SessionTypeWorkflow:
		if record.UserID != nil {
			base["user_id"] = *record.UserID
		}
		if record.WorkflowID != nil {
			base["workflow_id"] = *record.WorkflowID
		}
		if data := cloneMap(record.WorkflowData); data != nil {
			base["workflow_data"] = data
		}
		if sessionData := cloneMap(record.SessionData); sessionData != nil {
			base["session_data"] = sessionData
		}
	}

	return base
}

func extractMetrics(sessionData map[string]any) map[string]any {
	if sessionData == nil {
		return nil
	}
	if metrics, ok := sessionData["session_metrics"].(map[string]any); ok {
		return cloneMap(metrics)
	}
	return nil
}

func extractChatHistory(sessionData map[string]any) ([]any, bool) {
	if sessionData == nil {
		return nil, false
	}
	if history, ok := sessionData["chat_history"].([]any); ok {
		return cloneSlice(history), true
	}
	return nil, false
}

func cloneMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	buffer, err := json.Marshal(src)
	if err != nil {
		return nil
	}
	var clone map[string]any
	if err := json.Unmarshal(buffer, &clone); err != nil {
		return nil
	}
	return clone
}

func cloneState(src map[string]any) map[string]any {
	return cloneMap(src)
}

func cloneSlice(src []any) []any {
	buffer, err := json.Marshal(src)
	if err != nil {
		return nil
	}
	var clone []any
	if err := json.Unmarshal(buffer, &clone); err != nil {
		return nil
	}
	return clone
}

func ptr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
