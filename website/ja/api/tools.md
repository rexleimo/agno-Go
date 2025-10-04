# Tools APIリファレンス

## Calculator

基本的な数学演算を提供します。

**作成:**
```go
func New() *Calculator
```

**関数:**
- `add(a, b)`: 加算
- `subtract(a, b)`: 減算
- `multiply(a, b)`: 乗算
- `divide(a, b)`: 除算

**例:**
```go
calc := calculator.New()

ag, _ := agent.New(agent.Config{
    Toolkits: []toolkit.Toolkit{calc},
    // ...
})
```

## HTTP

GET/POSTリクエスト用のHTTPクライアント。

**作成:**
```go
func New() *HTTPToolkit
```

**関数:**
- `http_get(url)`: HTTP GETリクエスト
- `http_post(url, body)`: HTTP POSTリクエスト

**例:**
```go
http := http.New()

ag, _ := agent.New(agent.Config{
    Toolkits: []toolkit.Toolkit{http},
    Instructions: "You can make HTTP requests to fetch data.",
})
```

## File

安全制御付きのファイル操作。

**作成:**
```go
func New(config Config) *FileToolkit

type Config struct {
    AllowedPaths []string // 許可されたディレクトリのホワイトリスト
    MaxFileSize  int64    // 最大ファイルサイズ(バイト) (デフォルト: 10MB)
}
```

**関数:**
- `read_file(path)`: ファイル内容を読み取り
- `write_file(path, content)`: ファイルを書き込み
- `list_files(directory)`: ディレクトリを一覧表示
- `delete_file(path)`: ファイルを削除

**例:**
```go
file := file.New(file.Config{
    AllowedPaths: []string{"/data", "/tmp"},
    MaxFileSize:  5 * 1024 * 1024, // 5MB
})

ag, _ := agent.New(agent.Config{
    Toolkits: []toolkit.Toolkit{file},
})
```

## カスタムツール

カスタムツールを作成:

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
    // 入力を処理
    return result, nil
}
```
