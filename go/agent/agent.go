package agent

// ID uniquely identifies an Agent within a workflow or session.
type ID string

// ProviderID identifies a provider that an Agent can call.
type ProviderID string

// ToolID identifies a tool or operation that an Agent can use.
type ToolID string

// Schema is a minimal representation of an input or output schema.
// It is intentionally abstract and can be mapped to concrete validation
// mechanisms in higher layers.
type Schema struct {
	// Name is a human-readable name for the schema.
	Name string
	// Description describes the expected shape or semantics of the data.
	Description string
}

// MemoryPolicy describes how an Agent should handle conversational or task
// memory.
type MemoryPolicy struct {
	// Persist indicates whether memory should be persisted beyond a single run.
	Persist bool `json:"persist,omitempty"`
	// WindowSize indicates the preferred number of recent interactions to keep
	// in active context. Zero means implementation-defined default.
	WindowSize int `json:"windowSize,omitempty"`
	// SensitiveFiltering notes whether sensitive data should be filtered from
	// stored memory.
	SensitiveFiltering bool `json:"sensitiveFiltering,omitempty"`
}

// Agent represents an executable unit with a clear role and capabilities.
// It binds together configuration, allowed providers/tools, and memory policy.
type Agent struct {
	ID               ID
	Name             string
	Role             string
	Description      string
	AllowedProviders []ProviderID
	AllowedTools     []ToolID
	InputSchema      Schema
	OutputSchema     Schema
	MemoryPolicy     MemoryPolicy
}
