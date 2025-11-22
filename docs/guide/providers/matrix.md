
# Provider Capability Matrix

This page summarizes the high-level capabilities and configuration knobs for the providers supported by Agno-Go. It is not an exhaustive specification, but it should help you quickly see:

- Which providers can be used for chat.
- Which providers can be used for embeddings (where supported).
- Which providers support streaming responses.
- Which environment variables you need to configure.

> The exact models and capabilities available to you depend on your account, region, and quota with each provider. Always refer to the provider’s official documentation for the latest details.

## Summary Table

| Provider    | Chat Support          | Embedding Support             | Streaming Support              | Key Env Vars                                                                 |
|------------|-----------------------|-------------------------------|--------------------------------|------------------------------------------------------------------------------|
| OpenAI     | Yes (chat)           | Yes (embeddings)             | Yes (chat streaming)          | `OPENAI_API_KEY`, `OPENAI_ENDPOINT`, `OPENAI_ORG`, `OPENAI_API_VERSION`      |
| Gemini     | Yes (chat)           | Yes (embeddings)             | Yes (chat streaming)          | `GEMINI_API_KEY`, `GEMINI_ENDPOINT`, `VERTEX_PROJECT`, `VERTEX_LOCATION`     |
| GLM4       | Yes (chat)           | Limited / planned*           | Provider-dependent             | `GLM4_API_KEY`, `GLM4_ENDPOINT`                                              |
| OpenRouter | Yes (chat, router)   | Where underlying model supports it | Yes (where model supports it) | `OPENROUTER_API_KEY`, `OPENROUTER_ENDPOINT`                                  |
| SiliconFlow| Yes (chat)           | Yes (embeddings)             | Yes (chat streaming)          | `SILICONFLOW_API_KEY`, `SILICONFLOW_ENDPOINT`                                |
| Cerebras   | Yes (chat)           | Where supported              | Provider-dependent             | `CEREBRAS_API_KEY`, `CEREBRAS_ENDPOINT`                                      |
| ModelScope | Yes (chat)           | Where supported              | Provider-dependent             | `MODELSCOPE_API_KEY`, `MODELSCOPE_ENDPOINT`                                  |
| Groq       | Yes (chat)           | Limited / planned*           | Yes (chat streaming)          | `GROQ_API_KEY`, `GROQ_ENDPOINT`                                              |
| Ollama     | Yes (local chat)     | Where model supports it      | Yes (chat streaming, local)   | `OLLAMA_ENDPOINT`                                                            |

`*` Embedding support for some providers is still evolving. In cases where embeddings are not yet fully supported or are only available for certain models, the Go adapter will either skip unsupported calls in tests or clearly document deviations in the contracts.

## Configuration Notes

- All provider-related environment variables are defined in `.env.example`. Copy it to `.env` and fill only the providers you plan to use.  
- Leaving required keys empty will cause health checks and provider tests to skip that provider with a clear reason; the runtime will not attempt to call providers that are not configured.  
- `OPENAI_ENDPOINT`, `GEMINI_ENDPOINT`, and similar values default to the hosted APIs, but can be overridden to point to private gateways or proxies.  
- `OLLAMA_ENDPOINT` typically points to a locally running Ollama/vLLM instance (for example `http://localhost:11434/v1`), and is only used when you explicitly enable local models.  

For more detail on how providers are routed and how errors are normalized, see the
**Core Features & API** page and the contracts in the specs directory.

## Next steps

- Review [Configuration & Security](../config-and-security) for a deeper explanation of
  the environment variables listed here and recommended practices for managing keys.  
- Return to [Quickstart](../quickstart) if you want to extend the basic example with a
  different provider configuration.  
