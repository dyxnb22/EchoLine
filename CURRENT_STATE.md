# Current State

Current phase: Phase 2 in progress (Phase 1 core APIs implemented).

Current milestone: realtime push + sync/unread/ACK foundation (A019-D004, B005-B010, C004-C006, F001-F007 skeleton).

Last completed:

- Phase 1: auth, conversations, messages REST, refresh token, seed, OpenAPI skeleton, unified API errors.
- Phase 2: WS protocol envelope, message.send over WS, online push to conversation members, ping/pong, connection hub, WS error envelope, WS unit smoke.
- Phase 3: mark read API, unread in conversation list, sync endpoint, history pagination cursor.
- Reliability: message_deliveries table, ACK REST + WS, forward-only delivery state, client_msg_id idempotency.
- Infra skeleton: Redis client + presence TTL, in-memory event bus, worker process skeleton.

Tests:

- `cd backend && go test ./...` passed.
- `make test` passed.
- `RUN_WS_SMOKE=1 make smoke` passed.
- Full Postgres integration smoke not run (no Docker/DB in cloud VM).

Known blockers:

- Docker/PostgreSQL unavailable in cloud VM. Set external `DATABASE_URL` for integration tests.

Next actions:

1. E001-E005: group roles, channel model/APIs.
2. F005-F008: Kafka/Redpanda client + message.created consumer.
3. H001-H002: login/message rate limiting middleware.
4. Full integration smoke when Postgres/Redis available.
5. J001: frontend bootstrap after API freeze checkpoint.

Do not repeat:

- Do not re-implement auth/register/login/migrations skeleton.
- Do not rewrite Phase 0/long-run docs.
- Do not mark Phase 1/2 done without integration smoke.
