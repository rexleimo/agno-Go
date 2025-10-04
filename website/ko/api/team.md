# Team API 레퍼런스

## team.New

에이전트 팀을 생성합니다.

**함수 시그니처:**
```go
func New(config Config) (*Team, error)
```

**매개변수:**
```go
type Config struct {
    // 필수
    Agents []*agent.Agent // 팀 멤버

    // 선택 사항
    Name   string         // 팀 이름 (기본값: "Team")
    Mode   CoordinationMode // 협업 모드 (기본값: Sequential)
    Leader *agent.Agent   // 리더 (LeaderFollower 모드용)
}

type CoordinationMode string

const (
    ModeSequential     CoordinationMode = "sequential"
    ModeParallel       CoordinationMode = "parallel"
    ModeLeaderFollower CoordinationMode = "leader-follower"
    ModeConsensus      CoordinationMode = "consensus"
)
```

**반환값:**
- `*Team`: 생성된 팀 인스턴스
- `error`: 에이전트 목록이 비어있거나 설정이 유효하지 않은 경우 에러

**예제:**
```go
tm, err := team.New(team.Config{
    Name:   "Research Team",
    Agents: []*agent.Agent{researcher, writer, editor},
    Mode:   team.ModeSequential,
})
```

## Team.Run

입력으로 팀을 실행합니다.

**함수 시그니처:**
```go
func (t *Team) Run(ctx context.Context, input string) (*RunOutput, error)
```

**모드별 동작:**

- **Sequential**: 에이전트가 순차적으로 실행되며, 출력이 다음 에이전트로 전달됨
- **Parallel**: 모든 에이전트가 동시에 실행되며, 결과가 결합됨
- **LeaderFollower**: 리더가 팔로워에게 작업을 위임함
- **Consensus**: 에이전트들이 합의에 도달할 때까지 토론함

**예제:**
```go
output, err := tm.Run(context.Background(), "Write an article about AI")
```

## Team.AddAgent / RemoveAgent

팀 멤버를 동적으로 관리합니다.

**함수 시그니처:**
```go
func (t *Team) AddAgent(ag *agent.Agent)
func (t *Team) RemoveAgent(name string) error
func (t *Team) GetAgents() []*agent.Agent
```

**예제:**
```go
tm.AddAgent(newAgent)
tm.RemoveAgent("OldAgent")
agents := tm.GetAgents()
```
