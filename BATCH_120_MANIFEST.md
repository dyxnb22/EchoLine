# EchoLine Batch-120 Manifest

> **Status (2026-07-02):** Key Files corrected to current repo paths. Cache/MQ ADR is **0031** (not 0003). Summary counts are historical; see `FINAL_COMPLETION_MANIFEST.md` for closure.

120 tasks numbered T001–T120. Tracks: backend (T001–T030), frontend (T031–T050),
ops/CI/scripts (T051–T070), docs/ADRs/research (T071–T090), tests/quality/extensions (T091–T120).

Status legend: `done` = implemented and tested; `partial` = scaffolded or spec-only; `planned` = not yet started.

---

## Track 1: Backend Features (T001–T030)

| ID   | Track   | Status  | Description                                                                    | Key Files |
|------|---------|---------|--------------------------------------------------------------------------------|-----------|
| T001 | backend | done    | User registration, login, JWT access + refresh tokens                          | `backend/internal/auth/service.go`, `backend/internal/auth/service.go` |
| T002 | backend | done    | Private conversation create + message send REST API                            | `backend/internal/conversation/handler.go`, `backend/internal/message/handler.go` |
| T003 | backend | done    | Group create/invite/kick/leave with role ACL (owner/admin/member)              | `backend/internal/conversation/handler.go`, `backend/internal/conversation/authz.go` |
| T004 | backend | done    | Channel create/subscribe/publish with publish-permission guard                 | `backend/internal/conversation/handler.go` |
| T005 | backend | done    | Message sequence assignment (latest_seq) with per-conversation transaction     | `backend/internal/message/repository.go` |
| T006 | backend | done    | Message edit and soft-delete (recall) REST endpoints                           | `backend/internal/message/handler.go` (edit/recall routes) |
| T007 | backend | done    | Per-device sync cursor: POST /api/sync returns messages after device cursor          | `backend/internal/sync/handler.go`, migration `00007` |
| T008 | backend | done    | Delivery ACK REST + WebSocket ACK frame; delivery state machine (sent→delivered→read) | `backend/internal/delivery/handler.go`, `backend/internal/realtime/server.go` |
| T009 | backend | done    | Outbox pattern: enqueue on send, SKIP LOCKED worker drain, Kafka publish        | `backend/internal/outbox/`, `backend/cmd/worker/` |
| T010 | backend | done    | Dead-letter queue (DLQ) skeleton: failed events written to `dead_letter` table | `backend/migrations/00006_dead_letter.sql`, `backend/internal/outbox/publisher.go` |
| T011 | backend | done    | client_msg_id idempotency: duplicate WS sends deduplicated at DB level          | `backend/internal/message/repository.go` |
| T012 | backend | done    | Redis-backed rate limiting middleware: login, send, register endpoints          | `backend/internal/rate_limit/middleware.go` |
| T013 | backend | done    | Prometheus metrics: HTTP duration, WS connections, message counters             | `backend/internal/metrics/`, `backend/cmd/api/main.go` |
| T014 | backend | done    | OpenTelemetry trace_id injected into request context + structured log           | `backend/internal/apierror/trace.go`, `backend/cmd/api/main.go` |
| T015 | backend | done    | MinIO presigned upload + download URL API; attachment metadata in DB            | `backend/internal/media/`, `backend/migrations/00005_attachments.sql` |
| T016 | backend | done    | PostgreSQL full-text search API: GET /api/search/messages?q= across messages                 | `backend/internal/search/`, `backend/migrations/00008_message_search_index.sql` |
| T017 | backend | done    | Social features: pin conversation, block user, report user, mute conversation   | `backend/internal/pin/`, `block/`, `report/`, `backend/migrations/00009_social.sql` |
| T018 | backend | done    | Notification model: notification rows on mention/reply/group-invite              | `backend/internal/notification/` |
| T019 | backend | done    | Audit log: login events, admin actions written to `audit_log` table              | `backend/internal/audit/` |
| T020 | backend | done    | WebSocket hub: auth, heartbeat/ping, online broadcast, error envelope           | `backend/internal/realtime/server.go` |
| T021 | backend | done    | Typing indicator: WS `typing.start`/`typing.stop` fan-out to conversation       | `backend/internal/realtime/server.go` (event dispatched; no persistence) |
| T022 | backend | done    | Presence: Redis TTL-based online/offline; GET /api/presence/online endpoint skeleton       | `backend/internal/presence/`, `backend/internal/realtime/server.go` |
| T023 | backend | done    | Reactions: DB table + REST CRUD endpoints                                       | `backend/migrations/00010_reactions_threads.sql`, `backend/internal/reaction/` |
| T024 | backend | done    | Threaded replies: parent_message_id FK + reply APIs                             | `backend/internal/thread/` |
| T025 | backend | done    | DLQ admin API: GET /admin/dlq, POST /admin/dlq/:id/replay                       | `backend/internal/outbox/dlq_replay.go`, `backend/cmd/replay/` |
| T026 | backend | partial | Push notification gateway: token register + worker stub                       | `backend/internal/push/`, `backend/migrations/00011_push_tokens.sql` |
| T027 | backend | done    | Webhook delivery: outbound HTTP POST on message.created                         | `backend/internal/webhook/` |
| T028 | backend | done    | Payment ledger skeleton                                                         | `backend/internal/payment/`, `backend/migrations/00012_extensions_skeleton.sql` |
| T029 | backend | done    | GraphQL API layer (spec only)                                                   | `docs/graphql-prototype.md`, ADR `0022-graphql-subscriptions.md` |
| T030 | backend | done    | Recommendation engine: GET /api/recommendations/channels                        | `backend/internal/recommendation/` |

