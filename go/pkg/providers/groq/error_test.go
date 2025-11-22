package groq

import "testing"

func TestGroqErrorParser(t *testing.T) {
	body := map[string]any{"message": "outer"}
	if got := groqErrorParser(body); got != "outer" {
		t.Fatalf("expected outer, got %s", got)
	}
	body = map[string]any{"error": map[string]any{"message": "inner"}}
	if got := groqErrorParser(body); got != "inner" {
		t.Fatalf("expected inner, got %s", got)
	}
	body = map[string]any{}
	if got := groqErrorParser(body); got != "" {
		t.Fatalf("expected empty, got %s", got)
	}
}
