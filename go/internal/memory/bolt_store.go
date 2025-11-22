package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"

	"github.com/rexleimo/agno-go/internal/agent"
)

// BoltStore backs MemoryStore with a bbolt database for simple persistence.
type BoltStore struct {
	db   *bolt.DB
	path string
	mu   sync.Mutex
}

type storedSession struct {
	Messages    []agent.Message                 `json:"messages"`
	ToolResults map[string]agent.ToolCallResult `json:"toolResults"`
	UpdatedAt   time.Time                       `json:"updatedAt"`
}

// NewBoltStore opens (or creates) a bolt database at path.
func NewBoltStore(path string) (*BoltStore, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create bolt dir: %w", err)
	}
	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, fmt.Errorf("open bolt db: %w", err)
	}
	return &BoltStore{db: db, path: path}, nil
}

// Close releases the underlying db handle.
func (s *BoltStore) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *BoltStore) UpsertSession(ctx context.Context, agentID, sessionID uuid.UUID) error {
	return s.updateSession(agentID, sessionID, func(state *storedSession) error {
		// no-op; creation handled by updateSession
		return nil
	})
}

func (s *BoltStore) AppendMessage(ctx context.Context, agentID, sessionID uuid.UUID, msg agent.Message) error {
	return s.updateSession(agentID, sessionID, func(state *storedSession) error {
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

func (s *BoltStore) AppendToolResult(ctx context.Context, agentID, sessionID uuid.UUID, result agent.ToolCallResult) error {
	if result.ToolCallID == "" {
		return errors.New("tool call id is required")
	}
	return s.updateSession(agentID, sessionID, func(state *storedSession) error {
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

func (s *BoltStore) LoadHistory(ctx context.Context, agentID, sessionID uuid.UUID, opts agent.HistoryOptions) ([]agent.Message, error) {
	state, err := s.readSession(agentID, sessionID)
	if err != nil {
		return nil, err
	}
	return applyTokenWindow(state.Messages, opts.TokenWindow), nil
}

func (s *BoltStore) DeleteSession(ctx context.Context, agentID, sessionID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Update(func(tx *bolt.Tx) error {
		agentBucket := tx.Bucket([]byte(agentID.String()))
		if agentBucket == nil {
			return agent.ErrSessionNotFound
		}
		if agentBucket.Get([]byte(sessionID.String())) == nil {
			return agent.ErrSessionNotFound
		}
		if err := agentBucket.Delete([]byte(sessionID.String())); err != nil {
			return err
		}
		return nil
	})
}

func (s *BoltStore) readSession(agentID, sessionID uuid.UUID) (*storedSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var payload []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		agentBucket := tx.Bucket([]byte(agentID.String()))
		if agentBucket == nil {
			return agent.ErrSessionNotFound
		}
		val := agentBucket.Get([]byte(sessionID.String()))
		if val == nil {
			return agent.ErrSessionNotFound
		}
		payload = append([]byte(nil), val...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	var state storedSession
	if err := json.Unmarshal(payload, &state); err != nil {
		return nil, err
	}
	if state.ToolResults == nil {
		state.ToolResults = make(map[string]agent.ToolCallResult)
	}
	return &state, nil
}

func (s *BoltStore) updateSession(agentID, sessionID uuid.UUID, fn func(*storedSession) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Update(func(tx *bolt.Tx) error {
		agentBucket, err := tx.CreateBucketIfNotExists([]byte(agentID.String()))
		if err != nil {
			return err
		}
		raw := agentBucket.Get([]byte(sessionID.String()))
		state := storedSession{
			Messages:    []agent.Message{},
			ToolResults: make(map[string]agent.ToolCallResult),
		}
		if raw != nil {
			if err := json.Unmarshal(raw, &state); err != nil {
				return err
			}
			if state.ToolResults == nil {
				state.ToolResults = make(map[string]agent.ToolCallResult)
			}
		}
		if err := fn(&state); err != nil {
			return err
		}
		encoded, err := json.Marshal(state)
		if err != nil {
			return err
		}
		return agentBucket.Put([]byte(sessionID.String()), encoded)
	})
}

func cloneMessage(msg agent.Message) agent.Message {
	copyMsg := msg
	copyMsg.ToolCalls = make([]agent.ToolCall, len(msg.ToolCalls))
	for i, tc := range msg.ToolCalls {
		copyMsg.ToolCalls[i] = cloneToolCall(tc)
	}
	return copyMsg
}

func cloneToolCall(tc agent.ToolCall) agent.ToolCall {
	clone := tc
	clone.Args = cloneArgs(tc.Args)
	if tc.Result != nil {
		r := cloneResult(*tc.Result)
		clone.Result = &r
	}
	return clone
}

func cloneArgs(args map[string]any) map[string]any {
	if len(args) == 0 {
		return nil
	}
	out := make(map[string]any, len(args))
	for k, v := range args {
		out[k] = v
	}
	return out
}

func cloneResult(res agent.ToolCallResult) agent.ToolCallResult {
	return res
}

func applyTokenWindow(msgs []agent.Message, tokenWindow int) []agent.Message {
	if tokenWindow <= 0 || len(msgs) == 0 {
		return msgs
	}
	total := 0
	window := make([]agent.Message, 0, len(msgs))
	for i := len(msgs) - 1; i >= 0; i-- {
		msg := msgs[i]
		tokens := msg.Usage.PromptTokens + msg.Usage.CompletionTokens
		if tokens == 0 {
			tokens = 1
		}
		if total+tokens > tokenWindow && len(window) > 0 {
			break
		}
		total += tokens
		window = append(window, msg)
	}
	for i, j := 0, len(window)-1; i < j; i, j = i+1, j-1 {
		window[i], window[j] = window[j], window[i]
	}
	return window
}
