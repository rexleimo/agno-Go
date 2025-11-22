
# Advanced Guide: Knowledge Base Assistant

This guide shows how to design an assistant that can answer questions based on your own knowledge sources, while keeping the HTTP interaction surface the same as in the Quickstart.

The goals of this example are:

- Keep the client integration to a small set of HTTP endpoints.  
- Introduce a retrieval step (for example a vector store lookup) before or alongside the model call.  
- Make it clear which parts belong to “knowledge configuration” and which parts belong to the runtime API.  

## 1. Scenario

Imagine you want an assistant that can answer questions about your product documentation or internal guidelines. High-level flow:

1. Offline, you embed your documents into a vector store (not covered in detail here).  
2. At query time, you retrieve the most relevant passages for the user’s question.  
3. You pass the retrieved context into the agent as part of the message content or metadata.  

The runtime remains responsible for managing agents, sessions and messages.

## 2. Agent and session

You can reuse the same Quickstart pattern to create an agent and session:

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

## 3. Passing retrieved context

Once your application has retrieved relevant passages from a knowledge store, you can include them directly in the message content:

```bash
curl -X POST "http://localhost:8080/agents/<agent-id>/sessions/<session-id>/messages" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "user",
    "content": "Using the following context, answer the question.\\n\\n[CONTEXT]\\n...retrieved passages here...\\n\\nQuestion: How does our refund policy work?"
  }'
```

Alternatively, you can pass retrieval metadata in the `metadata` field when creating the session or as part of your own application-level state. The runtime API itself does not enforce a specific retrieval pattern.

## 4. Configuration and providers

When building a knowledge base assistant:

- Choose a provider and model with good long-context support, as summarized in the Provider Matrix.  
- Configure the necessary env vars in `.env` (for example `OPENAI_API_KEY`, `GEMINI_API_KEY`), and ensure they are documented on the Configuration & Security page.  
- Keep knowledge indexing and retrieval infrastructure (vector store, database, storage) outside of the runtime; treat it as a separate concern that feeds into message content.  

## 5. Testing and validation

To validate this pattern:

- Start with a small set of carefully curated documents and test questions.  
- Verify that the assistant can answer questions accurately when given the retrieved context.  
- Log or record cases where the answer is incomplete or incorrect, and use them to refine your retrieval strategy and prompt templates.  
