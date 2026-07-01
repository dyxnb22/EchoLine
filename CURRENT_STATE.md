# Current State

Current phase: Batch-Next-200 complete (T241–T440 manifest).

Milestone: 200 tasks in `BATCH_NEXT_200_MANIFEST.md`.

Last batch highlights:

- Backend: E2EE key API, webhook retry worker, GraphQL sendMessage mutation, last-seen, friend recommendations, migration 00015
- Frontend: LoginPage split, ConversationActions (pin/archive/export/forward), friend recs
- Ops: docker compose app profile, backup script, k8s secrets, Playwright CI job, strict integration tests
- Docs: ADR 0023–0026, iteration-06

Tests:

- `go test ./...` passed
- `npm run build` passed

Blocker:

- Docker/Postgres unavailable in cloud VM for full integration smoke

Next:

1. `make dev-up` + `make dev-app` + `make smoke-full`
2. react-router wiring in frontend
3. Channel entitlement enforcement
4. Playwright send-message E2E
