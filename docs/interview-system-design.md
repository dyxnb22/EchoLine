# Interview Guide: EchoLine System Design

This document prepares you to explain EchoLine's architecture in a system design interview. Each section maps to a typical interview question.

---

## 1. Problem Framing

**Q: Design a messaging platform like Telegram.**

Start by clarifying scope:

- Scale: 50M DAU, 1B messages/day, 500k concurrent WS connections.
- Features: 1-1 DM, group chat (up to 256 members), channels (unlimited subscribers), full-text search, read receipts, multi-device.
- Non-features (out of scope in interview): E2EE (mention as extension), payments, ads.

State constraints:
- Message delivery guarantee: at-least-once with idempotent deduplication at client.
- Ordering: per-conversation monotonic seq; no global ordering required.
- Storage: messages are retained indefinitely (tiering handles cost).

---

## 2. High-Level Architecture

```
clients (web, mobile, desktop)
       │
       ├── REST/HTTP ──► API Service (Go, Gin)
       │                      │
       └── WebSocket ──► WS Gateway (Go gorilla/websocket)
                               │
          ┌────────────────────┼────────────────────────┐
          ▼                    ▼                         ▼
     PostgreSQL            Redis                    Kafka / Redpanda
  (source of truth)   (cache + presence)         (async events / fanout)
          │                                             │
          └──── MinIO / S3 (media)     Worker Service ──┘
```

Key decisions:
- **PostgreSQL is the source of truth.** WebSocket push is best-effort; clients recover via sync.
- **Redis** is used for conversation list cache, presence, and rate limiting.
- **Kafka** decouples message persistence from fanout, search indexing, and notifications.
- **Modular monolith** today; extract WS gateway and notification service first when scaling.

---

## 3. Data Model

### Core Tables

```
users(id, username, password_hash, created_at)
devices(id, user_id, push_token, last_seen_at)
conversations(id, type[dm|group|channel], latest_seq, latest_message_at)
conversation_members(conversation_id, user_id, role, last_read_seq)
messages(id, conversation_id, sender_id, seq, client_msg_id, body, type, created_at)
message_deliveries(message_id, user_id, status[sent|delivered|read])
outbox(id, event_type, payload, status, created_at, attempts)
```

Key design choices:
- `seq` is per-conversation (not global). Allocated in a transaction: `UPDATE conversations SET latest_seq = latest_seq + 1 RETURNING latest_seq`.
- `client_msg_id` has a unique index on `(sender_id, client_msg_id)` for idempotency.
- `outbox` table enables transactional message publish: message insert and outbox enqueue in one transaction, worker drains outbox with `SKIP LOCKED`.

---

## 4. Message Send Flow (Critical Path)

**Step-by-step for `POST /api/conversations/{id}/messages`:**

1. JWT auth middleware validates token, extracts `user_id`.
2. Rate limit check: Redis `INCR` with 1-minute window.
3. Membership check: verify sender is a member of `conversation_id`.
4. Idempotency check: `SELECT id FROM messages WHERE sender_id = ? AND client_msg_id = ?`. If found, return existing message.
5. Transaction:
   - `UPDATE conversations SET latest_seq = latest_seq + 1 RETURNING latest_seq` → assign `seq`.
   - `INSERT INTO messages (seq, body, ...)` → persist message.
   - `INSERT INTO outbox (event_type='message.created', payload=...)` → enqueue event.
6. Commit transaction.
7. **Hot-path fanout**: for online recipients on this gateway instance, push `message.received` event over WS directly (sub-millisecond, no Kafka wait).
8. Return HTTP 200 with message payload.
9. **Async**: outbox worker picks up the event, publishes to Kafka, which drives search indexing, notification service, cross-instance WS delivery.

**Why DB before WS push?** If we push WS before persisting, a crash loses the message. Persist first; push is best-effort.

---

## 5. WebSocket Protocol

Connection lifecycle:
1. Client `GET /ws?token=<jwt>` → server upgrades to WS, registers connection in hub.
2. Server starts ping goroutine (30s interval); client responds with pong.
3. On message receive: WS event `{ "type": "message.received", "payload": {...}, "server_seq": 123 }`.
4. Client sends `{ "type": "ack", "message_id": "..." }` → server records delivery state.
5. On disconnect: hub unregisters; presence TTL expires in 10s.

