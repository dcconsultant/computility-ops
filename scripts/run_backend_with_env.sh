#!/usr/bin/env bash
set -euo pipefail

# Usage:
#   scripts/run_backend_with_env.sh ~/.secrets/computility-ops.env

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="${1:-$HOME/.secrets/computility-ops.env}"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "env file not found: $ENV_FILE"
  exit 1
fi

set -a
# shellcheck source=/dev/null
source "$ENV_FILE"
set +a

cd "${ROOT_DIR}/backend"
go run ./cmd/server
