package portkey

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
    p, err := New("gpt-4o-mini", Config{APIKey: "portkey-key"})
    if err != nil { t.Fatalf("New error: %v", err) }
    req := &models.InvokeRequest{ Messages: []*types.Message{ types.NewUserMessage("Hi") }, MaxTokens: 50 }
    cr := p.buildChatRequest(req)
    if cr.Model != p.ID { t.Errorf("model = %s, want %s", cr.Model, p.ID) }
    if cr.MaxTokens != 50 { t.Errorf("max_tokens = %v, want 50", cr.MaxTokens) }
}

