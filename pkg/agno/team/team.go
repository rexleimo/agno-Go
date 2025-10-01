package team

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/yourusername/agno-go/pkg/agno/agent"
	"github.com/yourusername/agno-go/pkg/agno/types"
)

// Team represents a group of agents working together
type Team struct {
	ID          string
	Name        string
	Agents      []*agent.Agent
	Leader      *agent.Agent // Optional team leader
	Mode        TeamMode
	MaxRounds   int // Maximum coordination rounds
	logger      *slog.Logger
	mu          sync.RWMutex
	taskResults map[string]*TaskResult
}

// TeamMode defines how agents collaborate
type TeamMode string

const (
	// ModeSequential - agents work one after another
	ModeSequential TeamMode = "sequential"
	// ModeParallel - all agents work simultaneously
	ModeParallel TeamMode = "parallel"
	// ModeLeaderFollower - leader delegates tasks to followers
	ModeLeaderFollower TeamMode = "leader_follower"
	// ModeConsensus - agents discuss until reaching consensus
	ModeConsensus TeamMode = "consensus"
)

// Config contains team configuration
type Config struct {
	ID        string
	Name      string
	Agents    []*agent.Agent
	Leader    *agent.Agent
	Mode      TeamMode
	MaxRounds int
	Logger    *slog.Logger
}

// TaskResult holds the result of an agent's task execution
type TaskResult struct {
	AgentID string
	Content string
	Error   error
}

