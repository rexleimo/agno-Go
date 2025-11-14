package agentos

// VectorSearchRequest 表示向量搜索请求
// VectorSearchRequest represents a vector search request
type VectorSearchRequest struct {
	// Query 查询文本
	// Query is the search query text
	Query string `json:"query" binding:"required"`

	// Limit 返回结果数量限制（默认: 10, 最大: 100）
	// Limit is the maximum number of results to return
	Limit int `json:"limit,omitempty"`

	// Offset 分页偏移量（默认: 0）
	// Offset is the pagination offset
	Offset int `json:"offset,omitempty"`

	// Filters 元数据过滤条件
	// Filters are metadata filter conditions
	Filters map[string]interface{} `json:"filters,omitempty"`

	// CollectionName 集合名称（可选，默认使用配置的集合）
	// CollectionName is the collection to search in
	CollectionName string `json:"collection_name,omitempty"`
}

// VectorSearchResult 表示单个搜索结果
// VectorSearchResult represents a single search result
type VectorSearchResult struct {
	// ID 文档 ID
	// ID is the document identifier
	ID string `json:"id"`

	// Content 文档内容
	// Content is the document content
	Content string `json:"content"`

	// Metadata 文档元数据
	// Metadata is the document metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Score 相似度得分（越高越相似）
	// Score is the similarity score (higher is better)
	Score float32 `json:"score"`

	// Distance 距离度量（越低越相似）
	// Distance is the distance metric (lower is better)
	Distance float32 `json:"distance"`
}

// VectorSearchResponse 表示搜索响应
// VectorSearchResponse represents the search response
type VectorSearchResponse struct {
	// Results 搜索结果列表
	// Results is the list of search results
	Results []VectorSearchResult `json:"results"`

	// Meta 分页元数据
	// Meta is pagination metadata
	Meta PaginationMeta `json:"meta"`
}

// PaginationMeta 表示分页元数据
// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	// TotalCount 总结果数（注意：向量数据库可能不支持精确计数）
	// TotalCount is the total number of results
	TotalCount int `json:"total_count"`

	// Page 当前页码（基于 offset 和 limit 计算）
	// Page is the current page number
	Page int `json:"page"`

	// PageSize 每页大小
	// PageSize is the number of results per page
	PageSize int `json:"page_size"`

	// TotalPages 总页数（基于 total_count 和 page_size 计算）
	// TotalPages is the total number of pages
	TotalPages int `json:"total_pages"`

	// Offset 当前偏移量
	Offset int `json:"offset"`

	// HasMore 是否还有更多结果
	HasMore bool `json:"has_more"`

	// NextOffset 下一页的偏移量
	NextOffset int `json:"next_offset,omitempty"`
}

// KnowledgeConfigResponse 表示知识库配置响应
// KnowledgeConfigResponse represents the knowledge configuration response
type KnowledgeConfigResponse struct {
	// AvailableChunkers 可用的分块器列表
	// AvailableChunkers is the list of available chunkers
	AvailableChunkers []ChunkerInfo `json:"available_chunkers"`

	// AvailableVectorDBs 可用的向量数据库列表
	// AvailableVectorDBs is the list of available vector databases
	AvailableVectorDBs []string `json:"available_vector_dbs"`

	// DefaultChunker 默认分块器
	// DefaultChunker is the default chunker
	DefaultChunker string `json:"default_chunker"`

	// DefaultVectorDB 默认向量数据库
	// DefaultVectorDB is the default vector database
	DefaultVectorDB string `json:"default_vector_db"`

	// EmbeddingModel 嵌入模型信息
	// EmbeddingModel is the embedding model information
	EmbeddingModel EmbeddingModelInfo `json:"embedding_model"`

	// Features 功能开关
	Features KnowledgeFeatures `json:"features"`

	// Limits 搜索限制
	Limits KnowledgeLimits `json:"limits"`

	// DefaultCollection 默认集合
	DefaultCollection string `json:"default_collection"`

	// AllowedCollections 允许的集合列表
	AllowedCollections []string `json:"allowed_collections,omitempty"`

	// AllowedSourceSchemes 允许的来源 URL scheme
	AllowedSourceSchemes []string `json:"allowed_source_schemes,omitempty"`
}

// KnowledgeFeatures describes enabled capabilities.
type KnowledgeFeatures struct {
	SearchEnabled    bool `json:"search_enabled"`
	IngestionEnabled bool `json:"ingestion_enabled"`
	HealthEnabled    bool `json:"health_enabled"`
}

// KnowledgeLimits captures pagination defaults and caps.
type KnowledgeLimits struct {
	DefaultLimit int `json:"default_limit"`
	MaxLimit     int `json:"max_limit"`
}

// ChunkerInfo 表示分块器信息
// ChunkerInfo represents chunker information
type ChunkerInfo struct {
	// Name 分块器名称
	// Name is the chunker name
	Name string `json:"name"`

	// Description 分块器描述
	// Description is the chunker description
	Description string `json:"description"`

	// DefaultChunkSize 默认块大小
	// DefaultChunkSize is the default chunk size
	DefaultChunkSize int `json:"default_chunk_size,omitempty"`

	// DefaultOverlap 默认重叠大小
	// DefaultOverlap is the default overlap size
	DefaultOverlap int `json:"default_overlap,omitempty"`

	// Metadata describes additional chunker-specific configuration.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// EmbeddingModelInfo 表示嵌入模型信息
// EmbeddingModelInfo represents embedding model information
type EmbeddingModelInfo struct {
	// Provider 提供商（openai, etc.）
	// Provider is the model provider
	Provider string `json:"provider"`

	// Model 模型名称
	// Model is the model name
	Model string `json:"model"`

	// Dimensions 嵌入维度
	// Dimensions is the embedding dimensions
	Dimensions int `json:"dimensions"`
}

// AddContentRequest 表示添加内容请求（P2 任务，预留）
// AddContentRequest represents an add content request (P2 task, reserved)
type AddContentRequest struct {
	// Content 文本内容
	// Content is the text content
	Content string `json:"content" binding:"required"`

	// Metadata 元数据
	// Metadata is the metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// ChunkerType 分块器类型（character/sentence/paragraph）
	// ChunkerType is the chunker type
	ChunkerType string `json:"chunker_type,omitempty"`

	// ChunkSize 块大小
	// ChunkSize is the chunk size
	ChunkSize int `json:"chunk_size,omitempty"`

	// ChunkOverlap specifies overlap between consecutive chunks
	ChunkOverlap int `json:"chunk_overlap,omitempty"`

	// CollectionName 集合名称
	// CollectionName is the collection name
	CollectionName string `json:"collection_name,omitempty"`
}
