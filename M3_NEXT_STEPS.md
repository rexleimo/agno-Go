# M3 后续步骤 - ChromaDB API 适配

**日期**: 2025-10-01
**状态**: ChromaDB 实现需要 API 适配
**优先级**: P1 (高)

---

## 🎯 当前状态

### ✅ 已完成
1. **VectorDB 接口设计** - 完整且合理
2. **OpenAI Embeddings** - 完全实现并测试通过
3. **Knowledge 包** - 文档加载器和分块器完成
4. **RAG 示例框架** - 代码结构完整
5. **完整文档** - README 和使用指南已准备

### ⏰ 需要修复
**ChromaDB 实现** - API 适配问题

---

## 🔧 需要修复的问题

### 问题清单

1. **Auth Provider 参数**
   ```go
   // 错误
   types.NewTokenAuthCredentialsProvider(apiKey)

   // 正确 (需要查看实际签名)
   types.NewTokenAuthCredentialsProvider(apiKey, headerType)
   ```

2. **Collection 创建方法**
   ```go
   // 实际 API
   client.CreateCollection(ctx, name, metadata, createOrGet, embeddingFunc, distanceFunc)
   // 不存在 GetOrCreateCollection
   ```

3. **Embedding 类型转换**
   ```go
   //我们的类型: [][]float32
   // ChromaDB 需要: []*types.Embedding

   // 使用转换函数
   embeddings := types.NewEmbeddingsFromFloat32(ourEmbeddings)
   ```

4. **Modify vs Update**
   ```go
   // Collection.Update() 是修改 collection 元数据
   // 修改文档内容应使用 Collection.Modify()
   ```

5. **Query 方法签名**
   ```go
   // 实际签名
   Query(ctx, queryTexts []string, nResults int32, where, whereDocuments, include)

   // 我们需要先生成 queryTexts embeddings,然后调用
   ```

---

## 📝 修复计划

### Step 1: 修复 Embedding 转换 (30分钟)

在 chromadb.go 中添加转换函数:

```go
// convertEmbeddings converts [][]float32 to []*types.Embedding
func convertEmbeddings(embeddings [][]float32) []*types.Embedding {
    return types.NewEmbeddingsFromFloat32(embeddings)
}

// extractEmbeddings converts []*types.Embedding to [][]float32
func extractEmbeddings(embeddings []*types.Embedding) [][]float32 {
    result := make([][]float32, len(embeddings))
    for i, emb := range embeddings {
        if emb != nil && emb.ArrayOfFloat32 != nil {
            result[i] = *emb.ArrayOfFloat32
        }
    }
    return result
}
```

### Step 2: 修复 Collection 操作 (45分钟)

```go
// CreateCollection - 使用正确的 API
func (c *ChromaDB) CreateCollection(ctx context.Context, name string, metadata map[string]interface{}) error {
    if name != "" {
        c.collectionName = name
    }

    distanceFunc := types.L2
    // 转换 distance function...

    collection, err := c.client.CreateCollection(
        ctx,
        c.collectionName,
        metadata,
        true, // createOrGet = true
        nil,  // embeddingFunction (我们自己处理)
        distanceFunc,
    )
    if err != nil {
        return fmt.Errorf("failed to create collection: %w", err)
    }

    c.collection = collection
    return nil
}
```

### Step 3: 修复 Add/Modify 方法 (30分钟)

```go
// Add documents
func (c *ChromaDB) Add(ctx context.Context, documents []vectordb.Document) error {
    // ... 准备数据 ...

    // 转换 embeddings
    chromaEmbeddings := types.NewEmbeddingsFromFloat32(embeddings)

    // 调用 ChromaDB API
    _, err := c.collection.Add(ctx, chromaEmbeddings, metadatas, contents, ids)
    return err
}

// Update - 使用 Modify 而不是 Update
func (c *ChromaDB) Update(ctx context.Context, documents []vectordb.Document) error {
    // ...
    chromaEmbeddings := types.NewEmbeddingsFromFloat32(embeddings)
    _, err := c.collection.Modify(ctx, chromaEmbeddings, metadatas, contents, ids)
    return err
}
```

### Step 4: 修复 Query 方法 (45分钟)

```go
// Query - 先生成 embedding,再查询
func (c *ChromaDB) Query(ctx context.Context, query string, limit int, filter map[string]interface{}) ([]vectordb.SearchResult, error) {
    if c.embeddingFunc == nil {
        return nil, fmt.Errorf("embedding function required")
    }

    // 生成 query embedding
    embedding, err := c.embeddingFunc.EmbedSingle(ctx, query)
    if err != nil {
        return nil, err
    }

    // 调用 QueryWithEmbedding
    return c.QueryWithEmbedding(ctx, embedding, limit, filter)
}

// QueryWithEmbedding - 使用 embedding 查询
func (c *ChromaDB) QueryWithEmbedding(ctx context.Context, embedding []float32, limit int, filter map[string]interface{}) ([]vectordb.SearchResult, error) {
    // ChromaDB Query API 使用 queryEmbeddings而不是 queryTexts
    chromaEmb := types.NewEmbeddingFromFloat32(embedding)

    // 注意: 需要查看实际 API 如何使用 query embeddings
    // 可能需要使用 QueryWithOptions
    queryOpts := []types.CollectionQueryOption{
        // 配置查询选项
    }

    result, err := c.collection.QueryWithOptions(ctx, queryOpts...)
    // ... 处理结果 ...
}
```