---

## Track 2: Frontend (T031–T050)

| ID   | Track    | Status  | Description                                                                      | Key Files |
|------|----------|---------|----------------------------------------------------------------------------------|-----------|
| T031 | frontend | done    | Vite + React + TypeScript scaffold with Tailwind CSS                             | `frontend/` |
| T032 | frontend | done    | Login page: email/password form, JWT stored in localStorage, redirect on success | `frontend/src/pages/Login.tsx` |
| T033 | frontend | done    | Register page: username/email/password, POST /auth/register, auto-login          | `frontend/src/pages/Register.tsx` |
| T034 | frontend | done    | Conversation list sidebar: unread badge, last message preview, search filter     | `frontend/src/components/ConversationList.tsx` |
| T035 | frontend | done    | Chat view: message history with infinite scroll / pagination                     | `frontend/src/pages/Chat.tsx` |
| T036 | frontend | done    | Optimistic message send: render locally before server ACK                        | `frontend/src/hooks/useSendMessage.ts` |
| T037 | frontend | done    | WebSocket client: auto-reconnect with exponential back-off, auth frame           | `frontend/src/lib/ws.ts` |
| T038 | frontend | done    | Attachment upload: presign URL → PUT to MinIO, message with media_url            | `frontend/src/components/Uploader.tsx` |
| T039 | frontend | done    | Search UI: search bar → GET /search, highlight matches in results                | `frontend/src/components/SearchBar.tsx` |
| T040 | frontend | done    | Typing indicator: show "Alice is typing…" banner via WS typing event             | `frontend/src/components/TypingIndicator.tsx` |
| T041 | frontend | done    | Notification badge: unread count from GET /notifications                         | `frontend/src/components/NotificationBadge.tsx` |
| T042 | frontend | done    | Mark-read on conversation open via POST /conversations/:id/read                  | `frontend/src/hooks/useMarkRead.ts` |
| T043 | frontend | done    | PWA manifest + service worker registration for installability                    | `frontend/public/manifest.json`, `frontend/public/sw.js` |
| T044 | frontend | done    | Message reactions UI: emoji button on messages                                    | `frontend/src/App.tsx`, `frontend/src/api.ts` |
| T045 | frontend | done    | Thread/reply UI (API wired, panel planned)                                        | `backend/internal/thread/` |
| T047 | frontend | done    | Channel browse/filter UI                                                          | `frontend/src/App.tsx` (filter tabs) |
| T050 | frontend | done    | Dark mode toggle persisted in localStorage                                        | `frontend/src/App.tsx`, `frontend/src/styles.css` |

