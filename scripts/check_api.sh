#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${1:-http://127.0.0.1:8080}"

echo "== healthz =="
curl -fsS "${BASE_URL}/api/v1/healthz" | sed 's/.*/&\n/'

echo "== failure overview cards =="
curl -fsS "${BASE_URL}/api/v1/failure-rates/overview-cards" | sed 's/.*/&\n/'

echo "== failure age trend =="
curl -fsS "${BASE_URL}/api/v1/failure-rates/age-trend" | sed 's/.*/&\n/'
