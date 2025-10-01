# RAG (Retrieval-Augmented Generation) Demo

A comprehensive example demonstrating the complete RAG pipeline in Agno-Go.

## What is RAG?

**Retrieval-Augmented Generation (RAG)** combines information retrieval with text generation to create more accurate and grounded AI responses. Instead of relying solely on the language model's training data, RAG:

1. **Retrieves** relevant documents from a knowledge base
2. **Augments** the prompt with retrieved context
3. **Generates** responses based on factual information

### Benefits of RAG

- ✅ **Reduced Hallucinations**: Answers based on real data
- ✅ **Up-to-date Information**: Query current knowledge bases
- ✅ **Source Citation**: Track where answers come from
- ✅ **Domain Expertise**: Add specialized knowledge easily

## Architecture

```
User Query
    ↓
[1. Generate Query Embedding]
    ↓
[2. Vector Search (ChromaDB)]
    ↓
[3. Retrieve Top-K Documents]
    ↓
[4. Agent + LLM (with context)]
    ↓
Grounded Response
```

## Prerequisites

### 1. ChromaDB Server

Start ChromaDB using Docker:

```bash
docker pull chromadb/chroma
docker run -p 8000:8000 chromadb/chroma
```

Or run locally:

```bash
pip install chromadb
chroma run --path /db_path
```

### 2. OpenAI API Key

```bash
export OPENAI_API_KEY=your-openai-api-key
```

## Running the Demo

```bash
# Make sure ChromaDB is running
docker run -p 8000:8000 chromadb/chroma

# Set API key
export OPENAI_API_KEY=your-api-key

# Run the demo
go run cmd/examples/rag_demo/main.go
```

## What the Demo Does

### Step-by-Step Process

1. **Creates Embedding Function**
   - Uses OpenAI `text-embedding-3-small` (1536 dimensions)
   - Converts text to vector representations

2. **Sets Up ChromaDB**
   - Connects to local ChromaDB server
   - Creates a collection for the knowledge base

3. **Loads Sample Documents**
   - 5 documents about AI, ML, and RAG
   - Chunks large documents into smaller pieces

4. **Generates Embeddings**
   - Automatically embeds all document chunks
   - Stores vectors in ChromaDB

5. **Tests Retrieval**
   - Performs semantic search
   - Shows top relevant documents

6. **Creates RAG Agent**
   - Agent with access to `search_knowledge` tool
   - Can query the vector database

7. **Interactive Q&A**
   - Demonstrates RAG in action
   - Shows how agent uses retrieved context

## Example Output

```
🚀 RAG (Retrieval-Augmented Generation) Demo

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
   1. [Score: 0.8523] Machine Learning (ML) is a subset of AI that focuses on creating...
   2. [Score: 0.7891] Artificial Intelligence (AI) is the simulation of human intelligence...

🤖 Step 6: Creating RAG-powered Agent...
   ✅ Agent created with RAG capabilities

💬 Step 7: Interactive Q&A (RAG in action)
================================================================

[Question 1] User: What is artificial intelligence?
Assistant: Artificial Intelligence (AI) is the simulation of human intelligence by
machines. AI systems can perform tasks that typically require human intelligence,
such as visual perception, speech recognition, decision-making, and language
translation. Modern AI is based on machine learning algorithms.

[Question 2] User: Explain the difference between AI and machine learning
Assistant: AI is the broad field of creating intelligent machines, while Machine
Learning is a specific subset of AI. ML focuses on creating systems that learn from
data rather than being explicitly programmed. ML is one approach to achieving AI.

✅ RAG Demo completed successfully!
```

## Code Highlights

### Creating RAG Toolkit

```go
type RAGToolkit struct {
    *toolkit.BaseToolkit
    vectorDB vectordb.VectorDB
}

func NewRAGToolkit(db vectordb.VectorDB) *RAGToolkit {
    t := &RAGToolkit{
        BaseToolkit: toolkit.NewBaseToolkit("knowledge_retrieval"),
        vectorDB:    db,
    }

    t.RegisterFunction(&toolkit.Function{
        Name:        "search_knowledge",
        Description: "Search the knowledge base for relevant information",
        Parameters: map[string]toolkit.Parameter{
            "query": {
                Type:        "string",
                Description: "The search query",
                Required:    true,
            },
        },
        Handler: t.searchKnowledge,
    })

    return t
}
```

