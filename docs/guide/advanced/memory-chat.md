
# Advanced Guide: Memory-Augmented Chat

This guide illustrates how to build a chat experience that can take advantage of memory across turns and sessions. The goal is not to prescribe a specific storage backend, but to show how to use the existing HTTP API and metadata fields to make memory useful.

## 1. Types of memory

At a high level, you can think of memory in three layers:

- **Conversation history**: recent messages in the current session.  
- **User profile**: long-lived information about a user (preferences, profile fields).  
- **Knowledge records**: domain-specific facts (for example past interactions or important events).  

Agno-Go provides primitives for the first layer out of the box（Session + Message），and allows you to connect the other two via configuration and your own services.

## 2. Creating a memory-capable agent

In Go, you will typically talk to the AgentOS runtime over HTTP. A minimal example
that mirrors the Quickstart flow looks like this:

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

  // In a real app you would decode the response to get agentId
  // and then create sessions / send messages as shown later.
}
```

If you prefer to experiment with the raw HTTP surface (for example in a terminal
or API client), the equivalent `curl` command is:

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

The key differences for a memory-capable agent will be how you structure sessions
and what metadata you pass along, which we cover next.

## 3. Using sessions and metadata

When creating a session, you can attach user-specific identifiers and metadata:

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

Your application can use `userId` and `metadata` to look up or update user profile records in your own storage, then include relevant information in future messages.

## 4. Incorporating memory into prompts

When sending messages, you can pass prior facts and context into the `content` field:

```bash
curl -X POST "http://localhost:8080/agents/<agent-id>/sessions/<session-id>/messages" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "user",
    "content": "You previously suggested a reading plan for me. Based on that and my preference for short sessions, suggest a plan for this week."
  }'
```

Your backend can:

- Load past interactions or notes from your memory store.  
- Add a brief summary or key facts to the prompt content.  
- Let the runtime handle the rest via the standard message endpoint.  

## 5. Configuration and storage

For memory-heavy scenarios:

- Use Configuration & Security docs to decide which memory backend to enable（for example an on-disk store vs purely in-memory）.  
- Document any additional infrastructure（databases, caches, queues）in your own operational docs; the runtime itself stays focused on HTTP behavior and contracts.  
- Make sure your `.env` and `config/default.yaml` settings are consistent with the guidance in the official docs, especially around retention and data locality.  

## 6. Testing and evolution

To validate memory-augmented chat:

- Design a small test plan that exercises both “short-term” and “long-term” memory behaviors.  
- Use the same `/health` and Quickstart flows to ensure the runtime remains healthy while memory usage grows.  
- Monitor latency and resource usage; adjust your memory strategy（for example summarization frequency）based on empirical results.  
