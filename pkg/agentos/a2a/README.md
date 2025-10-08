# A2A (Agent-to-Agent) Interface

A2A 接口为 Agno Go 提供标准化的 Agent 间通信协议，基于 JSON-RPC 2.0。

The A2A interface provides standardized agent-to-agent communication for Agno Go, based on JSON-RPC 2.0.

## 特性 / Features

- ✅ **JSON-RPC 2.0** - 标准化协议 / Standardized protocol
- ✅ **REST API** - HTTP 端点 / HTTP endpoints  
- ✅ **流式支持** - Server-Sent Events / Server-Sent Events
- ✅ **多媒体支持** - 文本、图片、文件 / Text, images, files
- ✅ **简单集成** - 几行代码即可暴露 Agent / Expose agents in few lines

## 快速开始 / Quick Start

### 1. 创建 Agent / Create an Agent

```go
type MyAgent struct {
    ID   string
    Name string
}

func (a *MyAgent) Run(ctx context.Context, input string) (interface{}, error) {
    return &a2a.RunOutput{
        Content: "Hello from agent!",
    }, nil
}

func (a *MyAgent) GetID() string { return a.ID }
func (a *MyAgent) GetName() string { return a.Name }
```

### 2. 创建 A2A 接口 / Create A2A Interface

```go
a2aInterface, err := a2a.New(a2a.Config{
    Agents: []a2a.Entity{myAgent},
    Prefix: "/a2a",
})
```

### 3. 注册路由 / Register Routes

```go
router := gin.Default()
a2aInterface.RegisterRoutes(router)
router.Run(":7777")
```

### 4. 调用 Agent / Call the Agent

```bash
curl -X POST http://localhost:7777/a2a/message/send \
  -H 'Content-Type: application/json' \
  -d '{
    "jsonrpc": "2.0",
    "method": "message/send",
    "id": "req-1",
    "params": {
      "message": {
        "messageId": "msg-1",
        "role": "user",
        "agentId": "my-agent",
        "contextId": "session-123",
        "parts": [{"kind": "text", "text": "Hello!"}]
      }
    }
  }'
```

## API 端点 / API Endpoints

### POST /a2a/message/send

发送消息到 Agent（非流式）/ Send message to agent (non-streaming)

**请求 / Request**:
```json
{
  "jsonrpc": "2.0",
  "method": "message/send",
  "id": "request-id",
  "params": {
    "message": {
      "messageId": "msg-id",
      "role": "user",
      "agentId": "agent-id",
      "contextId": "context-id",
      "parts": [
        {"kind": "text", "text": "Hello"}
      ]
    }
  }
}
```

**响应 / Response**:
```json
{
  "jsonrpc": "2.0",
  "id": "request-id",
  "result": {
    "task": {
      "id": "task-123",
      "context_id": "context-id",
      "status": "completed",
      "history": [...]
    }
  }
}
```

### POST /a2a/message/stream

发送消息到 Agent（流式）/ Send message to agent (streaming)

使用 Server-Sent Events 返回实时响应。

Uses Server-Sent Events for real-time responses.

## 消息部分类型 / Message Part Types

### Text / 文本

```json
{
  "kind": "text",
  "text": "消息内容"
}
```

### File (URI) / 文件（URI）

```json
{
  "kind": "file",
  "file": {
    "uri": "https://example.com/image.png",
    "mimeType": "image/png",
    "name": "image.png"
  }
}
```

### File (Bytes) / 文件（字节）

```json
{
  "kind": "file",
  "file": {
    "bytes": "base64-encoded-content",
    "mimeType": "image/png",
    "name": "image.png"
  }
}
```

### Data / 数据

```json
{
  "kind": "data",
  "data": {
    "content": "{\"key\": \"value\"}",
    "mimeType": "application/json"
  }
}
```

## 错误处理 / Error Handling

A2A 使用标准 JSON-RPC 2.0 错误码：

A2A uses standard JSON-RPC 2.0 error codes:

| 错误码 / Code | 含义 / Meaning |
|--------------|----------------|
| -32700 | 解析错误 / Parse error |
| -32600 | 无效请求 / Invalid request |
| -32601 | 方法未找到 / Method not found |
| -32602 | 无效参数 / Invalid params |
| -32603 | 内部错误 / Internal error |
| -32000 | 服务器错误 / Server error |

**错误响应示例 / Error Response Example**:
```json
{
  "jsonrpc": "2.0",
  "id": "req-1",
  "error": {
    "code": -32600,
    "message": "Invalid request: agentId is required"
  }
}
```

## 完整示例 / Complete Example

查看 `cmd/examples/a2a_server/main.go` 获取完整的工作示例。

See `cmd/examples/a2a_server/main.go` for a complete working example.

## 架构 / Architecture

```
┌─────────────┐
│HTTP Client  │
└──────┬──────┘
       │ JSON-RPC 2.0 Request
       ▼
┌──────────────────────┐
│  A2A Interface       │
│  ┌────────────────┐  │
│  │ Validator      │  │ Validate request
│  └────────┬───────┘  │
│           ▼          │
│  ┌────────────────┐  │
│  │ Mapper         │  │ A2A → RunInput
│  └────────┬───────┘  │
│           ▼          │
│  ┌────────────────┐  │
│  │ Entity         │  │ Agent/Team/Workflow
│  │ (Run)          │  │
│  └────────┬───────┘  │
│           ▼          │
│  ┌────────────────┐  │
│  │ Mapper         │  │ RunOutput → A2A
│  └────────┬───────┘  │
└───────────┼──────────┘
            ▼
     JSON-RPC 2.0 Response
```

## 兼容性 / Compatibility

兼容 Python Agno 的 A2A 实现，可以与 Python Agent 互操作。

Compatible with Python Agno's A2A implementation, can interoperate with Python agents.

## 许可证 / License

Apache License 2.0
