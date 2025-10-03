# Team API Reference

## team.New

Create a new team of agents.

**Signature:**
```go
func New(config Config) (*Team, error)
```

**Parameters:**
```go
type Config struct {
    // Required
    Agents []*agent.Agent // Team members

    // Optional
    Name   string         // Team name (default: "Team")
    Mode   CoordinationMode // Coordination mode (default: Sequential)
    Leader *agent.Agent   // Leader (for LeaderFollower mode)
}

type CoordinationMode string

const (
    ModeSequential     CoordinationMode = "sequential"
    ModeParallel       CoordinationMode = "parallel"
    ModeLeaderFollower CoordinationMode = "leader-follower"
    ModeConsensus      CoordinationMode = "consensus"
)
```

**Returns:**
- `*Team`: Created team instance
- `error`: Error if agents list is empty or invalid config

**Example:**
```go
tm, err := team.New(team.Config{
    Name:   "Research Team",
    Agents: []*agent.Agent{researcher, writer, editor},
    Mode:   team.ModeSequential,
})
```

## Team.Run

Execute the team with input.

**Signature:**
```go
func (t *Team) Run(ctx context.Context, input string) (*RunOutput, error)
```

**Behavior by Mode:**

- **Sequential**: Agents execute one after another, output feeds to next
- **Parallel**: All agents execute simultaneously, results combined
- **LeaderFollower**: Leader delegates tasks to followers
- **Consensus**: Agents discuss until reaching agreement

**Example:**
```go
output, err := tm.Run(context.Background(), "Write an article about AI")
```

## Team.AddAgent / RemoveAgent

Manage team members dynamically.

**Signatures:**
```go
func (t *Team) AddAgent(ag *agent.Agent)
func (t *Team) RemoveAgent(name string) error
func (t *Team) GetAgents() []*agent.Agent
```

**Example:**
```go
tm.AddAgent(newAgent)
tm.RemoveAgent("OldAgent")
agents := tm.GetAgents()
```