// New creates a new team
func New(config Config) (*Team, error) {
	if len(config.Agents) == 0 {
		return nil, types.NewInvalidConfigError("team must have at least one agent", nil)
	}

	if config.ID == "" {
		config.ID = fmt.Sprintf("team-%s", config.Name)
	}

	if config.Name == "" {
		config.Name = config.ID
	}

	if config.Mode == "" {
		config.Mode = ModeSequential
	}

	if config.MaxRounds <= 0 {
		config.MaxRounds = 3
	}

	if config.Logger == nil {
		config.Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	// Validate leader-follower mode
	if config.Mode == ModeLeaderFollower && config.Leader == nil {
		return nil, types.NewInvalidConfigError("leader_follower mode requires a leader agent", nil)
	}

	return &Team{
		ID:          config.ID,
		Name:        config.Name,
		Agents:      config.Agents,
		Leader:      config.Leader,
		Mode:        config.Mode,
		MaxRounds:   config.MaxRounds,
		logger:      config.Logger,
		taskResults: make(map[string]*TaskResult),
	}, nil
}

// RunOutput contains the team execution result
type RunOutput struct {
	Content      string                 `json:"content"`
	AgentOutputs []*AgentOutput         `json:"agent_outputs"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// AgentOutput contains output from a single agent
type AgentOutput struct {
	AgentID string `json:"agent_id"`
	Content string `json:"content"`
}

// Run executes the team with the given input
func (t *Team) Run(ctx context.Context, input string) (*RunOutput, error) {
	if input == "" {
		return nil, types.NewInvalidInputError("input cannot be empty", nil)
	}

	t.logger.Info("team run started", "team_id", t.ID, "mode", t.Mode, "agents", len(t.Agents))

	var output *RunOutput
	var err error

	switch t.Mode {
	case ModeSequential:
		output, err = t.runSequential(ctx, input)
	case ModeParallel:
		output, err = t.runParallel(ctx, input)
	case ModeLeaderFollower:
		output, err = t.runLeaderFollower(ctx, input)
	case ModeConsensus:
		output, err = t.runConsensus(ctx, input)
	default:
		return nil, types.NewInvalidConfigError(fmt.Sprintf("unknown team mode: %s", t.Mode), nil)
	}

	if err != nil {
		t.logger.Error("team run failed", "error", err)
		return nil, err
	}

	t.logger.Info("team run completed", "team_id", t.ID)
	return output, nil
}

// runSequential executes agents one after another, passing output as input to next
func (t *Team) runSequential(ctx context.Context, input string) (*RunOutput, error) {
	currentInput := input
	agentOutputs := make([]*AgentOutput, 0, len(t.Agents))

	for i, ag := range t.Agents {
		t.logger.Info("running agent", "agent_id", ag.ID, "sequence", i+1)

		result, err := ag.Run(ctx, currentInput)
		if err != nil {
			return nil, types.NewError(types.ErrCodeUnknown, fmt.Sprintf("agent %s failed", ag.ID), err)
		}

		agentOutputs = append(agentOutputs, &AgentOutput{
			AgentID: ag.ID,
			Content: result.Content,
		})

		// Pass output to next agent
		currentInput = result.Content
	}

	// Last agent's output is the final result
	finalContent := ""
	if len(agentOutputs) > 0 {
		finalContent = agentOutputs[len(agentOutputs)-1].Content
	}

	return &RunOutput{
		Content:      finalContent,
		AgentOutputs: agentOutputs,
		Metadata: map[string]interface{}{
			"mode":        string(ModeSequential),
			"agent_count": len(t.Agents),
		},
	}, nil
}

// runParallel executes all agents simultaneously
func (t *Team) runParallel(ctx context.Context, input string) (*RunOutput, error) {
	var wg sync.WaitGroup
	results := make(chan *AgentOutput, len(t.Agents))
	errors := make(chan error, len(t.Agents))

	for _, ag := range t.Agents {
		wg.Add(1)
		go func(a *agent.Agent) {
			defer wg.Done()

			t.logger.Info("running agent in parallel", "agent_id", a.ID)

			result, err := a.Run(ctx, input)
			if err != nil {
				errors <- types.NewError(types.ErrCodeUnknown, fmt.Sprintf("agent %s failed", a.ID), err)
				return
			}

			results <- &AgentOutput{
				AgentID: a.ID,
				Content: result.Content,
			}
		}(ag)
	}

	wg.Wait()
	close(results)
	close(errors)

	// Check for errors
	if len(errors) > 0 {
		return nil, <-errors
	}

	// Collect all results
	agentOutputs := make([]*AgentOutput, 0, len(t.Agents))
	for output := range results {
		agentOutputs = append(agentOutputs, output)
	}

	// Combine outputs (simple concatenation for now)
	combinedContent := ""
	for i, output := range agentOutputs {
		if i > 0 {
			combinedContent += "\n\n"
		}
		combinedContent += fmt.Sprintf("[%s]: %s", output.AgentID, output.Content)
	}

	return &RunOutput{
		Content:      combinedContent,
		AgentOutputs: agentOutputs,
		Metadata: map[string]interface{}{
			"mode":        string(ModeParallel),
			"agent_count": len(t.Agents),
		},
	}, nil
}

// runLeaderFollower uses leader to delegate tasks and synthesize results
func (t *Team) runLeaderFollower(ctx context.Context, input string) (*RunOutput, error) {
	// Step 1: Leader plans and delegates
	t.logger.Info("leader planning", "leader_id", t.Leader.ID)

	planPrompt := fmt.Sprintf(`You are a team leader. Break down this task into subtasks for your team members:
Task: %s

Respond with a JSON array of subtasks, one for each team member.
Example: ["subtask1", "subtask2", "subtask3"]`, input)

	planResult, err := t.Leader.Run(ctx, planPrompt)
	if err != nil {
		return nil, types.NewError(types.ErrCodeUnknown, "leader planning failed", err)
	}

	// For simplicity, assign the same task to all followers
	// In a real implementation, parse planResult.Content to extract subtasks
	var wg sync.WaitGroup
	results := make(chan *AgentOutput, len(t.Agents))
	errors := make(chan error, len(t.Agents))

	// Step 2: Followers execute
	for _, ag := range t.Agents {
		wg.Add(1)
		go func(a *agent.Agent) {
			defer wg.Done()

			t.logger.Info("follower executing", "agent_id", a.ID)

			result, err := a.Run(ctx, input) // Use original input for now
			if err != nil {
				errors <- types.NewError(types.ErrCodeUnknown, fmt.Sprintf("agent %s failed", a.ID), err)
				return
			}

			results <- &AgentOutput{
				AgentID: a.ID,
				Content: result.Content,
			}
		}(ag)
	}

	wg.Wait()
	close(results)
	close(errors)

	if len(errors) > 0 {
		return nil, <-errors
	}

	// Collect follower outputs
	followerOutputs := make([]*AgentOutput, 0, len(t.Agents))
	combinedResults := ""
	for output := range results {
		followerOutputs = append(followerOutputs, output)
		combinedResults += fmt.Sprintf("\n[%s]: %s", output.AgentID, output.Content)
	}

	// Step 3: Leader synthesizes results
	synthesisPrompt := fmt.Sprintf(`You are a team leader. Synthesize these team member outputs into a final answer:

Original Task: %s

Team Outputs:%s

Provide a comprehensive final answer.`, input, combinedResults)

	finalResult, err := t.Leader.Run(ctx, synthesisPrompt)
	if err != nil {
		return nil, types.NewError(types.ErrCodeUnknown, "leader synthesis failed", err)
	}

	// Include leader outputs
	allOutputs := append([]*AgentOutput{{
		AgentID: t.Leader.ID + "_plan",
		Content: planResult.Content,
	}}, followerOutputs...)
	allOutputs = append(allOutputs, &AgentOutput{
		AgentID: t.Leader.ID + "_final",
		Content: finalResult.Content,
	})

	return &RunOutput{
		Content:      finalResult.Content,
		AgentOutputs: allOutputs,
		Metadata: map[string]interface{}{
			"mode":        string(ModeLeaderFollower),
			"leader_id":   t.Leader.ID,
			"agent_count": len(t.Agents),
		},
	}, nil
}

// runConsensus runs multiple rounds until agents reach consensus
func (t *Team) runConsensus(ctx context.Context, input string) (*RunOutput, error) {
	allOutputs := make([]*AgentOutput, 0)
	previousOutputs := ""

	for round := 1; round <= t.MaxRounds; round++ {
		t.logger.Info("consensus round", "round", round, "max_rounds", t.MaxRounds)

		roundPrompt := input
		if round > 1 {
			roundPrompt = fmt.Sprintf(`Original task: %s

Previous round outputs:
%s

Consider the previous outputs and provide your refined answer. If you agree with a previous answer, state so clearly.`, input, previousOutputs)
		}

		// Run all agents in parallel
		var wg sync.WaitGroup
		results := make(chan *AgentOutput, len(t.Agents))
		errors := make(chan error, len(t.Agents))

		for _, ag := range t.Agents {
			wg.Add(1)
			go func(a *agent.Agent) {
				defer wg.Done()

				result, err := a.Run(ctx, roundPrompt)
				if err != nil {
					errors <- err
					return
				}

				results <- &AgentOutput{
					AgentID: fmt.Sprintf("%s_round%d", a.ID, round),
					Content: result.Content,
				}
			}(ag)
		}

		wg.Wait()
		close(results)
		close(errors)

		if len(errors) > 0 {
			return nil, <-errors
		}

		// Collect round outputs
		roundOutputs := make([]*AgentOutput, 0, len(t.Agents))
		previousOutputs = ""
		for output := range results {
			roundOutputs = append(roundOutputs, output)
			allOutputs = append(allOutputs, output)
			previousOutputs += fmt.Sprintf("\n[%s]: %s", output.AgentID, output.Content)
		}

		// Simple consensus check: if all outputs are similar length (placeholder logic)
		// In real implementation, use semantic similarity or voting
		if round >= 2 {
			// Consider consensus reached if we've done at least 2 rounds
			break
		}
	}

	// Use last round's first agent output as final
	finalContent := ""
	if len(allOutputs) > 0 {
		finalContent = allOutputs[len(allOutputs)-1].Content
	}

	return &RunOutput{
		Content:      finalContent,
		AgentOutputs: allOutputs,
		Metadata: map[string]interface{}{
			"mode":        string(ModeConsensus),
			"rounds":      t.MaxRounds,
			"agent_count": len(t.Agents),
		},
	}, nil
}

// AddAgent adds an agent to the team
func (t *Team) AddAgent(ag *agent.Agent) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Agents = append(t.Agents, ag)
}

// RemoveAgent removes an agent from the team
func (t *Team) RemoveAgent(agentID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, ag := range t.Agents {
		if ag.ID == agentID {
			t.Agents = append(t.Agents[:i], t.Agents[i+1:]...)
			return
		}
	}
}

// GetAgents returns all agents in the team
func (t *Team) GetAgents() []*agent.Agent {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Return a copy to prevent external modification
	agents := make([]*agent.Agent, len(t.Agents))
	copy(agents, t.Agents)
	return agents
}
