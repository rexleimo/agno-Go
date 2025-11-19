#!/usr/bin/env bash

set -u
set -o pipefail

# Minimal parity statistics script for migration scenarios.
# It currently tracks:
#   - teams-basic-coordination-us1
#   - custom-internal-search-us3
#
# For each scenario it runs a corresponding Go test that encapsulates the
# Python+Go parity check or configuration check, then reports coverage and
# pass rate.

ROOT_DIR="$(CDPATH="" cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

SCENARIOS=(
  "teams-basic-coordination-us1 must_match go test ./go/agent -run TestUS1ParityConfigScript"
  "custom-internal-search-us3 must_match go test ./go/providers -run TestUS3CustomProviderParity"
)

total=0
must_match_total=0
implemented=0
passed=0

echo "Running parity checks for defined scenarios..."
echo

for entry in "${SCENARIOS[@]}"; do
  total=$((total + 1))
  IFS=' ' read -r scenario_id severity cmd_head cmd_rest <<<"$entry"

  # Rebuild command from the remaining fields
  cmd="$cmd_head $cmd_rest"

  echo "Scenario: $scenario_id (severity=$severity)"
  echo "  Command: $cmd"

  # Count must_match scenarios
  if [ "$severity" = "must_match" ]; then
    must_match_total=$((must_match_total + 1))
  fi

  # Run the command and capture exit code
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

coverage_pct=0
passrate_pct=0

if [ "$total" -gt 0 ]; then
  coverage_pct=$((implemented * 100 / total))
fi

if [ "$must_match_total" -gt 0 ]; then
  passrate_pct=$((passed * 100 / must_match_total))
fi

echo "  Coverage (implemented/total): ${coverage_pct}%"
echo "  Pass rate (passed/must_match): ${passrate_pct}%"

