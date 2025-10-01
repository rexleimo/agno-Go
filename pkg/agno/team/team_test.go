package team

import (
	"context"
	"strings"
	"testing"

	"github.com/yourusername/agno-go/pkg/agno/agent"
	"github.com/yourusername/agno-go/pkg/agno/models"
	"github.com/yourusername/agno-go/pkg/agno/types"
)

// MockModel for testing
type MockModel struct {
	models.BaseModel
	InvokeFunc func(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error)
}

func (m *MockModel) Invoke(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
	if m.InvokeFunc != nil {
		return m.InvokeFunc(ctx, req)
	}
	return &types.ModelResponse{
		ID:      "test-response",
		Content: "mock response",
		Model:   "test",
	}, nil
}

func (m *MockModel) InvokeStream(ctx context.Context, req *models.InvokeRequest) (<-chan types.ResponseChunk, error) {
	return nil, nil
}

func createMockAgent(id string, responseContent string) *agent.Agent {
	model := &MockModel{
		BaseModel: models.BaseModel{ID: id, Provider: "mock"},
		InvokeFunc: func(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
			return &types.ModelResponse{
				ID:      "test-" + id,
				Content: responseContent,
				Model:   id,
			}, nil
		},
	}

	ag, _ := agent.New(agent.Config{
		ID:    id,
		Name:  "Agent " + id,
		Model: model,
	})

	return ag
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid sequential team",
			config: Config{
				Name:   "test-team",
				Agents: []*agent.Agent{createMockAgent("a1", "response1")},
				Mode:   ModeSequential,
			},
			wantErr: false,
		},
		{
			name: "empty agents list",
			config: Config{
				Name:   "test-team",
				Agents: []*agent.Agent{},
			},
			wantErr: true,
		},
		{
			name: "leader-follower without leader",
			config: Config{
				Name:   "test-team",
				Agents: []*agent.Agent{createMockAgent("a1", "response1")},
				Mode:   ModeLeaderFollower,
			},
			wantErr: true,
		},
		{
			name: "leader-follower with leader",
			config: Config{
				Name:   "test-team",
				Agents: []*agent.Agent{createMockAgent("a1", "response1")},
				Leader: createMockAgent("leader", "leader response"),
				Mode:   ModeLeaderFollower,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			team, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && team == nil {
				t.Error("New() returned nil team")
			}
		})
	}
}

func TestTeam_RunSequential(t *testing.T) {
	agent1 := createMockAgent("agent1", "step1 complete")
	agent2 := createMockAgent("agent2", "step2 complete")
	agent3 := createMockAgent("agent3", "step3 complete")

	team, err := New(Config{
		Name:   "sequential-team",
		Agents: []*agent.Agent{agent1, agent2, agent3},
		Mode:   ModeSequential,
	})

	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	output, err := team.Run(context.Background(), "start task")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if output == nil {
		t.Fatal("Run() returned nil output")
	}

	// Should have outputs from all 3 agents
	if len(output.AgentOutputs) != 3 {
		t.Errorf("Expected 3 agent outputs, got %d", len(output.AgentOutputs))
	}

	// Final content should be from last agent
	if output.Content != "step3 complete" {
		t.Errorf("Final content = %v, want 'step3 complete'", output.Content)
	}

	// Verify metadata
	if output.Metadata["mode"] != string(ModeSequential) {
		t.Errorf("Metadata mode = %v, want sequential", output.Metadata["mode"])
	}
}

func TestTeam_RunParallel(t *testing.T) {
	agent1 := createMockAgent("agent1", "parallel result 1")
	agent2 := createMockAgent("agent2", "parallel result 2")

	team, err := New(Config{
		Name:   "parallel-team",
		Agents: []*agent.Agent{agent1, agent2},
		Mode:   ModeParallel,
	})

	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	output, err := team.Run(context.Background(), "parallel task")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Should have outputs from both agents
	if len(output.AgentOutputs) != 2 {
		t.Errorf("Expected 2 agent outputs, got %d", len(output.AgentOutputs))
	}

	// Combined content should contain both results
	if !strings.Contains(output.Content, "parallel result 1") || !strings.Contains(output.Content, "parallel result 2") {
		t.Errorf("Combined content missing expected results: %v", output.Content)
	}

	if output.Metadata["mode"] != string(ModeParallel) {
		t.Errorf("Metadata mode = %v, want parallel", output.Metadata["mode"])
	}
}

func TestTeam_RunLeaderFollower(t *testing.T) {
	leader := createMockAgent("leader", "delegation plan")
	follower1 := createMockAgent("follower1", "task1 done")
	follower2 := createMockAgent("follower2", "task2 done")

	team, err := New(Config{
		Name:   "leader-team",
		Agents: []*agent.Agent{follower1, follower2},
		Leader: leader,
		Mode:   ModeLeaderFollower,
	})

	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	output, err := team.Run(context.Background(), "complex task")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Should have plan + follower outputs + synthesis
	if len(output.AgentOutputs) < 3 {
		t.Errorf("Expected at least 3 outputs (plan + followers + synthesis), got %d", len(output.AgentOutputs))
	}

	if output.Metadata["mode"] != string(ModeLeaderFollower) {
		t.Errorf("Metadata mode = %v, want leader_follower", output.Metadata["mode"])
	}

	if output.Metadata["leader_id"] != "leader" {
		t.Errorf("Metadata leader_id = %v, want 'leader'", output.Metadata["leader_id"])
	}
}