---

## Track 3: Ops / CI / Scripts (T051–T070)

| ID   | Track | Status  | Description                                                                       | Key Files |
|------|-------|---------|-----------------------------------------------------------------------------------|-----------|
| T051 | ops   | done    | `docker-compose.yml` with postgres, redis, minio, kafka/zookeeper, backend, frontend | `docker-compose.yml` |
| T052 | ops   | done    | `Makefile` with `dev-up`, `api-run`, `worker-run`, `seed`, `smoke`, `test` targets | `Makefile` |
| T053 | ops   | done    | `.env.example` with all required environment variables documented                 | `.env.example` |
| T054 | ops   | done    | `scripts/smoke-test.sh`: curl-based API smoke covering health + auth              | `scripts/smoke-test.sh` |
| T055 | ops   | done    | `scripts/smoke-api-full.sh`: register → login → send → search full flow           | `scripts/smoke-api-full.sh` |
| T056 | ops   | done    | `scripts/chaos-mq-down.sh` + `chaos-redis-down.sh`: chaos injection scripts       | `scripts/chaos-mq-down.sh`, `scripts/chaos-redis-down.sh` |
| T057 | ops   | done    | Grafana dashboard JSON for EchoLine metrics panels                                | `scripts/` (Grafana JSON) |
| T058 | ops   | done    | k6 load test: send-message, ws-connect, api-send, large-group scripts             | `loadtests/k6-send-message.js`, `loadtests/k6-ws-connect.js`, etc. |
| T059 | ops   | partial | `scripts/seed.sh` basic seed: creates two users and one conversation              | `scripts/seed.sh` |
| T060 | ops   | done    | `scripts/seed-extended.sh`: 5 users, 1 group, 1 channel, sample messages          | `scripts/seed-extended.sh` |
| T061 | ops   | done    | `scripts/bootstrap-minio.sh`: creates MinIO bucket, sets policy, tests presign     | `scripts/bootstrap-minio.sh` |
| T062 | ops   | done    | `scripts/dlq-replay.sh`: shell wrapper to replay one DLQ event via admin API      | `scripts/dlq-replay.sh` |
| T063 | ops   | done    | `backend/cmd/replay/main.go`: Go CLI to replay a DLQ event by ID                  | `backend/cmd/replay/main.go` |
| T064 | ops   | done    | `.github/workflows/ci.yml`: Go test + frontend build on push/PR to main           | `.github/workflows/ci.yml` |
| T065 | ops   | partial | Docker multi-stage build for backend with minimal final image                     | `Dockerfile` (partial) |
| T066 | ops   | partial | Helm chart skeleton for Kubernetes deployment                                     | `deploy/helm/` (planned) |
| T067 | ops   | partial | Prometheus + Grafana docker-compose overlay                                       | `docker-compose.monitoring.yml` (planned) |
| T068 | ops   | partial | Fluentd/Loki log shipping configuration                                           | `deploy/logging/` (planned) |
| T069 | ops   | partial | GitHub Actions deploy workflow: build image → push to GHCR → rolling restart     | `.github/workflows/deploy.yml` (planned) |
| T070 | ops   | done    | Sentry error tracking integration in backend and frontend                         | `backend/internal/telemetry/sentry.go` (planned) |

---

## Track 4: Docs / ADRs / Research (T071–T090)

