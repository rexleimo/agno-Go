# Team API 参考 / Team API Reference

## team.New

创建一个新的智能体团队。/ Create a new team of agents.

**签名 / Signature:**
```go
func New(config Config) (*Team, error)
```

**参数 / Parameters:**
```go
type Config struct {
    // 必需 / Required
    Agents []*agent.Agent // 团队成员 / Team members

    // 可选 / Optional
    Name   string         // 团队名称 (默认: "Team") / Team name (default: "Team")
    Mode   CoordinationMode // 协作模式 (默认: Sequential) / Coordination mode (default: Sequential)
    Leader *agent.Agent   // 领导者 (用于 LeaderFollower 模式) / Leader (for LeaderFollower mode)
}

type CoordinationMode string

const (
    ModeSequential     CoordinationMode = "sequential"
    ModeParallel       CoordinationMode = "parallel"
    ModeLeaderFollower CoordinationMode = "leader-follower"
    ModeConsensus      CoordinationMode = "consensus"
)
```

**返回值 / Returns:**
- `*Team`: 创建的团队实例 / Created team instance
- `error`: 如果智能体列表为空或配置无效则返回错误 / Error if agents list is empty or invalid config

**示例 / Example:**
```go
tm, err := team.New(team.Config{
    Name:   "Research Team",
    Agents: []*agent.Agent{researcher, writer, editor},
    Mode:   team.ModeSequential,
})
```

## Team.Run

使用输入执行团队。/ Execute the team with input.

**签名 / Signature:**
```go
func (t *Team) Run(ctx context.Context, input string) (*RunOutput, error)
```

**按模式的行为 / Behavior by Mode:**

- **Sequential**: 智能体依次执行,输出传递给下一个 / Agents execute one after another, output feeds to next
- **Parallel**: 所有智能体同时执行,结果合并 / All agents execute simultaneously, results combined
- **LeaderFollower**: 领导者将任务委派给跟随者 / Leader delegates tasks to followers
- **Consensus**: 智能体讨论直到达成一致 / Agents discuss until reaching agreement

**示例 / Example:**
```go
output, err := tm.Run(context.Background(), "Write an article about AI")
```

## Team.AddAgent / RemoveAgent

动态管理团队成员。/ Manage team members dynamically.

**签名 / Signatures:**
```go
func (t *Team) AddAgent(ag *agent.Agent)
func (t *Team) RemoveAgent(name string) error
func (t *Team) GetAgents() []*agent.Agent
```

**示例 / Example:**
```go
tm.AddAgent(newAgent)
tm.RemoveAgent("OldAgent")
agents := tm.GetAgents()
```
