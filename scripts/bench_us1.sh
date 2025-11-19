#!/usr/bin/env bash

set -euo pipefail

# Simple benchmark script for the US1 scenario.
# It runs:
#   - A Go benchmark for RunUS1Example
#   - A single Python invocation of the US1 parity script

ROOT_DIR="$(CDPATH="" cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

echo "== Go benchmark: RunUS1Example =="
go test ./go/agent -run ^$ -bench BenchmarkRunUS1Example -benchtime=10x

if command -v python >/dev/null 2>&1 || command -v python3 >/dev/null 2>&1; then
  echo
  echo "== Python timing: us1_basic_coordination_parity =="
  cd "$ROOT_DIR/agno"
  PYTHON_BIN="$(command -v python || command -v python3)"
  "$PYTHON_BIN" - <<'PY'
import time
from cookbook.scripts.us1_basic_coordination_parity import run_parity

payload = {"query": "Write an article about the top 2 stories on hackernews"}

start = time.perf_counter()
result = run_parity(payload)
elapsed = time.perf_counter() - start
print(f"Python us1_basic_coordination_parity: {elapsed:.4f}s for single run")
PY
  # Note: above inline script assumes that the Python import path is configured
  # so that the 'cookbook' package is importable. If this is not the case in
  # the current environment, the Python timing will fail; such failures should
  # be treated as environment/setup issues rather than benchmark regressions.
else
  echo
  echo "Python executable not found; skipping Python timing."
fi
