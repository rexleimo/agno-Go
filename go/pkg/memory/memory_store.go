package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rexleimo/agno-go/internal/agent"
)

// InMemoryStore provides a threadsafe in-memory implementation of the memory.Store interface.
type InMemoryStore struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]map[uuid.UUID]*sessionState
}

type sessionState struct {
	messages    []agent.Message
	toolResults map[string]agent.ToolCallResult
	updatedAt   time.Time
}

// Ensure InMemoryStore satisfies the storage contract.
var _ agent.Store = (*InMemoryStore)(nil)

// NewInMemoryStore constructs a new empty store.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		sessions: make(map[uuid.UUID]map[uuid.UUID]*sessionState),
	}
}

// UpsertSession creates the session bucket if missing.
func (s *InMemoryStore) UpsertSession(ctx context.Context, agentID, sessionID uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.sessions[agentID]; !ok {
		s.sessions[agentID] = make(map[uuid.UUID]*sessionState)
	}
	if _, ok := s.sessions[agentID][sessionID]; !ok {
		s.sessions[agentID][sessionID] = &sessionState{
			messages:    []agent.Message{},
			toolResults: map[string]agent.ToolCallResult{},
			updatedAt:   time.Now(),
		}
	}
	return nil
}

// AppendMessage appends a message to history and merges any pending tool results.
func (s *InMemoryStore) AppendMessage(ctx context.Context, agentID, sessionID uuid.UUID, msg agent.Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := s.getSessionLocked(agentID, sessionID)
	if err != nil {
		return err
	}

	clone := cloneMessage(msg)
	for i, tc := range clone.ToolCalls {
		if res, ok := state.toolResults[tc.ToolCallID]; ok {
			r := cloneResult(res)
			clone.ToolCalls[i].Result = &r
		}
	}
	state.messages = append(state.messages, clone)
	state.updatedAt = time.Now()
	return nil
}

// AppendToolResult records tool output and attaches it to the latest matching tool call if present.
func (s *InMemoryStore) AppendToolResult(ctx context.Context, agentID, sessionID uuid.UUID, result agent.ToolCallResult) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if result.ToolCallID == "" {
		return errors.New("tool call id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := s.getSessionLocked(agentID, sessionID)
	if err != nil {
		return err
	}
	state.toolResults[result.ToolCallID] = cloneResult(result)
	// Attempt to update the most recent matching tool call in history.
	for idx := len(state.messages) - 1; idx >= 0; idx-- {
		for tcIdx, tc := range state.messages[idx].ToolCalls {
			if tc.ToolCallID == result.ToolCallID {
				r := cloneResult(result)
				state.messages[idx].ToolCalls[tcIdx].Result = &r
				break
			}
		}
	}
	state.updatedAt = time.Now()
	return nil
}

// LoadHistory returns copies of session messages honoring the token window.
func (s *InMemoryStore) LoadHistory(ctx context.Context, agentID, sessionID uuid.UUID, opts agent.HistoryOptions) ([]agent.Message, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	state, err := s.getSessionLocked(agentID, sessionID)
	if err != nil {
		s.mu.RUnlock()
		return nil, err
	}
	// Copy under lock to avoid races.
	messages := make([]agent.Message, len(state.messages))
	for i, msg := range state.messages {
		messages[i] = cloneMessage(msg)
	}
	s.mu.RUnlock()

	return applyTokenWindow(messages, opts.TokenWindow), nil
}

// DeleteSession removes session state entirely.
func (s *InMemoryStore) DeleteSession(ctx context.Context, agentID, sessionID uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.sessions[agentID]; !ok {
		return agent.ErrSessionNotFound
	}
	if _, ok := s.sessions[agentID][sessionID]; !ok {
		return agent.ErrSessionNotFound
	}
	delete(s.sessions[agentID], sessionID)
	if len(s.sessions[agentID]) == 0 {
		delete(s.sessions, agentID)
	}
	return nil
}

func (s *InMemoryStore) getSessionLocked(agentID, sessionID uuid.UUID) (*sessionState, error) {
	agentSessions, ok := s.sessions[agentID]
	if !ok {
		return nil, agent.ErrSessionNotFound
	}
	state, ok := agentSessions[sessionID]
	if !ok {
		return nil, agent.ErrSessionNotFound
	}
	return state, nil
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
			tokens = 1 // fallback to avoid dropping messages with unknown usage
		}
		if total+tokens > tokenWindow && len(window) > 0 {
			break
		}
		total += tokens
		window = append(window, msg)
	}
	// Reverse to chronological order.
	for i, j := 0, len(window)-1; i < j; i, j = i+1, j-1 {
		window[i], window[j] = window[j], window[i]
	}
	return window
}
