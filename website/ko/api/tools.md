# Tools API 레퍼런스

## Calculator

기본 수학 연산입니다.

**생성:**
```go
func New() *Calculator
```

**함수:**
- `add(a, b)`: 덧셈
- `subtract(a, b)`: 뺄셈
- `multiply(a, b)`: 곱셈
- `divide(a, b)`: 나눗셈

**예제:**
```go
calc := calculator.New()

ag, _ := agent.New(agent.Config{
    Toolkits: []toolkit.Toolkit{calc},
    // ...
})
```

## HTTP

GET/POST 요청을 위한 HTTP 클라이언트입니다.

**생성:**
```go
func New() *HTTPToolkit
```

**함수:**
- `http_get(url)`: HTTP GET 요청
- `http_post(url, body)`: HTTP POST 요청

**예제:**
```go
http := http.New()

ag, _ := agent.New(agent.Config{
    Toolkits: []toolkit.Toolkit{http},
    Instructions: "You can make HTTP requests to fetch data.",
})
```

## File

안전 제어가 포함된 파일 작업입니다.

**생성:**
```go
func New(config Config) *FileToolkit

type Config struct {
    AllowedPaths []string // 허용된 디렉토리의 화이트리스트
    MaxFileSize  int64    // 최대 파일 크기 (바이트) (기본값: 10MB)
}
```

**함수:**
- `read_file(path)`: 파일 내용 읽기
- `write_file(path, content)`: 파일 쓰기
- `list_files(directory)`: 디렉토리 목록
- `delete_file(path)`: 파일 삭제

**예제:**
```go
file := file.New(file.Config{
    AllowedPaths: []string{"/data", "/tmp"},
    MaxFileSize:  5 * 1024 * 1024, // 5MB
})

ag, _ := agent.New(agent.Config{
    Toolkits: []toolkit.Toolkit{file},
})
```

## 커스텀 도구

커스텀 도구를 생성합니다:

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
    // 입력 처리
    return result, nil
}
```
