# Current State

Current phase: Batch-Next-120 complete (T121–T240 manifest).

Milestone: 120 tasks in `BATCH_NEXT_120_MANIFEST.md`.

Last batch highlights:

- Backend: admin APIs + RBAC, webhook on send, GraphQL prototype, payment settle, ads frequency cap, push worker, migration 00014
- Frontend: thread panel, admin panel, reactions display, typing.stop
- Ops: Dockerfile, Helm skeleton, monitoring compose, deploy workflow, golangci-lint
- Tests: reaction/thread/webhook/push/admin/graph/replay/integration skeleton

Tests:

- `go test ./...` passed
- `npm run build` passed

Blocker:

- Docker/Postgres unavailable in cloud VM for full integration smoke

Next:

1. `make dev-up` + `make smoke-full` when Postgres available
2. Playwright CI wiring
3. GraphQL mutations/subscriptions
4. Frontend component split + routing
5. E2EE/payment/ads production paths
