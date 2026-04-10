#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

MAJOR="${APP_VERSION_MAJOR:-1}"
MINOR="${APP_VERSION_MINOR:-0}"
MSG="${1:-chore: release}" 

if [[ -n "$(git status --porcelain)" ]]; then
  echo "[release] 检测到未提交改动，继续打版本并提交。"
fi

VERSION=$(APP_VERSION_MAJOR="$MAJOR" APP_VERSION_MINOR="$MINOR" node ./scripts/bump-version.mjs)
echo "[release] version => ${VERSION}"

git add frontend/src/version.ts
if ! git diff --cached --quiet; then
  git commit -m "${MSG} (${VERSION})"
else
  echo "[release] version file unchanged, skip commit"
fi

echo "[release] push origin main"
if git push origin main; then
  echo "[release] push success"
  exit 0
fi

echo "[release] proxy push failed, retry without git http(s).proxy"
git -c http.proxy= -c https.proxy= push origin main

echo "[release] done"
