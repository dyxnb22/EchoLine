# Interview Guide: EchoLine Reliability Design

This document prepares you to explain EchoLine's reliability mechanisms in a system design or engineering interview.

---

## Core Reliability Goals

1. **No message loss**: A message that receives HTTP 200 is guaranteed to eventually be delivered or retrievable via sync.
2. **No duplicates visible to user**: Idempotent deduplication ensures a retried send does not create duplicate messages.
3. **Ordered delivery within a conversation**: Messages have monotonically increasing `seq` per conversation; clients sort by `seq`.
4. **Eventual delivery to offline clients**: Sync endpoint allows offline clients to recover all missed messages on reconnect.

---

## Mechanism 1: Persistence Before Push

**Problem**: If we push a WS event and then the server crashes before persisting to DB, the message is lost.

**Solution**: Always write to Postgres first, then push.

```
API handler:
  1. INSERT INTO messages          ← DB first
  2. INSERT INTO outbox            ← same transaction
  3. COMMIT
  4. hub.Broadcast(...)            ← WS push after commit
```

If the server crashes after step 3 but before step 4, the client never received the WS push. On reconnect, the client calls `/api/sync` and receives the persisted message. No data loss.

**Interview angle**: "The WS push is an optimization for low latency, not the delivery mechanism. Postgres is the delivery mechanism."

---

## Mechanism 2: Client Idempotency (`client_msg_id`)

**Problem**: Network timeout on message send — did the server receive it? Client retries and creates a duplicate.

**Solution**: Client generates a UUID (`client_msg_id`) before sending. Server has a unique index:

```sql
UNIQUE INDEX ON messages (sender_id, client_msg_id);
```

On retry with the same `client_msg_id`, `INSERT ... ON CONFLICT DO NOTHING RETURNING *` returns the existing message. Client receives the same message ID as if it were the first attempt.

**Interview angle**: "This is the same pattern Stripe uses for their payment idempotency keys. The client owns the deduplication key; the server enforces it."

---

## Mechanism 3: Transactional Outbox

**Problem**: Message is persisted to Postgres but Kafka publish fails. Downstream consumers never see the event. Search index, notification service, and cross-instance WS delivery all miss the message.

**Solution**: Transactional outbox pattern.

```
Transaction:
  INSERT INTO messages (...)
  INSERT INTO outbox (event_type='message.created', payload=...)
COMMIT

Outbox worker (separate process):
  SELECT ... FROM outbox WHERE status='pending' FOR UPDATE SKIP LOCKED LIMIT 10
  → For each row:
      Produce to Kafka
      UPDATE outbox SET status='published'
```

`SKIP LOCKED` ensures multiple worker instances don't process the same row. The message is guaranteed to eventually reach Kafka as long as Postgres is available.

**Interview angle**: "The outbox table is a staging area inside the same DB transaction. This converts a distributed transaction problem (DB + Kafka) into a local idempotency problem (worker retries)."

---

## Mechanism 4: Delivery State Machine

**States**: `sent` → `delivered` → `read`

Transitions only move forward (no regression from `read` to `delivered`).

```sql
UPDATE message_deliveries
SET status = $1, updated_at = now()
WHERE message_id = $2 AND user_id = $3 AND status < $1
-- Postgres enum ordering ensures no regression
```

**ACK sources**:
- WS push: server sends `message.received`; client sends `ack` WS event.
- REST fallback: `POST /api/messages/{id}/ack`.
- Read: `POST /api/conversations/{id}/read` marks all messages up to `seq` as read.

**Multi-device**: A conversation is marked read when **any** device marks it read. Per-device cursors track individual device positions for sync purposes.

---

## Mechanism 5: Dead Letter Queue (DLQ)

**Problem**: Outbox worker fails to publish a message after retries (Kafka down for extended period, malformed payload).

**Solution**: After 5 failed attempts, move to `dead_letter` table:

```sql
CREATE TABLE dead_letter (
  id           BIGSERIAL PRIMARY KEY,
  outbox_id    BIGINT NOT NULL,
  event_type   TEXT NOT NULL,
  payload      JSONB NOT NULL,
  error        TEXT,
  created_at   TIMESTAMPTZ DEFAULT now()
);
```

Operations team receives an alert. Manual replay: update `dead_letter.status = 'replay'`; a DLQ replay worker re-enqueues into outbox.

**Interview angle**: "DLQ is the circuit breaker for the outbox. It separates 'retryable' failures from 'permanently stuck' failures. Without DLQ, a single malformed message can block the entire worker indefinitely."

---

## Mechanism 6: Sync Compensation

**Problem**: Client was offline for 2 hours. WebSocket reconnects. How does it get the 47 messages it missed?

**Solution**: On reconnect, client sends per-conversation requests:

```
GET /api/sync?after_seq=1042&limit=100
```

Server returns messages with `seq > 1042` in the conversation. Client merges into local state. This is the catch-all recovery mechanism for any delivery failure.

The sync endpoint is safe to call repeatedly (idempotent reads). Clients should call it on every reconnect, not just after long absences.

---

## Failure Scenario Matrix

| Scenario | Detection | Recovery | User Experience |
|----------|-----------|---------|-----------------|
| DB write fails | HTTP 5xx | Client retries with same `client_msg_id` | Brief send failure, retry succeeds |
| DB write succeeds, WS push fails | Client misses push | Sync on reconnect | Message visible after reconnect |
| Kafka unavailable | Outbox worker logs error | Worker retries every 5s | Delayed search/notification; messages not lost |
| Redis down | Cache miss (conv list) | Fall back to Postgres | Slower response; no data loss |
| Worker crash mid-outbox | SKIP LOCKED ensures row stays locked to dead worker; lock released on crash | Surviving workers pick up on next poll | At-most-once Kafka publish within a worker run; at-least-once overall |
| Duplicate Kafka consume | `message.created` consumer is idempotent (checks if already indexed) | Duplicate event discarded | No visible effect |

---

## Testing the Reliability Mechanisms

1. **Idempotency test**: Send the same `client_msg_id` twice; assert single message in DB.
2. **Outbox test**: Disable Kafka mock; send a message; assert message in Postgres and outbox; start worker; assert outbox row transitions to `published`.
3. **DLQ test**: Set max attempts to 1; fail all Kafka publishes; assert row in `dead_letter` after 1 attempt.
4. **Sync compensation test**: Create two messages while client is "offline"; reconnect; call sync endpoint; assert both messages returned.
5. **Chaos test**: Kill DB mid-transaction; assert no partial outbox entries (transaction rollback).

---

## Files Involved

- `backend/internal/message/idempotency.go` — `client_msg_id` check
- `backend/internal/outbox/` — outbox enqueue and worker
- `backend/internal/worker/outbox.go` — SKIP LOCKED drainer
- `backend/internal/worker/dlq.go` — DLQ skeleton
- `backend/internal/delivery/state.go` — delivery state machine
- `backend/internal/sync/handler.go` — sync endpoint
- `docs/reliability.md` — reliability design reference
- `docs/dlq-replay.md` — DLQ replay runbook
- `docs/reliability-adr-suite.md` — ADR index for reliability decisions
