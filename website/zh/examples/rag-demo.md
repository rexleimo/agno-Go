# RAG (检索增强生成) 演示

## 概述

本示例演示如何使用 Agno-Go 构建 RAG 系统。RAG 将从知识库检索信息与 LLM 文本生成相结合,以提供准确、有根据的答案。该系统使用 ChromaDB 进行向量存储、OpenAI 嵌入和自定义 RAG toolkit 来实现语义搜索能力。

## 你将学到

- 如何创建和使用 OpenAI 嵌入
- 如何将 ChromaDB 设置为向量数据库
- 如何分块文档以实现最佳检索
- 如何为 Agent 构建自定义 RAG toolkit
- 如何创建具有知识检索能力的 Agent
- RAG 实现的最佳实践

## 前置要求

- Go 1.21 或更高版本
- OpenAI API key
- 本地运行的 ChromaDB (通过 Docker)

## 设置

1. 设置你的 OpenAI API key:
```bash
export OPENAI_API_KEY=sk-your-api-key-here
```

2. 使用 Docker 启动 ChromaDB:
```bash
docker pull chromadb/chroma
docker run -p 8000:8000 chromadb/chroma
```

3. 进入示例目录:
```bash
cd cmd/examples/rag_demo
```

## 完整代码

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/rexleimo/agno-go/pkg/agno/agent"
	openaiembed "github.com/rexleimo/agno-go/pkg/agno/embeddings/openai"
	"github.com/rexleimo/agno-go/pkg/agno/knowledge"
	openaimodel "github.com/rexleimo/agno-go/pkg/agno/models/openai"
	"github.com/rexleimo/agno-go/pkg/agno/tools/toolkit"
	"github.com/rexleimo/agno-go/pkg/agno/vectordb"
	"github.com/rexleimo/agno-go/pkg/agno/vectordb/chromadb"
)

// RAGToolkit provides knowledge retrieval tools for the agent
type RAGToolkit struct {
	*toolkit.BaseToolkit
	vectorDB vectordb.VectorDB
}

// NewRAGToolkit creates a new RAG toolkit
func NewRAGToolkit(db vectordb.VectorDB) *RAGToolkit {
	t := &RAGToolkit{
		BaseToolkit: toolkit.NewBaseToolkit("knowledge_retrieval"),
		vectorDB:    db,
	}

	// Register search function
	t.RegisterFunction(&toolkit.Function{
		Name:        "search_knowledge",
		Description: "Search the knowledge base for relevant information. Use this to find answers to user questions.",
		Parameters: map[string]toolkit.Parameter{
			"query": {
				Type:        "string",
				Description: "The search query or question",
				Required:    true,
			},
			"limit": {
				Type:        "integer",
				Description: "Maximum number of results to return (default: 3)",
				Required:    false,
			},
		},
		Handler: t.searchKnowledge,
	})

	return t
}

func (t *RAGToolkit) searchKnowledge(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	query, ok := args["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query must be a string")
	}

	limit := 3
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	results, err := t.vectorDB.Query(ctx, query, limit, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search knowledge base: %w", err)
	}

	// Format results for the agent
	var formattedResults []map[string]interface{}
	for i, result := range results {
		formattedResults = append(formattedResults, map[string]interface{}{
			"rank":     i + 1,
			"content":  result.Content,
			"score":    result.Score,
			"metadata": result.Metadata,
		})
	}

	return formattedResults, nil
}

