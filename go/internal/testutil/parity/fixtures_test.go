package parity

import (
	"testing"
)

func TestLoadJSONFixture(t *testing.T) {
	data := []byte(`{
		"fixture_id":"us1",
		"description":"demo",
		"workflow_template":{"id":"wf1"},
		"user_inputs":[{"role":"user","content":"hi"}]
	}`)

	fixture, err := LoadJSONFixture(data)
	if err != nil {
		t.Fatalf("LoadJSONFixture: %v", err)
	}
	if fixture.FixtureID != "us1" {
		t.Fatalf("unexpected fixture id: %s", fixture.FixtureID)
	}
}

func TestLoadYAMLFixture(t *testing.T) {
	yamlData := `
fixture_id: us1
workflow_template:
  id: wf1
user_inputs:
  - role: user
`
	if _, err := decodeData([]byte(yamlData)); err != nil {
		t.Fatalf("decode yaml: %v", err)
	}
}

func TestValidateMissingFields(t *testing.T) {
	fixture := Fixture{}
	if err := fixture.Validate(); err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestApplySeed(t *testing.T) {
	fixture := Fixture{
		UserInputs: []RunMessage{
			{Role: "user"},
			{Role: "assistant", RandomSeed: 99},
		},
	}
	fixture.ApplySeed(42)
	if fixture.UserInputs[0].RandomSeed != 42 {
		t.Fatalf("expected seed to be applied")
	}
	if fixture.UserInputs[1].RandomSeed != 99 {
		t.Fatalf("expected existing seed to remain")
	}
}

func TestDiffAssertion(t *testing.T) {
	if diff := DiffAssertion("outputs[0]", "foo", "foo", 0); diff != nil {
		t.Fatalf("expected nil diff when values match")
	}

	if diff := DiffAssertion("metrics", 1.0, 1.2, 0.1); diff == nil {
		t.Fatalf("expected diff when outside tolerance")
	}

	if diff := DiffAssertion("metrics", 1.0, "nope", 0.5); diff == nil {
		t.Fatalf("expected diff when actual non numeric")
	}
}
