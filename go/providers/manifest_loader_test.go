package providers

import (
	"testing"
)

func TestDecodeManifestYAML(t *testing.T) {
	data := []byte(`
providers:
  - id: openai-chat
    type: llm
    display_name: OpenAI GPT
    capabilities: ["generate"]
    config:
      model: gpt-5-mini
toolkits:
  - id: research
    description: Research toolkit
    providers: ["openai-chat"]
    guardrails: ["safe-output"]
guardrails:
  - id: safe-output
    type: regex
    enforcement: block
`)

	manifest, err := DecodeManifest(data)
	if err != nil {
		t.Fatalf("DecodeManifest: %v", err)
	}
	if len(manifest.Providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(manifest.Providers))
	}
	if manifest.Toolkits[0].Guardrails[0] != "safe-output" {
		t.Fatalf("unexpected guardrail reference")
	}

	providers := manifest.ProviderConfigs()
	if len(providers) != 1 {
		t.Fatalf("expected provider conversion")
	}
	if providers[0].ID != ID("openai-chat") {
		t.Fatalf("unexpected provider ID: %s", providers[0].ID)
	}
	if len(providers[0].Capabilities) != 1 || providers[0].Capabilities[0] != Capability("generate") {
		t.Fatalf("unexpected capabilities: %+v", providers[0].Capabilities)
	}
}

func TestDecodeManifestValidationError(t *testing.T) {
	data := []byte(`providers: [{type: llm}]`)
	if _, err := DecodeManifest(data); err == nil {
		t.Fatalf("expected validation error for missing provider id")
	}
}

func TestMergeProviders(t *testing.T) {
	base := []Provider{
		{ID: ID("openai"), DisplayName: "OpenAI"},
	}
	extra := []Provider{
		{ID: ID("openai"), DisplayName: "Overridden"},
		{ID: ID("anthropic"), DisplayName: "Claude"},
	}

	result := MergeProviders(base, extra)
	if len(result) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(result))
	}
	if result[0].DisplayName != "Overridden" {
		t.Fatalf("expected override to win, got %s", result[0].DisplayName)
	}
}
