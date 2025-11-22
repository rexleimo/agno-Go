package openrouter

import (
	"testing"

	"github.com/rexleimo/agno-go/pkg/providers/shared"
)

func TestOpenRouterErrorParser(t *testing.T) {
	body := map[string]any{
		"error": map[string]any{
			"message":     "blocked",
			"description": "desc",
		},
	}
	msg := openRouterErrorParser(body)
	if msg != "blocked" {
		t.Fatalf("expected message, got %q", msg)
	}
	body = map[string]any{
		"error": map[string]any{
			"description": "desc",
		},
	}
	if msg := openRouterErrorParser(body); msg != "desc" {
		t.Fatalf("expected description, got %q", msg)
	}
	_ = shared.OACompatChatResp{} // ensure import retained
}
