
# 高级案例：带持久记忆的对话代理

本指南展示如何构建一个能够利用记忆的对话体验。目标不是规定特定存储实现，而是说明如何结合现有 HTTP 接口与元数据字段，让记忆真正发挥作用。

## 1. 记忆类型

可以粗略将记忆分为三层：

- **会话历史**：当前 Session 内最近的多轮对话；  
- **用户画像**：关于用户的长期信息（偏好、重要字段）；  
- **知识记录**：与业务领域相关的事实（例如历史交互、关键事件）。  

Agno-Go 通过 Session + Message 原生支持第一层，并允许你通过配置与自有服务连接后两层。

## 2. 创建支持记忆的 Agent

在 Go 中，你通常通过 HTTP 与 AgentOS 运行时交互。下面是一个与 Quickstart
一致的最小示例，展示如何在 Go 代码里创建一个支持记忆的 Agent：

```go
package main

import (
  "bytes"
  "encoding/json"
  "log"
  "net/http"
  "time"
)

type Agent struct {
  Name        string                 `json:"name"`
  Description string                 `json:"description"`
  Model       map[string]any         `json:"model"`
  Tools       []map[string]any       `json:"tools"`
  Config      map[string]any         `json:"config"`
}

func main() {
  client := &http.Client{Timeout: 10 * time.Second}

  agent := Agent{
    Name:        "memory-chat-agent",
    Description: "A chat agent that uses session history and external memory.",
    Model: map[string]any{
      "provider": "openai",
      "modelId":  "gpt-4o-mini",
      "stream":   true,
    },
    Tools:  nil,
    Config: map[string]any{},
  }

  body, err := json.Marshal(agent)
  if err != nil {
    log.Fatalf("marshal agent: %v", err)
  }

  resp, err := client.Post("http://localhost:8080/agents", "application/json", bytes.NewReader(body))
  if err != nil {
    log.Fatalf("create agent: %v", err)
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusCreated {
    log.Fatalf("unexpected status: %s", resp.Status)
  }

  // 在真实应用中，你会在这里解析响应拿到 agentId，
  // 然后按后续章节创建 Session 并发送消息。
}
```

如果你只是想在终端或 API 客户端里直接试验 HTTP 接口，可以使用等价的 `curl`
命令：

```bash
curl -X POST http://localhost:8080/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "memory-chat-agent",
    "description": "A chat agent that uses session history and external memory.",
    "model": {
      "provider": "openai",
      "modelId": "gpt-4o-mini",
      "stream": true
    },
    "tools": [],
    "config": {}
  }'
```

真正决定“是否善用记忆”的关键，在于后续你如何组织 Session，以及如何在
Session 中传入与用户相关的上下文和元数据。

## 3. 利用 Session 与元数据

在创建 Session 时，可以附带用户标识与元数据：

```bash
curl -X POST http://localhost:8080/agents/<agent-id>/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-1234",
    "metadata": {
      "source": "advanced-memory-chat",
      "segment": "beta-testers"
    }
  }'
```

你的应用可以使用 `userId` 与 `metadata` 在自有存储中查找或更新用户画像，并在后续消息中加入相关信息。

## 4. 在提示中融合记忆

发送消息时，可以将已知事实和上下文合并到 `content` 中：

```bash
curl -X POST "http://localhost:8080/agents/<agent-id>/sessions/<session-id>/messages" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "user",
    "content": "你之前帮我制定过一个阅读计划。基于那次的建议以及我偏好短时间阅读，请为本周推荐一个计划。"
  }'
```

后端可以：

- 从记忆存储中加载历史交互或备注；  
- 将简要总结或关键事实附加到提示中；  
- 仍通过标准消息接口完成调用。  

## 5. 配置与存储

对于高度依赖记忆的场景：

- 使用“配置与安全实践”文档来决定启用哪种记忆后端（如内存 vs 本地持久化）；  
- 将额外基础设施（数据库、缓存、队列等）记录在内部运维文档中，AgentOS 运行时依旧专注于 HTTP 行为与契约；  
- 确保 `.env` 与 `config/default.yaml` 中的设置与官方文档中的建议保持一致，尤其是数据保留与数据所在区域相关配置。  

## 6. 测试与迭代

验证记忆增强型对话时：

- 设计一套同时覆盖“短期”与“长期”记忆行为的测试用例；  
- 使用 `/health` 与 Quickstart 流程确认在记忆使用量增加时运行时仍保持健康；  
- 监控延迟与资源占用，并基于实测结果调整记忆策略（例如摘要频率、回放长度）。  