### Step 5: 修复 Get 方法 (20分钟)

```go
// Get documents by IDs
func (c *ChromaDB) Get(ctx context.Context, ids []string) ([]vectordb.Document, error) {
    // 使用正确的参数
    result, err := c.collection.Get(
        ctx,
        nil,  // where
        nil,  // whereDocuments
        ids,  // ids
        []types.QueryEnum{types.Documents, types.Metadatas, types.Embeddings}, // include
    )
    if err != nil {
        return nil, err
    }

    // 转换结果
    documents := make([]vectordb.Document, len(result.Ids))
    for i, id := range result.Ids {
        documents[i] = vectordb.Document{
            ID:       id,
            Content:  result.Documents[i],
            Metadata: result.Metadatas[i],
        }
        if i < len(result.Embeddings) && result.Embeddings[i] != nil {
            documents[i].Embedding = *result.Embeddings[i].GetFloat32()
        }
    }

    return documents, nil
}
```

---

## ⏱️ 时间估算

| 任务 | 预计时间 | 优先级 |
|------|---------|-------|
| Embedding 转换函数 | 30分钟 | P0 |
| Collection 创建修复 | 45分钟 | P0 |
| Add/Modify 方法 | 30分钟 | P0 |
| Query 方法完整实现 | 45分钟 | P1 |
| Get 方法修复 | 20分钟 | P1 |
| Delete 方法验证 | 15分钟 | P2 |
| 单元测试更新 | 60分钟 | P1 |
| 集成测试验证 | 30分钟 | P1 |
| **总计** | **~4.5小时** | |

---

## 🧪 验证步骤

完成修复后,按以下顺序验证:

### 1. 编译验证
```bash
go build ./pkg/agno/vectordb/chromadb/...
```

### 2. 单元测试
```bash
go test ./pkg/agno/vectordb/chromadb/ -v
```

### 3. 集成测试
```bash
# 启动 ChromaDB
docker run -p 8000:8000 chromadb/chroma

# 运行集成测试
go test ./pkg/agno/vectordb/chromadb/ -v -run TestCreateCollection
go test ./pkg/agno/vectordb/chromadb/ -v -run TestAddAndQuery
```

### 4. RAG Demo 验证
```bash
export OPENAI_API_KEY=your-key
go run cmd/examples/rag_demo/main.go
```

---

## 📚 参考资料

### ChromaDB Go 客户端文档
- GitHub: https://github.com/amikos-tech/chroma-go
- Go Pkg Doc: https://pkg.go.dev/github.com/amikos-tech/chroma-go
- 官方文档: https://go-client.chromadb.dev/

### 关键类型和方法
```go
// Client
func NewClient(options ...ClientOption) (*Client, error)
func (c *Client) CreateCollection(...) (*Collection, error)
func (c *Client) GetCollection(...) (*Collection, error)

// Collection
func (c *Collection) Add(ctx, embeddings, metadatas, documents, ids) error
func (c *Collection) Modify(ctx, embeddings, metadatas, documents, ids) error
func (c *Collection) Query(ctx, queryTexts, nResults, where, whereDocuments, include) (*QueryResults, error)
func (c *Collection) Get(ctx, where, whereDocuments, ids, include) (*GetResults, error)
func (c *Collection) Delete(ctx, ids, where, whereDocuments) error

// Types
types.NewEmbeddingsFromFloat32([][]float32) []*types.Embedding
types.NewEmbeddingFromFloat32([]float32) *types.Embedding
embedding.GetFloat32() *[]float32
```

---

## ✅ 完成标准

ChromaDB 实现被认为完成的标准:

1. ✅ 所有方法编译通过
2. ✅ 单元测试通过 (至少 8/11 个测试)
3. ✅ 集成测试通过 (需要运行 ChromaDB 服务器)
4. ✅ RAG demo 可以运行并返回正确结果
5. ✅ 文档反映实际 API 使用方式

---

## 🔄 备选方案

如果 ChromaDB Go 客户端问题过多,考虑:

### 方案 B: 使用 chromem-go
- 纯 Go 实现,无需外部服务
- API 类似 ChromaDB
- 零依赖,可嵌入
- GitHub: https://github.com/philippgille/chromem-go

### 方案 C: 直接 HTTP API
- 使用 ChromaDB 的 REST API
- 更灵活的控制
- 需要更多代码

---

## 📞 下一步行动

**建议**: 先完成 API 适配 (预计 4.5 小时),这是解锁 RAG 功能的关键。

完成后,M3 将真正达到 100% 完成状态,可以进入 M4 阶段。

---

*生成日期: 2025-10-01*
*预计完成时间: 2025-10-02*
