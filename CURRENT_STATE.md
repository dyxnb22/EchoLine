# Current State

Current phase: Phase 1 in progress.

Current milestone: Phase 1 core APIs + Phase 2 WebSocket foundation (A001-C003, B001-B004).

Last completed:

- A001-A018: backend foundation, auth, conversations, messages.
- A011: refresh token endpoint.
- C003: conversation list API.
- B001-B004: WebSocket endpoint, token auth, connection hub, ping/pong heartbeat.

Tests:

- `cd backend && go test ./...` passed (integration tests skip without `DATABASE_URL`).
- `make test` passed.
- `make smoke` passed (unit-test based smoke).
- Docker/PostgreSQL integration smoke not run in this environment.

Known blockers:

- Docker unavailable in cloud VM (`make dev-up` fails). PostgreSQL not installed locally. DB integration tests and full API smoke require `DATABASE_URL` pointing to a running Postgres instance.

Next actions:

1. B005-B007: WS event envelope, message.send over WS, online push.
2. A019: history pagination tests with DB.
3. A020-A022: error envelope, OpenAPI, seed script.
4. Full integration smoke when Postgres available.

Do not repeat:

- Do not recreate Go module skeleton.
- Do not re-implement auth/register/login.
- Do not rewrite Phase 0 docs.
- Do not skip tests before marking Phase 1 done.
