
# 高级案例：结合知识库的助手

本指南展示如何在保持 Quickstart 风格 HTTP 接口的前提下，构建一个可以使用你自己的知识源回答问题的助手。

目标是：

- 对客户端而言，仍然只需使用少量 HTTP 接口；  
- 在模型调用之前或同时引入检索步骤（例如向量检索）；  
- 明确哪些属于“知识配置/检索”，哪些属于 AgentOS 运行时的职责。  

## 1. 场景描述

假设你希望构建一个能够回答产品文档或内部规范问题的助手。大致流程：

1. 离线阶段，将文档嵌入为向量并写入向量库（本指南不展开具体实现）；  
2. 查询阶段，根据用户问题从向量库检索最相关的片段；  
3. 将检索到的上下文作为消息内容的一部分传入 Agent。  

运行时仍然负责管理 Agent、Session 与 Message。

## 2. Agent 与会话

可以沿用 Quickstart 的模式创建 Agent 与会话：

```bash
curl -X POST http://localhost:8080/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "kb-assistant",
    "description": "Answers questions using knowledge base context.",
    "model": "openai:gpt-4o-mini",
    "tools": [],
    "config": {}
  }'
```

```bash
curl -X POST http://localhost:8080/agents/<agent-id>/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "kb-user",
    "metadata": {
      "source": "advanced-knowledge-base-assistant"
    }
  }'
```

## 3. 传入检索上下文

当应用从知识库中检索到相关片段后，可以直接加入消息内容：

```bash
curl -X POST "http://localhost:8080/agents/<agent-id>/sessions/<session-id>/messages" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "user",
    "content": "请使用以下上下文回答问题。\\n\\n[CONTEXT]\\n...检索到的片段...\\n\\n问题：我们的退款策略是怎样的？"
  }'
```

也可以在创建会话时或在你自己的应用层状态中通过 `metadata` 传入检索元信息。运行时本身不强制特定的检索方案。

## 4. 配置与供应商选择

在构建知识库助手时：

- 根据“模型供应商矩阵”选择具备较好长上下文能力的供应商与模型；  
- 在 `.env` 中配置必要的环境变量（如 `OPENAI_API_KEY`、`GEMINI_API_KEY`），并在“配置与安全实践”中说明；  
- 把知识索引与检索基础设施（向量库、数据库、存储等）视为运行时之外的独立组件，只需将检索结果注入消息内容即可。  

## 5. 测试与优化

验证这一模式时：

- 从一小批精心挑选的文档与测试问题开始；  
- 检查在给定检索上下文时，助手能否准确回答问题；  
- 对不完整或错误的回答进行记录，并用来迭代检索策略与提示设计。  
