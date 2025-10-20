package internlm

import (
    "testing"
    "github.com/rexleimo/agno-go/pkg/agno/models"
    "github.com/rexleimo/agno-go/pkg/agno/types"
)

func TestNew_RequiresBaseURLAndKey(t *testing.T) {
    if _, err := New("internlm2.5", Config{}); err == nil { t.Fatalf("expected error for missing key/baseURL") }
    if _, err := New("internlm2.5", Config{APIKey: "k"}); err == nil { t.Fatalf("expected error for missing baseURL") }
}

func TestBuildChatRequest_Basic(t *testing.T) {
    in, err := New("internlm2.5", Config{APIKey: "k", BaseURL: "https://example.com/v1"})
    if err != nil { t.Fatalf("New error: %v", err) }
    req := &models.InvokeRequest{ Messages: []*types.Message{ types.NewUserMessage("Hi") }, MaxTokens: 20 }
    cr := in.buildChatRequest(req)
    if cr.Model != in.ID { t.Errorf("model = %s, want %s", cr.Model, in.ID) }
    if cr.MaxTokens != 20 { t.Errorf("max_tokens = %v, want 20", cr.MaxTokens) }
}