### Using with Agent

```go
// Create agent with RAG capabilities
ag, err := agent.New(agent.Config{
    Name:     "RAG Assistant",
    Model:    model,
    Toolkits: []toolkit.Toolkit{ragToolkit},
    Instructions: `You are a helpful AI assistant with access to a knowledge base.
Use the search_knowledge tool to find relevant information before answering.`,
})

// Agent automatically uses RAG for answers
output, err := ag.Run(ctx, "What is machine learning?")
```

## Customization

### Adding Your Own Documents

Replace the sample documents with your own:

```go
// Load from files
loader := knowledge.NewTextLoader("path/to/document.txt")
docs, err := loader.Load()

// Or load from directory
dirLoader := knowledge.NewDirectoryLoader("./docs", "*.md", true)
docs, err := dirLoader.Load()
```

### Changing Chunk Size

```go
// Larger chunks for more context
chunker := knowledge.NewCharacterChunker(1000, 100)

// Sentence-based chunking
sentenceChunker := knowledge.NewSentenceChunker(1000, 200)
```

### Using Different Models

```go
// Use GPT-4 for better quality
model, err := openaimodel.New("gpt-4", openaimodel.Config{
    APIKey: openaiKey,
})

// Use larger embeddings for better search
embedFunc, err := openaiembed.New(openaiembed.Config{
    APIKey: openaiKey,
    Model:  "text-embedding-3-large", // 3072 dimensions
})
```

### Adding Metadata Filtering

```go
// Search with metadata filter
filter := map[string]interface{}{
    "topic": "Machine Learning",
    "date": map[string]interface{}{
        "$gte": "2025-01-01",
    },
}

results, err := db.Query(ctx, query, 5, filter)
```

## Production Considerations

### 1. Persistent Storage

ChromaDB supports persistent storage:

```bash
docker run -p 8000:8000 -v /path/to/data:/chroma/data chromadb/chroma
```

### 2. Embedding Caching

Cache embeddings to reduce API costs:

```go
type CachedEmbedding struct {
    embedFunc *openai.OpenAIEmbedding
    cache     map[string][]float32
    mu        sync.RWMutex
}

func (c *CachedEmbedding) EmbedSingle(ctx context.Context, text string) ([]float32, error) {
    c.mu.RLock()
    if emb, ok := c.cache[text]; ok {
        c.mu.RUnlock()
        return emb, nil
    }
    c.mu.RUnlock()

    emb, err := c.embedFunc.EmbedSingle(ctx, text)
    if err != nil {
        return nil, err
    }

    c.mu.Lock()
    c.cache[text] = emb
    c.mu.Unlock()

    return emb, nil
}
```

### 3. Error Handling

```go
output, err := ag.Run(ctx, question)
if err != nil {
    if strings.Contains(err.Error(), "rate_limit") {
        // Retry with backoff
        time.Sleep(5 * time.Second)
        output, err = ag.Run(ctx, question)
    }
}
```

### 4. Monitoring

Track RAG performance:

```go
type RAGMetrics struct {
    QueriesProcessed int
    AverageLatency   time.Duration
    RetrievalAccuracy float64
}
```

## Troubleshooting

### ChromaDB Connection Error

```
Error: failed to create ChromaDB client: connection refused
```

**Solution**: Ensure ChromaDB is running:
```bash
docker ps | grep chroma
```

### No Results Found

```
Warning: No relevant documents found
```

**Solutions**:
- Check if documents were added successfully
- Verify embedding function is working
- Try lowering similarity threshold
- Increase search limit

### High Latency

**Optimizations**:
- Use `text-embedding-3-small` instead of `3-large`
- Reduce chunk size
- Implement embedding caching
- Use batch operations

## Resources

- [RAG Overview (Anthropic)](https://www.anthropic.com/index/retrieval-augmented-generation)
- [ChromaDB Documentation](https://docs.trychroma.com/)
- [OpenAI Embeddings Guide](https://platform.openai.com/docs/guides/embeddings)

## License

Same as Agno-Go (MIT License)
