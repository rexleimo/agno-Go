# Tools - Extend Agent Capabilities

Give your agents access to external functionality with tools.

---

## What are Tools?

Tools are functions that agents can call to interact with external systems, perform calculations, fetch data, and more. Agno-Go provides a flexible toolkit system for extending agent capabilities.

### Built-in Tools

- **Calculator**: Basic math operations
- **HTTP**: Make web requests
- **File**: Read/write files with safety controls

---

## Using Tools

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/rexleimo/agno-go/pkg/agno/agent"
    "github.com/rexleimo/agno-go/pkg/agno/models/openai"
    "github.com/rexleimo/agno-go/pkg/agno/tools/calculator"
    "github.com/rexleimo/agno-go/pkg/agno/tools/toolkit"
)

func main() {
    model, _ := openai.New("gpt-4o-mini", openai.Config{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })

    agent, _ := agent.New(agent.Config{
        Name:     "Math Assistant",
        Model:    model,
        Toolkits: []toolkit.Toolkit{calculator.New()},
    })

    output, _ := agent.Run(context.Background(), "What is 23 * 47?")
    fmt.Println(output.Content) // Agent uses calculator automatically
}
```

---

## Calculator Tool

Perform mathematical operations.

### Operations

- `add(a, b)` - Addition
- `subtract(a, b)` - Subtraction
- `multiply(a, b)` - Multiplication
- `divide(a, b)` - Division

### Example

```go
import "github.com/rexleimo/agno-go/pkg/agno/tools/calculator"

agent, _ := agent.New(agent.Config{
    Model:    model,
    Toolkits: []toolkit.Toolkit{calculator.New()},
})

// Agent will automatically use calculator
output, _ := agent.Run(ctx, "Calculate 15% tip on $85")
```

---

## HTTP Tool

Make HTTP requests to external APIs.

### Methods

- `get(url)` - HTTP GET request
- `post(url, body)` - HTTP POST request

### Example

```go
import "github.com/rexleimo/agno-go/pkg/agno/tools/http"

agent, _ := agent.New(agent.Config{
    Model:    model,
    Toolkits: []toolkit.Toolkit{http.New()},
})

// Agent can fetch data from APIs
output, _ := agent.Run(ctx, "Get the latest GitHub status from https://www.githubstatus.com/api/v2/status.json")
```

### Configuration

Control allowed domains for security:

```go
httpTool := http.New(http.Config{
    AllowedDomains: []string{"api.github.com", "api.weather.com"},
    Timeout:        10 * time.Second,
})
```

---

## File Tool

Read and write files with built-in safety controls.

### Operations

- `read_file(path)` - Read file content
- `write_file(path, content)` - Write content to file
- `list_directory(path)` - List directory contents
- `delete_file(path)` - Delete a file

### Example

```go
import "github.com/rexleimo/agno-go/pkg/agno/tools/file"

fileTool := file.New(file.Config{
    AllowedPaths: []string{"/tmp", "./data"},  // Restrict access
    MaxFileSize:  1024 * 1024,                 // 1MB limit
})

agent, _ := agent.New(agent.Config{
    Model:    model,
    Toolkits: []toolkit.Toolkit{fileTool},
})

output, _ := agent.Run(ctx, "Read the contents of ./data/report.txt")
```

### Safety Features

- Path restrictions (whitelist)
- File size limits
- Read-only mode option
- Automatic path sanitization

---

## Multiple Tools

Agents can use multiple tools:

```go
agent, _ := agent.New(agent.Config{
    Name:  "Multi-Tool Agent",
    Model: model,
    Toolkits: []toolkit.Toolkit{
        calculator.New(),
        http.New(),
        file.New(file.Config{
            AllowedPaths: []string{"./data"},
        }),
    },
})

// Agent can now calculate, fetch data, and read files
output, _ := agent.Run(ctx,
    "Fetch weather data, calculate average temperature, and save to file")
```

---

## Creating Custom Tools

Build your own tools by implementing the Toolkit interface.

### Step 1: Create Toolkit Struct

```go
package mytool

import "github.com/rexleimo/agno-go/pkg/agno/tools/toolkit"

type MyToolkit struct {
    *toolkit.BaseToolkit
}

