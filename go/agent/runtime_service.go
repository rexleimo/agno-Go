package agent

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/agno-agi/agno-go/go/session"
)

// RuntimeService manages AgentRuntime registrations and coordinates
// session-policy side effects against the configured session store. It mirrors
// the responsibilities of the Python AgentRuntime service without pulling in
// Python dependencies.
type RuntimeService struct {
	store    session.Store
	runtimes map[ID]*AgentRuntime
	mu       sync.RWMutex
	now      func() time.Time
}

// RuntimeServiceOption allows callers to customize the service.
type RuntimeServiceOption func(*RuntimeService)

// WithClock overrides the clock used for setting timestamps. Primarily useful
// in tests.
func WithClock(clock func() time.Time) RuntimeServiceOption {
	return func(s *RuntimeService) {
		s.now = clock
	}
}

// NewRuntimeService constructs a RuntimeService. The provided store may be nil,
// in which case the default in-memory implementation is used.
func NewRuntimeService(store session.Store, opts ...RuntimeServiceOption) *RuntimeService {
	if store == nil {
		store = session.NewMemoryStore()
	}
	rs := &RuntimeService{
		store:    store,
		runtimes: map[ID]*AgentRuntime{},
		now:      time.Now,
	}
	for _, opt := range opts {
		opt(rs)
	}
	return rs
}

// RegisterAgentRuntime validates and registers (or updates) an AgentRuntime.
// The returned pointer is a copy of the stored runtime to avoid accidental
// mutation by callers.
func (s *RuntimeService) RegisterAgentRuntime(ctx context.Context, cfg AgentRuntime) (*AgentRuntime, error) {
	runtime, err := NewAgentRuntime(cfg)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.runtimes[runtime.ID] = runtime
	s.mu.Unlock()

	if err := s.applySessionPolicy(ctx, runtime); err != nil {
		return nil, err
	}
	return cloneRuntime(runtime), nil
}

// AgentRuntime returns a copy of the registered runtime with the given ID.
func (s *RuntimeService) AgentRuntime(id ID) (*AgentRuntime, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rt, ok := s.runtimes[id]
	if !ok {
		return nil, false
	}
	return cloneRuntime(rt), true
}

// AgentRuntimes returns a copy of all registered runtimes.
func (s *RuntimeService) AgentRuntimes() []AgentRuntime {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]AgentRuntime, 0, len(s.runtimes))
	for _, rt := range s.runtimes {
		result = append(result, *cloneRuntime(rt))
	}
	return result
}

func (s *RuntimeService) applySessionPolicy(ctx context.Context, runtime *AgentRuntime) error {
	if s.store == nil {
		return nil
	}
	sessionID := strings.TrimSpace(runtime.SessionPolicy.SessionID)
	if sessionID == "" {
		return nil
	}

	id := session.ID(sessionID)
	record, err := s.store.Get(ctx, id)
	switch {
	case err == nil:
		// existing record will be updated below
	case err == session.ErrNotFound:
		record = &session.Session{
			ID:        id,
			Status:    session.StatusPending,
			Context:   session.UserContext{Payload: map[string]any{}},
			CreatedAt: s.now(),
		}
	default:
		return err
	}

	if record.Context.Payload == nil {
		record.Context.Payload = map[string]any{}
	}
	record.Context.Payload["session_policy"] = map[string]any{
		"overwrite_db_session_state": runtime.SessionPolicy.OverwriteDBSessionState,
		"enable_agentic_state":       runtime.SessionPolicy.EnableAgenticState,
		"cache_session":              runtime.SessionPolicy.CacheSession,
	}
	record.UpdatedAt = s.now()
	return s.store.Save(ctx, record)
}

func cloneRuntime(src *AgentRuntime) *AgentRuntime {
	if src == nil {
		return nil
	}
	copy := *src
	copy.Toolkits = append([]ToolkitRef(nil), src.Toolkits...)
	copy.Hooks = cloneStringMap(src.Hooks)
	copy.Metadata = cloneStringMap(src.Metadata)
	return &copy
}

func cloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
