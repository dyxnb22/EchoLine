# Round 01 Test Results

**Date:** 2026-07-02

| Command | Result |
|---------|--------|
| `make dev-up` | PASS — postgres, redis, redpanda, minio running |
| `make dev-app` | PASS after migration fix |
| `make test` | PASS |
| `cd backend && go test ./...` | PASS |
| `cd backend && go vet ./...` | PASS |
| `cd frontend && npm run lint` | PASS |
| `cd frontend && npm run build` | PASS |
| `cd frontend && npm audit --omit=dev` | PASS (0 vulnerabilities) |
| `make smoke-full` | PASS (16 pass, 1 skip WS optional) |
| `cd frontend && npx playwright test` | PASS (4/4) |
| `k6 run --dry-run loadtests/k6-*.js` | SKIP — k6 not installed locally |

## Environment notes

- Docker (OrbStack) available and used for full stack verification
- API reachable at http://localhost:8080 after rebuild
- Migration 00018 applied successfully in container

## Initial failures (resolved)

1. API container exit: migrations path → fixed SLA-R01-001
2. smoke-full: invalid client_msg_id, wrong group endpoint → fixed SLA-R01-011
