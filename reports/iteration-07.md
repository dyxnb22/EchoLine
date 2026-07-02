# Iteration 07 - Full-stack smoke and WS live verification

Date: 2026-07-02

## Task

Close the remaining local verification gap after project completion: run the Docker compose stack, execute full API smoke, execute Postgres integration tests, and make live WebSocket smoke reliable.

## Changes

- `scripts/smoke-api-full.sh`
  - Added a portable WebSocket probe using Node 22+ built-in `WebSocket`.
  - Kept `wscat` as fallback and removed the hard dependency on GNU `timeout`.
  - Added the required `device_id` query parameter for `/ws`.
- `backend/internal/metrics/middleware.go`
  - Preserves `http.Hijacker`, `http.Flusher`, `http.Pusher`, and `Unwrap` through the metrics status recorder so WebSocket upgrades work behind middleware.
- `backend/internal/metrics/middleware_test.go`
  - Added a regression test for Hijacker passthrough.
- State files updated to remove stale local Docker/Postgres blocker guidance.

## Verification

```bash
make dev-up
make dev-app
cd backend && go test ./...
cd backend && go vet ./...
RUN_INTEGRATION=1 DATABASE_URL='postgres://echoline:echoline@localhost:5432/echoline?sslmode=disable' go test ./tests
make smoke
make smoke-full
RUN_WS_SMOKE=1 make smoke-full
```

Results:

- Full API smoke: pass.
- Live WebSocket smoke: pass, including valid connection and invalid token rejection.
- Postgres integration tests: pass.
- Backend unit tests and vet: pass.

## Notes

The previous cloud VM blocker remains useful historical context, but this workspace now has a working local Docker/OrbStack verification path.
