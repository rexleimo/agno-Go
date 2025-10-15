package agentos

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rexleimo/agno-go/pkg/agno/vectordb"
)

// KnowledgeService 封装知识库操作
// KnowledgeService encapsulates knowledge operations
type KnowledgeService struct {
	vectorDB      vectordb.VectorDB
	embeddingFunc vectordb.EmbeddingFunction
	config        KnowledgeServiceConfig
}

// KnowledgeServiceConfig 知识库服务配置
// KnowledgeServiceConfig is the knowledge service configuration
type KnowledgeServiceConfig struct {
	// DefaultLimit 默认返回结果数量
	// DefaultLimit is the default number of results to return
	DefaultLimit int

	// MaxLimit 最大返回结果数量
	// MaxLimit is the maximum number of results to return
	MaxLimit int

	// DefaultChunkerType 默认分块器类型
	// DefaultChunkerType is the default chunker type
	DefaultChunkerType string

	// DefaultChunkSize 默认块大小
	// DefaultChunkSize is the default chunk size
	DefaultChunkSize int

	// DefaultOverlap 默认重叠大小
	// DefaultOverlap is the default overlap size
	DefaultOverlap int

	// EmbeddingProvider 嵌入模型提供商
	// EmbeddingProvider is the embedding model provider
	EmbeddingProvider string

	// EmbeddingModel 嵌入模型名称
	// EmbeddingModel is the embedding model name
	EmbeddingModel string

	// EmbeddingDimensions 嵌入维度
	// EmbeddingDimensions is the embedding dimensions
	EmbeddingDimensions int
}

// NewKnowledgeService 创建知识库服务
// NewKnowledgeService creates a new knowledge service
func NewKnowledgeService(vdb vectordb.VectorDB, embFunc vectordb.EmbeddingFunction, config KnowledgeServiceConfig) *KnowledgeService {
	// 设置默认值
	// Set default values
	if config.DefaultLimit <= 0 {
		config.DefaultLimit = 10
	}
	if config.MaxLimit <= 0 {
		config.MaxLimit = 100
	}
	if config.DefaultChunkerType == "" {
		config.DefaultChunkerType = "character"
	}
	if config.DefaultChunkSize <= 0 {
		config.DefaultChunkSize = 1000
	}
	if config.DefaultOverlap < 0 {
		config.DefaultOverlap = 100
	}
	if config.EmbeddingProvider == "" {
		config.EmbeddingProvider = "openai"
	}
	if config.EmbeddingModel == "" {
		config.EmbeddingModel = "text-embedding-3-small"
	}
	if config.EmbeddingDimensions <= 0 {
		config.EmbeddingDimensions = 1536
	}

	return &KnowledgeService{
		vectorDB:      vdb,
		embeddingFunc: embFunc,
		config:        config,
	}
}

// handleKnowledgeSearch 处理知识库搜索请求
// handleKnowledgeSearch handles knowledge search requests
func (s *Server) handleKnowledgeSearch(c *gin.Context) {
	var req VectorSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	// 参数验证和默认值
	// Validate and set default values
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// 获取知识库服务
	// Get knowledge service
	knowledgeSvc := s.getKnowledgeService()
	if knowledgeSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "service_unavailable",
			"message": "knowledge service not configured",
		})
		return
	}

	// 创建超时上下文
	// Create timeout context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// 执行向量搜索
	// Perform vector search
	results, err := knowledgeSvc.vectorDB.Query(ctx, req.Query, req.Limit+req.Offset, req.Filters)
	if err != nil {
		s.logger.Error("knowledge search failed", "error", err, "query", req.Query)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "search_failed",
			"message": fmt.Sprintf("failed to search: %v", err),
		})
		return
	}

	// 应用分页（向量数据库返回的是前 N 条，我们需要切片）
	// Apply pagination (vector DB returns top N, we need to slice)
	totalCount := len(results)
	paginatedResults := results
	if req.Offset < len(results) {
		paginatedResults = results[req.Offset:]
		if len(paginatedResults) > req.Limit {
			paginatedResults = paginatedResults[:req.Limit]
		}
	} else {
		paginatedResults = []vectordb.SearchResult{}
	}

	// 转换为响应格式
	// Convert to response format
	searchResults := make([]VectorSearchResult, len(paginatedResults))
	for i, r := range paginatedResults {
		searchResults[i] = VectorSearchResult{
			ID:       r.ID,
			Content:  r.Content,
			Metadata: r.Metadata,
			Score:    r.Score,
			Distance: r.Distance,
		}
	}

	// 计算分页元数据
	// Calculate pagination metadata
	pageSize := req.Limit
	page := (req.Offset / pageSize) + 1
	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	response := VectorSearchResponse{
		Results: searchResults,
		Meta: PaginationMeta{
			TotalCount: totalCount,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	}

	c.JSON(http.StatusOK, response)
}

// handleKnowledgeConfig 处理知识库配置查询
// handleKnowledgeConfig handles knowledge configuration query
func (s *Server) handleKnowledgeConfig(c *gin.Context) {
	knowledgeSvc := s.getKnowledgeService()
	if knowledgeSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "service_unavailable",
			"message": "knowledge service not configured",
		})
		return
	}

	// 构建配置响应
	// Build configuration response
	response := KnowledgeConfigResponse{
		AvailableChunkers: []ChunkerInfo{
			{
				Name:             "character",
				Description:      "按字符数量分块，适合通用文本",
				DefaultChunkSize: 1000,
				DefaultOverlap:   100,
			},
			{
				Name:             "sentence",
				Description:      "按句子分块，适合对话和文档",
				DefaultChunkSize: 1000,
				DefaultOverlap:   0,
			},
			{
				Name:             "paragraph",
				Description:      "按段落分块，适合长文档",
				DefaultChunkSize: 2000,
				DefaultOverlap:   0,
			},
		},
		AvailableVectorDBs: []string{"chromadb"},
		DefaultChunker:     knowledgeSvc.config.DefaultChunkerType,
		DefaultVectorDB:    "chromadb",
		EmbeddingModel: EmbeddingModelInfo{
			Provider:   knowledgeSvc.config.EmbeddingProvider,
			Model:      knowledgeSvc.config.EmbeddingModel,
			Dimensions: knowledgeSvc.config.EmbeddingDimensions,
		},
	}

	c.JSON(http.StatusOK, response)
}

// getKnowledgeService 获取知识库服务实例（从 Server 配置中）
// getKnowledgeService gets the knowledge service instance from server config
func (s *Server) getKnowledgeService() *KnowledgeService {
	// 这个方法会在 Server 结构体添加 knowledgeService 字段后正常工作
	// This method will work properly after adding knowledgeService field to Server struct
	return s.knowledgeService
}
