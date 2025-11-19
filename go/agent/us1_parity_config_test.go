package agent

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/agno-agi/agno-go/go/internal/testutil/parity"
)

type parityScenario struct {
	ScenarioID  string `json:"scenario_id"`
	PythonEntry string `json:"python_entry"`
	GoEntry     string `json:"go_entry"`
}

type parityConfig struct {
	RunID     string           `json:"run_id"`
	Scenarios []parityScenario `json:"scenarios"`
}

func TestUS1ParityConfigScript(t *testing.T) {
	cmd := exec.Command("../../scripts/us1_parity_run.sh")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to run us1_parity_run.sh: %v", err)
	}

	var cfg parityConfig
	if err := json.Unmarshal(out, &cfg); err != nil {
		t.Fatalf("failed to unmarshal parity config: %v", err)
	}

	if cfg.RunID == "" {
		t.Fatalf("run_id must not be empty")
	}
	if len(cfg.Scenarios) != 1 {
		t.Fatalf("expected 1 scenario, got %d", len(cfg.Scenarios))
	}

	sc := cfg.Scenarios[0]
	if sc.ScenarioID != "teams-basic-coordination-us1" {
		t.Errorf("unexpected scenario_id: %q", sc.ScenarioID)
	}
	if sc.GoEntry != "github.com/agno-agi/agno-go/go/agent.RunUS1Example" {
		t.Errorf("unexpected go_entry: %q", sc.GoEntry)
	}
	if sc.PythonEntry == "" {
		t.Errorf("python_entry must not be empty")
	}
}

func TestUS1ParityFixtureIntegration(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "specs", "001-migrate-agno-core", "fixtures", "us1_basic_coordination.yaml")
	fixture, err := parity.Load(fixturePath)
	if err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}
	if fixture.WorkflowTemplate.ID == "" {
		t.Fatalf("fixture missing workflow_template.id: %+v", fixture.WorkflowTemplate)
	}
	if len(fixture.UserInputs) == 0 {
		t.Fatalf("fixture missing user inputs")
	}

	input := US1Input{Query: fixture.UserInputs[0].Content}
	result, err := RunUS1Example(input)
	if err != nil {
		t.Fatalf("RunUS1Example: %v", err)
	}
	if result.Result == nil {
		t.Fatalf("expected placeholder result")
	}
	if diff := parity.DiffAssertion("workflow_id", fixture.WorkflowTemplate.ID, result.Result["workflow_id"], 0); diff != nil {
		t.Fatalf("workflow mismatch: %+v", diff)
	}
}
