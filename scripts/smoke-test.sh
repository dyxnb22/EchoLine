#!/usr/bin/env sh
set -eu

echo "==> Running backend unit tests"
(cd backend && go test ./...)

if [ "${RUN_WS_SMOKE:-}" = "1" ]; then
  echo "==> Running WebSocket protocol unit smoke"
  cd backend && go test ./internal/realtime/... -run 'TestMarshal|TestHub' -count=1
fi

if [ "${RUN_API_SMOKE:-}" = "1" ]; then
  echo "==> Running API smoke against ${API_URL:-http://localhost:8080}"
  API_URL="${API_URL:-http://localhost:8080}"
  curl -fsS "${API_URL}/health" | grep -q '"status"'
  echo "health check passed"
fi

echo "Smoke checks passed."
