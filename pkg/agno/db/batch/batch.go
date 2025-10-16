package batch

import (
	"context"

	"github.com/rexleimo/agno-go/pkg/agno/session"
)

// BatchWriter 定义批量写入接口
// BatchWriter defines the interface for batch write operations
type BatchWriter interface {
	// UpsertSessions 批量插入或更新 sessions
	// UpsertSessions batch inserts or updates sessions
	// preserveUpdatedAt: true 保留原 updated_at, false 使用当前时间
	// preserveUpdatedAt: true preserves original updated_at, false uses current time
	UpsertSessions(ctx context.Context, sessions []*session.Session, preserveUpdatedAt bool) error

	// Close 关闭批量写入器并释放资源
	// Close closes the batch writer and releases resources
	Close() error
}

// Config 批量操作配置
// Config for batch operations
type Config struct {
	// BatchSize 批量大小,默认 5000
	// BatchSize is the batch size, default 5000
	BatchSize int

	// MaxRetries 最大重试次数,默认 3
	// MaxRetries is the maximum number of retries, default 3
	MaxRetries int

	// TimeoutSeconds 每批操作超时时间(秒),默认 30
	// TimeoutSeconds is the timeout for each batch operation in seconds, default 30
	TimeoutSeconds int
}

// DefaultConfig 返回默认配置
// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		BatchSize:      5000,
		MaxRetries:     3,
		TimeoutSeconds: 30,
	}
}
