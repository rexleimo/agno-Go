package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agno-agi/agno-go/go/internal/testutil/parity"
	"github.com/agno-agi/agno-go/go/session"
	"github.com/agno-agi/agno-go/go/workflow"
	"sigs.k8s.io/yaml"
)

func TestAgentRuntimeJSONMatchesSchema(t *testing.T) {
	schema, required := loadSchemaProperties(t, "AgentRuntime")
	runtime := AgentRuntime{
		ID:       ID("agent-1"),
		Name:     "Coordinator",
		ModelRef: "openai:gpt-5-mini",
		MemoryPolicy: MemoryPolicy{
			Persist:            true,
			WindowSize:         4,
			SensitiveFiltering: true,
		},
		SessionPolicy: SessionPolicy{
			SessionID:               "session-123",
			OverwriteDBSessionState: true,
		},
	}
	assertJSONContainsKeys(t, runtime, schema, required)
}

func TestWorkflowRunContracts(t *testing.T) {
	spec := loadOpenAPISpec(t)
	fixturePath := filepath.Join("..", "..", "specs", "001-migrate-agno-core", "fixtures", "us1_basic_coordination.yaml")
	fixture, err := parity.Load(fixturePath)
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	requestSchema, requestRequired := getSchemaProperties(t, spec, "WorkflowRunRequest")
	request := workflowRunRequestPayload{
		SessionID: fixture.FixtureID,
		Agents: []AgentRuntime{
			{ID: ID("agent-1"), ModelRef: "openai:gpt-5-mini", MemoryPolicy: MemoryPolicy{}, SessionPolicy: SessionPolicy{}},
		},
		WorkflowManifest: workflow.Workflow{ID: workflow.ID(fixture.WorkflowTemplate.ID), PatternType: workflow.PatternType(fixture.WorkflowTemplate.PatternType)},
		Steps:            []workflowStepInput{{StepID: workflow.StepID("step-1"), AgentID: "agent-1", Input: map[string]any{"query": fixture.UserInputs[0].Content}}},
		SessionState:     map[string]any{"seed": fixture.UserInputs[0].RandomSeed},
	}
	assertJSONContainsKeys(t, request, requestSchema, requestRequired)

	responseSchema, responseRequired := getSchemaProperties(t, spec, "WorkflowRunResult")
	now := time.Now()
	run := workflow.WorkflowRun{
		ID:          "run-id",
		SessionID:   "session-id",
		WorkflowRef: workflow.ID(fixture.WorkflowTemplate.ID),
		PatternType: workflow.PatternSequential,
		Status:      workflow.RunStatusCompleted,
		Steps: []workflow.StepRun{
			{
				ID:      workflow.StepID("search_hackernews"),
				AgentID: "agent-1",
				Status:  workflow.StepStatusCompleted,
				Output: workflow.StepIODigest{
					Summary: "done",
					Payload: map[string]any{"stories": fixture.ToolResponses[0].Outputs},
				},
			},
		},
		ResourcesUsed: workflow.ResourceUsage{
			PromptTokens: 10,
		},
		ReasoningTrace: []workflow.ReasoningStep{{Order: 1, Type: "thought", Text: "analysis"}},
		StartedAt:      now,
		UpdatedAt:      now,
	}
	sessionRecord := session.NewSessionRecordFromSession(&session.Session{
		ID:        session.ID("session-id"),
		Context:   session.UserContext{UserID: "user"},
		Status:    session.StatusCompleted,
		History:   []session.HistoryEntry{{Timestamp: now, Source: "agent", Message: "hello"}},
		Result:    &session.Result{Success: true, Reason: "ok"},
		CreatedAt: now,
		UpdatedAt: now,
	})
	response := workflowRunResultPayload{
		RunID:          run.ID,
		Status:         run.Status,
		Outputs:        run.Steps,
		ReasoningTrace: run.ReasoningTrace,
		Metrics:        run.ResourcesUsed,
		SessionRecord:  sessionRecord,
	}
	assertJSONContainsKeys(t, response, responseSchema, responseRequired)
}

type workflowRunRequestPayload struct {
	SessionID        string                 `json:"sessionId"`
	Agents           []AgentRuntime         `json:"agents"`
	WorkflowManifest workflow.Workflow      `json:"workflowManifest"`
	Steps            []workflowStepInput    `json:"steps"`
	SessionState     map[string]any         `json:"sessionState,omitempty"`
	History          []session.HistoryEntry `json:"history,omitempty"`
}

type workflowStepInput struct {
	StepID  workflow.StepID `json:"stepId"`
	AgentID string          `json:"agentId"`
	Input   map[string]any  `json:"input"`
}

type workflowRunResultPayload struct {
	RunID          string                   `json:"runId"`
	Status         workflow.RunStatus       `json:"status"`
	Outputs        []workflow.StepRun       `json:"outputs"`
	ReasoningTrace []workflow.ReasoningStep `json:"reasoningTrace,omitempty"`
	Metrics        workflow.ResourceUsage   `json:"metrics"`
	SessionRecord  session.SessionRecord    `json:"sessionRecord"`
}

func assertJSONContainsKeys(t *testing.T, v interface{}, schema map[string]interface{}, required []string) {
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal back: %v", err)
	}
	keys := required
	if len(keys) == 0 {
		for key := range schema {
			keys = append(keys, key)
		}
	}
	for _, key := range keys {
		if _, ok := payload[key]; !ok {
			t.Fatalf("missing key %q in payload %v", key, payload)
		}
	}
}

func loadOpenAPISpec(t *testing.T) map[string]interface{} {
	path := filepath.Join("..", "..", "specs", "001-migrate-agno-core", "contracts", "runtime-openapi.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	var spec map[string]interface{}
	if err := yaml.Unmarshal(data, &spec); err != nil {
		t.Fatalf("parse spec: %v", err)
	}
	return spec
}

func loadSchemaProperties(t *testing.T, schema string) (map[string]interface{}, []string) {
	return getSchemaProperties(t, loadOpenAPISpec(t), schema)
}

func getSchemaProperties(t *testing.T, spec map[string]interface{}, schema string) (map[string]interface{}, []string) {
	components, ok := spec["components"].(map[string]interface{})
	if !ok {
		t.Fatalf("spec missing components section")
	}
	schemas, ok := components["schemas"].(map[string]interface{})
	if !ok {
		t.Fatalf("spec missing schemas section")
	}
	entry, ok := schemas[schema].(map[string]interface{})
	if !ok {
		t.Fatalf("schema %q not found", schema)
	}
	props, ok := entry["properties"].(map[string]interface{})
	if !ok {
		t.Fatalf("schema %q missing properties", schema)
	}
	var required []string
	if req, ok := entry["required"].([]interface{}); ok {
		for _, r := range req {
			if rs, ok := r.(string); ok {
				required = append(required, rs)
			}
		}
	}
	return props, required
}