Reconnection:
1. Client re-authenticates via REST to get fresh JWT (or uses refresh token).
2. Client reconnects WS.
3. Client sends `GET /api/sync?after_seq={last_known_seq}` per conversation to pull missed messages.
4. Server returns delta; client merges into local state.

**Why not rely on WS for delivery guarantees?** WS is a TCP stream; any node restart loses the connection. The sync endpoint is the safety net.

---

## 6. Fanout Strategy

**Small groups (<= 256 members):**
- Synchronous hub fanout on the API goroutine.
- O(online_members) per message send.
- For 10-member group: ~10 WS writes in < 1ms.

**Large groups (> 256 members, channels):**
- API goroutine publishes `message.created` to Kafka and returns immediately.
- Fanout worker consumes from Kafka, looks up subscriber list, batches WS pushes.
- Online members on this gateway instance are still pushed synchronously; only cross-instance delivery is async.

**Channels with 100k subscribers:**
- Push to online members via Redis Pub/Sub (broadcast to all gateway instances).
- Offline members: when they reconnect, sync endpoint delivers missed messages.
- Pull fanout for reads: client requests channel feed; server queries messages table.

---

## 7. Reliability Mechanisms

| Failure | Mechanism |
|---------|-----------|
| Duplicate send | `client_msg_id` unique index; return existing message on conflict |
| WS push failure | Sync endpoint compensates on reconnect |
| Kafka publish failure | Outbox table; worker retries with exponential backoff |
| Worker crash | `FOR UPDATE SKIP LOCKED` prevents double-processing; worker restarts pick up unprocessed events |
| DLQ | After 5 failed attempts, move to `dead_letter` table; alert and manual replay |
| DB write failure | API returns 5xx; client retries (with same `client_msg_id` for idempotency) |

---

## 8. Scaling Milestones

| DAU | Changes |
|-----|---------|
| 10k | Monolith, single Postgres, single Redis |
| 100k | Add read replica for search, Redis cluster |
| 500k | Extract WS gateway, Redis Pub/Sub for cross-instance delivery |
| 2M | Extract notification service, add Kafka partitions |
| 10M | Shard Postgres by `conversation_id`, message tiering (hot/warm/cold) |
| 50M | Separate clusters per region, cross-region sync |

---

## 9. Common Interview Follow-Ups

**Q: How do you handle message ordering?**
Per-conversation `seq` allocated in a transaction. Clients sort by `seq`. No global ordering needed — messages only need to be ordered within a conversation.

**Q: How do you handle a user with 1000 devices?**
Multi-device ACK aggregation: `conversation.read_at` is updated when any device marks read. Per-device sync cursors allow each device to independently track its position.

**Q: How does search work?**
PostgreSQL `tsvector` full-text index on `messages.body`. Scoped to conversations the user is a member of. For larger scale, export to OpenSearch/Elasticsearch via the `message.created` Kafka event and a search indexing worker.

**Q: How do you prevent spam?**
Redis rate limiting per user per endpoint (login: 5/min, register: 3/min, send: 60/min). Account-level bans stored in Postgres. Content moderation via Kafka consumer with ML model (deferred).

**Q: What's the delivery guarantee?**
At-least-once. The outbox worker retries on failure. The client deduplicates by `message_id`. The sync endpoint provides eventual delivery for offline clients.

---

## Files Involved

- Per-domain REST handlers under `backend/internal/*/handler.go` (e.g. `message`, `conversation`, `sync`)
- `backend/internal/realtime/` — WS gateway
- `backend/internal/worker/` — outbox drainer, fanout worker
- `backend/internal/mq/` — Kafka producer/consumer
- `backend/internal/cache/` — Redis cache
- `backend/internal/presence/` — Redis presence
- `docs/architecture.md` — architecture overview
- `docs/data-model.md` — full schema reference
- `docs/websocket-protocol.md` — WS event reference
- `docs/reliability.md` — reliability mechanisms
