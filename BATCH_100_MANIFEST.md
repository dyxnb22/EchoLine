# EchoLine Batch-100 Task Manifest

This manifest records 100 tasks (B001–B100) executed across the EchoLine backlog.
Status: **done** = fully implemented and tested; **partial** = scaffolded or implemented without full integration test coverage (blocked by Docker/Postgres unavailability in cloud VM).

| ID   | Track | Status  | Description                                         | Key File(s)                                                                 |
|------|-------|---------|-----------------------------------------------------|-----------------------------------------------------------------------------|
| B001 | WS    | done    | WebSocket endpoint (`/ws`)                          | `backend/internal/realtime/server.go`                                       |
| B002 | WS    | done    | WS auth handshake (JWT token query param)           | `backend/internal/realtime/server.go`                                       |
| B003 | WS    | done    | Connection manager registry                         | `backend/internal/realtime/server.go`                                       |
| B004 | WS    | done    | Ping/pong heartbeat & idle-connection cleanup       | `backend/internal/realtime/server.go`                                       |
| B005 | WS    | done    | Server-event envelope (type/payload/request_id)     | `backend/internal/realtime/events.go`                                       |
| B006 | WS    | done    | `message.send` over WebSocket                       | `backend/internal/realtime/server.go`                                       |
| B007 | WS    | done    | Online-member push on message creation              | `backend/internal/realtime/server.go`                                       |
| B008 | WS    | done    | WS error envelope with `request_id`                 | `backend/internal/realtime/events.go`                                       |
| B009 | WS    | done    | Reconnect fallback documentation                    | `docs/websocket-protocol.md`                                                |
| B010 | WS    | partial | WS smoke-test hook                                  | `scripts/smoke-test.sh`                                                     |
| B011 | WS    | partial | Typing indicator WS event (no DB persistence)       | `backend/internal/realtime/events.go`                                       |
| B012 | WS    | done    | Gateway multi-instance routing ADR                  | `docs/adr/0004-ws-gateway-routing.md`                                       |
| A001 | Core  | done    | Go module + API server skeleton                     | `backend/go.mod`, `backend/cmd/api/main.go`                                 |
| A002 | Core  | done    | Config loading & env validation                     | `backend/internal/config/config.go`                                         |
| A003 | Core  | done    | Health endpoint `/health`                           | `backend/internal/api/health.go`                                            |
| A004 | Core  | done    | PostgreSQL connection pool                          | `backend/internal/db/db.go`                                                 |
| A005 | Core  | done    | DB migration tooling (golang-migrate)               | `backend/migrations/`                                                       |
| A006 | Core  | done    | Users table + repository                            | `backend/internal/user/repo.go`                                             |
| A007 | Core  | done    | Password hashing (bcrypt)                           | `backend/internal/auth/password.go`                                         |
| A008 | Core  | done    | Register API (`POST /api/auth/register`)            | `backend/internal/api/auth.go`                                              |
| A009 | Core  | done    | Login API (`POST /api/auth/login`)                  | `backend/internal/api/auth.go`                                              |
| A010 | Core  | done    | JWT middleware                                      | `backend/internal/middleware/auth.go`                                       |
| A011 | Core  | done    | Refresh token skeleton                              | `backend/internal/api/auth.go`                                              |
| A012 | Core  | done    | Devices table + repository                          | `backend/internal/device/repo.go`                                           |
| A013 | Core  | done    | Conversations table + repository                    | `backend/internal/conversation/repo.go`                                     |
| A014 | Core  | done    | Conversation members table                          | `backend/internal/conversation/members.go`                                  |
| A015 | Core  | done    | Messages table + repository                         | `backend/internal/message/repo.go`                                          |
| A016 | Core  | done    | Create DM API (`POST /api/conversations/dm`)        | `backend/internal/api/conversation.go`                                      |
| A017 | Core  | done    | Create group API (`POST /api/conversations/group`)  | `backend/internal/api/conversation.go`                                      |
| A018 | Core  | done    | Send message REST (`POST /api/conversations/:id/messages`) | `backend/internal/api/message.go`                                    |
| A019 | Core  | done    | Message history pagination API                      | `backend/internal/api/message.go`                                           |
| A020 | Core  | done    | Unified error envelope                              | `backend/internal/api/errors.go`                                            |
| A021 | Core  | done    | OpenAPI skeleton                                    | `docs/openapi.yaml`                                                         |
| A022 | Core  | done    | Seed script                                         | `scripts/seed.sh`                                                           |
| C001 | Sync  | done    | `latest_seq` on conversation                        | `backend/migrations/`, `backend/internal/conversation/repo.go`              |
| C002 | Sync  | done    | Seq allocation in transaction                       | `backend/internal/message/repo.go`                                          |
| C003 | Sync  | done    | Conversation list API with recency sort             | `backend/internal/api/conversation.go`                                      |
| C004 | Sync  | done    | `last_read_seq` per member                          | `backend/internal/conversation/members.go`                                  |
| C005 | Sync  | done    | Unread count calculation                            | `backend/internal/conversation/unread.go`                                   |
| C006 | Sync  | done    | Sync endpoint (`GET /api/sync`)                     | `backend/internal/api/sync.go`                                              |
| C007 | Sync  | done    | Per-device sync cursor (`device_sync_cursors`)      | `backend/internal/device/sync.go`                                           |
| C008 | Sync  | done    | Message edit API (`PATCH /api/messages/:id`)        | `backend/internal/api/message.go`                                           |
| C009 | Sync  | done    | Message recall API (`POST /api/messages/:id/recall`)| `backend/internal/api/message.go`                                           |
| C010 | Sync  | partial | Pinned message API                                  | `backend/internal/api/message.go`                                           |
| D001 | Rel   | done    | `client_msg_id` idempotency constraint              | `backend/migrations/`                                                       |
| D002 | Rel   | done    | Idempotency repository                              | `backend/internal/message/idempotency.go`                                   |
| D003 | Rel   | done    | Message ACK API + WS event                          | `backend/internal/api/ack.go`                                               |
| D004 | Rel   | done    | Delivered/read state machine                        | `backend/internal/delivery/state.go`                                        |
| D005 | Rel   | partial | Multi-device ACK aggregation strategy               | `docs/reliability.md`                                                       |
| D006 | Rel   | done    | Outbox table                                        | `backend/migrations/`, `backend/internal/outbox/`                           |
| D007 | Rel   | done    | Outbox publisher worker (SKIP LOCKED)               | `backend/internal/worker/outbox.go`                                         |
| D008 | Rel   | partial | Dead-letter skeleton                                | `backend/internal/worker/dlq.go`                                            |
| D009 | Rel   | partial | Reliability fault-injection tests                   | `backend/internal/reliability/`                                             |
| D010 | Rel   | done    | Reliability ADR suite                               | `docs/reliability-adr-suite.md`                                             |
| E001 | Group | done    | Group member roles owner/admin/member               | `backend/internal/conversation/roles.go`                                    |
| E002 | Group | done    | Invite/kick/leave group APIs                        | `backend/internal/api/group.go`                                             |
| E003 | Group | done    | Channel data model                                  | `backend/migrations/`                                                       |
| E004 | Group | done    | Channel subscribe/unsubscribe APIs                  | `backend/internal/api/channel.go`                                           |
| E005 | Group | done    | Channel publish permission (owner/admin only)       | `backend/internal/api/channel.go`                                           |
| E006 | Group | done    | Small-group online fanout                           | `backend/internal/realtime/server.go`                                       |
| E007 | Group | done    | Large-group fanout ADR                              | `docs/adr/0003-large-group-fanout.md`                                       |
| E008 | Group | done    | Fanout worker skeleton                              | `backend/internal/worker/fanout.go`                                         |
| E009 | Group | partial | Hot-group detection metrics                         | `backend/internal/metrics/`                                                 |
| E010 | Group | done    | Large-group k6 load-test script                     | `loadtests/k6-large-group.js`                                               |
| F001 | Infra | done    | Redis client                                        | `backend/internal/cache/redis.go`                                           |
| F002 | Infra | done    | Redis rate limiter                                  | `backend/internal/middleware/ratelimit.go`                                  |
| F003 | Infra | done    | Redis presence TTL                                  | `backend/internal/presence/redis.go`                                        |
| F004 | Infra | done    | Redis conversation list cache (30 s TTL)            | `backend/internal/cache/convlist.go`                                        |
| F005 | Infra | done    | Kafka client                                        | `backend/internal/mq/kafka.go`                                              |
| F006 | Infra | done    | `message.created` publish                           | `backend/internal/mq/producer.go`                                           |
| F007 | Infra | done    | Worker skeleton consuming `message.created`         | `backend/cmd/worker/main.go`                                                |
| F008 | Infra | done    | Idempotent consumer handler                         | `backend/internal/worker/handlers.go`                                       |
| F009 | Infra | partial | MQ lag metrics (Kafka consumer lag)                 | `backend/internal/metrics/kafka.go`                                         |
| F010 | Infra | done    | Cache consistency ADR                               | `docs/adr/0005-cache-consistency.md`                                        |
| G001 | Media | done    | MinIO presign upload URL                            | `backend/internal/api/media.go`                                             |
| G002 | Media | done    | MinIO presign download URL                          | `backend/internal/api/media.go`                                             |
| G003 | Media | done    | Attachment metadata table                           | `backend/migrations/`                                                       |
| G004 | Media | done    | Attachment message type                             | `backend/internal/message/repo.go`                                          |
| G005 | Media | done    | Presigned download URL endpoint                     | `backend/internal/api/media.go`                                             |
| G006 | Media | done    | Message indexing worker on `message.created`        | `backend/internal/worker/search_index.go`                                   |
| G007 | Media | done    | Full-text search API (`GET /api/search/messages`)   | `backend/internal/api/search.go`                                            |
| G008 | Media | done    | Search scoped to member conversations               | `backend/internal/api/search.go`                                            |
| G009 | Media | partial | OpenSearch adapter (optional upgrade)               | `backend/internal/search/opensearch.go`                                     |
| G010 | Media | partial | Notification event table skeleton                   | `backend/migrations/`                                                       |
| H001 | Sec   | done    | Redis rate limit middleware (login)                 | `backend/internal/middleware/ratelimit.go`                                  |
| H002 | Sec   | done    | Rate limit on conv send                             | `backend/internal/middleware/ratelimit.go`                                  |
| H003 | Sec   | done    | Rate limit on register                              | `backend/internal/middleware/ratelimit.go`                                  |
| H004 | Sec   | done    | Audit log table + service                           | `backend/internal/audit/`                                                   |
| H005 | Sec   | done    | Login audit logging                                 | `backend/internal/api/auth.go`                                              |
| H006 | Sec   | done    | Recall audit logging                                | `backend/internal/api/message.go`                                           |
| H007 | Sec   | partial | Input sanitisation middleware                       | `backend/internal/middleware/sanitize.go`                                   |
| H008 | Sec   | partial | Attachment virus scan mock                          | `docs/virus-scan-mock.md`                                                   |
| H009 | Sec   | partial | TLS/mTLS notes                                      | `docs/security-checklist.md`                                                |
| H010 | Sec   | done    | Security checklist                                  | `docs/security-checklist.md`                                                |
| I001 | Obs   | done    | Structured JSON logs + `X-Trace-ID` middleware      | `backend/internal/middleware/trace.go`                                      |
| I002 | Obs   | done    | Request logger middleware                           | `backend/internal/middleware/logger.go`                                     |
| I003 | Obs   | done    | Prometheus `/metrics`                               | `backend/internal/metrics/prometheus.go`                                    |
| I004 | Obs   | done    | WS connection gauge metric                          | `backend/internal/metrics/ws.go`                                            |
| I005 | Obs   | done    | Message send latency histogram                      | `backend/internal/metrics/messages.go`                                      |
| I006 | Obs   | done    | k6 API load-test script                             | `loadtests/k6-api-send.js`                                                  |
| I007 | Obs   | done    | k6 WS load-test script                              | `loadtests/k6-ws-connect.js`                                                |
| I008 | Obs   | done    | Chaos: Redis down script                            | `scripts/chaos-redis-down.sh`                                               |
| I009 | Obs   | done    | Chaos: MQ down script                               | `scripts/chaos-mq-down.sh`                                                  |
| I010 | Obs   | done    | Grafana dashboard JSON                              | `grafana/echoline-dashboard.json`                                           |
| J001 | FE    | done    | Vite + React scaffold                               | `frontend/`                                                                 |
| J002 | FE    | done    | Login page                                          | `frontend/src/pages/Login.tsx`                                              |
| J003 | FE    | done    | Conversation list sidebar                           | `frontend/src/components/ConversationList.tsx`                              |
| J004 | FE    | done    | Chat window with message history                    | `frontend/src/components/ChatWindow.tsx`                                    |
| J005 | FE    | done    | Message history pagination                          | `frontend/src/components/ChatWindow.tsx`                                    |
| J006 | FE    | done    | WS reconnect logic                                  | `frontend/src/hooks/useWebSocket.ts`                                        |
| J007 | FE    | done    | Optimistic send UI                                  | `frontend/src/components/ChatWindow.tsx`                                    |
| J008 | FE    | done    | Attachment upload UI                                | `frontend/src/components/AttachmentUpload.tsx`                              |
| J009 | FE    | done    | Search UI in sidebar                                | `frontend/src/components/SearchBar.tsx`                                     |

