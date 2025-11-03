#!/usr/bin/env bash
set -euo pipefail

BASE_URL=${1:-http://localhost:8080}

echo "==> Health check"
curl -sf "${BASE_URL}/healthz" | jq .

echo "==> Listing sessions"
curl -sf "${BASE_URL}/sessions?type=agent" | jq '.meta'

echo "==> Creating session"
CREATE_PAYLOAD='{"session_id":"demo-session","session_name":"Demo Session","session_state":{"step":"init"},"agent_id":"demo-agent","user_id":"demo-user"}'
curl -sf -X POST "${BASE_URL}/sessions?type=agent" \
  -H 'Content-Type: application/json' \
  -d "${CREATE_PAYLOAD}" | jq '.session_id'

echo "==> Fetching session detail"
curl -sf "${BASE_URL}/sessions/demo-session?type=agent" | jq '.session_name'

echo "==> Renaming session"
RENAME_PAYLOAD='{"session_name":"Renamed Demo Session"}'
curl -sf -X POST "${BASE_URL}/sessions/demo-session/rename?type=agent" \
  -H 'Content-Type: application/json' \
  -d "${RENAME_PAYLOAD}" | jq '.session_name'

echo "==> Deleting session"
curl -sf -X DELETE "${BASE_URL}/sessions/demo-session?type=agent"
echo "Done"
