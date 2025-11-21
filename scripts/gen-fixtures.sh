#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SOURCE="${FIXTURE_SOURCE_DIR:-${ROOT}/specs/001-go-agno-rewrite/contracts/fixtures-src}"
DEST="${FIXTURE_DEST_DIR:-${ROOT}/specs/001-go-agno-rewrite/contracts/fixtures}"

command -v go >/dev/null || { echo "go is required to generate fixtures"; exit 1; }

echo "==> fixtures: ${SOURCE} -> ${DEST}"
cd "${ROOT}"

if [[ "${VERIFY_ONLY:-false}" == "true" ]]; then
  go run ./go/scripts/gen_fixtures.go --source="${SOURCE}" --dest="${DEST}" --verify-only
else
  go run ./go/scripts/gen_fixtures.go --source="${SOURCE}" --dest="${DEST}"
fi
