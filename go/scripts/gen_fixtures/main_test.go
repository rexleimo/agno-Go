package main

import "testing"

func TestIsFixtureFile(t *testing.T) {
	valid := []string{"a.json", "b.yaml", "c.YML"}
	for _, name := range valid {
		if !isFixtureFile(name) {
			t.Fatalf("expected fixture file %s", name)
		}
	}
	if isFixtureFile("readme.md") {
		t.Fatalf("expected non-fixture for md")
	}
}

func TestChoosePathPriority(t *testing.T) {
	t.Setenv("FIXTURE_SOURCE_DIR", "from-env")
	if got := choosePath("", "FIXTURE_SOURCE_DIR", "default"); got != "from-env" {
		t.Fatalf("expected env override, got %s", got)
	}
	if got := choosePath("flag", "FIXTURE_SOURCE_DIR", "default"); got != "flag" {
		t.Fatalf("expected flag override, got %s", got)
	}
}
