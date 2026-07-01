# EchoLine Iteration 02 Report

## Scope

Phase 1 completion push, Phase 2 realtime main path, Phase 3 sync/unread/ACK foundations, and Phase 6 infra skeleton.

## Completed Task IDs

- A019-A022: history pagination cursor, unified REST error envelope + request_id, OpenAPI skeleton, seed command
- B005-B010: WS protocol envelope, message.send, online push, WS error envelope, smoke script hooks
- C001-C006: latest_seq/seq allocation (existing), mark read, unread in list, sync endpoint
- D001-D004: client_msg_id idempotency, delivery table, ACK REST/WS, forward-only state machine
- F001-F003 partial: Redis client, presence TTL store, in-memory event bus + worker skeleton

## Key Files

- `backend/internal/realtime/` protocol + push + message.send/ack
- `backend/internal/message/service.go` shared send path for REST/WS
- `backend/internal/sync/`, `backend/internal/delivery/`
- `backend/internal/apierror/` request_id middleware
- `backend/cmd/seed`, `backend/cmd/worker`
- `backend/migrations/00003_deliveries.sql`
- `docs/openapi.yaml`, updated `docs/api.md`

## Tests

- `cd backend && go test ./...` pass
- `make test` pass
- `RUN_WS_SMOKE=1 make smoke` pass
- DB integration tests still skip without `DATABASE_URL`

## Risks / Blockers

- Cloud VM has no Docker/PostgreSQL; full end-to-end API/WS integration not executed here
- Kafka/Redpanda consumer not wired yet (memory bus only)
- Rate limit middleware not yet attached to routes

## Next

1. Full integration smoke with Postgres + Redis
2. B009 reconnect fallback doc polish
3. E001-E005 group/channel permissions
4. F005-F008 Kafka worker handlers
5. Frontend J001+ after API stabilizes
