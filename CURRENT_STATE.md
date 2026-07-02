# Current State

Current phase: **Documentation alignment** — post engineering review #03.

Milestone: `FINAL_COMPLETION_MANIFEST.md` (closed) + docs consistency pass.

Last session highlights:

- ADR index complete (0001–0031); duplicate ADR 0003 resolved; 0013 superseded by 0019
- `websocket-protocol.md` aligned with `realtime/protocol.go` (`message.edited`, typing events)
- `data-model.md` adds `outbox_events`
- State docs updated: `DONE.md`, `BACKLOG.md`, `ACCEPTANCE_MATRIX.md`, `TASKS.md` closure banners
- Navigation: `docs/README.md`, `README.md` interview links, `review-docs-consistency.md` refresh

Tests:

- `make verify` — go test + build + playwright (unchanged by doc-only pass)

Blocker:

- Docker/Postgres unavailable in cloud VM for `make smoke-full` — see `BLOCKERS.md`

Next (optional):

1. Migrate `conversation/handler` to `apierror` envelope (legacy `writeError`)
2. Local `make dev-up && make smoke-full`
3. Expand `docs/openapi.yaml` error response examples
