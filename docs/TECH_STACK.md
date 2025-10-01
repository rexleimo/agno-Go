# 技术栈选型 (KISS 原则)

## 核心原则
1. **标准库优先** - 减少依赖
2. **成熟稳定优先** - 避免新库
3. **简单够用优先** - 不过度设计

---

## 基础设施

### Go 版本
- **Go 1.21+**
- 原因: 支持 slog, context 改进, 性能优化

### 包管理
- **Go Modules**
- 原因: 官方工具

---

## 开发工具

| 类别 | 工具 | 说明 |
|------|------|------|
| 代码格式化 | `gofmt` / `goimports` | 标准工具 |
| 代码检查 | `golangci-lint` | 集成多种 linter |
| 静态分析 | `go vet` | 标准工具 |
| 测试 | `testing` + `testify` | 标准库 + 断言库 |
| 基准测试 | `testing` | 标准库 |
| 性能分析 | `pprof` | 标准工具 |

---

## 核心依赖

### HTTP 客户端
```go
// 标准库足够
import "net/http"

// 如需更强功能
import "github.com/go-resty/resty/v2"
```

### JSON 处理
```go
// 标准库
import "encoding/json"

// 动态 JSON 查询
import "github.com/tidwall/gjson"
```

### 数据验证
```go
import "github.com/go-playground/validator/v10"

type Config struct {
    APIKey string `validate:"required"`
    Port   int    `validate:"min=1,max=65535"`
}
```

### 日志
```go
// Go 1.21+ 标准库
import "log/slog"

logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
logger.Info("agent started", "id", agentID)
```

### 配置管理
```go
// 简单场景: 标准库
import "encoding/json"
import "os"

// 复杂场景
import "github.com/spf13/viper"
```

---

## LLM SDK

### OpenAI
```go
import "github.com/sashabaranov/go-openai"
// 9k+ stars, 活跃维护
```

### Anthropic
```go
// 自己封装 HTTP 客户端
// 参考: https://docs.anthropic.com/claude/reference/messages_post
```

### Google
```go
import "cloud.google.com/go/vertexai/genai"
// 官方 SDK
```

### AWS Bedrock
```go
import "github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
// 官方 SDK
```

### 其他 (OpenAI 兼容)
```go
// 使用 go-openai 改 BaseURL
config := openai.DefaultConfig(apiKey)
config.BaseURL = "https://api.groq.com/openai/v1"
```

---

## Web 框架

### API 服务
```go
import "github.com/gin-gonic/gin"
// 最流行, 性能好, 文档全

// 轻量替代
import "github.com/gofiber/fiber/v2"
```

### WebSocket
```go
import "github.com/gorilla/websocket"
// 标准选择
```

### 中间件
```go
// Gin 内置
router.Use(gin.Logger())
router.Use(gin.Recovery())

// 自定义
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 验证逻辑
    }
}
```

---

## 数据库

### SQL
```go
// 标准库 + 驱动
import "database/sql"
import _ "github.com/lib/pq" // PostgreSQL

// 轻量 ORM (可选)
import "github.com/jmoiron/sqlx"

// 不使用 GORM (太重)
```

### Redis
```go
import "github.com/redis/go-redis/v9"
// 官方推荐
```

### SQLite (测试用)
```go
import _ "github.com/mattn/go-sqlite3"
```

---

## 向量数据库

### PgVector
```go
import "github.com/pgvector/pgvector-go"
import "github.com/jackc/pgx/v5"
```

### Qdrant
```go
import "github.com/qdrant/go-client/qdrant"
```

### ChromaDB
```go
// HTTP API 调用
// 暂无成熟 Go SDK
```

### Milvus
```go
import "github.com/milvus-io/milvus-sdk-go/v2/client"
```

---

## 并发与同步

### 标准库完全够用
```go
import (
    "context"
    "sync"
    "time"
)

// Worker Pool
type WorkerPool struct {
    tasks   chan func()
    workers int
}

// Rate Limiter
import "golang.org/x/time/rate"
limiter := rate.NewLimiter(10, 1) // 10 req/s

// Errgroup
import "golang.org/x/sync/errgroup"
g, ctx := errgroup.WithContext(context.Background())
g.Go(func() error { /* ... */ })
g.Wait()
```

---

## 测试

### 单元测试
```go
import "testing"
import "github.com/stretchr/testify/assert"
import "github.com/stretchr/testify/mock"

func TestAgent_Run(t *testing.T) {
    assert := assert.New(t)
    assert.Equal(expected, actual)
}
```

### HTTP 测试
```go
import "net/http/httptest"

recorder := httptest.NewRecorder()
router.ServeHTTP(recorder, request)
assert.Equal(t, 200, recorder.Code)
```

### Mock
```go
// 使用接口 + mock 实现
type MockModel struct {
    mock.Mock
}

func (m *MockModel) Invoke(ctx context.Context, req *InvokeRequest) (*ModelResponse, error) {
    args := m.Called(ctx, req)
    return args.Get(0).(*ModelResponse), args.Error(1)
}
```

---

## 文档生成

### 代码文档
```go
// 使用 godoc
godoc -http=:6060
// 访问 http://localhost:6060/pkg/your-package/
```

### API 文档
```go
// Swagger (可选)
import "github.com/swaggo/gin-swagger"

// 注释生成
// @Summary Agent Run
// @Description Run agent with input
// @Tags agent
// @Accept json
// @Produce json
// @Param input body RunInput true "Input"
// @Success 200 {object} RunOutput
// @Router /agents/{id}/run [post]
```

---

## CLI 工具

### 简单场景
```go
import "flag"

var port = flag.Int("port", 8080, "server port")
flag.Parse()
```

### 复杂场景
```go
import "github.com/spf13/cobra"

rootCmd := &cobra.Command{
    Use:   "agno",
    Short: "Agno CLI",
}

runCmd := &cobra.Command{
    Use:   "run",
    Short: "Run agent",
    Run: func(cmd *cobra.Command, args []string) {
        // ...
    },
}
```

---

## 不使用的库

### 避免过度依赖
❌ **GORM** - 太重, 用 sqlx
❌ **Beego** - 太大, 用 Gin
❌ **各种 ORM** - 简单场景直接 SQL

### 避免不成熟的库
❌ 0.x 版本的库
❌ 长期无维护的库
❌ 只有个人维护的库

---

## 依赖管理

### go.mod 示例
```go
module github.com/yourusername/agno-go

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/sashabaranov/go-openai v1.17.9
    github.com/go-playground/validator/v10 v10.16.0
    github.com/stretchr/testify v1.8.4
)
```

### 更新策略
```bash
# 定期更新 (每月)
go get -u ./...

# 检查过期依赖
go list -u -m all
```

---

## 性能优化工具

### CPU Profiling
```bash
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof
```

### Memory Profiling
```bash
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

### Trace
```bash
go test -trace=trace.out
go tool trace trace.out
```

---

## CI/CD

### GitHub Actions
```yaml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go test -v -race -coverprofile=coverage.out ./...
      - run: golangci-lint run
```

---

## 总结

### 核心依赖 (必需)
- `github.com/gin-gonic/gin` - Web
- `github.com/sashabaranov/go-openai` - OpenAI
- `github.com/go-playground/validator/v10` - 验证
- `github.com/stretchr/testify` - 测试

### 可选依赖
- `github.com/spf13/viper` - 配置
- `github.com/spf13/cobra` - CLI
- `github.com/redis/go-redis/v9` - Redis
- 向量数据库客户端 (按需)

### 原则
**能用标准库就用标准库**
**能不加依赖就不加依赖**
**保持 go.mod 干净**

---

**少即是多**
