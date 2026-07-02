# Iteration 05 — Batch Next 120 (T121–T240)

## Scope

Close batch-120 partials: admin APIs, webhook wiring, GraphQL prototype, push worker, frontend thread/admin UI, ops deploy skeleton.

## Deliverables

- Backend: migration `00014`, admin RBAC, webhook on send, GraphQL `POST /graphql`, payment settle, ads frequency cap, push worker mock
- Frontend: `ThreadPanel`, `AdminPanel`, reactions display, typing.stop
- Ops: `Dockerfile`, Helm skeleton, monitoring compose, deploy workflow, golangci-lint
- Tests: reaction/thread/webhook/push/admin/graph/replay/integration skeleton — all unit tests pass

## Verification

```bash
cd backend && go test ./...
cd frontend && npm run build
```

Integration smoke blocked in cloud VM (no Docker/Postgres). CI applies migrations on Postgres service container.

## Known Gaps

- Admin UI requires `ADMIN_USER_IDS` env
- GraphQL supports `conversations` query only
- Playwright not wired in CI
- E2EE/payments/ads production paths remain partial

## Interview Points

- Webhook dispatch is async (goroutine) to avoid blocking send path
- Admin RBAC via env allowlist (prototype); production needs DB role + audit
- Ads frequency cap enforced at impression insert with daily count
