# Team APIリファレンス

## team.New

新しいエージェントチームを作成します。

**シグネチャ:**
```go
func New(config Config) (*Team, error)
```

**パラメータ:**
```go
type Config struct {
    // 必須
    Agents []*agent.Agent // チームメンバー

    // オプション
    Name   string         // チーム名 (デフォルト: "Team")
    Mode   CoordinationMode // 協調モード (デフォルト: Sequential)
    Leader *agent.Agent   // リーダー (LeaderFollowerモード用)
}

type CoordinationMode string

const (
    ModeSequential     CoordinationMode = "sequential"
    ModeParallel       CoordinationMode = "parallel"
    ModeLeaderFollower CoordinationMode = "leader-follower"
    ModeConsensus      CoordinationMode = "consensus"
)
```

**戻り値:**
- `*Team`: 作成されたチームインスタンス
- `error`: エージェントリストが空または設定が無効な場合のエラー

**例:**
```go
tm, err := team.New(team.Config{
    Name:   "Research Team",
    Agents: []*agent.Agent{researcher, writer, editor},
    Mode:   team.ModeSequential,
})
```

## Team.Run

入力でチームを実行します。

**シグネチャ:**
```go
func (t *Team) Run(ctx context.Context, input string) (*RunOutput, error)
```

**モード別の動作:**

- **Sequential**: エージェントが順番に実行され、出力が次のエージェントに渡される
- **Parallel**: すべてのエージェントが同時に実行され、結果が結合される
- **LeaderFollower**: リーダーがフォロワーにタスクを委任する
- **Consensus**: エージェントが合意に達するまで議論する

**例:**
```go
output, err := tm.Run(context.Background(), "Write an article about AI")
```

## Team.AddAgent / RemoveAgent

チームメンバーを動的に管理します。

**シグネチャ:**
```go
func (t *Team) AddAgent(ag *agent.Agent)
func (t *Team) RemoveAgent(name string) error
func (t *Team) GetAgents() []*agent.Agent
```

**例:**
```go
tm.AddAgent(newAgent)
tm.RemoveAgent("OldAgent")
agents := tm.GetAgents()
```
