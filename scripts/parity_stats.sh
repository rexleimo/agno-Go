#!/usr/bin/env bash

set -u
set -o pipefail

ROOT_DIR="$(CDPATH="" cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RESULTS_FILE="$ROOT_DIR/scripts/ci/.cache/parity_results.json"

usage() {
  cat <<'EOF'
Usage: scripts/parity_stats.sh [--results <file>]

Parses the JSON emitted by scripts/ci/cross-language-parity.sh and prints
per-fixture pass rates. If the results file is missing, the legacy scenario
commands will be executed as a fallback.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --results)
      RESULTS_FILE="$2"
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

if [[ -f "$RESULTS_FILE" ]]; then
  python3 - <<'PY' "$RESULTS_FILE"
import json, sys
from collections import OrderedDict

path = sys.argv[1]
with open(path, 'r', encoding='utf-8') as fh:
    data = json.load(fh)

if isinstance(data, dict):
    data = [data]

summary = OrderedDict()
for entry in data:
    fixture = entry.get('fixture', 'unknown')
    status = (entry.get('status') or '').lower()
    stats = summary.setdefault(fixture, {'runs': 0, 'pass': 0, 'fail': 0, 'last': status})
    stats['runs'] += 1
    if status == 'pass':
        stats['pass'] += 1
    else:
        stats['fail'] += 1
    stats['last'] = status or 'n/a'

if not summary:
    print(f"No parity entries found in {path}")
    raise SystemExit(0)

print(f"Parity results from {path}:\n")
print("| Fixture | Runs | Pass | Fail | Pass Rate | Last Status |")
print("|---------|------|------|------|-----------|-------------|")
for fixture, stats in summary.items():
    rate = 0.0
    if stats['runs']:
        rate = stats['pass'] * 100.0 / stats['runs']
    print(f"| {fixture} | {stats['runs']} | {stats['pass']} | {stats['fail']} | {rate:.1f}% | {stats['last']} |")
PY
  exit 0
fi

echo "[parity-stats] No parity results JSON found at $RESULTS_FILE"
echo "[parity-stats] Falling back to legacy scenario commands..."

cd "$ROOT_DIR"

SCENARIOS=(
  "teams-basic-coordination-us1 must_match go test ./go/agent -run TestUS1ParityConfigScript"
  "custom-internal-search-us3 must_match go test ./go/providers -run TestUS3CustomProviderParity"
)

total=0
must_match_total=0
implemented=0
passed=0

for entry in "${SCENARIOS[@]}"; do
  total=$((total + 1))
  IFS=' ' read -r scenario_id severity cmd_head cmd_rest <<<"$entry"
  cmd="$cmd_head $cmd_rest"
  echo "Scenario: $scenario_id (severity=$severity)"
  echo "  Command: $cmd"
  if [ "$severity" = "must_match" ]; then
    must_match_total=$((must_match_total + 1))
  fi
  if eval "$cmd" >/dev/null 2>&1; then
    echo "  Result: PASS"
    implemented=$((implemented + 1))
    passed=$((passed + 1))
  else
    echo "  Result: FAIL"
    implemented=$((implemented + 1))
  fi
  echo
done

echo "Summary:"
echo "  Total scenarios:          $total"
echo "  Implemented scenarios:    $implemented"
echo "  Must-match scenarios:     $must_match_total"
echo "  Passed scenarios:         $passed"
if [ "$total" -gt 0 ]; then
  coverage_pct=$((implemented * 100 / total))
else
  coverage_pct=0
fi
if [ "$must_match_total" -gt 0 ]; then
  passrate_pct=$((passed * 100 / must_match_total))
else
  passrate_pct=0
fi
echo "  Coverage (implemented/total): ${coverage_pct}%"
echo "  Pass rate (passed/must_match): ${passrate_pct}%"
