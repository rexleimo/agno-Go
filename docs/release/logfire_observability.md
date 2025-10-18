# Logfire Observability (Agno-Go)

This guide shows how to instrument an Agno-Go agent with OpenTelemetry and forward traces to Logfire. It pairs with the sample program located at `cmd/examples/logfire_observability`.

## Prerequisites

1. **OpenAI Reasoning access** – the demo uses `o1-preview`. You can change the model ID if needed.
2. **Logfire account** – grab a write token from the Logfire dashboard.
3. **Go tooling** – install Go 1.23+.

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `OPENAI_API_KEY` | LLM provider key | `sk-...` |
| `LOGFIRE_WRITE_TOKEN` | Logfire write token (used as `Authorization` header) | `lf_...` |
| `LOGFIRE_OTLP_ENDPOINT` | *(Optional)* OTLP endpoint (defaults to `logfire-eu.pydantic.dev`) | `logfire-us.pydantic.dev` |
| `LOGFIRE_ENV` | *(Optional)* Resource attribute for trace environment | `staging` |

```bash
export OPENAI_API_KEY="sk-..."
export LOGFIRE_WRITE_TOKEN="lf_..."
# Optional overrides:
# export LOGFIRE_OTLP_ENDPOINT="logfire-us.pydantic.dev"
# export LOGFIRE_ENV="staging"
```

## Run the Example

```bash
cd cmd/examples/logfire_observability
# install the OpenTelemetry dependencies (only required once)
go get go.opentelemetry.io/otel/sdk go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp

# run the example with the `logfire` build tag
go run -tags logfire .
```

What happens:

1. An OTLP/HTTP exporter is configured to send spans to Logfire.
2. The agent executes a reasoning-capable prompt.
3. The span records metadata such as runtime, loop count, token usage, and reasoning snippets.
4. Events and attributes appear in Logfire dashboards within seconds.

If `LOGFIRE_WRITE_TOKEN` is not provided, the example still runs and prints the agent’s response, but telemetry is kept local.

## Streaming & Event Filters

For real-time monitoring the AgentOS server exposes an SSE endpoint:

```
POST /api/v1/agents/{id}/run/stream?types=run_start,reasoning,token,complete
```

The handler now sends dedicated `reasoning` events and enriches the final `complete` payload with usage + reasoning summaries. These events can be forwarded to Logfire or any observability backend.

See [`pkg/agentos/events.go`](../../pkg/agentos/events.go) and [`pkg/agentos/events_handlers.go`](../../pkg/agentos/events_handlers.go) for the canonical event schema.

## Troubleshooting

- **Empty traces in Logfire** – double-check the endpoint (EU/US) and that the write token includes the `Authorization` prefix exactly as issued.
- **TLS errors** – the example enables HTTPS by default. For self-hosted endpoints ensure the certificate chain is valid or customise the TLS config.
- **Model errors** – reasoning models may require special access; switch to a standard OpenAI model if necessary.

Happy tracing! 🚀
