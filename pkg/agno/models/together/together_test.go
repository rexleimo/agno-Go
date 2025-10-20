package together

import (
    "testing"
    "github.com/rexleimo/agno-go/pkg/agno/models"
    "github.com/rexleimo/agno-go/pkg/agno/types"
)

func TestNew_MissingAPIKey(t *testing.T) {
    _, err := New("meta-llama/Meta-Llama-3.1-8B-Instruct-Turbo", Config{})
    if err == nil {
        t.Fatalf("expected error for missing API key")
    }
}

func TestBuildChatRequest_Basic(t *testing.T) {
    tg, err := New("meta-llama/Meta-Llama-3.1-8B-Instruct-Turbo", Config{APIKey: "test"})
    if err != nil {
        t.Fatalf("New error: %v", err)
    }

    req := &models.InvokeRequest{
        Messages: []*types.Message{
            types.NewSystemMessage("Be concise."),
            types.NewUserMessage("Hello"),
            types.NewAssistantMessage("Hi!"),
            types.NewUserMessage("Write a haiku about data"),
        },
        Temperature: 0.3,
        MaxTokens:   64,
    }

    cr := tg.buildChatRequest(req)
    if cr.Model != tg.ID {
        t.Errorf("model = %s, want %s", cr.Model, tg.ID)
    }
    if len(cr.Messages) != len(req.Messages) {
        t.Errorf("messages count = %d, want %d", len(cr.Messages), len(req.Messages))
    }
    if float32(0.3) != cr.Temperature {
        t.Errorf("temperature = %v, want 0.3", cr.Temperature)
    }
    if cr.MaxTokens != 64 {
        t.Errorf("max_tokens = %v, want 64", cr.MaxTokens)
    }
}

