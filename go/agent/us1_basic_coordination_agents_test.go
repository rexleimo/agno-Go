package agent

import "testing"

func TestUS1Agents(t *testing.T) {
	agents := US1Agents()
	if len(agents) != 2 {
		t.Fatalf("expected 2 agents for US1, got %d", len(agents))
	}

	var hasHN, hasReader bool
	for _, a := range agents {
		if a.ID == ID("hn_researcher") {
			hasHN = true
		}
		if a.ID == ID("article_reader") {
			hasReader = true
		}
		if a.Name == "" || a.Role == "" {
			t.Errorf("agent %q has empty name or role", a.ID)
		}
		if len(a.AllowedProviders) == 0 {
			t.Errorf("agent %q has no allowed providers", a.ID)
		}
	}

	if !hasHN || !hasReader {
		t.Fatalf("missing US1 agents: hn=%v article_reader=%v", hasHN, hasReader)
	}
}
