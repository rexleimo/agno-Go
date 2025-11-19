#!/usr/bin/env bash

set -euo pipefail

# Minimal incident statistics script.
# It expects a telemetry log file where each line is a JSON object containing,
# optionally, an "error_code" field. Any event with an "error_code" is treated
# as a severe incident for the purposes of SC-003.
#
# Usage:
#   ./scripts/incidents_stats.sh path/to/telemetry.jsonl
# If no path is provided, it defaults to logs/telemetry.jsonl

ROOT_DIR="$(CDPATH="" cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

FILE="${1:-logs/telemetry.jsonl}"

if [ ! -f "$FILE" ]; then
  echo "Telemetry file not found: $FILE"
  echo "Total requests:        0"
  echo "Severe incidents:      0"
  echo "Severe incident rate:  0% (no data)"
  exit 0
fi

total=$(wc -l < "$FILE" | tr -d ' ')
if [ "$total" -eq 0 ]; then
  echo "Telemetry file is empty: $FILE"
  echo "Total requests:        0"
  echo "Severe incidents:      0"
  echo "Severe incident rate:  0% (no data)"
  exit 0
fi

# Count lines that contain an "error_code" field as severe incidents.
severe=$(grep -c '"error_code"' "$FILE" || true)

rate=$((severe * 100 / total))

echo "Telemetry file:        $FILE"
echo "Total requests:        $total"
echo "Severe incidents:      $severe"
echo "Severe incident rate:  ${rate}%"

