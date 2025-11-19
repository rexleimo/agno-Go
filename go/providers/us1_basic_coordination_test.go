package providers

import "testing"

func TestUS1Providers(t *testing.T) {
	ps := US1Providers()
	if len(ps) != 3 {
		t.Fatalf("expected 3 providers for US1, got %d", len(ps))
	}

	var hasOpenAI, hasHN, hasNewspaper bool
	for _, p := range ps {
		switch p.ID {
		case ID("openai-chat-gpt-5-mini"):
			hasOpenAI = true
		case ID("hackernews-tools"):
			hasHN = true
		case ID("newspaper4k-tools"):
			hasNewspaper = true
		}
		if p.DisplayName == "" {
			t.Errorf("provider %q has empty display name", p.ID)
		}
		if len(p.Capabilities) == 0 {
			t.Errorf("provider %q has no capabilities", p.ID)
		}
	}

	if !hasOpenAI || !hasHN || !hasNewspaper {
		t.Fatalf("missing one or more US1 providers: openai=%v hn=%v newspaper=%v", hasOpenAI, hasHN, hasNewspaper)
	}
}
