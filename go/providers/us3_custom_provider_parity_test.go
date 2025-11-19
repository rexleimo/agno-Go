package providers

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

type us3PythonOutput struct {
	Query   string           `json:"query"`
	Results []CustomDocument `json:"results"`
}

func TestUS3CustomProviderParity(t *testing.T) {
	pythonPath, err := exec.LookPath("python")
	if err != nil {
		t.Skip("python executable not found; skipping US3 parity test")
	}

	// Determine repository root by walking up from the providers package
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	// wd is .../agno-Go/go/providers; repo root is two levels up
	repoRoot := filepath.Clean(filepath.Join(wd, "../.."))
	pythonScript := filepath.Join(repoRoot, "agno", "cookbook", "scripts", "us3_custom_provider_parity.py")

	cmd := exec.Command(pythonPath, pythonScript)
	cmd.Dir = filepath.Join(repoRoot, "agno")
	out, err := cmd.Output()
	if err != nil {
		t.Skipf("failed to run us3_custom_provider_parity.py: %v (skipping parity test)", err)
	}

	var py us3PythonOutput
	if err := json.Unmarshal(out, &py); err != nil {
		t.Fatalf("failed to unmarshal python output: %v", err)
	}

	goProvider := DefaultUS3CustomInternalSearch()
	goResults := goProvider.SearchDocuments(py.Query)

	if len(py.Results) != len(goResults) {
		t.Fatalf("result length mismatch: python=%d go=%d", len(py.Results), len(goResults))
	}

	// Compare IDs and titles in order; both implementations use the same
	// static dataset and simple filtering.
	for i := range py.Results {
		if py.Results[i].ID != goResults[i].ID {
			t.Errorf("result %d id mismatch: python=%q go=%q", i, py.Results[i].ID, goResults[i].ID)
		}
		if py.Results[i].Title != goResults[i].Title {
			t.Errorf("result %d title mismatch: python=%q go=%q", i, py.Results[i].Title, goResults[i].Title)
		}
	}
}
