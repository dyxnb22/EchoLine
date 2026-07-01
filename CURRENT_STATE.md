# Current State

Current phase: Batch-100 complete (social, notifications, observability, docs, loadtests).

Milestone: 100 tasks tracked in `BATCH_100_MANIFEST.md`.

Last batch highlights:

- Social: pins, blocks, reports, mute, notifications, spam checker
- Admin: health + DLQ list skeleton
- Realtime: typing indicator WS event
- Frontend: register, notifications badge, typing UI, mark-read, PWA manifest, Playwright skeleton
- Ops: k6 scripts, chaos scripts, Grafana dashboard, `.env.example`
- Docs: 12 ADRs, 10 research/interview docs, 9 review reports, iteration-03

Tests:

- `go test ./...` passed
- `npm run build` passed
- `RUN_WS_SMOKE=1 make smoke` passed

Blocker:

- Docker/Postgres unavailable in cloud VM for full integration smoke

Next:

1. `make dev-up` + `make smoke-full` when Postgres available
2. Playwright CI wiring (J010)
3. OpenSearch adapter (G009)
4. Future extensions X003+
