package agent

import (
	"time"

	"github.com/google/uuid"
)

// Provider enumerates supported model providers.
type Provider string

const (
	ProviderOllama      Provider = "ollama"
	ProviderGemini      Provider = "gemini"
	ProviderOpenAI      Provider = "openai"
	ProviderGLM4        Provider = "glm4"
	ProviderOpenRouter  Provider = "openrouter"
	ProviderSiliconFlow Provider = "siliconflow"
	ProviderCerebras    Provider = "cerebras"
	ProviderModelScope  Provider = "modelscope"
	ProviderGroq        Provider = "groq"
)

// SessionState describes the lifecycle states of a session.
type SessionState string

const (
	SessionIdle      SessionState = "idle"
	SessionStreaming SessionState = "streaming"
	SessionCompleted SessionState = "completed"
	SessionErrored   SessionState = "errored"
	SessionCancelled SessionState = "cancelled"
)

// MemoryStoreType represents available persistence backends.
type MemoryStoreType string

const (
	MemoryStoreInMemory MemoryStoreType = "memory"
	MemoryStoreBolt     MemoryStoreType = "bolt"
	MemoryStoreBadger   MemoryStoreType = "badger"
)

// Role aligns with OpenAPI roles for chat transcripts.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
)

// ToolCallStatus captures tool lifecycle outcomes.
type ToolCallStatus string

const (
	ToolStatusPending  ToolCallStatus = "pending"
	ToolStatusSuccess  ToolCallStatus = "success"
	ToolStatusError    ToolCallStatus = "error"
	ToolStatusTimeout  ToolCallStatus = "timeout"
	ToolStatusDisabled ToolCallStatus = "disabled"
)

// ModelConfig controls the provider/model selection and runtime behavior.
type ModelConfig struct {
	Provider    Provider `json:"provider" yaml:"provider"`
	ModelID     string   `json:"modelId" yaml:"modelId"`
	Stream      bool     `json:"stream,omitempty" yaml:"stream,omitempty"`
	Temperature float64  `json:"temperature,omitempty" yaml:"temperature,omitempty"`
	MaxTokens   *int     `json:"maxTokens,omitempty" yaml:"maxTokens,omitempty"`
	TimeoutMs   int      `json:"timeoutMs,omitempty" yaml:"timeoutMs,omitempty"`
}

// ToolConfig enumerates registered/enabled tools plus optional MCP endpoints.
type ToolConfig struct {
	Registered    []string `json:"registered,omitempty" yaml:"registered,omitempty"`
	Enabled       []string `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	MCPEndpoints  []string `json:"mcpEndpoints,omitempty" yaml:"mcpEndpoints,omitempty"`
	ToolTimeoutMs int      `json:"toolTimeoutMs,omitempty" yaml:"toolTimeoutMs,omitempty"`
}

// MemoryConfig captures persistence and eviction settings.
type MemoryConfig struct {
	StoreType   MemoryStoreType `json:"storeType,omitempty" yaml:"storeType,omitempty"`
	Namespace   string          `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Retention   time.Duration   `json:"retention,omitempty" yaml:"retention,omitempty"`
	TokenWindow int             `json:"tokenWindow,omitempty" yaml:"tokenWindow,omitempty"`
}

// Agent describes an agent runtime configuration.
type Agent struct {
	ID        uuid.UUID      `json:"id"`
	Name      string         `json:"name"`
	Model     ModelConfig    `json:"model"`
	Tools     ToolConfig     `json:"tools,omitempty"`
	Memory    MemoryConfig   `json:"memory,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"createdAt,omitempty"`
	UpdatedAt time.Time      `json:"updatedAt,omitempty"`
}

// Session tracks a conversation tied to an Agent.
type Session struct {
	ID             uuid.UUID      `json:"sessionId"`
	AgentID        uuid.UUID      `json:"agentId"`
	State          SessionState   `json:"state"`
	UserID         string         `json:"userId,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	LastActivityAt time.Time      `json:"lastActivityAt,omitempty"`
	ExpiredAt      *time.Time     `json:"expiredAt,omitempty"`
	History        []Message      `json:"history,omitempty"`
	CreatedAt      time.Time      `json:"createdAt,omitempty"`
}

// Message stores a single chat entry and optional tool calls.
type Message struct {
	ID        uuid.UUID  `json:"id,omitempty"`
	AgentID   uuid.UUID  `json:"agentId,omitempty"`
	SessionID uuid.UUID  `json:"sessionId,omitempty"`
	Role      Role       `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"toolCalls,omitempty"`
	Usage     Usage      `json:"usage,omitempty"`
	CreatedAt time.Time  `json:"createdAt,omitempty"`
}

// ToolCall captures a pending or completed tool invocation.
type ToolCall struct {
	ToolCallID string          `json:"toolCallId"`
	Name       string          `json:"name"`
	Args       map[string]any  `json:"args,omitempty"`
	IssuedAt   time.Time       `json:"issuedAt,omitempty"`
	Result     *ToolCallResult `json:"result,omitempty"`
}

// ToolCallResult records the outcome of a tool call.
type ToolCallResult struct {
	ToolCallID  string         `json:"toolCallId,omitempty"`
	Status      ToolCallStatus `json:"status"`
	Output      string         `json:"output,omitempty"`
	Error       string         `json:"error,omitempty"`
	DurationMs  int64          `json:"durationMs,omitempty"`
	CompletedAt time.Time      `json:"completedAt,omitempty"`
}

// Usage aggregates token and latency metadata for a message turn.
type Usage struct {
	PromptTokens     int `json:"promptTokens,omitempty"`
	CompletionTokens int `json:"completionTokens,omitempty"`
	LatencyMs        int `json:"latencyMs,omitempty"`
}

// RoleString returns the string value for the message role.
func (m Message) RoleString() string {
	return string(m.Role)
}
