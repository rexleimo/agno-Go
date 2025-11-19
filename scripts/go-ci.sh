#!/usr/bin/env bash

set -euo pipefail

# Go CI helper for agno-Go
# - Runs unit tests across all Go packages with coverage profile
# - Ensures coverage ≥85%
# - Runs basic static analysis using go vet
# - Executes the cross-language parity script and fails on diff
#
# Usage:
#   ./scripts/go-ci.sh

ROOT_DIR="$(CDPATH="" cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

COVER_PROFILE="/tmp/coverage.out"
FIXTURE_PATH="$ROOT_DIR/specs/001-migrate-agno-core/fixtures/us1_basic_coordination.yaml"
PARITY_RESULTS="$ROOT_DIR/scripts/ci/.cache/parity_results.json"

echo "[go-ci] Running go test ./... with coverage profile ${COVER_PROFILE}"
go test ./... -coverprofile="$COVER_PROFILE"

TOTAL_COVERAGE=$(go tool cover -func "$COVER_PROFILE" | awk '/total:/ {print $3}')
TOTAL_COVERAGE=${TOTAL_COVERAGE%%%}
python3 - <<'PY' "$TOTAL_COVERAGE"
import sys
coverage = float(sys.argv[1]) if sys.argv[1] else 0.0
if coverage < 85.0:
    raise SystemExit(f"go-ci: coverage {coverage:.2f}% is below required 85%")
print(f"[go-ci] Coverage OK: {coverage:.2f}%")
PY

echo "[go-ci] Running go vet ./..."
go vet ./...

echo "[go-ci] Running cross-language parity script"
./scripts/ci/cross-language-parity.sh \
  --fixture "$FIXTURE_PATH"

python3 - <<'PY' "$PARITY_RESULTS"
import json, sys, os
result_path = sys.argv[1]
if not os.path.exists(result_path):
    raise SystemExit(f"go-ci: parity result not found at {result_path}")
with open(result_path, 'r', encoding='utf-8') as fh:
    data = json.load(fh)
status = data.get('status', '').lower()
fixture = data.get('fixture')
if status != 'pass':
    raise SystemExit(f"go-ci: parity failed for fixture {fixture}: status={status}")
print(f"[go-ci] Parity OK for fixture {fixture}")
PY

echo "[go-ci] Done."
