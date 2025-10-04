# AgentOS Multi-Instance Example

[English](#english) | [中文](#中文)

---

## English

This example demonstrates how to run multiple AgentOS instances with different route prefixes on the same or different ports.

### Use Cases

1. **Multi-tenant Systems**: Host different agents for different customers
2. **Service Organization**: Separate agents by functionality (chat, math, search, etc.)
3. **Version Management**: Run different API versions side-by-side
4. **Load Distribution**: Distribute different agent types across ports

### Architecture

```
Port 8080 (Math Service)
├── /health                          → Health check
└── /math/api/v1/                    → Math agent API
    ├── /agents                      → List agents
    ├── /agents/:id/run              → Run agent
    └── /sessions                    → Session management

Port 8081 (Chat Service)
├── /health                          → Health check
└── /chat/api/v1/                    → Chat agent API
    ├── /agents                      → List agents
    ├── /agents/:id/run              → Run agent
    └── /sessions                    → Session management
```

### Running the Example

```bash
# Set your OpenAI API key
export OPENAI_API_KEY=sk-...

# Run the example
go run cmd/examples/agentos_multi_instance/main.go
```

### Example API Calls

**Math Service (Port 8080)**

```bash
# Health check
curl http://localhost:8080/health

# List agents
curl http://localhost:8080/math/api/v1/agents

# Run math calculation
curl -X POST http://localhost:8080/math/api/v1/agents/default/run \
     -H 'Content-Type: application/json' \
     -d '{"input": "What is 25 * 4 + 10?"}'

# Create session
curl -X POST http://localhost:8080/math/api/v1/sessions \
     -H 'Content-Type: application/json' \
     -d '{"agent_id": "default"}'
```

**Chat Service (Port 8081)**

```bash
# Health check
curl http://localhost:8081/health

# List agents
curl http://localhost:8081/chat/api/v1/agents

# Chat with agent
curl -X POST http://localhost:8081/chat/api/v1/agents/default/run \
     -H 'Content-Type: application/json' \
     -d '{"input": "Hello, how are you?"}'
```

### Configuration Options

```go
server, err := agentos.NewServer(&agentos.Config{
    Address: ":8080",        // Port to listen on
    Prefix:  "/math",        // Route prefix (e.g., "/math", "/v1", "/api")
    Debug:   true,           // Enable debug mode
    RequestTimeout: 30 * time.Second,
})
```

### Advanced: Single Port, Multiple Prefixes

You can also run multiple agent services on a **single port** by using different prefixes:

```go
// All on port 8080, but different prefixes
mathServer := agentos.NewServer(&agentos.Config{
    Address: ":8080",
    Prefix:  "/math",
})

chatServer := agentos.NewServer(&agentos.Config{
    Address: ":8080",
    Prefix:  "/chat",
})

// Access:
// http://localhost:8080/math/api/v1/agents
// http://localhost:8080/chat/api/v1/agents
```

---

## 中文

此示例演示如何在相同或不同端口上使用不同的路由前缀运行多个 AgentOS 实例。

### 使用场景

1. **多租户系统**: 为不同客户托管不同的 agent
2. **服务组织**: 按功能分离 agent (聊天、数学、搜索等)
3. **版本管理**: 并行运行不同的 API 版本
4. **负载分配**: 将不同类型的 agent 分布到不同端口

### 架构

```
端口 8080 (数学服务)
├── /health                          → 健康检查
└── /math/api/v1/                    → 数学 agent API
    ├── /agents                      → 列出 agents
    ├── /agents/:id/run              → 运行 agent
    └── /sessions                    → Session 管理

端口 8081 (聊天服务)
├── /health                          → 健康检查
└── /chat/api/v1/                    → 聊天 agent API
    ├── /agents                      → 列出 agents
    ├── /agents/:id/run              → 运行 agent
    └── /sessions                    → Session 管理
```

### 运行示例

```bash
# 设置你的 OpenAI API 密钥
export OPENAI_API_KEY=sk-...

# 运行示例
go run cmd/examples/agentos_multi_instance/main.go
```

### API 调用示例

**数学服务 (端口 8080)**

```bash
# 健康检查
curl http://localhost:8080/health

# 列出 agents
curl http://localhost:8080/math/api/v1/agents

# 运行数学计算
curl -X POST http://localhost:8080/math/api/v1/agents/default/run \
     -H 'Content-Type: application/json' \
     -d '{"input": "What is 25 * 4 + 10?"}'

# 创建 session
curl -X POST http://localhost:8080/math/api/v1/sessions \
     -H 'Content-Type: application/json' \
     -d '{"agent_id": "default"}'
```

**聊天服务 (端口 8081)**

```bash
# 健康检查
curl http://localhost:8081/health

# 列出 agents
curl http://localhost:8081/chat/api/v1/agents

# 与 agent 聊天
curl -X POST http://localhost:8081/chat/api/v1/agents/default/run \
     -H 'Content-Type: application/json' \
     -d '{"input": "Hello, how are you?"}'
```

### 配置选项

```go
server, err := agentos.NewServer(&agentos.Config{
    Address: ":8080",        // 监听端口
    Prefix:  "/math",        // 路由前缀 (例如: "/math", "/v1", "/api")
    Debug:   true,           // 启用调试模式
    RequestTimeout: 30 * time.Second,
})
```

### 高级: 单端口,多前缀

你也可以通过使用不同的前缀在**单个端口**上运行多个 agent 服务:

```go
// 全部在端口 8080,但使用不同的前缀
mathServer := agentos.NewServer(&agentos.Config{
    Address: ":8080",
    Prefix:  "/math",
})

chatServer := agentos.NewServer(&agentos.Config{
    Address: ":8080",
    Prefix:  "/chat",
})

// 访问:
// http://localhost:8080/math/api/v1/agents
// http://localhost:8080/chat/api/v1/agents
```

### 优势

✅ **灵活性**: 在同一端口上托管多个服务
✅ **隔离性**: 不同的前缀提供清晰的 API 边界
✅ **可扩展性**: 轻松添加新的 agent 服务
✅ **多租户**: 支持为不同租户提供独立的 API 路径

### 注意事项

- 健康检查端点 `/health` 始终在根级别 (不受前缀影响)
- 确保不同服务使用不同的端口或不同的前缀
- 所有 API 路由都在前缀下 (例如: `{prefix}/api/v1/...`)
