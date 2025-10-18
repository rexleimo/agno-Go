# Batch Operations for PostgreSQL / PostgreSQL 批量操作

高性能的 PostgreSQL 批量会话(Session)写入实现,使用 COPY 协议和临时表策略。

High-performance PostgreSQL batch session write implementation using COPY protocol and temporary table strategy.

## 特性 / Features

- ✅ **高性能**: 使用 PostgreSQL COPY 协议,支持 >10,000 records/sec
- ✅ **原子性**: 所有操作在事务中执行,确保数据一致性
- ✅ **灵活性**: 支持保留或自动更新 `updated_at` 时间戳
- ✅ **可配置**: 批量大小、重试次数、超时时间都可自定义
- ✅ **测试覆盖**: 80.9% 测试覆盖率,包含完整的单元测试和集成测试

- ✅ **High Performance**: Uses PostgreSQL COPY protocol, supports >10,000 records/sec
- ✅ **Atomicity**: All operations execute in transactions, ensuring data consistency
- ✅ **Flexibility**: Supports preserving or auto-updating `updated_at` timestamps
- ✅ **Configurable**: Batch size, retry count, timeout are all customizable
- ✅ **Test Coverage**: 80.9% test coverage with comprehensive unit and integration tests

## 架构设计 / Architecture

### COPY + Temporary Table Strategy

```
1. CREATE TEMPORARY TABLE temp_sessions
2. COPY data INTO temp_sessions (批量导入 / Bulk import)
3. INSERT INTO sessions ... FROM temp_sessions ON CONFLICT DO UPDATE
4. DROP temp_sessions (事务结束时自动 / Auto on transaction end)
```

### 优势 / Benefits

- **速度快**: COPY 比逐条 INSERT 快 10-100 倍
- **内存优化**: 临时表在事务结束后自动清理
- **UPSERT 支持**: 自动处理插入和更新

- **Fast**: COPY is 10-100x faster than individual INSERTs
- **Memory Optimized**: Temporary tables auto-cleanup after transaction
- **UPSERT Support**: Automatically handles both insert and update

## 安装 / Installation

```bash
go get github.com/rexleimo/agno-go
```

依赖 / Dependencies:
- `github.com/lib/pq` - PostgreSQL 驱动 / PostgreSQL driver
- Go 1.21+

## 快速开始 / Quick Start

### 基础使用 / Basic Usage

```go
package main

import (
    "context"
    "database/sql"
    "log"

    _ "github.com/lib/pq"
    "github.com/rexleimo/agno-go/pkg/agno/db/batch"
    "github.com/rexleimo/agno-go/pkg/agno/session"
)

func main() {
    // 1. 连接数据库 / Connect to database
    db, err := sql.Open("postgres", "postgres://user:pass@localhost/agno?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // 2. 创建批量写入器 / Create batch writer
    writer, err := batch.NewPostgresBatchWriter(db, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer writer.Close()

    // 3. 准备数据 / Prepare data
    sessions := []*session.Session{
        {
            SessionID: "session-1",
            AgentID:   "agent-001",
            UserID:    "user-001",
            Name:      "Example Session",
            // ... 其他字段 / other fields
        },
    }

    // 4. 批量写入 / Batch upsert
    ctx := context.Background()
    if err := writer.UpsertSessions(ctx, sessions, false); err != nil {
        log.Fatal(err)
    }

    log.Println("Success!")
}
```

### 自定义配置 / Custom Configuration

```go
config := &batch.Config{
    BatchSize:        1000,               // 每批处理 1000 条记录 / Process 1000 records per batch
    MinBatchSize:     200,                // 动态缩小时的下限 / Minimum batch size when shrinking
    MaxRetries:       5,                  // 失败时最多重试 5 次 / Retry up to 5 times on failure
    TimeoutSeconds:   60,                 // 每批超时 60 秒 / 60 seconds timeout per batch
    ThrottleInterval: 100 * time.Millisecond, // 批次之间休眠 100ms / Sleep 100ms between batches
}

writer, err := batch.NewPostgresBatchWriter(db, config)
```

### 保留时间戳 (数据迁移) / Preserve Timestamps (Data Migration)

```go
// 迁移历史数据时,保留原始的 updated_at 时间戳
// When migrating historical data, preserve original updated_at timestamps
err := writer.UpsertSessions(ctx, sessions, true) // preserveUpdatedAt=true
```

### 自动更新时间戳 (正常操作) / Auto-Update Timestamps (Normal Operations)

```go
// 正常插入/更新时,自动设置 updated_at 为当前时间
// For normal insert/update, auto-set updated_at to current time
err := writer.UpsertSessions(ctx, sessions, false) // preserveUpdatedAt=false
```

## API 文档 / API Documentation

### BatchWriter Interface

```go
type BatchWriter interface {
    // UpsertSessions 批量插入或更新 sessions
    // UpsertSessions batch inserts or updates sessions
    UpsertSessions(ctx context.Context, sessions []*session.Session, preserveUpdatedAt bool) error

    // Close 关闭批量写入器并释放资源
    // Close closes the batch writer and releases resources
    Close() error
}
```

### Config

```go
type Config struct {
    BatchSize        int           // 批量大小,默认 5000 / Batch size, default 5000
    MinBatchSize     int           // 最小批量大小,默认 500 / Minimum batch size, default 500
    MaxRetries       int           // 最大重试次数,默认 3 / Max retries, default 3
    TimeoutSeconds   int           // 超时时间(秒),默认 30 / Timeout (seconds), default 30
    ThrottleInterval time.Duration // 批次间休眠时间,默认 0 / Sleep duration between batches, default 0
}
```

