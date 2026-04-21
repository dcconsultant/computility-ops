#!/usr/bin/env bash
set -euo pipefail

# Usage:
#   scripts/init_mysql.sh \
#     --host 127.0.0.1 --port 3306 --user root --db computility_ops

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

MYSQL_HOST="127.0.0.1"
MYSQL_PORT="3306"
MYSQL_USER="root"
MYSQL_DB="computility_ops"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --host) MYSQL_HOST="$2"; shift 2 ;;
    --port) MYSQL_PORT="$2"; shift 2 ;;
    --user) MYSQL_USER="$2"; shift 2 ;;
    --db)   MYSQL_DB="$2"; shift 2 ;;
    *) echo "Unknown arg: $1"; exit 1 ;;
  esac
done

echo "[1/6] create database ${MYSQL_DB} ..."
mysql -h "${MYSQL_HOST}" -P "${MYSQL_PORT}" -u "${MYSQL_USER}" -p \
  -e "CREATE DATABASE IF NOT EXISTS ${MYSQL_DB} CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;"

echo "[2/6] apply mysql_v1.sql ..."
mysql -h "${MYSQL_HOST}" -P "${MYSQL_PORT}" -u "${MYSQL_USER}" -p "${MYSQL_DB}" < "${ROOT_DIR}/backend/migrations/mysql_v1.sql"

echo "[3/6] apply mysql_v2_failure_dashboard.sql ..."
mysql -h "${MYSQL_HOST}" -P "${MYSQL_PORT}" -u "${MYSQL_USER}" -p "${MYSQL_DB}" < "${ROOT_DIR}/backend/migrations/mysql_v2_failure_dashboard.sql"

echo "[4/6] apply mysql_v3_ops_repo_tables.sql ..."
mysql -h "${MYSQL_HOST}" -P "${MYSQL_PORT}" -u "${MYSQL_USER}" -p "${MYSQL_DB}" < "${ROOT_DIR}/backend/migrations/mysql_v3_ops_repo_tables.sql"

echo "[5/6] apply mysql_v10_renewal_unit_prices.sql ..."
mysql -h "${MYSQL_HOST}" -P "${MYSQL_PORT}" -u "${MYSQL_USER}" -p "${MYSQL_DB}" < "${ROOT_DIR}/backend/migrations/mysql_v10_renewal_unit_prices.sql"

echo "Done. DB initialized: ${MYSQL_DB}"


echo "[6/6] apply mysql_v11_renewal_settings.sql ..."
mysql -h "${MYSQL_HOST}" -P "${MYSQL_PORT}" -u "${MYSQL_USER}" -p "${MYSQL_DB}" < "${ROOT_DIR}/backend/migrations/mysql_v11_renewal_settings.sql"
