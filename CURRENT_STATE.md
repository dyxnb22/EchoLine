# Current State

Current phase: Batch-120 complete.

Milestone: 120 tasks in `BATCH_120_MANIFEST.md` (T001–T120).

Last batch highlights:

- Backend: reactions, threads, forward, presence API, OpenSearch skeleton, webhook, DLQ replay, archive/export, push/payment/ads/recommendation skeletons
- Migrations: 00010–00013
- Realtime: message.edited/recalled broadcast, typing.stop
- Ops: GitHub Actions CI, k6 mixed workload, replay CLI, seed-extended, chaos/bootstrap scripts
- Docs: 7 new ADRs (0016–0022), 9 prototype/research docs, iteration-04, review-fixes
- Frontend: reactions/report/block actions, channel filter, dark mode, PWA service worker, recommendations

Tests:

- `go test ./...` passed
- `npm run build` passed

Blocker:

- Docker/Postgres unavailable in cloud VM for full integration smoke

Next:

1. `make dev-up` + `make smoke-full` + CI green on Postgres
2. Frontend thread panel + admin UI (T045–T048)
3. GraphQL prototype implementation (T029)
4. Payment/ads production paths (T028+)