---

## Secondary Backlog Tasks

| ID   | Track     | Status  | Description                                  | Key File(s)                                   |
|------|-----------|---------|----------------------------------------------|-----------------------------------------------|
| SB01 | Backlog   | partial | Read receipts UI (`✓✓`)                     | `frontend/src/components/MessageStatus.tsx`   |
| SB02 | Backlog   | partial | Presence (online/offline dot in sidebar)     | `frontend/src/components/PresenceDot.tsx`     |
| SB03 | Backlog   | partial | Emoji reactions skeleton                     | `backend/internal/api/reactions.go`           |
| SB04 | Backlog   | partial | Thread/reply data model sketch               | `docs/data-model.md`                          |
| SB05 | Backlog   | partial | Push notification stub (FCM placeholder)     | `backend/internal/notification/`             |
| SB06 | Backlog   | partial | Admin panel API skeleton                     | `backend/internal/api/admin.go`               |
| SB07 | Backlog   | partial | Conversation archive/mute API                | `backend/internal/api/conversation.go`        |
| SB08 | Backlog   | partial | Dark mode CSS                                | `frontend/src/styles/theme.css`               |
| SB09 | Backlog   | partial | Message forward API skeleton                 | `backend/internal/api/message.go`             |
| SB10 | Backlog   | partial | User profile update API                      | `backend/internal/api/user.go`                |
| SB11 | Backlog   | done    | Smoke script: register/login/send/search     | `scripts/smoke-api-full.sh`                   |
| SB12 | Backlog   | done    | `.env.example` with all required vars        | `.env.example`                                |

