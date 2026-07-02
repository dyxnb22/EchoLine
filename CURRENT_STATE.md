# Current State

Current phase: **Full audit fix — Phases 0–4 complete**.

Last session highlights:

- **P0:** thread/forward via `message.Service`; RBAC on export/reaction/thread/forward/ads; frontend typing/sync/composer fixes
- **P1:** paid channel UI flow; DLQ replay requeue; reaction idempotent; apierror envelope; device WS touch; composite admin RBAC; fanout pagination; create-conversation modal
- **P2:** WS buffer 256 + drop metric; outbox publisher fix; sync `has_more`; client_msg_id UUID required; Kafka consumer in worker; outbox cleanup job; edit/recall outbox
- **P3:** integration RBAC tests; CI worker job; dual-client WS unit test; Playwright extended smoke
- **P4:** admin-panel.md, business-flows, websocket-protocol alignment

Tests:

- `go test ./...` — pass
- `npm run build` — pass

Blocker:

- Docker/Postgres unavailable in cloud VM for `make smoke-full` — see `BLOCKERS.md`

Next (optional):

1. Local `make dev-up && make smoke-full`
2. Full-stack Playwright against running compose stack
