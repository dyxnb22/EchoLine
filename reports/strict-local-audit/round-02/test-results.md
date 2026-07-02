# Round 02 Test Results

**Date:** 2026-07-02

| Command | Result |
|---------|--------|
| `make test` | PASS |
| `cd backend && go test ./...` | PASS |
| `cd backend && go vet ./...` | PASS |
| `cd frontend && npm run lint` | PASS |
| `cd frontend && npm run build` | PASS |
| `make smoke-full` | PASS (16 pass, 1 skip) |
| `cd frontend && npx playwright test` | PASS (4/4) |
| `k6 run --dry-run` | SKIP — k6 not installed |

All mandatory stop-condition tests passed.
