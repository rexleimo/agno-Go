package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"

	"github.com/rexleimo/agno-go/internal/agent"
)

// BadgerStore backs MemoryStore with badger to provide persistence.
type BadgerStore struct {
	db        *badger.DB
	mu        sync.Mutex
	retention time.Duration
}

// NewBadgerStore opens a badger database rooted at dir. retention>0 applies TTL to entries.
func NewBadgerStore(dir string, retention time.Duration) (*BadgerStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create badger dir: %w", err)
	}
	opts := badger.DefaultOptions(dir).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("open badger: %w", err)
	}
	return &BadgerStore{db: db, retention: retention}, nil
}

// Close closes the underlying DB.
func (s *BadgerStore) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *BadgerStore) UpsertSession(ctx context.Context, agentID, sessionID uuid.UUID) error {
	return s.saveSession(agentID, sessionID, storedSession{
		Messages:    []agent.Message{},
		ToolResults: make(map[string]agent.ToolCallResult),
		UpdatedAt:   time.Now(),
	})
}

func (s *BadgerStore) AppendMessage(ctx context.Context, agentID, sessionID uuid.UUID, msg agent.Message) error {
	return s.update(agentID, sessionID, func(state *storedSession) error {
		clone := cloneMessage(msg)
		for i, tc := range clone.ToolCalls {
			if res, ok := state.ToolResults[tc.ToolCallID]; ok {
				r := cloneResult(res)
				clone.ToolCalls[i].Result = &r
			}
		}
		state.Messages = append(state.Messages, clone)
		state.UpdatedAt = time.Now()
		return nil
	})
}

func (s *BadgerStore) AppendToolResult(ctx context.Context, agentID, sessionID uuid.UUID, result agent.ToolCallResult) error {
	return s.update(agentID, sessionID, func(state *storedSession) error {
		if result.ToolCallID == "" {
			return fmt.Errorf("tool call id required")
		}
		state.ToolResults[result.ToolCallID] = cloneResult(result)
		for idx := len(state.Messages) - 1; idx >= 0; idx-- {
			for tcIdx, tc := range state.Messages[idx].ToolCalls {
				if tc.ToolCallID == result.ToolCallID {
					r := cloneResult(result)
					state.Messages[idx].ToolCalls[tcIdx].Result = &r
					break
				}
			}
		}
		state.UpdatedAt = time.Now()
		return nil
	})
}

func (s *BadgerStore) LoadHistory(ctx context.Context, agentID, sessionID uuid.UUID, opts agent.HistoryOptions) ([]agent.Message, error) {
	state, err := s.read(agentID, sessionID)
	if err != nil {
		return nil, err
	}
	return applyTokenWindow(state.Messages, opts.TokenWindow), nil
}

func (s *BadgerStore) DeleteSession(ctx context.Context, agentID, sessionID uuid.UUID) error {
	key := buildKey(agentID, sessionID)
	return s.db.Update(func(txn *badger.Txn) error {
		if err := txn.Delete([]byte(key)); err != nil {
			if err == badger.ErrKeyNotFound {
				return agent.ErrSessionNotFound
			}
			return err
		}
		return nil
	})
}

func (s *BadgerStore) update(agentID, sessionID uuid.UUID, fn func(*storedSession) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, err := s.read(agentID, sessionID)
	if err != nil {
		if err != agent.ErrSessionNotFound {
			return err
		}
		state = &storedSession{
			Messages:    []agent.Message{},
			ToolResults: make(map[string]agent.ToolCallResult),
			UpdatedAt:   time.Now(),
		}
	}
	if err := fn(state); err != nil {
		return err
	}
	return s.saveSession(agentID, sessionID, *state)
}

func (s *BadgerStore) saveSession(agentID, sessionID uuid.UUID, state storedSession) error {
	key := buildKey(agentID, sessionID)
	raw, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return s.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), raw)
		if s.retention > 0 {
			e = e.WithTTL(s.retention)
		}
		return txn.SetEntry(e)
	})
}

func (s *BadgerStore) read(agentID, sessionID uuid.UUID) (*storedSession, error) {
	key := buildKey(agentID, sessionID)
	var raw []byte
	if err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return agent.ErrSessionNotFound
			}
			return err
		}
		return item.Value(func(val []byte) error {
			raw = append([]byte(nil), val...)
			return nil
		})
	}); err != nil {
		return nil, err
	}

	var state storedSession
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, err
	}
	if state.ToolResults == nil {
		state.ToolResults = make(map[string]agent.ToolCallResult)
	}
	return &state, nil
}

func buildKey(agentID, sessionID uuid.UUID) string {
	return filepath.Join(agentID.String(), sessionID.String())
}
