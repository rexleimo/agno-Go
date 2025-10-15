# AgentOS Server API 레퍼런스

## NewServer

HTTP 서버를 생성합니다.

**함수 시그니처:**
```go
func NewServer(config *Config) (*Server, error)

type Config struct {
    Address        string           // 서버 주소 (기본값: :8080)
    SessionStorage session.Storage  // 세션 저장소 (기본값: memory)
    Logger         *slog.Logger     // 로거 (기본값: slog.Default())
    Debug          bool             // 디버그 모드 (기본값: false)
    AllowOrigins   []string         // CORS origins
    AllowMethods   []string         // CORS methods
    AllowHeaders   []string         // CORS headers
    RequestTimeout time.Duration    // 요청 타임아웃 (기본값: 30s)
    MaxRequestSize int64            // 최대 요청 크기 (기본값: 10MB)

    // 지식 API (선택) / Knowledge API (optional)
    VectorDBConfig  *VectorDBConfig  // 벡터 DB 구성 (예: chromadb)
    EmbeddingConfig *EmbeddingConfig // 임베딩 모델 구성 (예: OpenAI)
}

type VectorDBConfig struct {
    Type           string // 예: "chromadb"
    BaseURL        string // 벡터 DB 엔드포인트
    CollectionName string // 기본 컬렉션
    Database       string // 선택 데이터베이스
    Tenant         string // 선택 테넌트
}

type EmbeddingConfig struct {
    Provider string // 예: "openai"
    APIKey   string
    Model    string // 예: "text-embedding-3-small"
    BaseURL  string // 예: "https://api.openai.com/v1"
}
```

**예제:**
```go
server, err := agentos.NewServer(&agentos.Config{
    Address: ":8080",
    Debug:   true,
    RequestTimeout: 60 * time.Second,
})
```

## Server.RegisterAgent

에이전트를 등록합니다.

**함수 시그니처:**
```go
func (s *Server) RegisterAgent(agentID string, ag *agent.Agent) error
```

**예제:**
```go
err := server.RegisterAgent("assistant", myAgent)
```

## Server.Start / Shutdown

서버를 시작하고 중지합니다.

**함수 시그니처:**
```go
func (s *Server) Start() error
func (s *Server) Shutdown(ctx context.Context) error
```

**예제:**
```go
go func() {
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}()

// 우아한 종료
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
server.Shutdown(ctx)
```

## API 엔드포인트

완전한 API 문서는 [OpenAPI 명세](../../pkg/agentos/openapi.yaml)를 참고하세요.

**핵심 엔드포인트:**
- `GET /health` - 헬스 체크
- `POST /api/v1/sessions` - 세션 생성
- `GET /api/v1/sessions/{id}` - 세션 조회
- `PUT /api/v1/sessions/{id}` - 세션 업데이트
- `DELETE /api/v1/sessions/{id}` - 세션 삭제
- `GET /api/v1/sessions` - 세션 목록
- `GET /api/v1/agents` - 에이전트 목록
- `POST /api/v1/agents/{id}/run` - 에이전트 실행

**지식 엔드포인트 (선택) / Knowledge Endpoints (optional):**
- `POST /api/v1/knowledge/search` — 지식 베이스에서 벡터 유사도 검색 / Vector similarity search
- `GET  /api/v1/knowledge/config` — 사용 가능한 청커, VectorDB, 임베딩 모델 정보 / Available chunkers, VectorDBs, embedding model

요청 예시 / Example:
```bash
curl -X POST http://localhost:8080/api/v1/knowledge/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "에이전트 생성 방법?",
    "limit": 5,
    "filters": {"source": "documentation"}
  }'
```

최소 서버 구성 (지식 API 활성화) / Minimal server config (enable Knowledge API):
```go
server, err := agentos.NewServer(&agentos.Config{
  Address: ":8080",
  VectorDBConfig: &agentos.VectorDBConfig{
    Type:           "chromadb",
    BaseURL:        os.Getenv("CHROMADB_URL"),
    CollectionName: "agno_knowledge",
  },
  EmbeddingConfig: &agentos.EmbeddingConfig{
    Provider: "openai",
    APIKey:   os.Getenv("OPENAI_API_KEY"),
    Model:    "text-embedding-3-small",
  },
})
```

실행 가능한 예시 / Runnable example: `cmd/examples/knowledge_api/`

## 모범 사례

### 1. 항상 Context 사용하기

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

output, err := ag.Run(ctx, input)
```

### 2. 적절한 에러 처리

```go
output, err := ag.Run(ctx, input)
if err != nil {
    switch {
    case types.IsInvalidInputError(err):
        // 잘못된 입력 처리
    case types.IsRateLimitError(err):
        // 백오프와 함께 재시도
    default:
        // 기타 에러 처리
    }
}
```

### 3. 메모리 관리

```go
// 새 주제를 시작할 때 초기화
ag.ClearMemory()

// 또는 제한된 메모리 사용
mem := memory.NewInMemory(50)
```

### 4. 적절한 타임아웃 설정

```go
server, _ := agentos.NewServer(&agentos.Config{
    RequestTimeout: 60 * time.Second, // 복잡한 에이전트용
})
```
