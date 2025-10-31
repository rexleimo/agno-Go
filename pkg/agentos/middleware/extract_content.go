package middleware

import (
	"bytes"
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

	// CustomExtractors 自定义内容抽取器，按顺序执行
	// CustomExtractors is an ordered list of custom extractors
	CustomExtractors []ContentExtractor

	// PreProcessors 内容预处理器，用于调整或重写原始 JSON
	// PreProcessors mutate the raw payload prior to extraction
	PreProcessors []ContentPreProcessor

	// Validators 内容验证器（可对 JSON 结果做 JSON Schema 校验）
	// Validators validate the processed payload (e.g. JSON Schema)
	Validators []ContentValidator
}

// ContentExtractor 自定义抽取器
// ContentExtractor defines pluggable extraction behaviour
type ContentExtractor func(ctx *gin.Context, body []byte, contentType string) (*ExtractedContent, map[string]interface{}, bool, error)

// ContentPreProcessor 在抽取前预处理原始数据
// ContentPreProcessor mutates the raw JSON payload prior to extraction
type ContentPreProcessor func(ctx *gin.Context, data map[string]interface{}) (map[string]interface{}, error)

// ContentValidator 验证处理后的数据
// ContentValidator validates the processed payload (e.g. JSON Schema)
type ContentValidator func(ctx *gin.Context, data map[string]interface{}) error

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

		contentType := c.ContentType()

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
		c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		var (
			extracted *ExtractedContent
			payload   map[string]interface{}
			handled   bool
		)

		for _, extractor := range config.CustomExtractors {
			if extractor == nil {
				continue
			}
			result, raw, ok, extractErr := extractor(c, bodyBytes, contentType)
			if extractErr != nil {
				writeValidationError(c, config.Logger, path, extractErr)
				return
			}
			if ok {
				extracted = result
				payload = copyMap(raw)
				handled = true
				break
			}
		}

		if !handled {
			switch contentType {
			case "application/json":
				var raw map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &raw); err != nil {
					config.Logger.Debug("failed to parse JSON, skipping extraction",
						"error", err,
						"path", path,
					)
					c.Next()
					return
				}
				payload = raw
				extracted = &ExtractedContent{}
			case "application/x-www-form-urlencoded":
				if err := c.Request.ParseForm(); err != nil {
					config.Logger.Debug("failed to parse form, skipping extraction",
						"error", err,
						"path", path,
					)
					c.Next()
					return
				}
				payload = formToMap(c)
				extracted = &ExtractedContent{}
			default:
				// 不支持的内容类型
				// Unsupported content type
				c.Next()
				return
			}
		}

		if extracted == nil {
			c.Next()
			return
		}

		if payload != nil {
			var processorErr error
			payload, processorErr = runPreProcessors(c, config.PreProcessors, payload)
			if processorErr != nil {
				writeValidationError(c, config.Logger, path, processorErr)
				return
			}

			if err := runValidators(c, config.Validators, payload); err != nil {
				writeValidationError(c, config.Logger, path, err)
				return
			}

			mergeExtractedFromMap(extracted, payload)
		}

		if config.EnableAudit {
			extracted.RawBody = bodyBytes
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
		if extractedPtr, ok := value.(*ExtractedContent); ok && extractedPtr != nil {
			return extractedPtr, true
		}
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

// ToolOutputPreProcessor 创建一个预处理器，将 tool_output 中的内容提升为一级字段
// ToolOutputPreProcessor lifts tool output content into top-level content/metadata
func ToolOutputPreProcessor(field string) ContentPreProcessor {
	if field == "" {
		field = "tool_output"
	}
	return func(ctx *gin.Context, data map[string]interface{}) (map[string]interface{}, error) {
		raw, ok := data[field]
		if !ok {
			return data, nil
		}

		switch value := raw.(type) {
		case map[string]interface{}:
			adoptToolOutput(value, data)
		case []interface{}:
			for _, item := range value {
				if entry, ok := item.(map[string]interface{}); ok {
					adoptToolOutput(entry, data)
				}
			}
		}
		return data, nil
	}
}

func adoptToolOutput(tool map[string]interface{}, target map[string]interface{}) {
	if content, ok := tool["content"].(string); ok && content != "" {
		if existing, okExisting := target["content"].(string); !okExisting || existing == "" {
			target["content"] = content
		}
	}
	if metadata, ok := tool["metadata"].(map[string]interface{}); ok {
		if target["metadata"] == nil {
			target["metadata"] = metadata
		} else if existing, ok := target["metadata"].(map[string]interface{}); ok {
			for k, v := range metadata {
				existing[k] = v
			}
		}
	}
}

func runPreProcessors(ctx *gin.Context, processors []ContentPreProcessor, data map[string]interface{}) (map[string]interface{}, error) {
	current := data
	for _, processor := range processors {
		if processor == nil {
			continue
		}
		next, err := processor(ctx, copyMap(current))
		if err != nil {
			return nil, err
		}
		if next == nil {
			current = make(map[string]interface{})
		} else {
			current = next
		}
	}
	return current, nil
}

func runValidators(ctx *gin.Context, validators []ContentValidator, data map[string]interface{}) error {
	for _, validator := range validators {
		if validator == nil {
			continue
		}
		if err := validator(ctx, copyMap(data)); err != nil {
			return err
		}
	}
	return nil
}

func mergeExtractedFromMap(dest *ExtractedContent, data map[string]interface{}) {
	if dest.Metadata == nil {
		dest.Metadata = map[string]interface{}{}
	}

	if content, ok := stringFromMap(data, "content"); ok && content != "" {
		dest.Content = content
	} else if dest.Content == "" {
		if input, ok := stringFromMap(data, "input"); ok && input != "" {
			dest.Content = input
		}
	}

	if userID, ok := stringFromMap(data, "user_id"); ok && userID != "" {
		dest.UserID = userID
	}
	if sessionID, ok := stringFromMap(data, "session_id"); ok && sessionID != "" {
		dest.SessionID = sessionID
	}

	if metadata, ok := data["metadata"].(map[string]interface{}); ok {
		dest.Metadata = metadata
	}
}

func formToMap(c *gin.Context) map[string]interface{} {
	result := make(map[string]interface{})
	for key, values := range c.Request.PostForm {
		if len(values) == 0 {
			continue
		}
		result[key] = values[0]
	}
	if metadataStr, ok := result["metadata"].(string); ok && metadataStr != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(metadataStr), &metadata); err == nil {
			result["metadata"] = metadata
		}
	}
	return result
}

func stringFromMap(data map[string]interface{}, key string) (string, bool) {
	if value, ok := data[key]; ok {
		if str, ok := value.(string); ok {
			return str, true
		}
	}
	return "", false
}

func copyMap(input map[string]interface{}) map[string]interface{} {
	if input == nil {
		return nil
	}
	clone := make(map[string]interface{}, len(input))
	for k, v := range input {
		clone[k] = v
	}
	return clone
}

func writeValidationError(c *gin.Context, logger *slog.Logger, path string, err error) {
	logger.Warn("content validation failed", "path", path, "error", err)
	c.JSON(http.StatusBadRequest, gin.H{
		"error":   "invalid_payload",
		"message": err.Error(),
	})
	c.Abort()
}
