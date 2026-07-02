# Current State

Current phase: **Post-audit full-stack verification complete — stop condition met (optional P3/P4 polish only)**.

Latest execution (2026-07-02):

- Started local Docker/OrbStack stack with Postgres, Redis, Redpanda, MinIO, API, and worker.
- Fixed live WebSocket smoke: `scripts/smoke-api-full.sh` now uses portable Node WebSocket probing, includes required `device_id`, and no longer depends on GNU `timeout`/global `wscat`.
- Fixed API WebSocket upgrade behind metrics middleware by preserving `http.Hijacker` on the status recorder.
- Rebuilt API/worker containers and ran full API + WS smoke successfully.
- Tests: `go test ./...`, `go vet ./...`, `RUN_INTEGRATION=1 go test ./tests`, `make smoke`, `make smoke-full`, `RUN_WS_SMOKE=1 make smoke-full`, `npm run lint`, `npm run build`, `npm audit --omit=dev`, `npm run test:e2e` all pass in this workspace.
- Report: `reports/iteration-07.md`.

Latest recheck (2026-07-02):

- No new P0/P1/P2 issues found.
- Revalidated strict-audit fixes around rate limiting, GraphQL authz, ads impression cap, media access, ACK binding, and frontend WS/token behavior.
- Tests: `go test ./...`, `go vet ./...`, targeted `go test -race`, `npm run lint`, `npm run build`, `npm audit --omit=dev`, `npm run test:e2e` all pass.
- Report: `reports/recheck-2026-07-02.md`.

Audit highlights:

- **P0:** sync cursor + frontend `has_more` pagination (message loss on reconnect)
- **P1:** idempotent send side effects, media member download, outbox processing claim, payment 402 detection, WS token refresh, sender dedup, upload pending, unread while viewing
- **P2:** ACK message binding, admin health RBAC, cache `can_publish`, WS edit/recall handlers
- **P3/P4:** presigned URL sharing (wontfix MVP), WS query token, Redis limiter Lua hardening, edit/recall HTTP semantics polish, loading polish

Tests:

- `go test ./...` — pass
- `npm run build` — pass
- `RUN_WS_SMOKE=1 make smoke-full` — pass

Blocker:

- No active blocker in this workspace. Historical cloud VM Docker/Postgres blocker is resolved locally; see `BLOCKERS.md`.

Reports:

- `reports/deep-review-final.md`
- `reports/iteration-07.md`

Next (optional):

1. Migrate `conversation/handler` legacy `writeError` to `apierror` envelope
2. Swap OTel stub for real exporter SDK
3. Expand `docs/openapi.yaml` request/response body schemas
