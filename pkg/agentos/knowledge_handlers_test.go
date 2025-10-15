package agentos

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rexleimo/agno-go/pkg/agno/vectordb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockVectorDB 是 VectorDB 接口的模拟实现
// MockVectorDB is a mock implementation of the VectorDB interface
type MockVectorDB struct {
	mock.Mock
}

func (m *MockVectorDB) CreateCollection(ctx context.Context, name string, metadata map[string]interface{}) error {
	args := m.Called(ctx, name, metadata)
	return args.Error(0)
}

func (m *MockVectorDB) DeleteCollection(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockVectorDB) Add(ctx context.Context, documents []vectordb.Document) error {
	args := m.Called(ctx, documents)
	return args.Error(0)
}

func (m *MockVectorDB) Update(ctx context.Context, documents []vectordb.Document) error {
	args := m.Called(ctx, documents)
	return args.Error(0)
}

func (m *MockVectorDB) Delete(ctx context.Context, ids []string) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *MockVectorDB) Query(ctx context.Context, query string, limit int, filter map[string]interface{}) ([]vectordb.SearchResult, error) {
	args := m.Called(ctx, query, limit, filter)
	if args.Get(0) == nil {
		return []vectordb.SearchResult{}, args.Error(1)
	}
	return args.Get(0).([]vectordb.SearchResult), args.Error(1)
}

func (m *MockVectorDB) QueryWithEmbedding(ctx context.Context, embedding []float32, limit int, filter map[string]interface{}) ([]vectordb.SearchResult, error) {
	args := m.Called(ctx, embedding, limit, filter)
	if args.Get(0) == nil {
		return []vectordb.SearchResult{}, args.Error(1)
	}
	return args.Get(0).([]vectordb.SearchResult), args.Error(1)
}

func (m *MockVectorDB) Get(ctx context.Context, ids []string) ([]vectordb.Document, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return []vectordb.Document{}, args.Error(1)
	}
	return args.Get(0).([]vectordb.Document), args.Error(1)
}

func (m *MockVectorDB) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockVectorDB) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockEmbeddingFunc 是 EmbeddingFunction 的模拟实现
// MockEmbeddingFunc is a mock implementation of the EmbeddingFunction interface
type MockEmbeddingFunc struct {
	mock.Mock
}

func (m *MockEmbeddingFunc) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	args := m.Called(ctx, texts)
	if args.Get(0) == nil {
		return [][]float32{}, args.Error(1)
	}
	return args.Get(0).([][]float32), args.Error(1)
}

func (m *MockEmbeddingFunc) EmbedSingle(ctx context.Context, text string) ([]float32, error) {
	args := m.Called(ctx, text)
	if args.Get(0) == nil {
		return []float32{}, args.Error(1)
	}
	return args.Get(0).([]float32), args.Error(1)
}

func TestHandleKnowledgeSearch_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建 mock
	// Create mocks
	mockVectorDB := new(MockVectorDB)
	mockEmbedding := new(MockEmbeddingFunc)

	// 模拟搜索结果
	// Mock search results
	mockResults := []vectordb.SearchResult{
		{
			ID:       "doc1",
			Content:  "test content 1",
			Score:    0.95,
			Distance: 0.05,
			Metadata: map[string]interface{}{"source": "test"},
		},
		{
			ID:       "doc2",
			Content:  "test content 2",
			Score:    0.85,
			Distance: 0.15,
			Metadata: map[string]interface{}{"source": "test"},
		},
	}

	mockVectorDB.On("Query", mock.Anything, "test query", 5, mock.Anything).
		Return(mockResults, nil)

	// 创建知识库服务
	// Create knowledge service
	knowledgeSvc := NewKnowledgeService(mockVectorDB, mockEmbedding, KnowledgeServiceConfig{
		DefaultLimit: 10,
		MaxLimit:     100,
	})

	// 创建服务器
	// Create server
	server := &Server{
		knowledgeService: knowledgeSvc,
		logger:           slog.Default(),
	}

	// 创建请求
	// Create request
	router := gin.New()
	router.POST("/search", server.handleKnowledgeSearch)

	req := VectorSearchRequest{
		Query:  "test query",
		Limit:  5,
		Offset: 0,
	}
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/search", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	// 断言
	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response VectorSearchResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(response.Results))
	assert.Equal(t, "doc1", response.Results[0].ID)
	assert.Equal(t, "test content 1", response.Results[0].Content)

	mockVectorDB.AssertExpectations(t)
}