---

## Stretch Backlog Tasks

| ID   | Track   | Status  | Description                                         | Key File(s)                                    |
|------|---------|---------|-----------------------------------------------------|------------------------------------------------|
| ST01 | Stretch | partial | Message scheduling (send_at)                        | `backend/internal/worker/scheduler.go`         |
| ST02 | Stretch | done    | Message tiering ADR (hot/warm/cold)                 | `docs/adr/0006-message-tiering.md`             |
| ST03 | Stretch | done    | DLQ replay design doc                               | `docs/dlq-replay.md`                           |
| ST04 | Stretch | partial | Conversation export API (JSON/CSV)                  | `backend/internal/api/export.go`               |
| ST05 | Stretch | done    | Conversation sharding ADR                           | `docs/adr/0007-conversation-sharding.md`       |
| ST06 | Stretch | done    | Virus scan mock design doc                          | `docs/virus-scan-mock.md`                      |
| ST07 | Stretch | done    | Chaos engineering playbook + OTel ADR               | `docs/chaos-playbook.md`, `docs/adr/0008-opentelemetry-tracing.md` |
| ST08 | Stretch | partial | Rate limit per conversation                         | `backend/internal/middleware/ratelimit.go`     |
| ST09 | Stretch | partial | Device trust score skeleton                         | `backend/internal/device/trust.go`             |
| ST10 | Stretch | partial | GraphQL subscription prototype notes                | `docs/extensions-roadmap.md`                  |

