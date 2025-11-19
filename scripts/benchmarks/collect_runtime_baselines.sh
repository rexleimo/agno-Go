#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(CDPATH='' cd -- "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DATA_DIR="$SCRIPT_DIR/data"

WORKFLOW=""
PYTHON_CMD="python agno/tests/benchmarks/run.py"
GO_CMD="go test ./go/agent -run TestUS1Bench -bench . -benchmem"

usage() {
    cat <<'EOF'
Usage: collect_runtime_baselines.sh --workflow <name> [--python "<cmd>"] [--go "<cmd>"]

Options:
  --workflow <name>   Workflow/scenario identifier (required)
  --python "<cmd>"    Command that emits Python metrics as JSON (default: python agno/tests/benchmarks/run.py)
  --go "<cmd>"        Command that runs Go benchmarks (default: go test ./go/agent -run TestUS1Bench -bench . -benchmem)
  -h, --help          Show this message

The Python command must emit a JSON object with keys latency_ms_p95, rss_mb, cpu_percent, tokens_per_second.
The Go benchmark command writes JSON to the file pointed by US1_BENCHMARK_EXPORT.
EOF
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --workflow)
            WORKFLOW="$2"
            shift 2
            ;;
        --python)
            PYTHON_CMD="$2"
            shift 2
            ;;
        --go)
            GO_CMD="$2"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "ERROR: unknown argument $1" >&2
            usage
            exit 1
            ;;
    esac
done

if [[ -z "$WORKFLOW" ]]; then
    echo "ERROR: --workflow is required" >&2
    usage
    exit 1
fi

mkdir -p "$DATA_DIR"
OUTPUT_FILE="$DATA_DIR/${WORKFLOW}.json"

TMP_DIR="$(mktemp -d)"
cleanup() {
    rm -rf "$TMP_DIR"
}
trap cleanup EXIT

PYTHON_OUTPUT="$TMP_DIR/python.json"
GO_OUTPUT="$TMP_DIR/go.json"

echo "[benchmarks] Running Python command: $PYTHON_CMD"
if ! eval "$PYTHON_CMD" >"$PYTHON_OUTPUT"; then
    echo "ERROR: Python benchmark command failed" >&2
    exit 1
fi

echo "[benchmarks] Running Go command: $GO_CMD"
if ! US1_BENCHMARK_EXPORT="$GO_OUTPUT" eval "$GO_CMD" >/dev/null; then
    echo "ERROR: Go benchmark command failed" >&2
    exit 1
fi

python3 - <<'PY' "$PYTHON_OUTPUT" "$GO_OUTPUT" "$OUTPUT_FILE" "$WORKFLOW"
import json
import sys
from datetime import datetime, timezone

python_path, go_path, output_path, workflow = sys.argv[1:5]

def load_json(path, label):
    with open(path, 'r', encoding='utf-8') as fh:
        try:
            return json.load(fh)
        except json.JSONDecodeError as exc:
            raise SystemExit(f"{label} JSON invalid: {exc}")

python_metrics = load_json(python_path, "Python")
go_payload = load_json(go_path, "Go")
go_metrics = go_payload.get('go', {})

summary = go_metrics.get('summary', {})
concurrency = go_metrics.get('concurrency', {})

result = {
    "scenario_id": workflow,
    "collected_at": datetime.now(tz=timezone.utc).isoformat(),
    "python": python_metrics,
    "go": {
        "summary": summary,
        "concurrency": concurrency,
    }
}

with open(output_path, 'w', encoding='utf-8') as fh:
    json.dump(result, fh, ensure_ascii=False, indent=2)

print(f"Baseline written to {output_path}")
PY
