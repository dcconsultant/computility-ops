#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"
mkdir -p logs

if [[ ! -f .env ]]; then
  echo "⚠️  .env not found. You can copy from .env.example and adjust MYSQL_DSN:"
  echo "   cp .env.example .env"
fi

docker compose up -d --build

echo "✅ computility-ops started"
echo "Frontend: http://localhost:18080"
echo "Audit log: ./logs/audit.log"
