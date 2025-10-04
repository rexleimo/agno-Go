# Tools - Agent 능력 확장

도구로 에이전트에 외부 기능에 대한 액세스를 제공하세요.

---

## 도구란?

도구는 에이전트가 외부 시스템과 상호작용하고, 계산을 수행하고, 데이터를 가져오는 등의 작업을 위해 호출할 수 있는 함수입니다. Agno-Go는 에이전트 능력을 확장하기 위한 유연한 툴킷 시스템을 제공합니다.

### 내장 도구

- **Calculator**: 기본 수학 연산
- **HTTP**: 웹 요청 만들기
- **File**: 안전 제어를 갖춘 파일 읽기/쓰기

---

## 도구 사용

### 기본 예제

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
    fmt.Println(output.Content) // 에이전트가 자동으로 계산기 사용
}
```

---

## Calculator 도구

수학 연산을 수행합니다.

### 연산

- `add(a, b)` - 덧셈
- `subtract(a, b)` - 뺄셈
- `multiply(a, b)` - 곱셈
- `divide(a, b)` - 나눗셈

### 예제

```go
import "github.com/rexleimo/agno-go/pkg/agno/tools/calculator"

agent, _ := agent.New(agent.Config{
    Model:    model,
    Toolkits: []toolkit.Toolkit{calculator.New()},
})

// 에이전트가 자동으로 계산기 사용
output, _ := agent.Run(ctx, "Calculate 15% tip on $85")
```

---

## HTTP 도구

외부 API에 HTTP 요청을 만듭니다.

### 메서드

- `get(url)` - HTTP GET 요청
- `post(url, body)` - HTTP POST 요청

### 예제

```go
import "github.com/rexleimo/agno-go/pkg/agno/tools/http"

agent, _ := agent.New(agent.Config{
    Model:    model,
    Toolkits: []toolkit.Toolkit{http.New()},
})

// 에이전트가 API에서 데이터 가져오기 가능
output, _ := agent.Run(ctx, "Get the latest GitHub status from https://www.githubstatus.com/api/v2/status.json")
```

### 구성

보안을 위한 허용 도메인 제어:

```go
httpTool := http.New(http.Config{
    AllowedDomains: []string{"api.github.com", "api.weather.com"},
    Timeout:        10 * time.Second,
})
```

---

## File 도구

내장 안전 제어로 파일을 읽고 씁니다.

### 연산

- `read_file(path)` - 파일 내용 읽기
- `write_file(path, content)` - 파일에 내용 쓰기
- `list_directory(path)` - 디렉토리 내용 나열
- `delete_file(path)` - 파일 삭제

### 예제

```go
import "github.com/rexleimo/agno-go/pkg/agno/tools/file"

fileTool := file.New(file.Config{
    AllowedPaths: []string{"/tmp", "./data"},  // 액세스 제한
    MaxFileSize:  1024 * 1024,                 // 1MB 제한
})

agent, _ := agent.New(agent.Config{
    Model:    model,
    Toolkits: []toolkit.Toolkit{fileTool},
})

output, _ := agent.Run(ctx, "Read the contents of ./data/report.txt")
```

### 안전 기능

- 경로 제한 (화이트리스트)
- 파일 크기 제한
- 읽기 전용 모드 옵션
- 자동 경로 정제

---

## 여러 도구

에이전트는 여러 도구를 사용할 수 있습니다:

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

// 에이전트는 이제 계산, 데이터 가져오기, 파일 읽기 가능
output, _ := agent.Run(ctx,
    "Fetch weather data, calculate average temperature, and save to file")
```

---

## 커스텀 도구 생성

Toolkit 인터페이스를 구현하여 자신만의 도구를 만드세요.

### 단계 1: Toolkit 구조체 생성

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

    // 함수 등록
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

### 단계 2: 핸들러 구현

```go
func (t *MyToolkit) greet(args map[string]interface{}) (interface{}, error) {
    name, ok := args["name"].(string)
    if !ok {
        return nil, fmt.Errorf("name must be a string")
    }

    return fmt.Sprintf("Hello, %s!", name), nil
}
```

### 단계 3: 도구 사용

```go
agent, _ := agent.New(agent.Config{
    Model:    model,
    Toolkits: []toolkit.Toolkit{mytool.New()},
})

output, _ := agent.Run(ctx, "Greet Alice")
// 에이전트가 greet("Alice")를 호출하고 "Hello, Alice!"로 응답
```

---

## 고급 커스텀 도구 예제

데이터베이스 쿼리 도구:

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

## 도구 모범 사례

### 1. 명확한 설명

에이전트가 도구를 언제 사용할지 이해하도록 돕기:

```go
// 좋음 ✅
Description: "Calculate the square root of a number. Use when user asks for square roots."

// 나쁨 ❌
Description: "Math function"
```

### 2. 입력 검증

항상 도구 매개변수 검증:

```go
func (t *MyToolkit) divide(args map[string]interface{}) (interface{}, error) {
    b, ok := args["divisor"].(float64)
    if !ok || b == 0 {
        return nil, fmt.Errorf("divisor must be a non-zero number")
    }
    // ... 나눗셈 수행
}
```

### 3. 오류 처리

의미 있는 오류 반환:

```go
func (t *MyToolkit) fetchData(args map[string]interface{}) (interface{}, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch data: %w", err)
    }
    // ... 응답 처리
}
```

### 4. 보안

도구 능력 제한:

```go
// 허용된 작업 화이트리스트
fileTool := file.New(file.Config{
    AllowedPaths: []string{"/safe/path"},
    ReadOnly:     true,  // 쓰기 방지
})

// 도메인 검증
httpTool := http.New(http.Config{
    AllowedDomains: []string{"api.trusted.com"},
})
```

---

## 도구 실행 흐름

1. 사용자가 에이전트에 요청 전송
2. 에이전트 (LLM)가 도구 필요 여부 결정
3. LLM이 매개변수로 도구 호출 생성
4. Agno-Go가 도구 함수 실행
5. 도구 결과가 LLM에 반환
6. LLM이 최종 응답 생성

### 예제 흐름

```
사용자: "What is 25 * 17?"
  ↓
LLM: "I need to use calculator"
  ↓
도구 호출: multiply(25, 17)
  ↓
도구 결과: 425
  ↓
LLM: "The answer is 425"
  ↓
사용자 수신: "The answer is 425"
```

---

## 문제 해결

### 에이전트가 도구를 사용하지 않음

명확한 지침 확인:

```go
agent, _ := agent.New(agent.Config{
    Model:        model,
    Toolkits:     []toolkit.Toolkit{calculator.New()},
    Instructions: "Use the calculator tool for any math operations.",
})
```

### 도구 오류

도구 등록 및 매개변수 타입 확인:

```go
// 도구는 숫자에 대해 float64 예상
args := map[string]interface{}{
    "value": 42.0,  // ✅ 올바름
    // "value": "42"  // ❌ 잘못된 타입
}
```

---

## 다음 단계

- 대화 상태를 위한 [Memory](/guide/memory) 배우기
- 전문 도구 에이전트로 [Teams](/guide/team) 구축
- 도구 오케스트레이션을 위한 [Workflow](/guide/workflow) 탐색
- 자세한 문서는 [API Reference](/api/tools) 확인

---

## 관련 예제

- [Simple Agent](/examples/simple-agent) - Calculator 도구 사용
- [Search Agent](/examples/search-agent) - 웹 검색을 위한 HTTP 도구
- [File Agent](/examples/file-agent) - 파일 작업