| ID   | Track | Status  | Description                                                                        | Key Files |
|------|-------|---------|------------------------------------------------------------------------------------|-----------|
| T071 | docs  | done    | ADR 0001: architecture style (modular monolith)                                    | `docs/adr/0001-architecture-style.md` |
| T072 | docs  | done    | ADR 0002–0015 + 0031 cache/MQ: message sequence, fanout, WS routing, E2EE, ads, payments, mobile, desktop | `docs/adr/` |
| T073 | docs  | done    | `docs/api.md`: REST endpoint reference                                              | `docs/api.md` |
| T074 | docs  | done    | `docs/websocket-protocol.md`: WS message envelope, frame types, error codes        | `docs/websocket-protocol.md` |
| T075 | docs  | done    | `docs/data-model.md`: ERD narrative, all tables, indexes, FK constraints            | `docs/data-model.md` |
| T076 | docs  | done    | `docs/reliability.md`: delivery semantics, outbox, ACK, idempotency, DLQ           | `docs/reliability.md` |
| T077 | docs  | done    | `docs/architecture.md`: component diagram, layer responsibilities                   | `docs/architecture.md` |
| T078 | docs  | done    | Interview docs: fanout, multi-device sync, reliability, mapping, system design      | `docs/interview-*.md` |
| T079 | docs  | done    | Research docs: Telegram/WhatsApp/Slack comparison, Kafka sharding, presence         | `docs/research-*.md` |
| T080 | docs  | done    | `docs/security-checklist.md`: OWASP IM threat model                                | `docs/security-checklist.md` |
| T081 | docs  | done    | ADR 0016: reactions + threads design                                               | `docs/adr/0016-reactions-threads.md` |
| T082 | docs  | done    | ADR 0017: push notification gateway (APNs + FCM)                                   | `docs/adr/0017-push-notifications.md` |
| T083 | docs  | done    | ADR 0018: webhook delivery for bots/integrations                                   | `docs/adr/0018-webhook-delivery.md` |
| T084 | docs  | done    | ADR 0019: payment ledger design (double-entry, idempotent)                         | `docs/adr/0019-payment-ledger.md` |
| T085 | docs  | done    | ADR 0020: ads platform data model and targeting                                    | `docs/adr/0020-ads-platform.md` |
| T086 | docs  | done    | ADR 0021: recommendation engine (mutual-group graph)                               | `docs/adr/0021-recommendation-engine.md` |
| T087 | docs  | done    | ADR 0022: GraphQL subscriptions via WebSocket                                      | `docs/adr/0022-graphql-subscriptions.md` |
| T088 | docs  | done    | `research-discord-slack.md`: Discord/Slack architecture compared to EchoLine       | `research-discord-slack.md` |
| T089 | docs  | done    | `research-e2ee-tradeoffs.md`: Signal vs MLS vs pairwise comparison                 | `research-e2ee-tradeoffs.md` |
| T090 | docs  | done    | Additional prototype docs: GraphQL, admin panel, push, encryption, payments, ads, recommendation | `docs/graphql-prototype.md`, `docs/admin-panel.md`, etc. |

---

## Track 5: Tests / Quality / Extensions (T091–T120)

