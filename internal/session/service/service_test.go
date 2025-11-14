package service

import (
	"context"
	"testing"

	"github.com/rexleimo/agno-go/internal/session/dto"
	"github.com/rexleimo/agno-go/internal/session/store"
)

type stubStore struct {
	lastGetType  dto.SessionType
	lastRunsType dto.SessionType
}

func (s *stubStore) UpsertSession(ctx context.Context, record *dto.SessionRecord, preserveCreated bool) (*dto.SessionRecord, error) {
	return record, nil
}

func (s *stubStore) ListSessions(ctx context.Context, opts store.ListSessionsOptions) ([]*dto.SessionRecord, int, error) {
	return []*dto.SessionRecord{}, 0, nil
}

func (s *stubStore) GetSession(ctx context.Context, sessionID string, sessionType dto.SessionType) (*dto.SessionRecord, error) {
	s.lastGetType = sessionType
	return &dto.SessionRecord{SessionID: sessionID, SessionType: sessionType}, nil
}

func (s *stubStore) DeleteSession(ctx context.Context, sessionID string, sessionType dto.SessionType) error {
	return nil
}

func (s *stubStore) RenameSession(ctx context.Context, sessionID string, sessionType dto.SessionType, sessionName string) (*dto.SessionRecord, error) {
	return &dto.SessionRecord{SessionID: sessionID, SessionType: sessionType}, nil
}

func TestGetSessionRespectsSessionType(t *testing.T) {
	st := &stubStore{}
	svc, err := New(Config{Stores: map[string]store.Store{"default": st}, DefaultDB: "default"})
	if err != nil {
		t.Fatalf("failed to init service: %v", err)
	}

	req := GetSessionInput{SessionID: "sess-1", SessionType: dto.SessionTypeWorkflow}
	if _, err := svc.GetSession(context.Background(), req); err != nil {
		t.Fatalf("GetSession returned error: %v", err)
	}

	if st.lastGetType != dto.SessionTypeWorkflow {
		t.Fatalf("expected store to receive workflow type, got %s", st.lastGetType)
	}
}

func TestGetSessionRunsRespectsSessionType(t *testing.T) {
	st := &stubStore{}
	svc, err := New(Config{Stores: map[string]store.Store{"default": st}, DefaultDB: "default"})
	if err != nil {
		t.Fatalf("failed to init service: %v", err)
	}

	input := GetSessionInput{SessionID: "sess-runs", SessionType: dto.SessionTypeTeam}
	if _, err := svc.GetSessionRuns(context.Background(), input); err != nil {
		t.Fatalf("GetSessionRuns returned error: %v", err)
	}

	if st.lastGetType != dto.SessionTypeTeam {
		t.Fatalf("expected store to use team session type, got %s", st.lastGetType)
	}
}
