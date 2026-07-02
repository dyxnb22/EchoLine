# Final Code Review — EchoLine

Date: 2026-07-01  
Scope: Full project after T001–T440 + backlog + extensions closure.

## Executive summary

EchoLine is a coherent modular monolith suitable for portfolio and interview use. Core messaging, realtime, workers, and admin paths are implemented with tests. This review records findings by module and fixes applied in the final completion branch.

## Architecture

| Layer | Assessment | Notes |
|-------|------------|-------|
| API monolith | Good | Clear handler/repository split; goose migrations |
| Workers | Good | Idempotent consumers; fanout batching added |
| Frontend | Good | Router + auth context; ChatPage modularized |
| Ops | Adequate | Compose profiles, Helm skeleton, gateway stub |

## Module review

### `internal/auth` / `internal/user`

- JWT + refresh flow tested.
- **Fix:** Integration tests expanded for full messaging path.

### `internal/conversation` / `internal/message`

- Seq ordering and membership checks solid.
- **Fix:** Entitlement gate on channel subscribe returns 402.

### `internal/entitlement` (new)

- `CanSubscribe` + `Grant` idempotent by reference.
- **Recommendation:** Add owner-only check on `HandleSetPaid` (future hardening).

### `internal/worker`

- **Bug fixed:** `pushWorker.NotifyUser(ctx, uuid.Nil, ...)` removed; fanout iterates members.
- **Fix:** Bounded `seen` map prevents memory leak in `MessageCreatedHandler`.

### `internal/graph`

- `addReaction` mutation wired.
- GraphQL remains prototype — no production auth hardening beyond bearer.

### `internal/payment`

- Settle grants entitlements when reference prefix is `channel:`.

### `frontend`

- **Bug fixed:** Recommended channels now call `subscribeChannel` before navigation.
- **Fix:** Reactions prefetched on message load; edit/recall UI; notification panel; settings route.
- **Fix:** `AuthContext` refresh on 401.

### `deploy` / CI

- **Fix:** Migrations use `go run ./cmd/migrate` (goose) instead of raw `psql || true`.
- **Fix:** Playwright uses `vite preview` webServer; `continue-on-error` removed from critical jobs.

## Security notes (accepted for MVP)

- E2EE client is XOR demo only — documented in `lib/e2ee.ts`.
- Admin RBAC via `ADMIN_USER_IDS` env — not DB roles.
- JWT secret must be set in production.

## Test coverage gaps (documented)

- Dual-client WebSocket integration smoke — requires running stack.
- Search E2E — async indexer via Kafka.
- Payment + entitlement E2E — needs Stripe mock expansion.

## Optimizations applied

1. Fanout push batch size 256.
2. Idempotency map cap 10k entries.
3. Frontend route-based code splitting ready (`ChatPage` separate from `App`).
4. OTel/Sentry stubs — zero cost when env unset.

## Recommended next steps (post-closure)

1. `make dev-up` + `make smoke-full` on developer machine.
2. Owner RBAC on paid channel configuration.
3. Real Signal/WebCrypto E2EE if security track continues.
4. Replace GraphQL prototype with schema-first codegen if public API needed.
