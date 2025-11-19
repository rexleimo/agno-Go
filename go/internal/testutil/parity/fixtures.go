package parity

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"sigs.k8s.io/yaml"
)

// Fixture represents the structured parity test input defined in specs.
type Fixture struct {
	FixtureID        string              `json:"fixture_id" yaml:"fixture_id"`
	Description      string              `json:"description" yaml:"description"`
	WorkflowTemplate WorkflowTemplate    `json:"workflow_template" yaml:"workflow_template"`
	UserInputs       []RunMessage        `json:"user_inputs" yaml:"user_inputs"`
	ToolResponses    []ToolResponse      `json:"tool_responses" yaml:"tool_responses"`
	Expected         []ExpectedAssertion `json:"expected_assertions" yaml:"expected_assertions"`
}

// WorkflowTemplate sketches the static workflow metadata embedded in fixtures.
type WorkflowTemplate struct {
	ID          string            `json:"id" yaml:"id"`
	Version     string            `json:"version" yaml:"version"`
	PatternType string            `json:"pattern_type" yaml:"pattern_type"`
	EntryPoints []string          `json:"entry_points" yaml:"entry_points"`
	Agents      []map[string]any  `json:"agents" yaml:"agents"`
	Metadata    map[string]string `json:"metadata" yaml:"metadata"`
}

// RunMessage mirrors contracts/RunMessage for parity fixtures.
type RunMessage struct {
	Role       string                 `json:"role" yaml:"role"`
	Content    string                 `json:"content" yaml:"content"`
	References []string               `json:"references" yaml:"references"`
	ToolResult map[string]any         `json:"tool_result" yaml:"tool_result"`
	Timestamp  time.Time              `json:"timestamp" yaml:"timestamp"`
	RandomSeed int64                  `json:"seed" yaml:"seed"`
	Metadata   map[string]interface{} `json:"metadata" yaml:"metadata"`
}

// ToolResponse captures mocked tool output loaded by the parity tests.
type ToolResponse struct {
	ToolName string                 `json:"tool_name" yaml:"tool_name"`
	RunID    string                 `json:"run_id" yaml:"run_id"`
	Outputs  map[string]interface{} `json:"outputs" yaml:"outputs"`
	Metadata map[string]interface{} `json:"metadata" yaml:"metadata"`
}

// ExpectedAssertion defines a single parity assertion.
type ExpectedAssertion struct {
	Type      string      `json:"type" yaml:"type"`
	Path      string      `json:"path" yaml:"path"`
	Expected  interface{} `json:"expected" yaml:"expected"`
	Tolerance float64     `json:"tolerance" yaml:"tolerance"`
}

// Load loads a fixture from YAML or JSON. It relies on an external YAML parser
// to avoid adding heavy dependencies; the CLI should convert YAML to JSON
// before invoking this helper.
func Load(path string) (*Fixture, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	return decodeData(data)
}

func decodeData(data []byte) (*Fixture, error) {
	var fixture Fixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		converted, convErr := yaml.YAMLToJSON(data)
		if convErr != nil {
			return nil, fmt.Errorf("decode fixture: %w", err)
		}
		if err := json.Unmarshal(converted, &fixture); err != nil {
			return nil, fmt.Errorf("decode fixture: %w", err)
		}
	}
	if err := fixture.Validate(); err != nil {
		return nil, err
	}
	return &fixture, nil
}

// Validate ensures the fixture has the mandatory fields.
func (f Fixture) Validate() error {
	var missing []string
	if strings.TrimSpace(f.FixtureID) == "" {
		missing = append(missing, "fixture_id")
	}
	if strings.TrimSpace(f.WorkflowTemplate.ID) == "" {
		missing = append(missing, "workflow_template.id")
	}
	if len(f.UserInputs) == 0 {
		missing = append(missing, "user_inputs")
	}
	if len(missing) > 0 {
		return fmt.Errorf("invalid fixture: missing fields %s", strings.Join(missing, ", "))
	}
	return nil
}

// ApplySeed sets deterministic seeds on user inputs when the fixture specifies a
// shared seed. Seeds are optional; when absent, the input remains unchanged.
func (f *Fixture) ApplySeed(seed int64) {
	if seed == 0 {
		return
	}
	for i := range f.UserInputs {
		if f.UserInputs[i].RandomSeed == 0 {
			f.UserInputs[i].RandomSeed = seed
		}
	}
}

// DiffResult captures differences between Go/Python outputs for later reporting.
type DiffResult struct {
	Path      string      `json:"path"`
	PythonVal interface{} `json:"python_value"`
	GoVal     interface{} `json:"go_value"`
	Message   string      `json:"message"`
	Severity  string      `json:"severity"`
}

// DiffAssertion compares actual values against expected assertions.
func DiffAssertion(path string, expected, actual interface{}, tolerance float64) *DiffResult {
	if tolerance > 0 {
		switch exp := expected.(type) {
		case float64:
			act, ok := toFloat(actual)
			if !ok {
				return &DiffResult{
					Path:      path,
					PythonVal: expected,
					GoVal:     actual,
					Message:   "actual is not numeric for tolerance comparison",
					Severity:  "error",
				}
			}
			if delta := act - exp; delta > tolerance || delta < -tolerance {
				return &DiffResult{
					Path:      path,
					PythonVal: expected,
					GoVal:     actual,
					Message:   fmt.Sprintf("value out of tolerance (±%f)", tolerance),
					Severity:  "fail",
				}
			}
			return nil
		}
	}
	if !isEqual(expected, actual) {
		return &DiffResult{
			Path:      path,
			PythonVal: expected,
			GoVal:     actual,
			Message:   "values differ",
			Severity:  "fail",
		}
	}
	return nil
}

func toFloat(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}

func isEqual(expected, actual interface{}) bool {
	return reflect.DeepEqual(expected, actual)
}

// LoadJSONFixture is a helper to load a fixture from a JSON byte slice. Useful
// in tests when YAML parsing is stubbed out.
func LoadJSONFixture(data []byte) (*Fixture, error) {
	return decodeData(data)
}
