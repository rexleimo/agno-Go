#!/usr/bin/env bash

set -euo pipefail

# Minimal driver script for the US1 parity scenario.
# This script is a placeholder that documents how parity runs
# will eventually be executed. It:
#   - describes the Python entrypoint module
#   - describes the Go entrypoint function
#   - prints a JSON stub compatible with ParityRun/ParityTestScenario.

ROOT_DIR="$(CDPATH="" cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

cat <<EOF
{
  "run_id": "us1-basic-coordination-demo",
  "scenarios": [
    {
      "scenario_id": "teams-basic-coordination-us1",
      "python_entry": "agno.cookbook.scripts.us1_basic_coordination_parity:run_parity",
      "go_entry": "github.com/agno-agi/agno-go/go/agent.RunUS1Example"
    }
  ]
}
EOF