func New() *MyToolkit {
    t := &MyToolkit{
        BaseToolkit: toolkit.NewBaseToolkit("my_tools"),
    }

    // Register functions
    t.RegisterFunction(&toolkit.Function{
        Name:        "greet",
        Description: "Greet a person by name",
        Parameters: map[string]toolkit.Parameter{
            "name": {
                Type:        "string",
                Description: "Person's name",
                Required:    true,
            },
        },
        Handler: t.greet,
    })

    return t
}
```

### Step 2: Implement Handler

```go
func (t *MyToolkit) greet(args map[string]interface{}) (interface{}, error) {
    name, ok := args["name"].(string)
    if !ok {
        return nil, fmt.Errorf("name must be a string")
    }

    return fmt.Sprintf("Hello, %s!", name), nil
}
```

### Step 3: Use Your Tool

```go
agent, _ := agent.New(agent.Config{
    Model:    model,
    Toolkits: []toolkit.Toolkit{mytool.New()},
})

output, _ := agent.Run(ctx, "Greet Alice")
// Agent calls greet("Alice") and responds with "Hello, Alice!"
```

---

## Advanced Custom Tool Example

Database query tool:

```go
type DatabaseToolkit struct {
    *toolkit.BaseToolkit
    db *sql.DB
}

func NewDatabaseToolkit(db *sql.DB) *DatabaseToolkit {
    t := &DatabaseToolkit{
        BaseToolkit: toolkit.NewBaseToolkit("database"),
        db:          db,
    }

    t.RegisterFunction(&toolkit.Function{
        Name:        "query_users",
        Description: "Query users from database",
        Parameters: map[string]toolkit.Parameter{
            "limit": {
                Type:        "integer",
                Description: "Maximum number of results",
                Required:    false,
            },
        },
        Handler: t.queryUsers,
    })

    return t
}

func (t *DatabaseToolkit) queryUsers(args map[string]interface{}) (interface{}, error) {
    limit := 10
    if l, ok := args["limit"].(float64); ok {
        limit = int(l)
    }

    rows, err := t.db.Query("SELECT id, name FROM users LIMIT ?", limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []map[string]interface{}
    for rows.Next() {
        var id int
        var name string
        rows.Scan(&id, &name)
        users = append(users, map[string]interface{}{
            "id":   id,
            "name": name,
        })
    }

    return users, nil
}
```

---

## Tool Best Practices

### 1. Clear Descriptions

Help the agent understand when to use tools:

```go
// Good ✅
Description: "Calculate the square root of a number. Use when user asks for square roots."

// Bad ❌
Description: "Math function"
```

### 2. Validate Input

Always validate tool parameters:

```go
func (t *MyToolkit) divide(args map[string]interface{}) (interface{}, error) {
    b, ok := args["divisor"].(float64)
    if !ok || b == 0 {
        return nil, fmt.Errorf("divisor must be a non-zero number")
    }
    // ... perform division
}
```

### 3. Error Handling

Return meaningful errors:

```go
func (t *MyToolkit) fetchData(args map[string]interface{}) (interface{}, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch data: %w", err)
    }
    // ... process response
}
```

### 4. Security

Restrict tool capabilities:

```go
// Whitelist allowed operations
fileTool := file.New(file.Config{
    AllowedPaths: []string{"/safe/path"},
    ReadOnly:     true,  // Prevent writes
})

// Validate domains
httpTool := http.New(http.Config{
    AllowedDomains: []string{"api.trusted.com"},
})
```

---

## Tool Execution Flow

1. User sends request to agent
2. Agent (LLM) decides if tools are needed
3. LLM generates tool call with parameters
4. Agno-Go executes tool function
5. Tool result returned to LLM
6. LLM generates final response

### Example Flow

```
User: "What is 25 * 17?"
  ↓
LLM: "I need to use calculator"
  ↓
Tool Call: multiply(25, 17)
  ↓
Tool Result: 425
  ↓
LLM: "The answer is 425"
  ↓
User receives: "The answer is 425"
```

---

## Troubleshooting

### Agent Not Using Tools

Ensure clear instructions:

```go
agent, _ := agent.New(agent.Config{
    Model:        model,
    Toolkits:     []toolkit.Toolkit{calculator.New()},
    Instructions: "Use the calculator tool for any math operations.",
})
```

### Tool Errors

Check tool registration and parameter types:

```go
// Tool expects float64 for numbers
args := map[string]interface{}{
    "value": 42.0,  // ✅ Correct
    // "value": "42"  // ❌ Wrong type
}
```

---

## Next Steps

- Learn about [Memory](/guide/memory) for conversation state
- Build [Teams](/guide/team) with specialized tool agents
- Explore [Workflow](/guide/workflow) for tool orchestration
- Check [API Reference](/api/tools) for detailed docs

---

## Related Examples

- [Simple Agent](/examples/simple-agent) - Calculator tool usage
- [Search Agent](/examples/search-agent) - HTTP tool for web search
- [File Agent](/examples/file-agent) - File operations
