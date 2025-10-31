package culture

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound 表示找不到指定的文化条目。
var ErrNotFound = errors.New("culture entry not found")

// Entry 表示一条文化知识记录。
type Entry struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Tags      []string               `json:"tags,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// Clone 返回条目的深拷贝，用于避免外部修改。
func (e *Entry) Clone() *Entry {
	if e == nil {
		return nil
	}

	clone := *e
	if len(e.Tags) > 0 {
		clone.Tags = append([]string(nil), e.Tags...)
	}
	if len(e.Metadata) > 0 {
		clone.Metadata = make(map[string]interface{}, len(e.Metadata))
		for k, v := range e.Metadata {
			clone.Metadata[k] = v
		}
	}
	return &clone
}

// Filter 用于筛选文化条目。
type Filter struct {
	Tags []string
}

// Store 定义文化存储接口，可接入不同后端（内存、数据库等）。
type Store interface {
	Create(ctx context.Context, entry *Entry) error
	Update(ctx context.Context, entry *Entry) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*Entry, error)
	List(ctx context.Context, filter Filter) ([]*Entry, error)
	Clear(ctx context.Context) error
}

// Result 表示异步操作的结果。
type Result struct {
	Entry *Entry
	Err   error
}

// Manager 管理文化条目，提供同步与异步 API。
type Manager struct {
	store       Store
	idGenerator func() string
	clock       func() time.Time
}

// Option 自定义 Manager 行为。
type Option func(*Manager)

// WithIDGenerator 指定自定义 ID 生成函数。
func WithIDGenerator(fn func() string) Option {
	return func(m *Manager) {
		if fn != nil {
			m.idGenerator = fn
		}
	}
}

// WithClock 指定时间来源，方便测试。
func WithClock(fn func() time.Time) Option {
	return func(m *Manager) {
		if fn != nil {
			m.clock = fn
		}
	}
}

// NewManager 创建文化管理器实例。
func NewManager(store Store, opts ...Option) (*Manager, error) {
	if store == nil {
		return nil, fmt.Errorf("culture store cannot be nil")
	}

	manager := &Manager{
		store:       store,
		idGenerator: uuid.NewString,
		clock:       time.Now,
	}

	for _, opt := range opts {
		opt(manager)
	}

	return manager, nil
}

// AddKnowledge 新增文化条目。
func (m *Manager) AddKnowledge(ctx context.Context, entry *Entry) (*Entry, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, fmt.Errorf("entry cannot be nil")
	}
	if strings.TrimSpace(entry.Content) == "" {
		return nil, fmt.Errorf("entry content cannot be empty")
	}

	now := m.clock().UTC()
	newEntry := entry.Clone()
	if newEntry.ID == "" {
		newEntry.ID = m.idGenerator()
	}
	newEntry.CreatedAt = now
	newEntry.UpdatedAt = now
	normaliseEntry(newEntry)

	if err := m.store.Create(ctx, newEntry); err != nil {
		return nil, err
	}
	return newEntry.Clone(), nil
}

// AddKnowledgeAsync 异步新增文化条目，返回结果通道。
func (m *Manager) AddKnowledgeAsync(ctx context.Context, entry *Entry) <-chan Result {
	resultCh := make(chan Result, 1)
	go func() {
		defer close(resultCh)
		newEntry, err := m.AddKnowledge(ctx, entry)
		resultCh <- Result{Entry: newEntry, Err: err}
	}()
	return resultCh
}

// UpdateKnowledge 更新文化条目。
func (m *Manager) UpdateKnowledge(ctx context.Context, entry *Entry) (*Entry, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}
	if entry == nil || entry.ID == "" {
		return nil, fmt.Errorf("entry ID is required")
	}

	existing, err := m.store.Get(ctx, entry.ID)
	if err != nil {
		return nil, err
	}

	updated := existing.Clone()
	if strings.TrimSpace(entry.Content) != "" {
		updated.Content = entry.Content
	}
	if entry.Tags != nil {
		updated.Tags = append([]string(nil), entry.Tags...)
	}
	if entry.Metadata != nil {
		updated.Metadata = make(map[string]interface{}, len(entry.Metadata))
		for k, v := range entry.Metadata {
			updated.Metadata[k] = v
		}
	}
	updated.UpdatedAt = m.clock().UTC()
	normaliseEntry(updated)

	if err := m.store.Update(ctx, updated); err != nil {
		return nil, err
	}
	return updated.Clone(), nil
}

// ListKnowledge 按筛选条件列出条目。
func (m *Manager) ListKnowledge(ctx context.Context, filter Filter) ([]*Entry, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}
	entries, err := m.store.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	results := make([]*Entry, 0, len(entries))
	for _, e := range entries {
		results = append(results, e.Clone())
	}
	return results, nil
}

// GetKnowledge 根据 ID 获取条目。
func (m *Manager) GetKnowledge(ctx context.Context, id string) (*Entry, error) {
	if err := ensureContext(ctx); err != nil {
		return nil, err
	}
	if id == "" {
		return nil, fmt.Errorf("entry ID is required")
	}
	entry, err := m.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return entry.Clone(), nil
}

// RemoveKnowledge 删除条目。
func (m *Manager) RemoveKnowledge(ctx context.Context, id string) error {
	if err := ensureContext(ctx); err != nil {
		return err
	}
	if id == "" {
		return fmt.Errorf("entry ID is required")
	}
	return m.store.Delete(ctx, id)
}

// Clear 清空所有条目。
func (m *Manager) Clear(ctx context.Context) error {
	if err := ensureContext(ctx); err != nil {
		return err
	}
	return m.store.Clear(ctx)
}

// InMemoryStore 简单内存实现，便于测试与开发。
type InMemoryStore struct {
	mu      sync.RWMutex
	entries map[string]*Entry
}

// NewInMemoryStore 创建内存存储。
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		entries: make(map[string]*Entry),
	}
}

// Create 保存新条目。
func (s *InMemoryStore) Create(_ context.Context, entry *Entry) error {
	if entry == nil {
		return fmt.Errorf("entry cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.entries[entry.ID]; exists {
		return fmt.Errorf("culture entry %s already exists", entry.ID)
	}

	s.entries[entry.ID] = entry.Clone()
	return nil
}

// Update 更新条目。
func (s *InMemoryStore) Update(_ context.Context, entry *Entry) error {
	if entry == nil {
		return fmt.Errorf("entry cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.entries[entry.ID]; !exists {
		return ErrNotFound
	}

	s.entries[entry.ID] = entry.Clone()
	return nil
}

// Delete 删除条目。
func (s *InMemoryStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.entries[id]; !ok {
		return ErrNotFound
	}
	delete(s.entries, id)
	return nil
}

// Get 获取单条记录。
func (s *InMemoryStore) Get(_ context.Context, id string) (*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.entries[id]
	if !ok {
		return nil, ErrNotFound
	}
	return entry.Clone(), nil
}

// List 根据过滤条件返回条目列表。
func (s *InMemoryStore) List(_ context.Context, filter Filter) ([]*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*Entry
	for _, entry := range s.entries {
		if matchesFilter(entry, filter) {
			results = append(results, entry.Clone())
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.Before(results[j].CreatedAt)
	})

	return results, nil
}

// Clear 清空存储。
func (s *InMemoryStore) Clear(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = make(map[string]*Entry)
	return nil
}

func matchesFilter(entry *Entry, filter Filter) bool {
	if len(filter.Tags) == 0 {
		return true
	}

	tagSet := make(map[string]struct{}, len(entry.Tags))
	for _, tag := range entry.Tags {
		tagSet[strings.ToLower(tag)] = struct{}{}
	}

	for _, tag := range filter.Tags {
		if _, ok := tagSet[strings.ToLower(tag)]; !ok {
			return false
		}
	}
	return true
}

func normaliseEntry(entry *Entry) {
	if entry.Tags != nil {
		normalized := make([]string, 0, len(entry.Tags))
		seen := make(map[string]struct{}, len(entry.Tags))
		for _, tag := range entry.Tags {
			tag = strings.TrimSpace(strings.ToLower(tag))
			if tag == "" {
				continue
			}
			if _, exists := seen[tag]; exists {
				continue
			}
			seen[tag] = struct{}{}
			normalized = append(normalized, tag)
		}
		sort.Strings(normalized)
		entry.Tags = normalized
	}
	if entry.Metadata == nil {
		entry.Metadata = make(map[string]interface{})
	}
}

func ensureContext(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