---

## Research Tasks

| ID   | Track    | Status  | Description                                      | Key File(s)                                     |
|------|----------|---------|--------------------------------------------------|-------------------------------------------------|
| RS01 | Research | done    | Telegram architecture deep-dive                  | `docs/research-telegram-whatsapp.md`            |
| RS02 | Research | done    | WhatsApp architecture comparison                 | `docs/research-telegram-whatsapp.md`            |
| RS03 | Research | done    | Fanout-on-write vs fanout-on-read analysis       | `docs/research-fanout-unread.md`                |
| RS04 | Research | done    | Unread count at scale                            | `docs/research-fanout-unread.md`                |
| RS05 | Research | done    | Kafka partition strategy for messaging           | `docs/research-kafka-sharding.md`               |
| RS06 | Research | done    | Consumer group lag and backpressure              | `docs/research-kafka-sharding.md`               |
| RS07 | Research | done    | Redpanda vs Kafka comparison                     | `docs/research-kafka-sharding.md`               |
| RS08 | Research | done    | Presence at scale design                         | `docs/research-presence-search-outbox.md`       |
| RS09 | Research | done    | Full-text search: PG tsvector vs OpenSearch      | `docs/research-presence-search-outbox.md`       |
| RS10 | Research | done    | Transactional outbox pattern deep-dive           | `docs/research-presence-search-outbox.md`       |

