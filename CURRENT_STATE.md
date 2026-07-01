# Current State

Current phase: Phase 4 in progress (group/channel permissions + infra).

Current milestone: E001-E005, F005-F008, H001-H005, D006 skeleton.

Last completed:

- E001-E002: group role checks, invite/kick/leave APIs.
- E003-E005: channel create, subscribe/unsubscribe, publish permission on send.
- F005-F008: Kafka publisher/consumer, message.created event schema, worker consumer.
- H001-H002: Redis rate limit on login and message send.
- H004-H005: audit_logs table + login audit.
- D006: outbox_events migration skeleton.

Tests:

- `cd backend && go test ./...` passed.
- `make test` passed.

Known blockers:

- Docker/PostgreSQL still unavailable in cloud VM for integration smoke.

Next actions:

1. E006: small group online fanout optimization.
2. F008 hardening: outbox publisher instead of direct Kafka publish.
3. H003: per-conversation send rate limit.
4. J001: frontend bootstrap.
5. Integration smoke when Postgres/Redis/Kafka available.

Do not repeat:

- Do not re-implement auth/migrations/core message path.
- Do not mark phases done without integration smoke.
