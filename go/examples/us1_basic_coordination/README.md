# US1 Basic Coordination (Go Runtime)

This example demonstrates the US1 cookbook scenario running entirely on the Go runtime. It wires together:

- `go/agent` – AgentRuntime configuration + runtime service
- `go/workflow` – sequential workflow runner for the hackernews → article-reader collaboration
- `go/session` – in-memory session store + parity-compatible session record mapper

## Prerequisites

1. Run `.specify/scripts/bash/check-prerequisites.sh --json --require-tasks --include-tasks` to verify Go 1.25.1 + Python ≥3.11 are installed (Python is only needed for parity checks).
2. Export any provider/tool credentials you need (e.g. `export OPENAI_API_KEY=...`).
3. Optional: set `GO_TELEMETRY_EXPORTER=stdout` if you want to stream telemetry while running the example.

## Run the example

```bash
go run ./go/examples/us1_basic_coordination \
  --providers go/providers/providers.go \
  --workflow go/workflow/us1_basic_coordination_workflow.go
```

The binary loads the agents/workflow from source, executes the two-step workflow, and prints the final output plus a serialized `SessionRecord` that matches the Python schema.

## Cross-language parity

Use the shared fixture at `specs/001-migrate-agno-core/fixtures/us1_basic_coordination.yaml` to compare Python vs Go outputs:

```bash
./scripts/ci/cross-language-parity.sh \
  --fixture specs/001-migrate-agno-core/fixtures/us1_basic_coordination.yaml \
  --python "python -m agno.tests.contracts.run" \
  --go "go test ./go/agent -run TestUS1ParityFixtureIntegration"
```

Parity results (including diffs) are written to `scripts/ci/.cache/parity_results.json` for CI review.

## Telemetry & observability

1. Register the handlers defined in `go/internal/telemetry/http_handlers.go` to expose `/v1/runtime/go/telemetry/events`.
2. Run the example with `GO_TELEMETRY_EXPORTER=stdout` or your recorder implementation.
3. Query the events endpoint with filters, e.g. `curl 'http://localhost:8080/v1/runtime/go/telemetry/events?sessionId=us1-demo&runtime=go'`.

Telemetry events now emit RunStarted → RunCompleted, ReasoningStep, ToolCall, and SessionSummary with a forced `runtime=go` tag and fall back to `unknown_event` when needed.

## Performance baselines

The Go benchmark `go/agent/us1_basic_coordination_bench_test.go` enforces latency/RSS/CPU/tokens/sec targets at 1 and 100 concurrent runs. Refresh the baseline JSON before pushing changes:

```bash
./scripts/benchmarks/collect_runtime_baselines.sh \
  --workflow us1_basic_coordination \
  --python "python agno/tests/benchmarks/run.py --workflow us1_basic_coordination" \
  --go "go test ./go/agent -run TestUS1Bench -bench . -benchmem"
```

The script writes `scripts/benchmarks/data/us1_basic_coordination.json`, which the benchmark automatically loads and uses for gating.

Refer to `agno/docs/go-runtime-support-matrix.md` for the latest list of Go-supported cookbook scenarios and any remaining Python-only modules.
