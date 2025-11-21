# Phase 0 Research - Go 版 Agno 重构

## Transport + Streaming 框架（AgentOS API）
Decision: 使用标准库 `net/http` 搭配 `github.com/go-chi/chi/v5` 路由，流式输出采用 `http.Flusher` 的 chunked/SSE 响应（与 Python 版流式语义对齐），客户端/CLI 复用相同 HTTP 接口。  
Rationale: 轻量无外部依赖，易于将 Python 版的 REST/事件语义映射到 Go；chi 支持中间件链（限流、鉴权、日志）、易于在基准场景下保持低开销；chunked/SSE 兼容多数代理/负载均衡器，便于 p95 延迟压测和背压提示。  
Alternatives considered: Gin/Fiber（性能相近但引入额外 DSL 和依赖，增加迁移成本）；gRPC（需要定义 proto 且与现有 HTTP 接口不一致，契约覆盖需重建）；纯 `net/http` 手写路由（最小依赖但中间件与分组管理成本高）。

## 模型供应商客户端策略（9 家的 Chat/Embedding）
Decision: 统一接口在 `/Users/rex/cool.cnb/agno-Go/go/internal/model/` 定义，适配器放在 `/Users/rex/cool.cnb/agno-Go/go/pkg/providers/<provider>/`；优先使用官方/稳定的 Go SDK（OpenAI、Groq），其余走自研轻量 REST 客户端（签名、重试、限流、流式解析均在本仓封装），所有路径支持流式与非流式两个 code path。  
Rationale: 官方 SDK 覆盖面有限且行为差异大，混用会导致依赖矩阵膨胀；自研 REST 客户端可统一错误规约与重试/限流策略，并确保无 cgo/子进程；对需要 SSE 的供应商（OpenAI/OpenRouter/Groq/SiliconFlow）可复用统一流式解析器，对不支持流式的供应商走非流式分支。  
Alternatives considered: 全量采用各供应商 SDK（依赖碎片化、版本漂移风险高，难以共用重试/流式解析）；通过 OpenRouter 统一入口（无法覆盖嵌入、且与直连差异会影响契约匹配）；引入多语言 sidecar（违反纯 Go 宪章）。

## 会话与记忆存储
Decision: 定义 `MemoryStore` 接口抽象（会话历史、工具调用结果、向量存储），默认提供线程安全的内存实现 + 可选嵌入式 Bolt/Badger 持久化实现；基准与契约测试默认用内存实现，长跑压测可切换到持久化以观察 GC/IO 行为。  
Rationale: Python 版的记忆可配置（内存/持久化）；内存实现满足快速启动与契约测试，嵌入式 KV（纯 Go，无外部服务）可覆盖持久化与崩溃恢复场景；接口分层便于后续扩展（如向量存储、分布式缓存）且不引入非 Go 依赖。  
Alternatives considered: 直接依赖外部 Redis/PostgreSQL（提高部署复杂度且不符合“纯 Go / 无额外依赖”的最小可行目标）；仅内存实现（无法验证持久化与恢复边界，压测结果不完整）；基于文件 append log（实现简单但并发/查询能力弱）。*** End Patch​
