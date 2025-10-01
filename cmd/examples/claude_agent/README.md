# Claude Agent Example

This example demonstrates how to use Anthropic's Claude models with the Agno-Go framework.

## Features

- Integration with Claude 3 family models (Opus, Sonnet, Haiku)
- Tool usage with calculator
- Conversation memory
- Error handling

## Available Claude Models

- `claude-3-opus-20240229` - Most capable model, best for complex tasks
- `claude-3-sonnet-20240229` - Balanced performance and speed
- `claude-3-haiku-20240307` - Fastest model, best for simple tasks

## Prerequisites

You need an Anthropic API key. Get one at: https://console.anthropic.com/

## Setup

1. Set your Anthropic API key:
```bash
export ANTHROPIC_API_KEY=your-api-key-here
```

2. Run the example:
```bash
go run main.go
```

Or build and run:
```bash
go build -o claude_agent
./claude_agent
```

## Example Output

```
=== Example 1: Simple Conversation ===
Agent: I'm Claude, an AI assistant created by Anthropic to be helpful, harmless, and honest.

=== Example 2: Calculator Tool Usage ===
Agent: Let me calculate that for you. 156 × 23 = 3,588, and 3,588 - 100 = 3,488.

=== Example 3: Complex Calculation ===
Agent: I'll solve this step by step:
1. First, add 45 + 67 = 112
2. Then multiply 112 × 3 = 336
3. Finally subtract 89: 336 - 89 = 247

The answer is 247.

=== Example 4: Mathematical Reasoning ===
Agent: Starting with $500:
- After spending $123: $500 - $123 = $377
- After earning $250: $377 + $250 = $627

You would have $627.
```

## Configuration Options

### Model Configuration

```go
model, err := anthropic.New("claude-3-opus-20240229", anthropic.Config{
    APIKey:      apiKey,
    BaseURL:     "https://api.anthropic.com/v1", // Optional, custom endpoint
    Temperature: 0.7,                             // 0.0 to 1.0, controls randomness
    MaxTokens:   2000,                            // Maximum response length
})
```

### Agent Configuration

```go
agent, err := agent.New(agent.Config{
    Name:         "Claude Assistant",
    Model:        model,
    Toolkits:     []toolkit.Toolkit{calc, search, db}, // Multiple tools
    Instructions: "Custom system prompt...",
    MaxLoops:     10, // Maximum tool calling loops
})
```

## Streaming Response

For streaming responses, use `InvokeStream`:

```go
chunks, err := model.InvokeStream(ctx, req)
if err != nil {
    log.Fatal(err)
}

for chunk := range chunks {
    if chunk.Error != nil {
        log.Printf("Error: %v", chunk.Error)
        break
    }
    fmt.Print(chunk.Content)
    if chunk.Done {
        break
    }
}
```

## Error Handling

The example includes proper error handling for:
- Missing API key
- API errors
- Invalid configuration
- Network issues

## Notes

- Claude models have a context window of 200K tokens (Opus/Sonnet) or 100K tokens (Haiku)
- The API uses HTTP/1.1 with JSON payloads
- Streaming is supported via Server-Sent Events (SSE)
- Tool calling is supported natively by Claude 3 models

## Related Examples

- `simple_agent` - Basic agent with OpenAI
- `team_demo` - Multi-agent collaboration
- `workflow_demo` - Workflow engine

## Documentation

- [Anthropic API Docs](https://docs.anthropic.com/claude/reference)
- [Claude Models Overview](https://docs.anthropic.com/claude/docs/models-overview)
- [Tool Use Guide](https://docs.anthropic.com/claude/docs/tool-use)
