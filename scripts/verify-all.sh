#!/usr/bin/env bash
# Local verification equivalent to CI critical jobs.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "==> Backend: go test"
(cd "$ROOT/backend" && go test ./...)

echo "==> Backend: go vet"
(cd "$ROOT/backend" && go vet ./...)

echo "==> Frontend: build"
(cd "$ROOT/frontend" && npm run build)

echo "==> Frontend: Playwright smoke"
(cd "$ROOT/frontend" && npx playwright test)

echo "All verify checks passed."
