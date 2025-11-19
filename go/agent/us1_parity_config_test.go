package agent

import (
	"encoding/json"
	"os/exec"
	"testing"
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
