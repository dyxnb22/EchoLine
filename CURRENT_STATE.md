# Current State

Current phase: **Audit issue remediation complete (31 items)**.

Last session highlights:

- P0: archived list API contract fixed (frontend + e2e)
- P1: media download membership check + UI; WS JWT refresh; edit/recall/ack handlers; sync has_more loop; GraphQL client_msg_id; owner paid channel UI; payment settle validation
- P2: unsubscribe/unpin/mute/unblock UI; admin health RBAC; ads list membership; JWT min length; register rate limit; doc sync
- P3: integration.spec.ts placeholder; manifest residual notes updated

Tests:

- `go test ./...` — pass
- `npm run build` — pass

Blocker:

- Docker/Postgres unavailable in cloud VM for `make smoke-full` — see `BLOCKERS.md`

Next (optional):

1. Local `make dev-up && make smoke-full`
2. Full-stack Playwright against running compose stack
