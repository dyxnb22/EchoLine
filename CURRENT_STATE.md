# Current State

Current phase: Phase 5/7/8 partial (reliability + media/search + observability + frontend).

Current milestone: batch-20 (G005-G008, C007-C009, D007, E007-E008, F004/F008, H006, I001-I005, J007-J009).

Last completed (20 tasks):

1. G005: presigned download URL (`POST /api/media/download-url`)
2. G006: message indexing worker on `message.created`
3. G007: `GET /api/search/messages` with PostgreSQL tsvector
4. G008: search scoped to member conversations
5. D007: outbox `FOR UPDATE SKIP LOCKED` fetch
6. F008: idempotent `message.created` consumer handler
7. E007: large group fanout ADR (`docs/adr/0003-large-group-fanout.md`)
8. E008: fanout worker skeleton
9. F004: Redis conversation list cache (30s TTL)
10. C007: `device_sync_cursors` + sync persistence
11. C008: message edit API (`PATCH`)
12. C009: message recall API (`POST .../recall`)
13. H006: recall audit logging
14. I001/I002: structured logs + `X-Trace-ID` middleware
15. I003: Prometheus `/metrics`
16. I004: WS connection gauge
17. I005: message send latency histogram
18. J007: optimistic send UI
19. J008: attachment upload UI
20. J009: search UI in sidebar

Tests:

- `go test ./...` passed
- `make test` passed
- `RUN_WS_SMOKE=1 make smoke` passed
- `npm run build` passed

Known blockers:

- Docker/PostgreSQL unavailable in cloud VM for integration smoke.

Next actions:

1. Integration smoke when Postgres available
2. F009 MQ lag metrics, I006 k6 load test
3. C010 pinned messages, B011 typing indicator
4. OpenSearch adapter (optional upgrade from pg tsvector)

Do not repeat:

- Do not re-implement auth/migrations/core message path.
- Do not mark phases done without integration smoke.
