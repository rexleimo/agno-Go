
# 核心功能与 API 概览

本页从概念和接口两个角度介绍 Agno-Go 的核心能力，帮助你在不阅读 Go 源码的前提下理解“有哪些东西可以用”和“应该调用哪些 HTTP 接口”。

## 核心概念

### Agent（代理）

**Agent** 表示针对某类任务或产品场景的“智能体配置”，通常包括：

- 名称与描述（用于区分用途）  
- 默认使用的模型（例如 OpenAI、Gemini、Groq 等供应商的具体模型）  
- 可以调用的工具列表  
- 可选的行为配置（如温度、路由策略等）  

Agent 通过 `/agents` 系列接口创建与读取。

### Session（会话）

**会话** 表示用户（或系统）与某个 Agent 之间的一段持续交互，它：

- 始终隶属某一个 Agent；  
- 携带 `userId` 与可选的 `metadata`（例如渠道、实验分组）；  
- 为一系列消息提供稳定的上下文。  

会话通过 `/agents/{agentId}/sessions` 创建。

### Message（消息）

**消息** 是会话中的单轮对话，包含：

- 角色（例如 `user`、`assistant`）；  
- 文本内容，以及在需要时包含的工具调用信息；  
- 可根据 `stream` 查询参数选择一次性 JSON 响应或事件流。  

消息通过 `/agents/{agentId}/sessions/{sessionId}/messages` 发送。

### Tool（工具）

工具让 Agent 能够调用外部系统（例如 HTTP API、数据库或搜索）。运行时提供：

- 为 Agent 注册工具的方法；  
- 启用/禁用某个工具的能力。  

工具状态可通过 `/agents/{agentId}/tools/{toolName}` 切换。

### Memory（记忆）与状态

记忆描述了跨会话持久化状态的方式。在 Agno-Go 中：

- 短期对话状态保存在 Session 与 Message 中；  
- 长期状态（如用户画像或知识库）通过配置的存储后端持久化。  

具体采用的存储技术（内存、Bolt、Badger 等）由环境变量与 `config/default.yaml` 决定，并在“配置与安全实践”页面中统一说明。

### Provider（模型供应商）

**供应商** 是提供模型能力的后端（例如 OpenAI、Gemini、GLM4、OpenRouter、SiliconFlow、Cerebras、ModelScope、Groq、Ollama）。每个供应商：

- 在 Go 中实现统一的 chat/embedding 接口；  
- 需要特定的环境变量完成鉴权与 endpoint 配置；  
- 可能支持非流式与流式输出。  

具体能力与配置项可在“模型供应商矩阵”页面中查看。

## HTTP API 概览（运行时）

运行时暴露的 HTTP 接口与上述概念一一对应。`contracts` 目录中的 OpenAPI 文档给出了完整细节，这里仅列出最常用的几个。

### Health Check（健康检查）

- **接口**：`GET /health`  
- **用途**：确认运行时是否就绪，并查看版本与供应商状态等元信息。  
- **典型场景**：活性/就绪探针、监控、手动验证部署是否成功。  

### Agents（代理）

- **创建 Agent**  
  - **接口**：`POST /agents`  
  - **请求体**：Agent 定义（名称、描述、模型、工具、配置）。  
  - **响应**：包含 `agentId` 的 JSON 对象。  
- **获取 Agent**  
  - **接口**：`GET /agents/{agentId}`  
  - **用途**：查看已创建 Agent 的配置，便于调试与审计。  

### Sessions（会话）

- **创建会话**  
  - **接口**：`POST /agents/{agentId}/sessions`  
  - **请求体**：可选 `userId` 与 `metadata` 字段。  
  - **响应**：包含会话 ID 的会话对象。  
- **关系**：一个会话只属于一个 Agent；一个 Agent 可以拥有多个会话。  

### Messages（消息）

- **发送非流式消息**  
  - **接口**：`POST /agents/{agentId}/sessions/{sessionId}/messages`  
  - **查询参数**：无 `stream` 或 `stream=false`。  
  - **请求体**：包含 `role` 与 `content` 的消息对象。  
  - **响应**：包含 `messageId`、`content`、`toolCalls`、`usage`、`state` 等字段的 JSON 对象。  
- **发送流式消息**  
  - **接口**：`POST /agents/{agentId}/sessions/{sessionId}/messages?stream=true`  
  - **响应**：以 Server-Sent Events（SSE）形式返回逐步输出。  

Quickstart 页面展示了使用这些接口完成最小闭环调用的示例。

### Tools（工具）

- **切换工具状态**  
  - **接口**：`PATCH /agents/{agentId}/tools/{toolName}`  
  - **请求体**：`{ "enabled": true | false }`  
  - **响应**：返回工具名称、当前状态以及完整工具列表。  

在需要演示工具工作流或动态开关工具的高级案例中，这个接口尤为重要。

## 配置概览

本页不限定具体的部署拓扑，而是假定：

- 运行时配置保存在 `config/default.yaml` 等文件中；  
- 模型供应商的凭据与 endpoint 通过 `.env` 中的环境变量注入；  
- `.env.example` 列出了所有支持的供应商变量，并以注释区分必填与可选。  

如果你想集中了解配置、环境变量与安全实践，请参考 **Configuration & Security Practices** 页面；若希望了解各供应商的能力差异与配置字段，请参考 **模型供应商矩阵** 页面。