func main() {
	fmt.Println("🚀 RAG (Retrieval-Augmented Generation) Demo")
	fmt.Println("This example demonstrates:")
	fmt.Println("1. Loading documents from files")
	fmt.Println("2. Chunking text into smaller pieces")
	fmt.Println("3. Generating embeddings with OpenAI")
	fmt.Println("4. Storing in ChromaDB vector database")
	fmt.Println("5. Using RAG with an Agent to answer questions")
	fmt.Println()

	// Check environment variables
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	ctx := context.Background()

	// Step 1: Create embedding function
	fmt.Println("📊 Step 1: Creating OpenAI embedding function...")
	embedFunc, err := openaiembed.New(openaiembed.Config{
		APIKey: openaiKey,
		Model:  "text-embedding-3-small",
	})
	if err != nil {
		log.Fatalf("Failed to create embedding function: %v", err)
	}
	fmt.Printf("   ✅ Created embedding function (model: %s, dimensions: %d)\n\n",
		embedFunc.GetModel(), embedFunc.GetDimensions())

	// Step 2: Create ChromaDB vector database
	fmt.Println("💾 Step 2: Connecting to ChromaDB...")
	db, err := chromadb.New(chromadb.Config{
		BaseURL:           "http://localhost:8000",
		CollectionName:    "rag_demo",
		EmbeddingFunction: embedFunc,
	})
	if err != nil {
		log.Fatalf("Failed to create ChromaDB: %v", err)
	}
	defer db.Close()

	// Create collection
	err = db.CreateCollection(ctx, "", map[string]interface{}{
		"description": "RAG demo knowledge base",
	})
	if err != nil {
		log.Fatalf("Failed to create collection: %v", err)
	}
	fmt.Println("   ✅ Connected to ChromaDB and created collection")

	// Step 3: Load and process documents
	fmt.Println("📚 Step 3: Loading and processing documents...")

	// Sample documents about AI and ML
	sampleDocs := []knowledge.Document{
		{
			ID:      "doc1",
			Content: "Artificial Intelligence (AI) is the simulation of human intelligence by machines. AI systems can perform tasks that typically require human intelligence, such as visual perception, speech recognition, decision-making, and language translation. Modern AI is based on machine learning algorithms that can learn from data.",
			Metadata: map[string]interface{}{
				"topic": "AI Overview",
				"date":  "2025-01-01",
			},
		},
		{
			ID:      "doc2",
			Content: "Machine Learning (ML) is a subset of AI that focuses on creating systems that learn from data. Instead of being explicitly programmed, ML models improve their performance through experience. Common ML algorithms include neural networks, decision trees, and support vector machines.",
			Metadata: map[string]interface{}{
				"topic": "Machine Learning",
				"date":  "2025-01-01",
			},
		},
		{
			ID:      "doc3",
			Content: "Vector databases are specialized databases designed to store and query high-dimensional vector embeddings. They enable semantic search by finding similar vectors using distance metrics like cosine similarity or Euclidean distance. Vector databases are essential for RAG (Retrieval-Augmented Generation) systems.",
			Metadata: map[string]interface{}{
				"topic": "Vector Databases",
				"date":  "2025-01-01",
			},
		},
		{
			ID:      "doc4",
			Content: "Retrieval-Augmented Generation (RAG) combines information retrieval with text generation. It first retrieves relevant documents from a knowledge base, then uses a language model to generate responses based on the retrieved context. RAG improves accuracy and reduces hallucinations in AI systems.",
			Metadata: map[string]interface{}{
				"topic": "RAG",
				"date":  "2025-01-01",
			},
		},
		{
			ID:      "doc5",
			Content: "Large Language Models (LLMs) like GPT-4 are neural networks trained on vast amounts of text data. They can understand and generate human-like text, perform reasoning, answer questions, and even write code. LLMs are the foundation of modern AI assistants and chatbots.",
			Metadata: map[string]interface{}{
				"topic": "Large Language Models",
				"date":  "2025-01-01",
			},
		},
	}

	// Chunk documents (optional, useful for large documents)
	chunker := knowledge.NewCharacterChunker(500, 50)
	var allChunks []knowledge.Chunk
	for _, doc := range sampleDocs {
		chunks, err := chunker.Chunk(doc)
		if err != nil {
			log.Printf("Warning: Failed to chunk document %s: %v", doc.ID, err)
			continue
		}
		allChunks = append(allChunks, chunks...)
	}
	fmt.Printf("   ✅ Loaded %d documents, created %d chunks\n", len(sampleDocs), len(allChunks))

	// Step 4: Generate embeddings and store in vector DB
	fmt.Println("\n🔢 Step 4: Generating embeddings and storing in ChromaDB...")

	var vdbDocs []vectordb.Document
	for _, chunk := range allChunks {
		vdbDocs = append(vdbDocs, vectordb.Document{
			ID:       chunk.ID,
			Content:  chunk.Content,
			Metadata: chunk.Metadata,
			// Embedding will be generated automatically by ChromaDB
		})
	}

	err = db.Add(ctx, vdbDocs)
	if err != nil {
		log.Fatalf("Failed to add documents to vector DB: %v", err)
	}

	count, _ := db.Count(ctx)
	fmt.Printf("   ✅ Stored %d documents in vector database\n\n", count)

	// Step 5: Test retrieval
	fmt.Println("🔍 Step 5: Testing knowledge retrieval...")
	testQuery := "What is machine learning?"
	results, err := db.Query(ctx, testQuery, 2, nil)
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}

	fmt.Printf("   Query: \"%s\"\n", testQuery)
	fmt.Printf("   Found %d relevant documents:\n", len(results))
	for i, result := range results {
		fmt.Printf("   %d. [Score: %.4f] %s\n", i+1, result.Score,
			truncate(result.Content, 80))
	}
	fmt.Println()

	// Step 6: Create RAG-powered Agent
	fmt.Println("🤖 Step 6: Creating RAG-powered Agent...")

	// Create OpenAI model
	model, err := openaimodel.New("gpt-4o-mini", openaimodel.Config{
		APIKey:      openaiKey,
		Temperature: 0.7,
		MaxTokens:   500,
	})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// Create RAG toolkit
	ragToolkit := NewRAGToolkit(db)

	// Create agent with RAG capabilities
	ag, err := agent.New(agent.Config{
		Name:     "RAG Assistant",
		Model:    model,
		Toolkits: []toolkit.Toolkit{ragToolkit},
		Instructions: `You are a helpful AI assistant with access to a knowledge base.
When users ask questions:
1. Use the search_knowledge tool to find relevant information
2. Base your answer on the retrieved information
3. Cite the sources when possible
4. If you can't find relevant information, say so

Always be helpful, accurate, and concise.`,
		MaxLoops: 5,
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}
	fmt.Println("   ✅ Agent created with RAG capabilities")

	// Step 7: Interactive Q&A
	fmt.Println("💬 Step 7: Interactive Q&A (RAG in action)")
	fmt.Println("=" + string(make([]byte, 60)) + "=")

	questions := []string{
		"What is artificial intelligence?",
		"Explain the difference between AI and machine learning",
		"What are vector databases used for?",
		"How does RAG improve AI systems?",
	}

	for i, question := range questions {
		fmt.Printf("\n[Question %d] User: %s\n", i+1, question)

		output, err := ag.Run(ctx, question)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}

		fmt.Printf("Assistant: %s\n", output.Content)
	}

	fmt.Println("\n" + string(make([]byte, 60)) + "=")
	fmt.Println("\n✅ RAG Demo completed successfully!")
	fmt.Println("\nKey Takeaways:")
	fmt.Println("• Documents are chunked and embedded automatically")
	fmt.Println("• Vector database enables semantic search")
	fmt.Println("• Agent uses RAG to provide accurate, grounded answers")
	fmt.Println("• Citations and sources improve trustworthiness")

	// Cleanup
	fmt.Println("\n🧹 Cleaning up...")
	err = db.DeleteCollection(ctx, "rag_demo")
	if err != nil {
		log.Printf("Warning: Failed to delete collection: %v", err)
	} else {
		fmt.Println("   ✅ Deleted demo collection")
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
```

## 代码解释

### 1. 自定义 RAG Toolkit

```go
type RAGToolkit struct {
	*toolkit.BaseToolkit
	vectorDB vectordb.VectorDB
}
```

RAG toolkit 封装了向量数据库并暴露一个 `search_knowledge` 函数供 Agent 调用:

- 接受查询字符串和可选的限制数
- 使用语义相似度搜索向量数据库
- 返回带有相关性分数的格式化结果

### 2. OpenAI 嵌入

```go
embedFunc, err := openaiembed.New(openaiembed.Config{
	APIKey: openaiKey,
	Model:  "text-embedding-3-small",
})
```

- 使用 OpenAI 的 `text-embedding-3-small` 模型 (1536 维)
- 将文本转换为密集向量表示
- 实现语义相似度搜索

### 3. ChromaDB 向量数据库

```go
db, err := chromadb.New(chromadb.Config{
	BaseURL:           "http://localhost:8000",
	CollectionName:    "rag_demo",
	EmbeddingFunction: embedFunc,
})
```

- 连接到本地 ChromaDB 实例
- 创建用于存储文档嵌入的集合
- 添加文档时自动生成嵌入

### 4. 文档分块

```go
chunker := knowledge.NewCharacterChunker(500, 50)
```

- 将文档分割为 500 字符的块
- 块之间有 50 字符的重叠
- 保留跨块边界的上下文
- 改善长文档的检索准确性

### 5. 具有 RAG 能力的 Agent

```go
ag, err := agent.New(agent.Config{
	Name:     "RAG Assistant",
	Model:    model,
	Toolkits: []toolkit.Toolkit{ragToolkit},
	Instructions: `You are a helpful AI assistant with access to a knowledge base.
When users ask questions:
1. Use the search_knowledge tool to find relevant information
2. Base your answer on the retrieved information
3. Cite the sources when possible
4. If you can't find relevant information, say so`,
	MaxLoops: 5,
})
```

指令告诉 Agent:
- 何时使用知识检索工具
- 如何整合检索到的信息
- 引用来源以提高透明度
- 在信息不可用时保持诚实

## 运行示例

```bash
# 确保 ChromaDB 正在运行
docker run -p 8000:8000 chromadb/chroma

# 运行演示
go run main.go
```

## 预期输出

```
🚀 RAG (Retrieval-Augmented Generation) Demo
This example demonstrates:
1. Loading documents from files
2. Chunking text into smaller pieces
3. Generating embeddings with OpenAI
4. Storing in ChromaDB vector database
5. Using RAG with an Agent to answer questions

📊 Step 1: Creating OpenAI embedding function...
   ✅ Created embedding function (model: text-embedding-3-small, dimensions: 1536)

💾 Step 2: Connecting to ChromaDB...
   ✅ Connected to ChromaDB and created collection

📚 Step 3: Loading and processing documents...
   ✅ Loaded 5 documents, created 5 chunks

🔢 Step 4: Generating embeddings and storing in ChromaDB...
   ✅ Stored 5 documents in vector database

🔍 Step 5: Testing knowledge retrieval...
   Query: "What is machine learning?"
   Found 2 relevant documents:
   1. [Score: 0.8234] Machine Learning (ML) is a subset of AI that focuses on creating syste...
   2. [Score: 0.7156] Artificial Intelligence (AI) is the simulation of human intelligence b...

🤖 Step 6: Creating RAG-powered Agent...
   ✅ Agent created with RAG capabilities

💬 Step 7: Interactive Q&A (RAG in action)
============================================================

[Question 1] User: What is artificial intelligence?
Assistant: Based on the knowledge base, Artificial Intelligence (AI) is the simulation of human intelligence by machines. AI systems can perform tasks that typically require human intelligence, such as visual perception, speech recognition, decision-making, and language translation. Modern AI is based on machine learning algorithms that can learn from data.

[Question 2] User: Explain the difference between AI and machine learning
Assistant: AI (Artificial Intelligence) is the broader concept of machines simulating human intelligence to perform various tasks. Machine Learning (ML) is a subset of AI that specifically focuses on creating systems that learn from data rather than being explicitly programmed. So while all ML is AI, not all AI is ML. ML improves performance through experience, using algorithms like neural networks and decision trees.

[Question 3] User: What are vector databases used for?
Assistant: According to the knowledge base, vector databases are specialized databases designed to store and query high-dimensional vector embeddings. They enable semantic search by finding similar vectors using distance metrics like cosine similarity or Euclidean distance. Vector databases are essential for RAG (Retrieval-Augmented Generation) systems.

[Question 4] User: How does RAG improve AI systems?
Assistant: RAG (Retrieval-Augmented Generation) improves AI systems by combining information retrieval with text generation. The process first retrieves relevant documents from a knowledge base, then uses a language model to generate responses based on the retrieved context. This approach improves accuracy and reduces hallucinations in AI systems by grounding responses in actual knowledge.

============================================================

✅ RAG Demo completed successfully!

Key Takeaways:
• Documents are chunked and embedded automatically
• Vector database enables semantic search
• Agent uses RAG to provide accurate, grounded answers
• Citations and sources improve trustworthiness

🧹 Cleaning up...
   ✅ Deleted demo collection
```

## 核心概念

### RAG 管道

1. **摄取**: 加载和分块文档
2. **嵌入**: 将块转换为向量嵌入
3. **存储**: 将嵌入存储在向量数据库中
4. **检索**: 用户查询被嵌入并找到相似文档
5. **生成**: LLM 基于检索到的上下文生成答案

### 语义搜索

与关键词搜索不同,语义搜索基于含义查找文档:

- 查询: "What is ML?"
- 匹配: 关于 "Machine Learning"、"neural networks"、"training models" 的文档
- 优于精确关键词匹配

### 分块策略

文档分块参数影响检索质量:

- **块大小 (500)**: 在上下文和精确性之间平衡
  - 太小: 丢失上下文
  - 太大: 检索不相关内容

- **重叠 (50)**: 防止分割重要信息
  - 确保跨块的连续性
  - 对于跨越边界的句子至关重要

### 自定义 Toolkit

RAG toolkit 演示了如何构建自定义工具:

```go
t.RegisterFunction(&toolkit.Function{
	Name:        "search_knowledge",
	Description: "Search the knowledge base...",
	Parameters: map[string]toolkit.Parameter{
		"query": {
			Type:        "string",
			Description: "The search query",
			Required:    true,
		},
	},
	Handler: t.searchKnowledge,
})
```

关键要素:
- 清晰的名称和描述供 LLM 理解
- 定义良好的参数和类型
- 实现逻辑的处理函数

## 高级特性

### 元数据过滤

你可以按元数据过滤结果:

```go
results, err := db.Query(ctx, query, limit, map[string]interface{}{
	"topic": "Machine Learning",
	"date": map[string]interface{}{
		"$gte": "2025-01-01",
	},
})
```

### 混合搜索

结合语义和关键词搜索:

```go
// 未来功能 - 尚未实现
results, err := db.HybridQuery(ctx, query, limit, HybridConfig{
	SemanticWeight: 0.7,
	KeywordWeight:  0.3,
})
```

### 重排序

使用更强大的模型重排序以改善结果:

```go
// 未来功能 - 尚未实现
reranked := reranker.Rerank(results, query, limit)
```

## 最佳实践

1. **块大小**: 从 500-1000 字符开始,根据你的内容调整
2. **重叠**: 使用块大小的 10-20% 作为重叠
3. **嵌入**: 可用时使用领域特定的嵌入
4. **元数据**: 包含元数据用于过滤和引用
5. **错误处理**: 总是优雅地处理检索失败
6. **缓存**: 考虑缓存频繁访问的嵌入
7. **监控**: 跟踪检索质量并调整参数

## 故障排除

**错误: "Failed to connect to ChromaDB"**
- 确保 ChromaDB 正在运行: `docker ps | grep chroma`
- 检查端口 (默认: 8000): `curl http://localhost:8000/api/v1/heartbeat`

**错误: "OPENAI_API_KEY environment variable is required"**
- 设置你的 API key: `export OPENAI_API_KEY=sk-...`

**检索质量低**
- 调整块大小和重叠
- 尝试不同的嵌入模型
- 添加更多相关文档
- 使用元数据过滤

**高延迟**
- 使用更小的嵌入模型
- 减少结果数量 (limit 参数)
- 考虑缓存嵌入
- 使用本地嵌入模型 (未来功能)

## 下一步

- 探索 [Simple Agent](./simple-agent.md) 了解基本 Agent 用法
- 了解 [Team 协作](./team-demo.md) 使用多个 Agent
- 尝试 [Workflow 引擎](./workflow-demo.md) 构建复杂的 RAG 管道
- 使用 [AgentOS API](../api/agentos.md) 构建生产 RAG

## 其他资源

- [OpenAI Embeddings 指南](https://platform.openai.com/docs/guides/embeddings)
- [ChromaDB 文档](https://docs.trychroma.com/)
- [RAG 最佳实践](https://docs.agno.com/advanced/rag)
- [向量数据库比较](https://docs.agno.com/storage/vector-dbs)
