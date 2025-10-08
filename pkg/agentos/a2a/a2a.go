package a2a

import (
	"context"
	"fmt"
	"sync"
)

// Entity represents any runnable entity (Agent, Team, or Workflow)
// Entity 表示任何可运行的实体（Agent、Team 或 Workflow）
type Entity interface {
	// Run executes the entity with the given input
	// Run 使用给定输入执行实体
	Run(ctx context.Context, input string) (interface{}, error)

	// GetID returns the entity's unique identifier
	// GetID 返回实体的唯一标识符
	GetID() string

	// GetName returns the entity's display name
	// GetName 返回实体的显示名称
	GetName() string
}

// A2AInterface manages A2A protocol endpoints for agents, teams, and workflows
// A2AInterface 管理 agents、teams 和 workflows 的 A2A 协议端点
type A2AInterface struct {
	entities map[string]Entity // entityID -> Entity
	prefix   string            // URL prefix (default: "/a2a")
	mu       sync.RWMutex
}

// Config for A2A interface
// A2A接口配置
type Config struct {
	// Agents to expose via A2A
	// 通过 A2A 暴露的 Agents
	Agents []Entity

	// Teams to expose via A2A
	// 通过 A2A 暴露的 Teams
	Teams []Entity

	// Workflows to expose via A2A
	// 通过 A2A 暴露的 Workflows
	Workflows []Entity

	// URL prefix for A2A endpoints (default: "/a2a")
	// A2A 端点的 URL 前缀（默认: "/a2a"）
	Prefix string
}

// New creates a new A2A interface
// New 创建新的 A2A 接口
func New(config Config) (*A2AInterface, error) {
	// Validate config
	// 验证配置
	totalEntities := len(config.Agents) + len(config.Teams) + len(config.Workflows)
	if totalEntities == 0 {
		return nil, fmt.Errorf("at least one agent, team, or workflow is required")
	}

	// Set default prefix
	// 设置默认前缀
	if config.Prefix == "" {
		config.Prefix = "/a2a"
	}

	a := &A2AInterface{
		entities: make(map[string]Entity),
		prefix:   config.Prefix,
	}

	// Register all entities
	// 注册所有实体
	for _, agent := range config.Agents {
		if err := a.registerEntity(agent); err != nil {
			return nil, fmt.Errorf("failed to register agent %s: %w", agent.GetID(), err)
		}
	}

	for _, team := range config.Teams {
		if err := a.registerEntity(team); err != nil {
			return nil, fmt.Errorf("failed to register team %s: %w", team.GetID(), err)
		}
	}

	for _, workflow := range config.Workflows {
		if err := a.registerEntity(workflow); err != nil {
			return nil, fmt.Errorf("failed to register workflow %s: %w", workflow.GetID(), err)
		}
	}

	return a, nil
}

// registerEntity registers an entity (Agent, Team, or Workflow)
// registerEntity 注册一个实体（Agent、Team 或 Workflow）
func (a *A2AInterface) registerEntity(entity Entity) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	id := entity.GetID()
	if id == "" {
		return fmt.Errorf("entity ID cannot be empty")
	}

	if _, exists := a.entities[id]; exists {
		return fmt.Errorf("entity with ID '%s' already registered", id)
	}

	a.entities[id] = entity
	return nil
}

// FindEntity finds an entity by ID
// FindEntity 通过 ID 查找实体
func (a *A2AInterface) FindEntity(entityID string) (Entity, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	entity, exists := a.entities[entityID]
	if !exists {
		return nil, fmt.Errorf("entity '%s' not found", entityID)
	}

	return entity, nil
}

// GetPrefix returns the URL prefix for A2A endpoints
// GetPrefix 返回 A2A 端点的 URL 前缀
func (a *A2AInterface) GetPrefix() string {
	return a.prefix
}

// ListEntities returns a list of all registered entity IDs
// ListEntities 返回所有已注册实体 ID 的列表
func (a *A2AInterface) ListEntities() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	ids := make([]string, 0, len(a.entities))
	for id := range a.entities {
		ids = append(ids, id)
	}

	return ids
}
