
# Advanced Guide: Multi-provider Routing

This guide shows how to design a higher-level workflow that can route requests across multiple model providers, while keeping the HTTP contract simple and aligned with the core runtime API.

The goals of this example are:

- Use a single AgentOS runtime and HTTP surface.  
- Route requests to different providers based on simple rules (for example model name or task type).  
- Keep configuration in `config/default.yaml` and `.env`, without coupling the client to any single provider.  

## 1. When to use multi-provider routing

Typical scenarios:

- You want to use one provider for general-purpose chat, and another for cost-sensitive or latency-sensitive workloads.  
- You want a “fallback” model when a primary provider is unavailable.  
- You want to experiment with new models while keeping the client integration stable.  

## 2. High-level design

At a high level, the routing logic lives in the Agent configuration and server-side runtime, not in the client:

1. You define one or more agents, each with a model field that indicates which provider/model to use.  
2. The runtime resolves the `model` into a concrete provider client based on configuration.  
3. The client always calls the same HTTP endpoints (`/agents`, `/sessions`, `/messages`).  

Example model naming convention:

- `openai:gpt-4o-mini`  
- `gemini:flash-1.5`  
- `groq:llama3-70b`  

The exact mapping is handled on the server side.

## 3. Example flow

1. **Create a routing-aware agent**

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

2. **Create a session**

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

3. **Send a message**

   ```bash
   curl -X POST "http://localhost:8080/agents/<agent-id>/sessions/<session-id>/messages" \
     -H "Content-Type: application/json" \
     -d '{
       "role": "user",
       "content": "For a small internal tool, which provider/model would you recommend and why?"
     }'
   ```

If the primary provider is unavailable, the runtime can fall back to the `fallbackModel` defined in `config` or in server-side configuration, without changing the client code.

## 4. Configuration considerations

In addition to `OPENAI_API_KEY` and `GEMINI_API_KEY`, you should:

- Keep provider-specific details (endpoints, keys, timeouts) in `.env` and `config/default.yaml`.  
- Ensure that the Provider Matrix page is used as a reference when deciding which providers to enable.  
- Avoid hard-coding provider-specific logic in clients; treat the runtime as the single integration surface.  

## 5. Testing and validation

Before using multi-provider routing in production:

- Verify basic behavior with Quickstart-style calls against the routing agent.  
- Temporarily disable one provider (for example by removing its key) and confirm that fallback behavior works as intended.  
- Record any known limitations (for example differences in tokenization or latency) in your internal runbooks or team docs.
