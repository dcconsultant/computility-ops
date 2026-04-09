#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
mkdir -p logs

docker compose up -d --build

echo "✅ computility-ops started"
echo "Frontend: http://localhost:18080"
echo "Audit log: ./logs/audit.log"
