# Simple Agent Example

This example demonstrates the basic usage of Agno-Go framework with an agent that can perform mathematical calculations using tools.

## Prerequisites

- Go 1.21 or higher
- OpenAI API key

## Setup

1. Set your OpenAI API key:
```bash
export OPENAI_API_KEY="your-api-key-here"
```

2. Run the example:
```bash
go run main.go
```

Or build and run:
```bash
go build -o simple_agent
./simple_agent
```

## What It Does

This example creates an agent that:
1. Uses GPT-4o-mini model
2. Has access to calculator tools (add, subtract, multiply, divide)
3. Can execute multi-step calculations
4. Automatically calls tools as needed

## Expected Output

The agent will:
1. Understand the mathematical query
2. Call the `multiply` tool with arguments (25, 4) → 100
3. Call the `add` tool with arguments (100, 15) → 115
4. Return the final answer: "115"

## Code Structure

- **Model Setup**: Creates OpenAI model instance
- **Tools**: Registers calculator toolkit
- **Agent Creation**: Configures agent with model, tools, and instructions
- **Execution**: Runs agent with user input and displays results
