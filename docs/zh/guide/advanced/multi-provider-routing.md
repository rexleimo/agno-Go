
# 高级案例：多模型回退与路由

本指南展示如何在保持统一 HTTP 接口的前提下，在多个模型供应商之间进行路由与回退。

目标是：

- 只暴露一套 AgentOS 运行时与 HTTP 接口给调用方；  
- 在服务端根据简单规则（如任务类型、模型名称）在不同供应商之间路由；  
- 当主供应商不可用时自动回退到备用模型，而无需修改客户端代码。  

## 1. 适用场景

常见需求包括：

- 通用对话使用一个供应商，成本敏感或低延迟场景使用另一个供应商；  
- 主供应商不可用时希望自动切换到备用供应商；  
- 希望在稳定的客户端集成之上做模型实验或 A/B 测试。  

## 2. 高层设计

路由逻辑应尽量放在 Agent 配置与服务端运行时，而不是客户端：

1. 定义一个或多个 Agent，通过 `model` 字段表达“首选模型/供应商”；  
2. 运行时根据 `model` 字段与配置，将请求路由到具体供应商；  
3. 客户端始终调用相同的 HTTP 接口（`/agents`、`/sessions`、`/messages`）。  

示例模型命名约定：

- `openai:gpt-4o-mini`  
- `gemini:flash-1.5`  
- `groq:llama3-70b`  

具体映射由服务端配置负责。

## 3. 示例流程

1. **创建支持路由的 Agent**

   ```bash
   curl -X POST http://localhost:8080/agents \
     -H "Content-Type: application/json" \
     -d '{
       "name": "routing-agent",
       "description": "An agent that routes across providers based on task type.",
       "model": "openai:gpt-4o-mini",
       "tools": [],
       "config": {
         "fallbackModel": "gemini:flash-1.5"
       }
     }'
   ```

2. **创建会话**

   ```bash
   curl -X POST http://localhost:8080/agents/<agent-id>/sessions \
     -H "Content-Type: application/json" \
     -d '{
       "userId": "routing-user",
       "metadata": {
         "source": "advanced-multi-provider-routing"
       }
     }'
   ```

3. **发送消息**

   ```bash
   curl -X POST "http://localhost:8080/agents/<agent-id>/sessions/<session-id>/messages" \
     -H "Content-Type: application/json" \
     -d '{
       "role": "user",
       "content": "对于一个内部小工具，你会推荐使用哪个供应商/模型？请说明原因。"
     }'
   ```

如果主供应商出现不可用，运行时可以根据配置回退到 `fallbackModel`，而客户端调用方式保持不变。

## 4. 配置要点

- 在 `.env` 与 `config/default.yaml` 中统一管理各供应商的 key、endpoint 与超时等配置；  
- 结合“模型供应商矩阵”页面选择合适的供应商与能力组合；  
- 避免在客户端硬编码供应商特有逻辑，将 Agno-Go 运行时视为唯一集成面。  

## 5. 测试与验证

上线前建议：

- 按 Quickstart 流程对路由 Agent 做一次完整调用；  
- 临时移除某个供应商的 key，观察是否按预期回退到备用模型；  
- 记录与不同供应商相关的已知限制（如 token 计数、延迟差异），在内部 runbook 中说明。  
