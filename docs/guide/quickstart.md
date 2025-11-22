# Quickstart: Agno-Go in 10 Minutes

This guide walks you through a minimal end-to-end flow using Agno-Go:

1. Start the AgentOS runtime.
2. Create an agent.
3. Create a session.
4. Send a message and inspect the response.

> All paths are relative to your project root (for example `<your-project-root>/go/cmd/agno` and `<your-project-root>/config/default.yaml`). Replace placeholders with your own paths and configuration.

```bash
cd <your-project-root>
go run ./go/cmd/agno --config ./config/default.yaml
```

Check that the server is running:

```bash
curl http://localhost:8080/health
```

Create a minimal agent:

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

Create a session for the agent:

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

Send a message in the session:

```bash
curl -X POST "http://localhost:8080/agents/<agent-id>/sessions/<session-id>/messages" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "user",
    "content": "Introduce Agno-Go briefly."
  }'
```

You should see a JSON response with `messageId`, `content`, `usage`, and `state` fields.

## Next steps

- Review [Configuration & Security](./config-and-security) to understand how to set provider
  keys, endpoints and runtime options safely.
- Explore [Core Features & API](./core-features-and-api) and
  [Provider Matrix](./providers/matrix) for a broader view of capabilities.
- Try one of the [Advanced Guides](./advanced/multi-provider-routing) once you are
  comfortable with the basic flow.
