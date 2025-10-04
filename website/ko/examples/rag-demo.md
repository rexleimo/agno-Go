# RAG (검색 증강 생성) 데모

## 개요

이 예제는 Agno-Go를 사용하여 RAG 시스템을 구축하는 방법을 보여줍니다. RAG는 지식 베이스에서의 정보 검색과 LLM 텍스트 생성을 결합하여 정확하고 근거 있는 답변을 제공합니다. 이 시스템은 벡터 저장을 위한 ChromaDB, OpenAI 임베딩, 그리고 의미론적 검색 기능을 활성화하는 사용자 정의 RAG 툴킷을 사용합니다.

## 학습 내용

- OpenAI 임베딩을 생성하고 사용하는 방법
- 벡터 데이터베이스로 ChromaDB를 설정하는 방법
- 최적의 검색을 위해 문서를 청크로 나누는 방법
- Agent용 사용자 정의 RAG 툴킷을 구축하는 방법
- 지식 검색 기능을 가진 Agent를 만드는 방법
- RAG 구현의 모범 사례

## 사전 요구 사항

- Go 1.21 이상
- OpenAI API 키
- ChromaDB가 로컬에서 실행 중 (Docker를 통해)

## 설정

1. OpenAI API 키 설정:
```bash
export OPENAI_API_KEY=sk-your-api-key-here
```

2. Docker를 사용하여 ChromaDB 시작:
```bash
docker pull chromadb/chroma
docker run -p 8000:8000 chromadb/chroma
```

3. 예제 디렉토리로 이동:
```bash
cd cmd/examples/rag_demo
```

## 전체 코드

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

## 코드 설명

### 1. 사용자 정의 RAG 툴킷

```go
type RAGToolkit struct {
	*toolkit.BaseToolkit
	vectorDB vectordb.VectorDB
}
```

RAG 툴킷은 벡터 데이터베이스를 래핑하고 Agent가 호출할 수 있는 `search_knowledge` 함수를 노출합니다:

- 쿼리 문자열과 선택적 제한을 받음
- 의미론적 유사성을 사용하여 벡터 데이터베이스 검색
- 관련성 점수와 함께 형식화된 결과 반환

### 2. OpenAI 임베딩

```go
embedFunc, err := openaiembed.New(openaiembed.Config{
	APIKey: openaiKey,
	Model:  "text-embedding-3-small",
})
```

- OpenAI의 `text-embedding-3-small` 모델 사용 (1536 차원)
- 텍스트를 밀집 벡터 표현으로 변환
- 의미론적 유사성 검색 활성화

### 3. ChromaDB 벡터 데이터베이스

```go
db, err := chromadb.New(chromadb.Config{
	BaseURL:           "http://localhost:8000",
	CollectionName:    "rag_demo",
	EmbeddingFunction: embedFunc,
})
```

- 로컬 ChromaDB 인스턴스에 연결
- 문서 임베딩 저장을 위한 컬렉션 생성
- 문서가 추가될 때 자동으로 임베딩 생성

### 4. 문서 청킹

```go
chunker := knowledge.NewCharacterChunker(500, 50)
```

- 문서를 500자 청크로 분할
- 청크 간 50자 중복
- 청크 경계를 넘어 컨텍스트 보존
- 긴 문서의 검색 정확도 향상

### 5. RAG 기능을 가진 Agent

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

지침은 Agent에게 다음을 알려줍니다:
- 지식 검색 도구를 사용할 때
- 검색된 정보를 통합하는 방법
- 투명성을 위해 출처를 인용할 것
- 정보를 사용할 수 없을 때 정직할 것

## 예제 실행

```bash
# ChromaDB가 실행 중인지 확인
docker run -p 8000:8000 chromadb/chroma

# 데모 실행
go run main.go
```

## 예상 출력

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

## 주요 개념

### RAG 파이프라인

1. **수집**: 문서가 로드되고 청크로 분할됨
2. **임베딩**: 청크가 벡터 임베딩으로 변환됨
3. **저장**: 임베딩이 벡터 데이터베이스에 저장됨
4. **검색**: 사용자 쿼리가 임베딩되고 유사한 문서가 발견됨
5. **생성**: LLM이 검색된 컨텍스트를 기반으로 답변 생성