func TestHandleKnowledgeSearch_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockVectorDB := new(MockVectorDB)
	mockEmbedding := new(MockEmbeddingFunc)

	knowledgeSvc := NewKnowledgeService(mockVectorDB, mockEmbedding, KnowledgeServiceConfig{})

	server := &Server{
		knowledgeService: knowledgeSvc,
		logger:           slog.Default(),
	}

	router := gin.New()
	router.POST("/search", server.handleKnowledgeSearch)

	// 缺少必需的 query 字段
	// Missing required query field
	req := VectorSearchRequest{
		Limit: 10,
	}
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/search", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleKnowledgeSearch_Pagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockVectorDB := new(MockVectorDB)
	mockEmbedding := new(MockEmbeddingFunc)

	// 模拟 3 个结果
	// Mock 3 results
	mockResults := []vectordb.SearchResult{
		{ID: "doc1", Content: "content1", Score: 0.9, Distance: 0.1},
		{ID: "doc2", Content: "content2", Score: 0.8, Distance: 0.2},
		{ID: "doc3", Content: "content3", Score: 0.7, Distance: 0.3},
	}

	mockVectorDB.On("Query", mock.Anything, "test query", 3, mock.Anything).
		Return(mockResults, nil)

	knowledgeSvc := NewKnowledgeService(mockVectorDB, mockEmbedding, KnowledgeServiceConfig{})

	server := &Server{
		knowledgeService: knowledgeSvc,
		logger:           slog.Default(),
	}

	router := gin.New()
	router.POST("/search", server.handleKnowledgeSearch)

	// 测试 offset=1, limit=2，应该返回 doc2 和 doc3
	// Test offset=1, limit=2, should return doc2 and doc3
	req := VectorSearchRequest{
		Query:  "test query",
		Limit:  2,
		Offset: 1,
	}
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/search", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response VectorSearchResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(response.Results))
	assert.Equal(t, "doc2", response.Results[0].ID)
	assert.Equal(t, "doc3", response.Results[1].ID)
	assert.Equal(t, 3, response.Meta.TotalCount)
	assert.Equal(t, 1, response.Meta.Page) // page = (offset / limit) + 1 = (1 / 2) + 1 = 1 (整数除法)

	mockVectorDB.AssertExpectations(t)
}

func TestHandleKnowledgeSearch_ServiceNotConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 服务器没有配置知识库服务
	// Server without knowledge service configured
	server := &Server{
		knowledgeService: nil,
		logger:           slog.Default(),
	}

	router := gin.New()
	router.POST("/search", server.handleKnowledgeSearch)

	req := VectorSearchRequest{
		Query: "test query",
	}
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/search", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestHandleKnowledgeConfig_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockVectorDB := new(MockVectorDB)
	mockEmbedding := new(MockEmbeddingFunc)

	knowledgeSvc := NewKnowledgeService(mockVectorDB, mockEmbedding, KnowledgeServiceConfig{
		DefaultChunkerType:  "character",
		DefaultChunkSize:    1000,
		EmbeddingProvider:   "openai",
		EmbeddingModel:      "text-embedding-3-small",
		EmbeddingDimensions: 1536,
	})

	server := &Server{
		knowledgeService: knowledgeSvc,
		logger:           slog.Default(),
	}

	router := gin.New()
	router.GET("/config", server.handleKnowledgeConfig)

	httpReq := httptest.NewRequest(http.MethodGet, "/config", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response KnowledgeConfigResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, 3, len(response.AvailableChunkers))
	assert.Contains(t, response.AvailableVectorDBs, "chromadb")
	assert.Equal(t, "character", response.DefaultChunker)
	assert.Equal(t, "openai", response.EmbeddingModel.Provider)
	assert.Equal(t, "text-embedding-3-small", response.EmbeddingModel.Model)
	assert.Equal(t, 1536, response.EmbeddingModel.Dimensions)
}

func TestHandleKnowledgeConfig_ServiceNotConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)

	server := &Server{
		knowledgeService: nil,
		logger:           slog.Default(),
	}

	router := gin.New()
	router.GET("/config", server.handleKnowledgeConfig)

	httpReq := httptest.NewRequest(http.MethodGet, "/config", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestNewKnowledgeService_Defaults(t *testing.T) {
	mockVectorDB := new(MockVectorDB)
	mockEmbedding := new(MockEmbeddingFunc)

	// 测试默认值设置
	// Test default value setting
	svc := NewKnowledgeService(mockVectorDB, mockEmbedding, KnowledgeServiceConfig{})

	assert.Equal(t, 10, svc.config.DefaultLimit)
	assert.Equal(t, 100, svc.config.MaxLimit)
	assert.Equal(t, "character", svc.config.DefaultChunkerType)
	assert.Equal(t, 1000, svc.config.DefaultChunkSize)
	assert.Equal(t, 0, svc.config.DefaultOverlap) // 传入的是 0，代码只在 < 0 时设置默认值
	assert.Equal(t, "openai", svc.config.EmbeddingProvider)
	assert.Equal(t, "text-embedding-3-small", svc.config.EmbeddingModel)
	assert.Equal(t, 1536, svc.config.EmbeddingDimensions)
}
