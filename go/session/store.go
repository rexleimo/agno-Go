package session

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/agno-agi/agno-go/go/workflow"
)

// ErrNotFound indicates the requested session record does not exist.
var ErrNotFound = errors.New("session: record not found")

// ErrDriverNotRegistered indicates that a requested driver has no registered factory.
var ErrDriverNotRegistered = errors.New("session: store driver not registered")

// Driver enumerates the supported backing implementations for the Store interface.
type Driver string

const (
	DriverMemory   Driver = "memory"
	DriverSQLite   Driver = "sqlite"
	DriverPostgres Driver = "postgres"
	DriverRedis    Driver = "redis"
)

// Options describes generic configuration knobs passed to driver factories.
type Options struct {
	// DSN holds a driver-specific connection string (for SQL/Redis backends).
	DSN string
	// Namespace scopes keys or tables, enabling multi-tenant installs.
	Namespace string
	// Extra carries driver-specific configuration and is intentionally opaque.
	Extra map[string]any
}

// Store exposes CRUD + search semantics similar to Python AgentSession storage.
type Store interface {
	Save(ctx context.Context, record *Session) error
	Get(ctx context.Context, id ID) (*Session, error)
	Delete(ctx context.Context, id ID) error
	Search(ctx context.Context, filter SearchFilter) ([]*Session, error)
}

// SearchFilter constrains the search space.
type SearchFilter struct {
	UserID   string
	Workflow workflow.ID
	Statuses []Status
	Since    time.Time
	Limit    int
}

// Factory builds a Store backed by the desired driver (Sqlite/Postgres/Redis/etc).
type Factory func(opts Options) (Store, error)

type driverRegistry struct {
	mu        sync.RWMutex
	factories map[Driver]Factory
}

var defaultRegistry = &driverRegistry{
	factories: map[Driver]Factory{},
}

// RegisterDriver registers a new Store factory for the given driver.
func RegisterDriver(driver Driver, factory Factory) {
	defaultRegistry.mu.Lock()
	defer defaultRegistry.mu.Unlock()
	defaultRegistry.factories[driver] = factory
}

// RegisterSQLite allows consumers to plug their Sqlite implementation.
func RegisterSQLite(factory Factory) {
	RegisterDriver(DriverSQLite, factory)
}

// RegisterPostgres allows consumers to plug their Postgres implementation.
func RegisterPostgres(factory Factory) {
	RegisterDriver(DriverPostgres, factory)
}

// RegisterRedis allows consumers to plug their Redis implementation.
func RegisterRedis(factory Factory) {
	RegisterDriver(DriverRedis, factory)
}

// NewStore constructs a Store given a registered driver.
func NewStore(driver Driver, opts Options) (Store, error) {
	defaultRegistry.mu.RLock()
	factory, ok := defaultRegistry.factories[driver]
	defaultRegistry.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrDriverNotRegistered, driver)
	}
	return factory(opts)
}

func init() {
	RegisterDriver(DriverMemory, func(Options) (Store, error) {
		return NewMemoryStore(), nil
	})
}

// MemoryStore is an in-memory Store implementation designed for tests.
type MemoryStore struct {
	mu      sync.RWMutex
	records map[ID]*Session
}

// NewMemoryStore constructs a MemoryStore with no initial records.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		records: map[ID]*Session{},
	}
}

// Save inserts or updates the provided session record.
func (m *MemoryStore) Save(ctx context.Context, record *Session) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if record == nil {
		return errors.New("session: record cannot be nil")
	}
	if record.ID == "" {
		return errors.New("session: record must have an ID")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records[record.ID] = cloneSession(record)
	return nil
}

// Get retrieves a session record by ID.
func (m *MemoryStore) Get(ctx context.Context, id ID) (*Session, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	rec, ok := m.records[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneSession(rec), nil
}

// Delete removes a session record.
func (m *MemoryStore) Delete(ctx context.Context, id ID) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.records[id]; !ok {
		return ErrNotFound
	}
	delete(m.records, id)
	return nil
}

// Search returns records matching the supplied filter.
func (m *MemoryStore) Search(ctx context.Context, filter SearchFilter) ([]*Session, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []*Session
	statusSet := map[Status]struct{}{}
	for _, st := range filter.Statuses {
		statusSet[st] = struct{}{}
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}

	for _, rec := range m.records {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if filter.UserID != "" && rec.Context.UserID != filter.UserID {
			continue
		}
		if filter.Workflow != "" && rec.Workflow != filter.Workflow {
			continue
		}
		if len(statusSet) > 0 {
			if _, ok := statusSet[rec.Status]; !ok {
				continue
			}
		}
		if !filter.Since.IsZero() && rec.UpdatedAt.Before(filter.Since) {
			continue
		}
		results = append(results, cloneSession(rec))
		if len(results) >= limit {
			break
		}
	}
	return results, nil
}

func cloneSession(src *Session) *Session {
	if src == nil {
		return nil
	}
	dst := *src
	dst.Context = UserContext{
		UserID:    src.Context.UserID,
		Channel:   src.Context.Channel,
		Locale:    src.Context.Locale,
		Payload:   cloneMap(src.Context.Payload),
		StartedAt: src.Context.StartedAt,
	}
	dst.History = cloneHistory(src.History)
	if src.Result != nil {
		dst.Result = &Result{
			Success: src.Result.Success,
			Reason:  src.Result.Reason,
			Data:    cloneMap(src.Result.Data),
		}
	}
	return &dst
}

func cloneHistory(entries []HistoryEntry) []HistoryEntry {
	if len(entries) == 0 {
		return nil
	}
	out := make([]HistoryEntry, len(entries))
	for i, entry := range entries {
		out[i] = HistoryEntry{
			Timestamp: entry.Timestamp,
			Source:    entry.Source,
			Message:   entry.Message,
			Metadata:  cloneMap(entry.Metadata),
		}
	}
	return out
}

func cloneMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
