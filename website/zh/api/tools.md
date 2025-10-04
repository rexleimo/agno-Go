# Tools API 参考 / Tools API Reference

## Calculator

基础数学运算。/ Basic math operations.

**创建 / Create:**
```go
func New() *Calculator
```

**函数 / Functions:**
- `add(a, b)`: 加法 / Addition
- `subtract(a, b)`: 减法 / Subtraction
- `multiply(a, b)`: 乘法 / Multiplication
- `divide(a, b)`: 除法 / Division

**示例 / Example:**
```go
calc := calculator.New()

ag, _ := agent.New(agent.Config{
    Toolkits: []toolkit.Toolkit{calc},
    // ...
})
```

## HTTP

用于 GET/POST 请求的 HTTP 客户端。/ HTTP client for GET/POST requests.

**创建 / Create:**
```go
func New() *HTTPToolkit
```

**函数 / Functions:**
- `http_get(url)`: HTTP GET 请求 / HTTP GET request
- `http_post(url, body)`: HTTP POST 请求 / HTTP POST request

**示例 / Example:**
```go
http := http.New()

ag, _ := agent.New(agent.Config{
    Toolkits: []toolkit.Toolkit{http},
    Instructions: "You can make HTTP requests to fetch data.",
})
```

## File

带安全控制的文件操作。/ File operations with safety controls.

**创建 / Create:**
```go
func New(config Config) *FileToolkit

type Config struct {
    AllowedPaths []string // 允许的目录白名单 / Whitelist of allowed directories
    MaxFileSize  int64    // 最大文件大小(字节) (默认: 10MB) / Max file size in bytes (default: 10MB)
}
```

**函数 / Functions:**
- `read_file(path)`: 读取文件内容 / Read file content
- `write_file(path, content)`: 写入文件 / Write file
- `list_files(directory)`: 列出目录 / List directory
- `delete_file(path)`: 删除文件 / Delete file

**示例 / Example:**
```go
file := file.New(file.Config{
    AllowedPaths: []string{"/data", "/tmp"},
    MaxFileSize:  5 * 1024 * 1024, // 5MB
})

ag, _ := agent.New(agent.Config{
    Toolkits: []toolkit.Toolkit{file},
})
```

## 自定义工具 / Custom Tools

创建自定义工具 / Create custom tools:

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
    // 处理输入 / Process input
    return result, nil
}
```
