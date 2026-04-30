#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
PORT="${PORT:-$((18080 + RANDOM % 1000))}"
APP_ADDR="${APP_ADDR:-:${PORT}}"
BASE_URL="${BASE_URL:-http://127.0.0.1:${PORT}}"

TMP_LOG="$(mktemp -t ops-backend-log.XXXXXX)"
cleanup() {
  if [[ -n "${SERVER_PID:-}" ]] && kill -0 "$SERVER_PID" 2>/dev/null; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
  rm -f "$TMP_LOG"
}
trap cleanup EXIT

cd "$BACKEND_DIR"

# Start backend with memory driver for deterministic E2E contract checks.
APP_ADDR="$APP_ADDR" STORAGE_DRIVER=memory go run ./cmd/server >"$TMP_LOG" 2>&1 &
SERVER_PID=$!

# Wait for health endpoint (max ~20s)
READY=0
for _ in $(seq 1 40); do
  if curl -fsS "${BASE_URL}/api/v1/healthz" >/dev/null 2>&1; then
    READY=1
    break
  fi
  sleep 0.5
done

if [[ "$READY" -ne 1 ]]; then
  echo "[E2E] backend not ready"
  echo "---- backend log ----"
  cat "$TMP_LOG"
  exit 1
fi

check_endpoint() {
  local path="$1"
  local body
  body="$(curl -fsS "${BASE_URL}${path}")"

  # Contract: envelope code/message exists, code must be 0.
  echo "$body" | grep -q '"code":0' || {
    echo "[E2E] ${path} failed envelope code check"
    echo "$body"
    exit 1
  }

  # Contract: data field exists.
  echo "$body" | grep -q '"data"' || {
    echo "[E2E] ${path} missing data field"
    echo "$body"
    exit 1
  }
}

check_endpoint "/api/v1/ops/decisions/replacement"
check_endpoint "/api/v1/ops/decisions/reconfig"
check_endpoint "/api/v1/ops/decisions/self-repair"

echo "[E2E] ops decision endpoints passed"
