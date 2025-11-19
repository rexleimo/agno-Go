package errors

import "testing"

func TestNewNotMigrated(t *testing.T) {
	err := NewNotMigrated("feature not yet available in Go")
	if err.Code != CodeNotMigrated {
		t.Fatalf("expected CodeNotMigrated, got %v", err.Code)
	}
	if err.Message == "" {
		t.Fatalf("expected non-empty message")
	}
}
