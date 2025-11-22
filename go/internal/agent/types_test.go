package agent

import "testing"

func TestMessageRoleString(t *testing.T) {
	m := Message{Role: RoleAssistant}
	if got := m.RoleString(); got != "assistant" {
		t.Fatalf("expected assistant, got %s", got)
	}
}

func TestSessionStateValues(t *testing.T) {
	states := []SessionState{SessionIdle, SessionStreaming, SessionCompleted, SessionErrored, SessionCancelled}
	if len(states) != 5 {
		t.Fatalf("unexpected session states length: %d", len(states))
	}
}
