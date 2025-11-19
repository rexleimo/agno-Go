package providers

import (
	"strings"
	"testing"
)

func TestRegistryLoadsToolkit(t *testing.T) {
	manifestYAML := []byte(`
providers:
  - id: openai-chat
    type: llm
    display_name: OpenAI
toolkits:
  - id: research
    description: Research toolkit
    providers: ["openai-chat"]
    guardrails: []
    metadata:
      notes: demo
knowledge:
  - id: knowledge-1
    description: Articles
    provider: openai-chat
memories:
  - id: memory-1
    description: Memory store
    provider: openai-chat
`)
	manifest, err := DecodeManifest(manifestYAML)
	if err != nil {
		t.Fatalf("DecodeManifest: %v", err)
	}
	registry := NewRegistry(nil)
	if err := registry.ApplyManifest(manifest); err != nil {
		t.Fatalf("ApplyManifest: %v", err)
	}
	toolkit, ok := registry.Toolkit("research")
	if !ok {
		t.Fatalf("expected toolkit to be retrievable")
	}
	if len(toolkit.Providers()) != 1 {
		t.Fatalf("expected providers to be loaded")
	}
	if _, ok := registry.Knowledge("knowledge-1"); !ok {
		t.Fatalf("expected knowledge binding")
	}
	if _, ok := registry.Memory("memory-1"); !ok {
		t.Fatalf("expected memory binding")
	}
}

func TestRegistryNotMigratedProviderError(t *testing.T) {
	manifestYAML := []byte(`
toolkits:
  - id: legacy
    description: Legacy toolkit
    providers: ["python-only"]
    metadata:
      fallback_provider: "legacy-py"
      migration_doc: "https://docs.example/legacy"
`)
	manifest, err := DecodeManifest(manifestYAML)
	if err != nil {
		t.Fatalf("DecodeManifest: %v", err)
	}
	registry := NewRegistry(nil)
	err = registry.ApplyManifest(manifest)
	if err == nil {
		t.Fatalf("expected not migrated error")
	}
	nmErr := NotMigratedError{}
	if !strings.Contains(err.Error(), "not migrated") {
		t.Fatalf("expected not migrated message, got %v", err)
	}
	if cast, ok := err.(NotMigratedError); ok {
		nmErr = cast
	} else if wrapped, ok := err.(interface{ Unwrap() error }); ok {
		if inner, ok := wrapped.Unwrap().(NotMigratedError); ok {
			nmErr = inner
		}
	}
	if nmErr.Fallback != "legacy-py" {
		t.Fatalf("expected fallback propagated, got %+v", nmErr)
	}
}
