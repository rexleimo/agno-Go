
# Core Features & API Overview

This page gives a high-level overview of the main concepts in Agno-Go and how they map to the HTTP API that the runtime exposes. It is intended for developers who want to understand *what* the system does and *which endpoints* they should call, without needing to read the Go implementation.

## Core Concepts

### Agent

An **agent** defines how the system should behave for a given task or product surface. It encapsulates:

- A name and description.
- The default model to use (for example an OpenAI, Gemini, Groq, or other provider model).
- Which tools the agent is allowed to call.
- Optional configuration such as temperature or routing policies.

Agents are created and retrieved via the `/agents` endpoints.

### Session

A **session** represents an ongoing interaction between a user (or system) and an agent. It:

- Is always tied to a single agent.
- Carries a `userId` and optional metadata (for example channel, experiment group).
- Provides a stable context for a sequence of messages.

Sessions are created via `/agents/{agentId}/sessions`.

### Message

A **message** is a single turn in a conversation within a session. It:

- Has a role (for example `user`, `assistant`).
- Contains textual content and, when applicable, tool call information.
- Can be returned as a single JSON payload or as a stream of events, depending on the `stream` query parameter.

Messages are sent via `/agents/{agentId}/sessions/{sessionId}/messages`.

### Tool

Tools allow agents to call out to external systems (for example HTTP APIs, databases or search). The runtime exposes:

- A way to register tools against an agent.
- A way to enable or disable tools for a given agent.

Tools can be toggled via `/agents/{agentId}/tools/{toolName}`.

### Memory and State

Memory describes how state is persisted across sessions and interactions. In Agno-Go:

- Short-term conversation state is held in sessions and messages.
- Long-term state (for example user profiles or knowledge bases) can be stored via the configured memory backends.

The exact storage technology (in-memory, Bolt, Badger, etc.) is configured via environment variables and `config/default.yaml` and is documented on the Configuration & Security page.

### Providers

Providers are model backends (for example OpenAI, Gemini, GLM4, OpenRouter, SiliconFlow, Cerebras, ModelScope, Groq, Ollama). Each provider:

- Implements a common chat/embedding interface in Go.
- Requires specific environment variables for authentication and endpoints.
- May support non-streaming and streaming responses.

The Provider Matrix page lists the supported capabilities and required environment variables for each provider.

## HTTP API Surface (Runtime)

The runtime exposes a small set of HTTP endpoints that correspond to the concepts above. The OpenAPI document in the `contracts` directory describes them in detail; this section only covers the most common ones.

### Health Check

- **Endpoint**: `GET /health`  
- **Purpose**: Confirm that the runtime is up and to inspect basic metadata such as version and provider status.  
- **Typical use**: Liveness/readiness checks, monitoring, and quick manual verification after deployment.

### Agents

- **Create agent**  
  - **Endpoint**: `POST /agents`  
  - **Payload**: Agent definition (name, description, model, tools, config).  
  - **Response**: A JSON object containing `agentId`.  
- **Get agent**  
  - **Endpoint**: `GET /agents/{agentId}`  
  - **Purpose**: Retrieve the stored agent definition, for inspection or debugging.

### Sessions

- **Create session**  
  - **Endpoint**: `POST /agents/{agentId}/sessions`  
  - **Payload**: Optional `userId` and `metadata` object.  
  - **Response**: Session object including its unique ID.  
- **Relationship**: Each session is attached to exactly one agent; a single agent can have many sessions.

### Messages

- **Send message (non-streaming)**  
  - **Endpoint**: `POST /agents/{agentId}/sessions/{sessionId}/messages`  
  - **Query**: no `stream` parameter, or `stream=false`.  
  - **Payload**: Message request with `role` and `content`.  
  - **Response**: A single JSON object containing `messageId`, `content`, `toolCalls`, `usage`, and `state`.  
- **Send message (streaming)**  
  - **Endpoint**: `POST /agents/{agentId}/sessions/{sessionId}/messages?stream=true`  
  - **Response**: Server-Sent Events (SSE) stream of partial outputs.  

The Quickstart page demonstrates a minimal call sequence using these endpoints.

### Tools

- **Toggle tool**  
  - **Endpoint**: `PATCH /agents/{agentId}/tools/{toolName}`  
  - **Payload**: `{ "enabled": true | false }`  
  - **Response**: Confirmation of the new tool state and the resulting tool list.  

This endpoint is especially relevant for advanced guides that demonstrate tool-based workflows or dynamic tool enabling/disabling.

## Configuration Overview

This page deliberately avoids prescribing a specific deployment topology. Instead, it assumes:

- The runtime configuration is stored in files such as `config/default.yaml`.  
- Provider credentials and endpoints are injected via environment variables defined in `.env`.  
- The `.env.example` file lists all supported provider variables, with comments explaining required vs optional values.

For a consolidated view of configuration, environment variables and security practices, refer to the **Configuration & Security Practices** page. For provider-specific details and capability differences, refer to the **Provider Matrix** page.
