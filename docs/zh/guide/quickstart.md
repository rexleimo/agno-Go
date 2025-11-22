# 快速开始：10 分钟体验 Agno-Go

本指南帮助你在 10 分钟内完成一次从“启动服务”到“收到模型响应”的完整体验。

1. 启动 AgentOS 运行时  
2. 创建一个最小 Agent  
3. 创建会话  
4. 发送一条消息并查看响应  

> 所有路径均相对于你的项目根目录（例如 `<your-project-root>/go/cmd/agno`、`<your-project-root>/config/default.yaml`）。请根据本地路径替换占位符，不要复制他人机器上的绝对路径。

在项目根目录下启动服务：

```bash
cd <your-project-root>
go run ./go/cmd/agno --config ./config/default.yaml
```

检查服务是否就绪：

```bash
curl http://localhost:8080/health
```

创建一个最小 Agent：

```bash
curl -X POST http://localhost:8080/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "quickstart-agent",
    "description": "A minimal agent created from the docs quickstart.",
    "model": "openai:gpt-4o-mini",
    "tools": [],
    "config": {}
  }'
```

为该 Agent 创建会话：

```bash
curl -X POST http://localhost:8080/agents/<agent-id>/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "quickstart-user",
    "metadata": {
      "source": "docs-quickstart"
    }
  }'
```

在会话中发送一条消息：

```bash
curl -X POST "http://localhost:8080/agents/<agent-id>/sessions/<session-id>/messages" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "user",
    "content": "你好，帮我简要介绍一下 Agno-Go。"
  }'
```

你应当看到一个包含 `messageId`、`content`、`usage` 与 `state` 字段的 JSON 响应。

## 下一步

- 查看[配置与安全实践](./config-and-security)，了解如何安全地配置各模型供应商的 Key、端点与运行参数。  
- 继续阅读[核心功能与 API](./core-features-and-api) 和[模型供应商矩阵](./providers/matrix)，系统性了解可用能力。  
- 在熟悉基础流程后，可以尝试[高级指南](./advanced/multi-provider-routing) 中的案例，构建更复杂的工作流。  
