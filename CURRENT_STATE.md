# Current State

Current phase: Phase 4/5/7 partial (group/channel + reliability + media + frontend).

Current milestone: D007-D008, E006, G001-G004, H003, J001-J006.

Last completed:

- D007: transactional outbox enqueue on message create; worker drains to Kafka/memory.
- D008: dead_letter_events migration + outbox DLQ on publish failure threshold.
- E006: fanout unit test excluding sender.
- F008: outbox publisher worker (no direct Kafka on hot path).
- G001-G004: MinIO presign, attachment metadata repo, attachment message send.
- H003: per-conversation send rate limit (`conv_send`, 60/min).
- J001-J006: Vite React frontend with login, list, chat, pagination, WS reconnect.

Tests:

- `cd backend && go test ./...` passed.
- `make test` passed.
- `RUN_WS_SMOKE=1 make smoke` passed.
- `cd frontend && npm run build` passed.

Known blockers:

- Docker/PostgreSQL still unavailable in cloud VM for integration smoke.

Next actions:

1. Integration smoke when Postgres/Redis/Kafka available.
2. G005-G007: attachment download URL, search skeleton.
3. I001-I003: structured logs/metrics.
4. J007-J008: optimistic send UI, attachment upload UI.

Do not repeat:

- Do not re-implement auth/migrations/core message path.
- Do not mark phases done without integration smoke.