func TestTeam_RunConsensus(t *testing.T) {
	agent1 := createMockAgent("agent1", "consensus view 1")
	agent2 := createMockAgent("agent2", "consensus view 2")

	team, err := New(Config{
		Name:      "consensus-team",
		Agents:    []*agent.Agent{agent1, agent2},
		Mode:      ModeConsensus,
		MaxRounds: 2,
	})

	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	output, err := team.Run(context.Background(), "reach consensus")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Should have outputs from multiple rounds
	if len(output.AgentOutputs) < 2 {
		t.Errorf("Expected at least 2 outputs from consensus rounds, got %d", len(output.AgentOutputs))
	}

	if output.Metadata["mode"] != string(ModeConsensus) {
		t.Errorf("Metadata mode = %v, want consensus", output.Metadata["mode"])
	}
}

func TestTeam_RunEmptyInput(t *testing.T) {
	agent1 := createMockAgent("agent1", "response")

	team, err := New(Config{
		Name:   "test-team",
		Agents: []*agent.Agent{agent1},
		Mode:   ModeSequential,
	})

	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	_, err = team.Run(context.Background(), "")
	if err == nil {
		t.Error("Run() with empty input should return error")
	}
}

func TestTeam_AddAgent(t *testing.T) {
	agent1 := createMockAgent("agent1", "response1")

	team, err := New(Config{
		Name:   "test-team",
		Agents: []*agent.Agent{agent1},
	})

	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	initialCount := len(team.GetAgents())

	agent2 := createMockAgent("agent2", "response2")
	team.AddAgent(agent2)

	if len(team.GetAgents()) != initialCount+1 {
		t.Errorf("AddAgent() failed, expected %d agents, got %d", initialCount+1, len(team.GetAgents()))
	}
}

func TestTeam_RemoveAgent(t *testing.T) {
	agent1 := createMockAgent("agent1", "response1")
	agent2 := createMockAgent("agent2", "response2")

	team, err := New(Config{
		Name:   "test-team",
		Agents: []*agent.Agent{agent1, agent2},
	})

	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	team.RemoveAgent("agent1")

	agents := team.GetAgents()
	if len(agents) != 1 {
		t.Errorf("RemoveAgent() failed, expected 1 agent, got %d", len(agents))
	}

	if agents[0].ID != "agent2" {
		t.Errorf("RemoveAgent() removed wrong agent, got %s", agents[0].ID)
	}
}

func TestTeam_GetAgents(t *testing.T) {
	agent1 := createMockAgent("agent1", "response1")
	agent2 := createMockAgent("agent2", "response2")

	team, err := New(Config{
		Name:   "test-team",
		Agents: []*agent.Agent{agent1, agent2},
	})

	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	agents := team.GetAgents()

	if len(agents) != 2 {
		t.Errorf("GetAgents() returned %d agents, want 2", len(agents))
	}

	// Verify it's a copy (modifying returned slice shouldn't affect team)
	agents[0] = nil
	if team.Agents[0] == nil {
		t.Error("GetAgents() should return a copy, not the original slice")
	}
}

func TestTeam_DefaultValues(t *testing.T) {
	agent1 := createMockAgent("agent1", "response1")

	team, err := New(Config{
		Agents: []*agent.Agent{agent1},
	})

	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	// Test default values
	if team.Mode != ModeSequential {
		t.Errorf("Default mode = %v, want sequential", team.Mode)
	}

	if team.MaxRounds != 3 {
		t.Errorf("Default MaxRounds = %v, want 3", team.MaxRounds)
	}

	if team.ID == "" {
		t.Error("Default ID should not be empty")
	}

	if team.logger == nil {
		t.Error("Default logger should not be nil")
	}
}

func TestTeam_AgentError(t *testing.T) {
	// Create agent that always returns error
	errorModel := &MockModel{
		BaseModel: models.BaseModel{ID: "error", Provider: "mock"},
		InvokeFunc: func(ctx context.Context, req *models.InvokeRequest) (*types.ModelResponse, error) {
			return nil, types.NewError(types.ErrCodeUnknown, "intentional error", nil)
		},
	}

	errorAgent, _ := agent.New(agent.Config{
		ID:    "error-agent",
		Model: errorModel,
	})

	team, err := New(Config{
		Name:   "error-team",
		Agents: []*agent.Agent{errorAgent},
		Mode:   ModeSequential,
	})

	if err != nil {
		t.Fatalf("Failed to create team: %v", err)
	}

	_, err = team.Run(context.Background(), "test")
	if err == nil {
		t.Error("Run() should return error when agent fails")
	}
}
