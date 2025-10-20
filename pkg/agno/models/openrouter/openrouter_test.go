package openrouter

import (
    "testing"
    "github.com/rexleimo/agno-go/pkg/agno/models"
    "github.com/rexleimo/agno-go/pkg/agno/types"
)

func TestNew_MissingAPIKey(t *testing.T) {
    _, err := New("openrouter/auto", Config{})
    if err == nil {
        t.Fatalf("expected error for missing API key")
    }
}

func TestBuildChatRequest_Basic(t *testing.T) {
    or, err := New("openrouter/auto", Config{APIKey: "test-key"})
    if err != nil {
        t.Fatalf("New error: %v", err)
    }

    req := &models.InvokeRequest{
        Messages: []*types.Message{
            types.NewSystemMessage("Behave."),
            types.NewUserMessage("Ping"),
            types.NewAssistantMessage("Pong"),
            types.NewUserMessage("Another ping"),
        },
        Temperature: 0.2,
        MaxTokens:   32,
    }
    cr := or.buildChatRequest(req)
    if cr.Model != or.ID {
        t.Errorf("model = %s, want %s", cr.Model, or.ID)
    }
    if len(cr.Messages) != len(req.Messages) {
        t.Errorf("messages = %d, want %d", len(cr.Messages), len(req.Messages))
    }
    if cr.MaxTokens != 32 {
        t.Errorf("max_tokens = %v, want 32", cr.MaxTokens)
    }
}

