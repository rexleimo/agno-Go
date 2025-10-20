package sambanova

import (
    "testing"
    "github.com/rexleimo/agno-go/pkg/agno/models"
    "github.com/rexleimo/agno-go/pkg/agno/types"
)

func TestNew_MissingAPIKey(t *testing.T) {
    _, err := New("Meta-Llama-3.1-70B-Instruct", Config{})
    if err == nil { t.Fatalf("expected error for missing API key") }
}

func TestBuildChatRequest_Basic(t *testing.T) {
    s, err := New("Meta-Llama-3.1-70B-Instruct", Config{APIKey: "sn-key"})
    if err != nil { t.Fatalf("New error: %v", err) }
    req := &models.InvokeRequest{ Messages: []*types.Message{ types.NewUserMessage("Hi") }, Temperature: 0.4, MaxTokens: 40 }
    cr := s.buildChatRequest(req)
    if cr.Model != s.ID { t.Errorf("model = %s, want %s", cr.Model, s.ID) }
    if cr.MaxTokens != 40 { t.Errorf("max_tokens = %v, want 40", cr.MaxTokens) }
}

