# Workflow Demo

This example demonstrates the **Workflow** functionality of agno-Go, showing how to build complex multi-step processes with control flow.

## Workflow Primitives Demonstrated

### 1. Step - Basic Sequential Execution
The fundamental building block that executes an agent.

**Example**: Content Pipeline
- Research → Analyze → Write
- Each step processes the previous step's output

### 2. Condition - Conditional Branching
Branches execution based on a condition function.

**Example**: Sentiment Handler
- Classify sentiment
- If positive → Thank user
- If negative → Apologize and offer help

### 3. Loop - Iterative Execution
Repeats a step until a condition is met.

**Example**: Iterative Refinement
- Refine text 3 times
- Each iteration improves the previous output

### 4. Parallel - Concurrent Execution
Executes multiple steps simultaneously.

**Example**: Multi-Perspective Analysis
- Tech, Business, Ethics agents analyze in parallel
- Results are collected and merged

### 5. Router - Dynamic Routing
Selects execution path based on runtime data.

**Example**: Smart Task Router
- Determine task type (calculation vs general)
- Route to appropriate specialized agent

## Running the Demo

```bash
# Set your OpenAI API key
export OPENAI_API_KEY=your-api-key-here

# Run the demo
go run cmd/examples/workflow_demo/main.go
```

## What You'll See

The demo runs 5 different workflow patterns:
1. Sequential pipeline (research → analyze → write)
2. Conditional branching (positive/negative sentiment)
3. Loop-based refinement (3 iterations)
4. Parallel analysis (tech/biz/ethics)
5. Dynamic routing (calculation vs general task)

## Key Features

- **Control Flow**: Full control over execution order
- **Composability**: Nest primitives to build complex logic
- **Execution Context**: Share data between steps
- **Parallel Execution**: Run independent steps concurrently
- **Dynamic Routing**: Make runtime decisions

## Architecture

```
Workflow
└── Steps: []Node
    ├── Step (executes agent)
    ├── Condition (if-else branching)
    ├── Loop (repeat until condition)
    ├── Parallel (concurrent execution)
    └── Router (dynamic routing)

ExecutionContext
├── Input: string
├── Output: string
├── Data: map[string]interface{} (shared state)
└── Metadata: map[string]interface{}
```

## Comparison: Workflow vs Team

| Feature | Workflow | Team |
|---------|----------|------|
| Control | Explicit, step-by-step | Implicit, coordination modes |
| Use Case | Deterministic processes | Autonomous collaboration |
| Flexibility | Custom control flow | Fixed coordination patterns |
| Complexity | Can build any logic | Simpler, fewer choices |

**Rule of Thumb**:
- Use **Workflow** when you need precise control flow (if-else, loops, parallel, routing)
- Use **Team** when agents should collaborate autonomously (sequential, parallel, consensus)

## Customization Ideas

1. **Add Error Handling**: Retry failed steps
2. **Conditional Loops**: Loop until quality threshold
3. **Nested Workflows**: Use a workflow as a step in another workflow
4. **Data Passing**: Use ExecutionContext to pass structured data
5. **Monitoring**: Track step execution time and success rate

## Advanced Example: E-commerce Order Processing

```go
// 1. Validate order (step)
// 2. Check inventory (step)
// 3. If in stock:
//      a. Process payment (step)
//      b. If payment success:
//           - Ship order (step)
//           - Send confirmation (step)
//      c. Else: Refund (step)
// 4. Else: Backorder (step)
```

This can be built using nested Conditions and Steps!

## Next Steps

- Try building your own workflow patterns
- Combine Workflow with Team for hybrid approaches
- Add custom Node implementations
- Explore error handling and retry logic