---

## Extension Tasks

| ID   | Track | Status  | Description                                    | Key File(s)                                   |
|------|-------|---------|------------------------------------------------|-----------------------------------------------|
| X001 | Ext   | done    | E2EE threat model ADR                          | `docs/adr/0010-e2ee-threat-model.md`          |
| X002 | Ext   | done    | E2EE key management ADR                        | `docs/adr/0011-e2ee-key-management.md`        |
| X003 | Ext   | partial | E2EE protocol stub (Signal double ratchet)     | `docs/extensions-roadmap.md`                  |
| X004 | Ext   | done    | Microservices split ADR                        | `docs/adr/0009-microservices-split.md`         |
| X005 | Ext   | partial | Service mesh notes (Istio/Linkerd)             | `docs/extensions-roadmap.md`                  |
| X006 | Ext   | partial | Recommendation engine notes                    | `docs/extensions-roadmap.md`                  |
| X007 | Ext   | done    | Ads data model ADR                             | `docs/adr/0012-ads-data-model.md`             |
| X008 | Ext   | partial | Ads targeting privacy notes                    | `docs/extensions-roadmap.md`                  |
| X009 | Ext   | done    | Payments ledger ADR                            | `docs/adr/0013-payments-ledger.md`            |
| X010 | Ext   | partial | Payments compliance notes                      | `docs/extensions-roadmap.md`                  |

---

## Knowledge / Interview / Maintenance Tasks

| ID   | Track | Status  | Description                                    | Key File(s)                                              |
|------|-------|---------|------------------------------------------------|----------------------------------------------------------|
| K001 | Know  | done    | Mobile prototype ADR                           | `docs/adr/0014-mobile-prototype.md`                      |
| K002 | Know  | partial | iOS prototype notes                            | `docs/extensions-roadmap.md`                            |
| K003 | Know  | partial | Android prototype notes                        | `docs/extensions-roadmap.md`                            |
| K004 | Know  | done    | Desktop prototype ADR (Electron/Tauri)         | `docs/adr/0015-desktop-prototype.md`                     |
| L001 | Learn | done    | Architecture doc                               | `docs/architecture.md`                                   |
| L002 | Learn | done    | Data model doc                                 | `docs/data-model.md`                                     |
| L003 | Learn | done    | WebSocket protocol doc                         | `docs/websocket-protocol.md`                             |
| L004 | Learn | done    | Scaling doc                                    | `docs/scaling.md`                                        |
| L005 | Learn | done    | Interview: system design                       | `docs/interview-system-design.md`                        |
| L006 | Learn | done    | Interview: reliability                         | `docs/interview-reliability.md`                          |
| L007 | Learn | done    | Interview: fanout                              | `docs/interview-fanout.md`                               |
| L008 | Learn | done    | Interview: multi-device sync                   | `docs/interview-multi-device-sync.md`                    |
| L009 | Learn | done    | Iteration 03 report                            | `reports/iteration-03.md`                                |
| M001 | Maint | done    | Review: API consistency                        | `reports/review-api-consistency.md`                      |
| M002 | Maint | done    | Review: DB schema                              | `reports/review-db-schema.md`                            |
| M003 | Maint | done    | Review: concurrency                            | `reports/review-concurrency.md`                          |
| M004 | Maint | done    | Review: reliability                            | `reports/review-reliability.md`                          |
| M005 | Maint | done    | Review: performance                            | `reports/review-performance.md`                          |
| M006 | Maint | done    | Review: security                               | `reports/review-security.md`                             |
| M007 | Maint | done    | Review: test coverage                          | `reports/review-test-coverage.md`                        |
| M008 | Maint | done    | Review: docs consistency                       | `reports/review-docs-consistency.md`                     |
