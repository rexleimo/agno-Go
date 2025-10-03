# Tools API Reference

## Calculator

Basic math operations.

**Create:**
```go
func New() *Calculator
```

**Functions:**
- `add(a, b)`: Addition
- `subtract(a, b)`: Subtraction
- `multiply(a, b)`: Multiplication
- `divide(a, b)`: Division

**Example:**
```go
calc := calculator.New()

ag, _ := agent.New(agent.Config{
    Toolkits: []toolkit.Toolkit{calc},
    // ...
})
```

## HTTP

HTTP client for GET/POST requests.

**Create:**
```go
func New() *HTTPToolkit
```

**Functions:**
- `http_get(url)`: HTTP GET request
- `http_post(url, body)`: HTTP POST request

**Example:**
```go
http := http.New()

ag, _ := agent.New(agent.Config{
    Toolkits: []toolkit.Toolkit{http},
    Instructions: "You can make HTTP requests to fetch data.",
})
```

## File

File operations with safety controls.

**Create:**
```go
func New(config Config) *FileToolkit

type Config struct {
    AllowedPaths []string // Whitelist of allowed directories
    MaxFileSize  int64    // Max file size in bytes (default: 10MB)
}
```

**Functions:**
- `read_file(path)`: Read file content
- `write_file(path, content)`: Write file
- `list_files(directory)`: List directory
- `delete_file(path)`: Delete file

**Example:**
```go
file := file.New(file.Config{
    AllowedPaths: []string{"/data", "/tmp"},
    MaxFileSize:  5 * 1024 * 1024, // 5MB
})

ag, _ := agent.New(agent.Config{
    Toolkits: []toolkit.Toolkit{file},
})
```

## Custom Tools

Create custom tools:

```go
type MyToolkit struct {
    *toolkit.BaseToolkit
}

func NewMyToolkit() *MyToolkit {
    t := &MyToolkit{
        BaseToolkit: toolkit.NewBaseToolkit("my_tools"),
    }

    t.RegisterFunction(&toolkit.Function{
        Name:        "my_function",
        Description: "Description of what this function does",
        Parameters: map[string]toolkit.Parameter{
            "input": {
                Type:        "string",
                Description: "Input parameter description",
                Required:    true,
            },
            "optional": {
                Type:        "number",
                Description: "Optional parameter",
                Required:    false,
            },
        },
        Handler: t.myHandler,
    })

    return t
}

func (t *MyToolkit) myHandler(args map[string]interface{}) (interface{}, error) {
    input := args["input"].(string)
    // Process input
    return result, nil
}
```
