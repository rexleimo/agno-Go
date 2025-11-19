package providers

import (
	"fmt"
	"time"

	internalerrors "github.com/agno-agi/agno-go/go/internal/errors"
	"github.com/agno-agi/agno-go/go/internal/telemetry"
)

// ID uniquely identifies a provider configuration (for example, a model or
// retriever backend).
type ID string

// Type describes the high-level capability type of a provider.
type Type string

const (
	TypeLLM          Type = "llm"
	TypeRetriever    Type = "retriever"
	TypeToolExecutor Type = "tool-executor"
	TypeBusinessAPI  Type = "business-api"
)

// Capability describes a specific operation that a provider can perform.
type Capability string

const (
	CapabilityGenerate   Capability = "generate"
	CapabilityEmbed      Capability = "embed"
	CapabilitySearch     Capability = "search"
	CapabilityInvokeTool Capability = "invoke_tool"
)

// ErrorCode is a coarse-grained classification of provider errors, aligned
// with the specification's error_semantics.
type ErrorCode string

const (
	ErrorTimeout       ErrorCode = "timeout"
	ErrorRateLimit     ErrorCode = "rate_limit"
	ErrorUnauthorized  ErrorCode = "unauthorized"
	ErrorInternal      ErrorCode = "internal"
	ErrorUnimplemented ErrorCode = "unimplemented"
)

// Config is a generic key–value configuration bag. Concrete providers can
// map these fields to environment variables, SDK options, or HTTP settings.
type Config map[string]any

// Provider describes the static configuration and capabilities of a provider.
type Provider struct {
	ID           ID
	Type         Type
	DisplayName  string
	Config       Config
	Capabilities []Capability
}

// US1 providers for the "basic coordination" scenario.
//
// These are intentionally minimal and focus on describing configuration and
// capability shape; concrete network calls and SDK integrations are handled
// by higher layers.
var (
	US1OpenAIChat = Provider{
		ID:          ID("openai-chat-gpt-5-mini"),
		Type:        TypeLLM,
		DisplayName: "OpenAI Chat gpt-5-mini",
		Config: Config{
			"model": "gpt-5-mini",
		},
		Capabilities: []Capability{CapabilityGenerate},
	}

	US1HackerNewsTools = Provider{
		ID:          ID("hackernews-tools"),
		Type:        TypeToolExecutor,
		DisplayName: "HackerNews Tools",
		Config:      Config{},
		Capabilities: []Capability{
			CapabilityInvokeTool,
		},
	}

	US1Newspaper4kTools = Provider{
		ID:          ID("newspaper4k-tools"),
		Type:        TypeToolExecutor,
		DisplayName: "Newspaper4k Tools",
		Config:      Config{},
		Capabilities: []Capability{
			CapabilityInvokeTool,
		},
	}
)

// US1Providers returns the providers required by the US1 "basic coordination"
// scenario.
func US1Providers() []Provider {
	return []Provider{
		US1OpenAIChat,
		US1HackerNewsTools,
		US1Newspaper4kTools,
	}
}

// NewNotMigratedError returns a classified error indicating that the given
// provider has not yet been migrated to the Go implementation.
func NewNotMigratedError(providerID ID) *internalerrors.Error {
	return internalerrors.NewNotMigrated(
		fmt.Sprintf("provider %q is not yet migrated to Go", providerID),
	)
}

// RecordNotMigrated emits a telemetry event for an attempt to use a provider
// that has not yet been migrated to Go. Callers are expected to pass in a
// Recorder implementation; in early stages this can be telemetry.NoopRecorder.
func RecordNotMigrated(rec telemetry.Recorder, providerID ID) {
	rec.Record(telemetry.Event{
		ID:         "provider-not-migrated",
		Timestamp:  time.Now(),
		ProviderID: string(providerID),
		Type:       telemetry.EventProviderError,
		Payload: map[string]any{
			"error_code": string(internalerrors.CodeNotMigrated),
		},
	})
}
