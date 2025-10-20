package lmstudio

import (
    "testing"
    "github.com/rexleimo/agno-go/pkg/agno/models"
    "github.com/rexleimo/agno-go/pkg/agno/types"
)

func TestNew_Defaults(t *testing.T) {
    // no API key required
    lm, err := New("local-model", Config{})
    if err != nil {
        t.Fatalf("New error: %v", err)
    }
    if lm.config.BaseURL == "" {
        t.Fatalf("expected default base URL")
    }
}

func TestBuildChatRequest_Basic(t *testing.T) {
    lm, _ := New("local-model", Config{})
    req := &models.InvokeRequest{
        Messages: []*types.Message{
            types.NewSystemMessage("Local test"),
            types.NewUserMessage("Hello"),
        },
        Temperature: 0.1,
        MaxTokens:   16,
    }
    cr := lm.buildChatRequest(req)
    if cr.Model != "local-model" {
        t.Errorf("model = %s, want local-model", cr.Model)
    }
    if len(cr.Messages) != 2 {
        t.Errorf("messages = %d, want 2", len(cr.Messages))
    }
    if cr.MaxTokens != 16 {
        t.Errorf("max_tokens = %v, want 16", cr.MaxTokens)
    }
}

