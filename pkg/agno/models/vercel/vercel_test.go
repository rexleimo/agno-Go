package vercel

import (
    "testing"
    "github.com/rexleimo/agno-go/pkg/agno/models"
    "github.com/rexleimo/agno-go/pkg/agno/types"
)

func TestNew_MissingAPIKey(t *testing.T) {
    _, err := New("gpt-4o-mini", Config{})
    if err == nil { t.Fatalf("expected error for missing API key") }
}

func TestBuildChatRequest_Basic(t *testing.T) {
    v, err := New("gpt-4o-mini", Config{APIKey: "vercel-key", BaseURL: "https://example.com/v1"})
    if err != nil { t.Fatalf("New error: %v", err) }
    req := &models.InvokeRequest{ Messages: []*types.Message{ types.NewUserMessage("Hello") }, Temperature: 0.5, MaxTokens: 100 }
    cr := v.buildChatRequest(req)
    if cr.Model != v.ID { t.Errorf("model = %s, want %s", cr.Model, v.ID) }
    if cr.MaxTokens != 100 { t.Errorf("max_tokens = %v, want 100", cr.MaxTokens) }
}

