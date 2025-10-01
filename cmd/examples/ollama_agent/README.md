# Ollama Agent Example

This example demonstrates how to use Ollama local models with the Agno-Go framework.

## Features

- Run AI models locally with Ollama
- No API keys required
- Privacy-focused (all data stays local)
- Support for various models (Llama 2, Mistral, CodeLlama, etc.)

## Prerequisites

1. **Install Ollama**: Download from https://ollama.ai/

2. **Pull a model**:
```bash
# Llama 2 (7B) - general purpose
ollama pull llama2

# Mistral (7B) - faster, good quality
ollama pull mistral

# CodeLlama (7B) - for coding tasks
ollama pull codellama

# Llama 3 (8B) - latest model
ollama pull llama3

# Other models
ollama pull gemma         # Google's model
ollama pull phi           # Microsoft's small model
ollama pull neural-chat   # Intel's model
```

3. **Start Ollama server**:
```bash
ollama serve
```

The server will run on `http://localhost:11434` by default.

## Available Models

Check available models: `ollama list`

Popular models:
- `llama2` - Meta's Llama 2 (7B, 13B, 70B)
- `llama3` - Meta's Llama 3 (8B, 70B)
- `mistral` - Mistral AI (7B)
- `mixtral` - Mixture of Experts (8x7B)
- `codellama` - Code-specialized Llama (7B, 13B, 34B)
- `gemma` - Google Gemma (2B, 7B)
- `phi` - Microsoft Phi (2.7B)

## Setup

1. Make sure Ollama is running:
```bash
ollama serve
```

2. Run the example:
```bash
go run main.go
```

Or build and run:
```bash
go build -o ollama_agent
./ollama_agent
```

## Example Output

```
=== Example 1: Simple Conversation ===
Agent: I'm an AI assistant running on Ollama, here to help you with questions and calculations.

=== Example 2: Calculator Tool Usage ===
Agent: 456 × 789 = 359,784

=== Example 3: Complex Calculation ===
Agent: Let me calculate that step by step:
1. 100 + 50 = 150
2. 150 × 2 = 300
3. 300 - 75 = 225

The answer is 225.

✅ All examples completed successfully!
```

## Configuration

### Model Selection

Change the model by modifying the model ID:

```go
// Use Llama 3
model, err := ollama.New("llama3", ollama.Config{
    BaseURL:     "http://localhost:11434",
    Temperature: 0.7,
    MaxTokens:   2000,
})

// Use Mistral
model, err := ollama.New("mistral", ollama.Config{
    BaseURL:     "http://localhost:11434",
    Temperature: 0.5,
    MaxTokens:   1000,
})

// Use CodeLlama for coding
model, err := ollama.New("codellama", ollama.Config{
    BaseURL:     "http://localhost:11434",
    Temperature: 0.3,
    MaxTokens:   4096,
})
```

### Custom Ollama Server

If running Ollama on a different host/port:

```go
model, err := ollama.New("llama2", ollama.Config{
    BaseURL:     "http://192.168.1.100:11434",
    Temperature: 0.7,
    MaxTokens:   2000,
})
```

### Temperature Settings

- `0.0` - Deterministic, focused responses
- `0.3-0.5` - Balanced, good for coding
- `0.7-0.9` - Creative, varied responses
- `1.0+` - Very creative, unpredictable

## Streaming Response

For streaming responses:

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

## Performance Tips

1. **Model Size**: Smaller models (7B) are faster but less capable
2. **GPU**: Enable GPU acceleration for better performance
3. **Context Length**: Reduce MaxTokens for faster responses
4. **Model Loading**: First run loads model into memory (slow), subsequent runs are fast

## Checking Ollama Status

```bash
# Check if Ollama is running
curl http://localhost:11434/api/tags

# List loaded models
ollama list

# Check model info
ollama show llama2
```

## Troubleshooting

**Error: connection refused**
- Make sure Ollama server is running: `ollama serve`

**Error: model not found**
- Pull the model first: `ollama pull llama2`

**Slow responses**
- Use smaller models (7B instead of 70B)
- Enable GPU if available
- Reduce MaxTokens

**Out of memory**
- Use smaller models
- Close other applications
- Check: `ollama ps` to see loaded models

## Resource Requirements

Minimum requirements by model size:
- 7B models: 8GB RAM
- 13B models: 16GB RAM
- 70B models: 64GB RAM (or quantized versions)

GPU memory (VRAM) for faster inference:
- 7B models: 6GB VRAM
- 13B models: 12GB VRAM
- 70B models: 48GB+ VRAM

## Related Examples

- `simple_agent` - Basic agent with OpenAI
- `claude_agent` - Anthropic Claude integration
- `team_demo` - Multi-agent collaboration

## Documentation

- [Ollama Documentation](https://github.com/ollama/ollama)
- [Model Library](https://ollama.ai/library)
- [API Reference](https://github.com/ollama/ollama/blob/main/docs/api.md)