### NewPostgresBatchWriter

```go
func NewPostgresBatchWriter(db *sql.DB, config *Config) (*PostgresBatchWriter, error)
```

创建 PostgreSQL 批量写入器。如果 `config` 为 `nil`,使用默认配置。

Creates a PostgreSQL batch writer. If `config` is `nil`, uses default configuration.

## 性能基准 / Performance Benchmarks

### 测试环境 / Test Environment
- PostgreSQL 15
- Table: 13 columns (5 JSONB, 2 timestamps)
- Network: localhost

### 结果 / Results

| Records | Time | Throughput |
|---------|------|------------|
| 1,000   | ~80ms | 12,500 records/sec |
| 5,000   | ~350ms | 14,285 records/sec |
| 10,000  | ~680ms | 14,706 records/sec |

> 💡 实际性能取决于网络延迟、表结构复杂度、数据库负载等因素。
>
> 💡 Actual performance depends on network latency, table complexity, database load, etc.

## 数据库表结构 / Database Schema

```sql
CREATE TABLE sessions (
    session_id VARCHAR(255) PRIMARY KEY,
    agent_id VARCHAR(255),
    team_id VARCHAR(255),
    workflow_id VARCHAR(255),
    user_id VARCHAR(255),
    name VARCHAR(255),
    metadata JSONB,
    state JSONB,
    agent_data JSONB,
    runs JSONB,
    summary JSONB,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);
```

## 示例 / Examples

完整的示例代码在 [`examples/batch_upsert/`](../../../examples/batch_upsert/) 目录。

Complete example code is in [`examples/batch_upsert/`](../../../examples/batch_upsert/) directory.

运行示例 / Run Example:

```bash
# 1. 启动 PostgreSQL / Start PostgreSQL
docker-compose up -d postgres

# 2. 运行示例 / Run example
go run examples/batch_upsert/main.go
```

## 测试 / Testing

### 单元测试 / Unit Tests

```bash
go test ./pkg/agno/db/batch/...
```

### 测试覆盖率 / Test Coverage

```bash
go test -cover ./pkg/agno/db/batch/...
# coverage: 80.9% of statements
```

### 集成测试 / Integration Tests

```bash
# 需要运行的 PostgreSQL 实例 / Requires running PostgreSQL instance
go test -tags=integration ./pkg/agno/db/batch/...
```

### 竞态检测 / Race Detection

```bash
go test -race ./pkg/agno/db/batch/...
```

## 错误处理 / Error Handling

所有错误都使用 `fmt.Errorf` 包装,可以使用 `errors.Unwrap` 获取原始错误。

All errors are wrapped using `fmt.Errorf`, can use `errors.Unwrap` to get original error.

```go
err := writer.UpsertSessions(ctx, sessions, false)
if err != nil {
    // 错误已经包含上下文信息 / Error already includes context
    log.Printf("Failed to upsert: %v", err)

    // 可以检查是否是特定错误 / Can check for specific errors
    if errors.Is(err, context.DeadlineExceeded) {
        log.Println("Operation timed out")
    }
}
```

## 最佳实践 / Best Practices

### 1. 批量大小 / Batch Size

- **默认 5000** 适合大多数场景
- 网络延迟高时可以增大批量
- 内存有限时可以减小批量

- **Default 5000** works for most scenarios
- Increase for high network latency
- Decrease for limited memory

### 2. 事务超时 / Transaction Timeout

```go
// 对于大批量操作,增加超时时间 / For large batches, increase timeout
config := &batch.Config{
    TimeoutSeconds: 120, // 2 minutes
}
```

### 3. 错误重试 / Error Retry

```go
// 网络不稳定时增加重试次数 / Increase retries for unstable networks
config := &batch.Config{
    MaxRetries: 5,
}
```

### 4. Context 取消 / Context Cancellation

```go
// 使用带超时的 context / Use context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := writer.UpsertSessions(ctx, sessions, false)
```

## 注意事项 / Notes

1. **连接池**: 确保数据库连接池配置合理
2. **索引**: `session_id` 需要是主键或唯一索引
3. **JSONB**: 使用 JSONB 类型可以在 PostgreSQL 中直接查询 JSON 字段
4. **事务**: 所有操作在单个事务中,失败会自动回滚

1. **Connection Pool**: Ensure database connection pool is properly configured
2. **Index**: `session_id` must be primary key or unique index
3. **JSONB**: Using JSONB type allows direct JSON field queries in PostgreSQL
4. **Transaction**: All operations in single transaction, auto-rollback on failure

## 故障排查 / Troubleshooting

### 错误: "db cannot be nil"

确保传入了有效的 `*sql.DB` 实例。

Ensure a valid `*sql.DB` instance is passed.

### 错误: "failed to create temp table"

检查数据库用户是否有创建临时表的权限。

Check if database user has permission to create temporary tables.

### 性能不达预期 / Performance Below Expectations

1. 检查网络延迟 / Check network latency
2. 查看数据库负载 / Check database load
3. 优化表索引 / Optimize table indexes
4. 调整批量大小 / Adjust batch size

## 贡献 / Contributing

欢迎提交 Issue 和 Pull Request!

Issues and Pull Requests are welcome!

## 许可证 / License

MIT License - 查看 [LICENSE](../../../../LICENSE) 文件

MIT License - See [LICENSE](../../../../LICENSE) file