### 의미론적 검색

키워드 검색과 달리 의미론적 검색은 의미를 기반으로 문서를 찾습니다:

- 쿼리: "What is ML?"
- 일치: "Machine Learning", "neural networks", "training models"에 대한 문서
- 정확한 키워드 매칭보다 나음

### 청킹 전략

문서 청킹 매개변수는 검색 품질에 영향을 미칩니다:

- **청크 크기 (500)**: 컨텍스트와 정밀도 사이의 균형
  - 너무 작음: 컨텍스트 손실
  - 너무 큼: 관련 없는 콘텐츠 검색

- **중복 (50)**: 중요한 정보 분할 방지
  - 청크 간 연속성 보장
  - 경계를 넘어가는 문장에 중요

### 사용자 정의 툴킷

RAG 툴킷은 사용자 정의 도구를 구축하는 방법을 보여줍니다:

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

주요 요소:
- LLM이 이해할 수 있는 명확한 이름과 설명
- 타입이 잘 정의된 매개변수
- 로직을 구현하는 핸들러 함수

## 고급 기능

### 메타데이터 필터링

메타데이터로 결과를 필터링할 수 있습니다:

```go
results, err := db.Query(ctx, query, limit, map[string]interface{}{
	"topic": "Machine Learning",
	"date": map[string]interface{}{
		"$gte": "2025-01-01",
	},
})
```

### 하이브리드 검색

의미론적 검색과 키워드 검색 결합:

```go
// 향후 기능 - 아직 구현되지 않음
results, err := db.HybridQuery(ctx, query, limit, HybridConfig{
	SemanticWeight: 0.7,
	KeywordWeight:  0.3,
})
```

### 재순위

더 강력한 모델로 재순위하여 결과 향상:

```go
// 향후 기능 - 아직 구현되지 않음
reranked := reranker.Rerank(results, query, limit)
```

## 모범 사례

1. **청크 크기**: 500-1000자로 시작하여 콘텐츠에 따라 조정
2. **중복**: 청크 크기의 10-20% 중복 사용
3. **임베딩**: 가능한 경우 도메인별 임베딩 사용
4. **메타데이터**: 필터링 및 인용을 위해 메타데이터 포함
5. **오류 처리**: 항상 검색 실패를 우아하게 처리
6. **캐싱**: 자주 액세스하는 임베딩 캐싱 고려
7. **모니터링**: 검색 품질을 추적하고 매개변수 조정

## 문제 해결

**오류: "Failed to connect to ChromaDB"**
- ChromaDB가 실행 중인지 확인: `docker ps | grep chroma`
- 포트 확인 (기본값: 8000): `curl http://localhost:8000/api/v1/heartbeat`

**오류: "OPENAI_API_KEY environment variable is required"**
- API 키 설정: `export OPENAI_API_KEY=sk-...`

**낮은 검색 품질**
- 청크 크기와 중복 조정
- 다른 임베딩 모델 시도
- 더 많은 관련 문서 추가
- 메타데이터 필터링 사용

**높은 지연 시간**
- 더 작은 임베딩 모델 사용
- 결과 수 줄이기 (limit 매개변수)
- 임베딩 캐싱 고려
- 로컬 임베딩 모델 사용 (향후 기능)

## 다음 단계

- 기본 Agent 사용을 위해 [Simple Agent](./simple-agent.md) 탐색
- 여러 Agent와 [Team 협업](./team-demo.md)에 대해 배우기
- 복잡한 RAG 파이프라인을 위한 [Workflow 엔진](./workflow-demo.md) 시도
- [AgentOS API](../api/agentos.md)로 프로덕션 RAG 구축

## 추가 리소스

- [OpenAI 임베딩 가이드](https://platform.openai.com/docs/guides/embeddings)
- [ChromaDB 문서](https://docs.trychroma.com/)
- [RAG 모범 사례](https://docs.agno.com/advanced/rag)
- [벡터 데이터베이스 비교](https://docs.agno.com/storage/vector-dbs)
