#!/usr/bin/env bash

set -euo pipefail

# Go CI helper for agno-Go
# - Runs unit tests across all Go packages
# - Runs basic static analysis using go vet
#
# Usage:
#   ./scripts/go-ci.sh

ROOT_DIR="$(CDPATH="" cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

echo "[go-ci] Running go test ./..."
go test ./...

echo "[go-ci] Running go vet ./..."
go vet ./...

echo "[go-ci] Done."

