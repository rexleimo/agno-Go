#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(CDPATH='' cd -- "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DATA_DIR="$SCRIPT_DIR/data"
TICKETS_FILE_DEFAULT="$DATA_DIR/support_tickets.json"
ALERT_FILE_DEFAULT="$(cd "$SCRIPT_DIR/../.." && pwd)/specs/001-migrate-agno-core/logs/support-alerts.md"

SINCE_DAYS=14
TICKETS_FILE="${SUPPORT_TICKETS_FILE:-$TICKETS_FILE_DEFAULT}"
ALERT_FILE="${SUPPORT_ALERTS_FILE:-$ALERT_FILE_DEFAULT}"

usage() {
    cat <<'EOF'
Usage: export_support_stats.sh [--since <days>] [--tickets <file>] [--alerts <file>]

Parses support ticket JSON (array of {"id", "created_at", ...}) and prints
counts for the selected window. When the matched tickets exceed 2, an entry is
appended to specs/001-migrate-agno-core/logs/support-alerts.md.

Environment overrides:
  SUPPORT_TICKETS_FILE  Path to JSON file (default scripts/support-metrics/data/support_tickets.json)
  SUPPORT_ALERTS_FILE   Alert markdown path (default specs/.../logs/support-alerts.md)
EOF
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --since)
            SINCE_DAYS="$2"
            shift 2
            ;;
        --tickets)
            TICKETS_FILE="$2"
            shift 2
            ;;
        --alerts)
            ALERT_FILE="$2"
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

mkdir -p "$DATA_DIR"

if [[ ! -f "$TICKETS_FILE" ]]; then
    echo "No ticket file found at $TICKETS_FILE (nothing to report)." >&2
    exit 0
fi

python3 - <<'PY' "$TICKETS_FILE" "$SINCE_DAYS" "$ALERT_FILE"
import json
import sys
import os
from datetime import datetime, timedelta, timezone

ticket_file, since_days, alert_file = sys.argv[1:4]
try:
    since_days = int(since_days)
except ValueError:
    raise SystemExit("--since expects an integer number of days")

with open(ticket_file, "r", encoding="utf-8") as fh:
    tickets = json.load(fh)

now = datetime.now(timezone.utc)
cutoff = now - timedelta(days=since_days)
window = [t for t in tickets if datetime.fromisoformat(t["created_at"]).astimezone(timezone.utc) >= cutoff]

weekly_counts = {}
for item in window:
    dt = datetime.fromisoformat(item["created_at"]).astimezone(timezone.utc)
    iso = dt.isocalendar()
    key = f"{iso.year}-W{iso.week:02d}"
    weekly_counts.setdefault(key, 0)
    weekly_counts[key] += 1

print(f"Tickets since {cutoff.isoformat()} (total {len(window)}):")
for key in sorted(weekly_counts):
    print(f"  {key}: {weekly_counts[key]}")

if len(window) > 2:
    os.makedirs(os.path.dirname(alert_file), exist_ok=True)
    header = "| 触发时间 | 观察窗口 | 支持票数量 | 说明 |\n"
    if not os.path.exists(alert_file):
        with open(alert_file, "w", encoding="utf-8") as fh:
            fh.write("# Support Alerts — Go Runtime Pilot\n\n")
            fh.write("| 触发时间 | 观察窗口 | 支持票数量 | 说明 |\n")
    timestamp = now.isoformat()
    desc = f">2 tickets detected (window={since_days}d)"
    with open(alert_file, "a", encoding="utf-8") as fh:
        fh.write(f"| {timestamp} | {since_days}d | {len(window)} | {desc} |\n")
PY
