package middleware

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ExtractedContent 表示从请求中抽取的内容
// ExtractedContent represents content extracted from the request
type ExtractedContent struct {
	// Content 主要内容（文本）
	// Content is the main text content
	Content string `json:"content"`

	// Metadata 元数据
	// Metadata is additional metadata
	Metadata map[string]interface{} `json:"metadata"`

	// UserID 用户 ID
	// UserID is the user identifier
	UserID string `json:"user_id"`

	// SessionID 会话 ID
	// SessionID is the session identifier
	SessionID string `json:"session_id"`

	// RawBody 原始请求体（用于审计）
	// RawBody is the raw request body (for auditing)
	RawBody []byte `json:"-"`
}

// ExtractContentConfig 内容抽取中间件配置
// ExtractContentConfig is the configuration for content extraction middleware
type ExtractContentConfig struct {
	// MaxRequestSize 最大请求大小（字节），默认 10MB
	// MaxRequestSize is the maximum request size in bytes (default 10MB)
	MaxRequestSize int64

	// Logger 日志记录器
	// Logger is the logger instance
	Logger *slog.Logger

	// EnableAudit 启用审计日志（记录原始请求体）
	// EnableAudit enables audit logging (logs raw request body)
	EnableAudit bool

	// SkipPaths 跳过内容抽取的路径列表
	// SkipPaths is a list of paths to skip content extraction
	SkipPaths []string
}

// ExtractContentMiddleware 创建内容抽取中间件
// ExtractContentMiddleware creates a content extraction middleware
func ExtractContentMiddleware(config ExtractContentConfig) gin.HandlerFunc {
	// 设置默认值
	// Set default values
	if config.MaxRequestSize == 0 {
		config.MaxRequestSize = 10 * 1024 * 1024 // 10MB
	}
	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	return func(c *gin.Context) {
		// 检查是否需要跳过此路径
		// Check if this path should be skipped
		path := c.Request.URL.Path
		for _, skipPath := range config.SkipPaths {
			if path == skipPath {
				c.Next()
				return
			}
		}

		// 只处理 POST 和 PUT 请求
		// Only process POST and PUT requests
		if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut {
			c.Next()
			return
		}

		// 检查 Content-Type
		// Check Content-Type
		contentType := c.ContentType()
		if contentType != "application/json" && contentType != "application/x-www-form-urlencoded" {
			c.Next()
			return
		}

		// 限制请求大小
		// Limit request size
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, config.MaxRequestSize)

		// 读取请求体
		// Read request body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			config.Logger.Error("failed to read request body",
				"error", err,
				"path", path,
				"method", c.Request.Method,
			)
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":   "request_too_large",
				"message": "request body exceeds maximum size",
			})
			c.Abort()
			return
		}

		// 重新设置请求体（供后续处理器使用）
		// Reset request body (for subsequent handlers)
		c.Request.Body = io.NopCloser(io.Reader(newBytesReader(bodyBytes)))

		// 尝试解析为 JSON
		// Try to parse as JSON
		var extracted ExtractedContent
		var rawData map[string]interface{}

		if contentType == "application/json" {
			if err := json.Unmarshal(bodyBytes, &rawData); err != nil {
				// JSON 解析失败，跳过抽取但继续处理
				// JSON parsing failed, skip extraction but continue
				config.Logger.Debug("failed to parse JSON, skipping extraction",
					"error", err,
					"path", path,
				)
				c.Next()
				return
			}

			// 抽取标准字段
			// Extract standard fields
			if content, ok := rawData["content"].(string); ok {
				extracted.Content = content
			}
			if metadata, ok := rawData["metadata"].(map[string]interface{}); ok {
				extracted.Metadata = metadata
			}
			if userID, ok := rawData["user_id"].(string); ok {
				extracted.UserID = userID
			}
			if sessionID, ok := rawData["session_id"].(string); ok {
				extracted.SessionID = sessionID
			}

			// 如果没有直接的字段，尝试从嵌套结构中提取
			// If no direct fields, try extracting from nested structures
			if extracted.Content == "" {
				// 尝试从 input 字段提取
				// Try to extract from input field
				if input, ok := rawData["input"].(string); ok {
					extracted.Content = input
				}
			}

			// 保存原始数据（用于审计）
			// Save raw data (for auditing)
			if config.EnableAudit {
				extracted.RawBody = bodyBytes
			}

		} else if contentType == "application/x-www-form-urlencoded" {
			// 处理表单数据
			// Handle form data
			if err := c.Request.ParseForm(); err != nil {
				config.Logger.Debug("failed to parse form, skipping extraction",
					"error", err,
					"path", path,
				)
				c.Next()
				return
			}

			extracted.Content = c.Request.FormValue("content")
			extracted.UserID = c.Request.FormValue("user_id")
			extracted.SessionID = c.Request.FormValue("session_id")

			// 尝试解析 metadata JSON
			// Try to parse metadata JSON
			if metadataStr := c.Request.FormValue("metadata"); metadataStr != "" {
				var metadata map[string]interface{}
				if err := json.Unmarshal([]byte(metadataStr), &metadata); err == nil {
					extracted.Metadata = metadata
				}
			}

			if config.EnableAudit {
				extracted.RawBody = bodyBytes
			}
		}

		// 注入到 context
		// Inject into context
		c.Set("extracted_content", extracted)

		// 如果有 content，也单独设置方便访问
		// If content exists, also set it separately for easy access
		if extracted.Content != "" {
			c.Set("content", extracted.Content)
		}
		if extracted.UserID != "" {
			c.Set("user_id", extracted.UserID)
		}
		if extracted.SessionID != "" {
			c.Set("session_id", extracted.SessionID)
		}
		if extracted.Metadata != nil {
			c.Set("metadata", extracted.Metadata)
		}

		// 记录日志
		// Log extraction
		config.Logger.Debug("content extracted",
			"path", path,
			"has_content", extracted.Content != "",
			"has_metadata", extracted.Metadata != nil,
			"user_id", extracted.UserID,
			"session_id", extracted.SessionID,
		)

		c.Next()
	}
}

// GetExtractedContent 从 Gin Context 中获取抽取的内容
// GetExtractedContent retrieves extracted content from Gin Context
func GetExtractedContent(c *gin.Context) (*ExtractedContent, bool) {
	if value, exists := c.Get("extracted_content"); exists {
		if extracted, ok := value.(ExtractedContent); ok {
			return &extracted, true
		}
	}
	return nil, false
}

// GetContent 从 Gin Context 中获取内容字符串
// GetContent retrieves content string from Gin Context
func GetContent(c *gin.Context) (string, bool) {
	if value, exists := c.Get("content"); exists {
		if content, ok := value.(string); ok {
			return content, true
		}
	}
	return "", false
}

// newBytesReader 创建一个字节读取器（内部辅助函数）
// newBytesReader creates a bytes reader (internal helper)
func newBytesReader(b []byte) io.Reader {
	return &bytesReader{data: b, pos: 0}
}

// bytesReader 简单的字节读取器实现
// bytesReader is a simple bytes reader implementation
type bytesReader struct {
	data []byte
	pos  int
}

func (r *bytesReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
