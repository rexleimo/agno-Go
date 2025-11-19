#!/usr/bin/env bash

set -euo pipefail

# Example routing script showing how to select between Python and Go
# implementations for a given scenario based on an environment variable.
#
# Usage:
#   RUNTIME_TARGET=python ./scripts/route_python_go_example.sh
#   RUNTIME_TARGET=go ./scripts/route_python_go_example.sh

TARGET="${RUNTIME_TARGET:-python}"

case "$TARGET" in
  python)
    echo "Routing to Python implementation..."
    echo "python: agno.cookbook.scripts.us1_basic_coordination_parity:run_parity"
    ;;
  go)
    echo "Routing to Go implementation..."
    echo "go: github.com/agno-agi/agno-go/go/agent.RunUS1Example"
    ;;
  *)
    echo "Unknown RUNTIME_TARGET: $TARGET" >&2
    exit 1
    ;;
esac

