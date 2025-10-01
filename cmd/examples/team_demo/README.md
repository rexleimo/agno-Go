# Team Demo

This example demonstrates the **Team** functionality of agno-Go, showing how multiple agents can collaborate using different coordination modes.

## Team Modes Demonstrated

### 1. Sequential Mode
Agents work one after another, each taking the previous agent's output as input.

**Example**: Content Pipeline
- Researcher → Analyst → Writer
- Each agent builds on the previous agent's work

### 2. Parallel Mode
All agents work simultaneously on the same input.

**Example**: Multi-Perspective Analysis
- Tech Specialist, Business Specialist, Ethics Specialist all analyze independently
- Results are combined

### 3. Leader-Follower Mode
A leader agent delegates tasks to follower agents and synthesizes their results.

**Example**: Project Team
- Leader delegates subtasks
- Calculator Agent and Data Agent execute
- Leader synthesizes final report

### 4. Consensus Mode
Agents discuss over multiple rounds until reaching consensus.

**Example**: Decision Team
- Optimist, Realist, Critic discuss
- Multiple rounds of refinement
- Converge on a consensus

## Running the Demo

```bash
# Set your OpenAI API key
export OPENAI_API_KEY=your-api-key-here

# Run the demo
go run cmd/examples/team_demo/main.go
```

## What You'll See

The demo will run all four team modes sequentially and show:
- How agents coordinate
- Different collaboration patterns
- Final team outputs
- Metadata (agent count, rounds, etc.)

## Key Features

- **Flexible Coordination**: Choose the right mode for your use case
- **Concurrent Execution**: Parallel and consensus modes use goroutines
- **Rich Metadata**: Track which agents participated and how
- **Error Handling**: Graceful handling of agent failures

## Customization Ideas

1. **Add More Agents**: Increase team size for richer collaboration
2. **Different Models**: Use different LLMs for different agents
3. **Tool Integration**: Add specialized tools to agents (calculator, search, etc.)
4. **Custom Instructions**: Fine-tune each agent's personality and expertise
5. **Adjust MaxRounds**: Control consensus iteration depth

## Architecture

```
Team
├── Mode: Sequential/Parallel/LeaderFollower/Consensus
├── Agents: [Agent1, Agent2, ...]
├── Leader: Optional leader agent
└── MaxRounds: For consensus mode

Agent
├── Model: LLM provider
├── Toolkits: Available tools
├── Instructions: Agent personality
└── Memory: Conversation history
```

## Next Steps

- Try modifying agent instructions to create different team dynamics
- Experiment with mixing tools across agents
- Build your own custom team modes
- Combine with Workflow (see workflow_demo) for complex orchestration
