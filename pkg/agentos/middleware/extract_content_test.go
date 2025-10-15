package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestExtractContentMiddleware_JSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := ExtractContentConfig{
		MaxRequestSize: 1024 * 1024, // 1MB
		Logger:         slog.Default(),
		EnableAudit:    true,
	}

	router := gin.New()
	router.Use(ExtractContentMiddleware(config))
	router.POST("/test", func(c *gin.Context) {
		extracted, exists := GetExtractedContent(c)
		assert.True(t, exists)
		c.JSON(http.StatusOK, extracted)
	})

	// 测试数据
	// Test data
	requestData := map[string]interface{}{
		"content":    "Test content",
		"user_id":    "user123",
		"session_id": "session456",
		"metadata": map[string]interface{}{
			"source": "test",
			"page":   1,
		},
	}

	body, _ := json.Marshal(requestData)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ExtractedContent
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Test content", response.Content)
	assert.Equal(t, "user123", response.UserID)
	assert.Equal(t, "session456", response.SessionID)
	assert.NotNil(t, response.Metadata)
	assert.Equal(t, "test", response.Metadata["source"])
}

func TestExtractContentMiddleware_InputField(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := ExtractContentConfig{
		Logger: slog.Default(),
	}

	router := gin.New()
	router.Use(ExtractContentMiddleware(config))
	router.POST("/test", func(c *gin.Context) {
		content, exists := GetContent(c)
		assert.True(t, exists)
		c.JSON(http.StatusOK, gin.H{"content": content})
	})

	// 测试使用 input 字段而不是 content 字段
	// Test using input field instead of content field
	requestData := map[string]interface{}{
		"input": "Test input content",
	}

	body, _ := json.Marshal(requestData)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Test input content", response["content"])
}

func TestExtractContentMiddleware_Form(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := ExtractContentConfig{
		Logger: slog.Default(),
	}

	router := gin.New()
	router.Use(ExtractContentMiddleware(config))
	router.POST("/test", func(c *gin.Context) {
		extracted, exists := GetExtractedContent(c)
		assert.True(t, exists)
		c.JSON(http.StatusOK, extracted)
	})

	// 测试表单数据
	// Test form data
	form := url.Values{}
	form.Add("content", "Form content")
	form.Add("user_id", "user789")
	form.Add("session_id", "session012")
	form.Add("metadata", `{"source":"form","type":"test"}`)

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ExtractedContent
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Form content", response.Content)
	assert.Equal(t, "user789", response.UserID)
	assert.Equal(t, "session012", response.SessionID)
	assert.NotNil(t, response.Metadata)
	assert.Equal(t, "form", response.Metadata["source"])
}

func TestExtractContentMiddleware_SkipGET(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := ExtractContentConfig{
		Logger: slog.Default(),
	}

	router := gin.New()
	router.Use(ExtractContentMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		_, exists := GetExtractedContent(c)
		assert.False(t, exists) // GET 请求不应抽取内容
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestExtractContentMiddleware_SkipPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := ExtractContentConfig{
		Logger:    slog.Default(),
		SkipPaths: []string{"/health", "/metrics"},
	}

	router := gin.New()
	router.Use(ExtractContentMiddleware(config))
	router.POST("/health", func(c *gin.Context) {
		_, exists := GetExtractedContent(c)
		assert.False(t, exists) // 跳过的路径不应抽取内容
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	requestData := map[string]interface{}{
		"content": "Should be ignored",
	}
	body, _ := json.Marshal(requestData)
	req := httptest.NewRequest(http.MethodPost, "/health", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestExtractContentMiddleware_MaxSize(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := ExtractContentConfig{
		MaxRequestSize: 100, // 只允许 100 字节
		Logger:         slog.Default(),
	}

	router := gin.New()
	router.Use(ExtractContentMiddleware(config))
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// 创建超过限制的请求
	// Create request exceeding limit
	largeData := map[string]interface{}{
		"content": string(make([]byte, 200)), // 200 字节的内容
	}
	body, _ := json.Marshal(largeData)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
}

func TestExtractContentMiddleware_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := ExtractContentConfig{
		Logger: slog.Default(),
	}

	router := gin.New()
	router.Use(ExtractContentMiddleware(config))
	router.POST("/test", func(c *gin.Context) {
		// 无效 JSON 应该跳过抽取但继续处理
		// Invalid JSON should skip extraction but continue processing
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// 发送无效 JSON
	// Send invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestExtractContentMiddleware_ContextValues(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := ExtractContentConfig{
		Logger: slog.Default(),
	}

	var capturedContent string
	var capturedUserID string
	var capturedSessionID string
	var capturedMetadata map[string]interface{}

	router := gin.New()
	router.Use(ExtractContentMiddleware(config))
	router.POST("/test", func(c *gin.Context) {
		// 测试单独设置的值
		// Test individually set values
		content, _ := c.Get("content")
		userID, _ := c.Get("user_id")
		sessionID, _ := c.Get("session_id")
		metadata, _ := c.Get("metadata")

		capturedContent = content.(string)
		capturedUserID = userID.(string)
		capturedSessionID = sessionID.(string)
		capturedMetadata = metadata.(map[string]interface{})

		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	requestData := map[string]interface{}{
		"content":    "Test content",
		"user_id":    "user999",
		"session_id": "session888",
		"metadata": map[string]interface{}{
			"key": "value",
		},
	}

	body, _ := json.Marshal(requestData)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Test content", capturedContent)
	assert.Equal(t, "user999", capturedUserID)
	assert.Equal(t, "session888", capturedSessionID)
	assert.Equal(t, "value", capturedMetadata["key"])
}

func TestGetExtractedContent_NotExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	extracted, exists := GetExtractedContent(c)
	assert.False(t, exists)
	assert.Nil(t, extracted)
}

func TestGetContent_NotExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	content, exists := GetContent(c)
	assert.False(t, exists)
	assert.Empty(t, content)
}

func TestBytesReader(t *testing.T) {
	data := []byte("Hello, World!")
	reader := newBytesReader(data)

	// 读取所有数据
	// Read all data
	result, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, data, result)

	// 再次读取应该返回 EOF
	// Reading again should return EOF
	buf := make([]byte, 10)
	n, err := reader.Read(buf)
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, 0, n)
}
