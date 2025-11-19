package testutil

import "testing"

// RequireNoError is a minimal helper to fail a test immediately when err is
// non-nil. It is intentionally tiny to avoid pulling in external assertion
// libraries.
func RequireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