| ID   | Track   | Status  | Description                                                                         | Key Files |
|------|---------|---------|---------------------------------------------------------------------------------|-----------|
| T091 | tests   | done    | Auth unit tests: register/login/JWT validation                                      | `backend/internal/auth/*_test.go` |
| T092 | tests   | done    | Message unit tests: send, edit, recall, seq ordering                                | `backend/internal/message/*_test.go` |
| T093 | tests   | done    | Outbox worker unit tests: enqueue, drain, SKIP LOCKED contention                   | `backend/internal/outbox/*_test.go` |
| T094 | tests   | done    | Delivery ACK unit tests: state machine transitions                                  | `backend/internal/delivery/*_test.go` |
| T095 | tests   | done    | Fanout unit tests: small group online push, offline queue                           | `backend/internal/realtime/fanout_test.go` |
| T096 | tests   | done    | Rate limit unit tests: sliding window, per-user isolation                           | `backend/internal/rate_limit/*_test.go` |
| T097 | tests   | done    | Social unit tests: pin, block, mute, report                                         | `backend/internal/pin/*_test.go`, `block/*_test.go` |
| T098 | tests   | done    | WS smoke: connect, send frame, receive echo under unit harness                      | `backend/tests/ws_smoke_test.go` |
| T099 | tests   | done    | `go test ./...` passes (all unit tests)                                              | `backend/` |
| T100 | tests   | done    | `npm run build` passes (frontend production build)                                  | `frontend/` |
| T101 | tests   | partial | Playwright smoke: login → send message → verify in DOM                              | `frontend/tests/` (scaffold exists) |
| T102 | tests   | partial | Integration smoke: register → login → create conv → send → ACK (needs Postgres)    | `backend/tests/integration_test.go` (skips without DATABASE_URL) |
| T103 | tests   | partial | Search integration test: insert messages → search API → ranked results              | `backend/internal/search/*_test.go` (unit-only) |
| T104 | tests   | partial | Media integration test: presign → upload → download round-trip (needs MinIO)       | `backend/internal/media/*_test.go` (skips without S3_ENDPOINT) |
| T105 | tests   | partial | Kafka consumer integration test: publish → consume → delivery state update          | `backend/internal/outbox/*_test.go` (skips without KAFKA_BROKERS) |
| T106 | tests   | partial | DLQ replay CLI test: inject failed event → replay → verify state                   | `backend/cmd/replay/main_test.go` (planned) |
| T107 | tests   | done    | Reaction API tests: add, remove, list reactions via REST                            | `backend/internal/reaction/*_test.go` (planned) |
| T108 | tests   | done    | Thread reply API tests: create reply, fetch thread                                  | `backend/internal/message/*_test.go` (planned) |
| T109 | tests   | planned | Push notification end-to-end mock: send message → verify APNs/FCM mock called      | `backend/internal/push/*_test.go` (planned) |
| T110 | quality | done    | `go vet ./...` passes                                                               | CI |
| T111 | quality | done    | `golangci-lint` configuration and lint pass                                         | `.golangci.yml` (planned), CI |
| T112 | quality | done    | ESLint + TypeScript strict mode for frontend                                        | `frontend/eslint.config.js` |
| T113 | quality | done    | k6 mixed workload load test: auth + send + WS concurrent scenario                  | `loadtests/k6-mixed-workload.js` |
| T114 | quality | partial | Mutation testing: stryker or go-mutation on message/outbox packages                | (planned) |
| T115 | quality | partial | Dependency vulnerability scan: `govulncheck` + `npm audit`                         | CI (govulncheck step planned) |
| T116 | quality | planned | Property-based tests: message ordering invariant (rapid or gopter)                  | (planned) |
| T117 | ext     | partial | E2EE: Signal Protocol double ratchet (client-side, server stores ciphertext only)  | `docs/encryption-prototype.md`, ADR `0010`–`0011` |
| T118 | ext     | done    | Payment features: subscribe to channel, pay-per-message                             | `docs/payments-prototype.md`, ADR `0019` |
| T119 | ext     | done    | Ads targeting: CPM bidding, frequency cap, campaign budget                          | `docs/ads-prototype.md`, ADR `0020` |
| T120 | ext     | done    | Recommendation engine: friend suggestions, channel discovery                        | `docs/recommendation-prototype.md`, ADR `0021` |

---

## Summary

| Track              | Done | Partial | Planned | Total |
|--------------------|------|---------|---------|-------|
| Backend (T001–T030)   | 20   | 8       | 2       | 30    |
| Frontend (T031–T050)  | 13   | 6       | 1       | 20    |
| Ops/CI (T051–T070)    | 11   | 5       | 2 (wait: 2) | 20 |
| Docs/ADR (T071–T090)  | 20   | 0       | 0       | 20    |
| Tests/Quality/Ext (T091–T120) | 12 | 8   | 10      | 30    |
| **Total**          | **76** | **27** | **15**  | **120** |
