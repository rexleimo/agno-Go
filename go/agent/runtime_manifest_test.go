package agent

import (
	"errors"
	"testing"
)

func TestNewAgentRuntimeSuccess(t *testing.T) {
	cfg := AgentRuntime{
		ID:       ID("agent-researcher"),
		Name:     "Researcher",
		ModelRef: "openai:gpt-5-mini",
		Toolkits: []ToolkitRef{"search_perplexity"},
		MemoryPolicy: MemoryPolicy{
			Persist:            true,
			WindowSize:         10,
			SensitiveFiltering: true,
		},
		SessionPolicy: SessionPolicy{
			SessionID:               "session-123",
			OverwriteDBSessionState: false,
			EnableAgenticState:      true,
		},
		Timeouts: TimeoutPolicy{
			RunMilliseconds:      5_000,
			ToolCallMilliseconds: 1_000,
		},
	}

	runtime, err := NewAgentRuntime(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if runtime == nil {
		t.Fatalf("expected runtime, got nil")
	}
	if runtime.Hooks == nil {
		t.Fatalf("expected hooks map to be initialized")
	}
	if runtime.Metadata == nil {
		t.Fatalf("expected metadata map to be initialized")
	}
}

func TestAgentRuntimeValidateFailures(t *testing.T) {
	tests := []struct {
		name      string
		cfg       AgentRuntime
		wantField string
	}{
		{
			name: "missing id",
			cfg: AgentRuntime{
				ModelRef:     "openai:gpt-5-mini",
				MemoryPolicy: MemoryPolicy{},
				SessionPolicy: SessionPolicy{
					SessionID: "s",
				},
			},
			wantField: "id",
		},
		{
			name: "session conflict",
			cfg: AgentRuntime{
				ID:       ID("agent-1"),
				ModelRef: "openai:gpt-5-mini",
				MemoryPolicy: MemoryPolicy{
					WindowSize: 0,
				},
				SessionPolicy: SessionPolicy{
					OverwriteDBSessionState: true,
					EnableAgenticState:      true,
				},
			},
			wantField: "sessionPolicy",
		},
		{
			name: "duplicate toolkit",
			cfg: AgentRuntime{
				ID:       ID("agent-1"),
				ModelRef: "openai:gpt-5-mini",
				Toolkits: []ToolkitRef{"search", "search"},
				MemoryPolicy: MemoryPolicy{
					WindowSize: 1,
				},
				SessionPolicy: SessionPolicy{
					SessionID: "s",
				},
			},
			wantField: "toolkits[1]",
		},
		{
			name: "negative timeout",
			cfg: AgentRuntime{
				ID:       ID("agent-1"),
				ModelRef: "openai:gpt-5-mini",
				MemoryPolicy: MemoryPolicy{
					WindowSize: 1,
				},
				SessionPolicy: SessionPolicy{
					SessionID: "s",
				},
				Timeouts: TimeoutPolicy{
					RunMilliseconds: -1,
				},
			},
			wantField: "timeouts.runMilliseconds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAgentRuntime(tt.cfg)
			if err == nil {
				t.Fatalf("expected error for %s", tt.name)
			}
			var validationErrs AgentRuntimeValidationErrors
			if !errors.As(err, &validationErrs) {
				t.Fatalf("expected AgentRuntimeValidationErrors, got %T", err)
			}
			found := false
			for _, v := range validationErrs {
				if v.Field == tt.wantField {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("expected error field %q in %+v", tt.wantField, validationErrs)
			}
		})
	}
}
