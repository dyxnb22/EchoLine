# EchoLine Reliability ADR Suite (D010)

This document serves as the index and synthesis of all reliability-related architectural decisions in EchoLine. It can be used as a single reference for system reliability review, incident post-mortems, and interview preparation.

---

## Reliability Philosophy

EchoLine's reliability model is built on four principles:

1. **Persistence before push**: A message is only considered "sent" when it is durable in Postgres. WebSocket push is an optimization for low latency, not the delivery mechanism.
2. **Idempotency by design**: Every mutation in the system is keyed by a client-provided or system-generated idempotency key. Retries are always safe.
3. **At-least-once with deduplication**: The system delivers every message at least once. The client deduplicates using `message_id` and `seq`.
4. **Sync compensation**: Any delivery failure (WS push failure, Kafka failure, client offline) is compensated by the sync endpoint, which acts as the final delivery mechanism.

---

## ADR Index by Reliability Concern

### Message Durability

| ADR | Decision | Reliability Guarantee |
|-----|----------|----------------------|
| (core design) | Postgres INSERT before WS push | Message is durable before client receives any confirmation |
| D001-D002 | `client_msg_id` unique index on `(sender_id, client_msg_id)` | Network retries never create duplicate messages |
| D006-D007 | Transactional outbox: message + event in same DB transaction | Event delivery (search, notification, fanout) always eventually happens |

**Key invariant**: If the API returns HTTP 200, the message is in Postgres. If the API returns 5xx, it is not (no partial writes).

### Event Delivery (Outbox)

| ADR | Decision | Reliability Guarantee |
|-----|----------|----------------------|
| (research RS10) | Transactional outbox pattern | Kafka events are exactly correlated with message inserts |
| D007 | SKIP LOCKED outbox drainer | Multiple worker instances never process the same event |
| D008 | DLQ after `max_attempts` | Permanently-failing events do not block the outbox forever |
| (docs/dlq-replay.md) | Manual DLQ replay procedure | Failed events can be recovered by operations |

**Key invariant**: Every `INSERT INTO messages` is accompanied by `INSERT INTO outbox`. The worker eventually delivers the outbox event to Kafka, or moves it to DLQ with an alert.

### Message Ordering

| ADR | Decision | Reliability Guarantee |
|-----|----------|----------------------|
| C001-C002 | Per-conversation `latest_seq` incremented in transaction | Messages are globally ordered within a conversation |
| (Kafka) | Partition by `conversation_id` | All events for a conversation are processed in order by the consumer |

**Key invariant**: For any two messages A and B in the same conversation, `A.seq < B.seq` if and only if A was committed before B. No two messages share the same seq.

### Delivery Status

| ADR | Decision | Reliability Guarantee |
|-----|----------|----------------------|
| D003-D004 | Delivery state machine: `sent → delivered → read` | State only moves forward; no regression |
| D005 | Account-level read (any device marks read) | Read state is consistent across devices for the sender's receipts |
| C007 | Per-device sync cursors | Each device independently tracks its position; no device is left behind |

**Key invariant**: A message's delivery status for a user never regresses (once read, always read).

### Sync Compensation

| ADR | Decision | Reliability Guarantee |
|-----|----------|----------------------|
| C006 | `/api/sync` endpoint accepts `after_seq` cursor | Any offline client can recover all missed messages |
| C007 | `device_sync_cursors` per (device, conversation) | Each device can independently recover from its last known position |
| (reliability.md) | WS reconnect triggers sync | Reconnecting clients always call sync before resuming normal operation |

**Key invariant**: If a client calls sync with the correct `after_seq`, it will receive all messages it missed, regardless of how they were delivered (hot-path push, fanout worker, or direct DB write by admin tools).

### Cross-Instance Delivery

| ADR | Decision | Reliability Guarantee |
|-----|----------|----------------------|
| ADR 0004 | Redis Pub/Sub for cross-gateway WS delivery | Messages reach connected clients on any gateway instance |
| (reliability) | Redis Pub/Sub failure degrades to sync | Redis outage degrades to delayed delivery (sync on reconnect), not message loss |

**Key invariant**: A Redis Pub/Sub failure causes a WS push miss, not a message loss. The message is in Postgres; sync compensates.

### Cache Consistency

| ADR | Decision | Reliability Guarantee |
|-----|----------|----------------------|
| ADR 0005 | Write-through invalidation for conv list | Foreground writes (user-initiated) are always reflected immediately in the next read |
| ADR 0005 | 30s TTL for background updates | Background async writes (e.g., fanout) are visible within 30 seconds |

---

## Failure Mode Analysis

| Component Failure | Impact | Recovery | Data Loss? |
|------------------|--------|---------|-----------|
| API process crash | In-flight requests fail (5xx) | Client retries with `client_msg_id` | No (transaction rollback) |
| Postgres primary failure | All writes fail | Promote read replica; replay WAL | Possible if WAL replication is async and lagging |
| Redis failure | Cache miss, rate limit bypassed, WS cross-instance delivery delayed | Cache falls through to Postgres; WS uses sync on reconnect | No |
| Kafka / MQ failure | Events accumulate in outbox; search/notification delayed | Outbox worker retries when Kafka recovers | No (outbox holds events) |
| Worker crash mid-outbox | Locked rows released on crash; reprocessed by next worker | SKIP LOCKED ensures no double-processing if crash is clean | No |
| WS gateway crash | All WS connections dropped | Clients reconnect + sync | No |
| MinIO / S3 failure | Media upload/download fails | User retries; presign URLs re-generated | No (metadata in Postgres) |

---

## Reliability Metrics (Prometheus)

| Metric | Alert Threshold | What It Means |
|--------|----------------|---------------|
| `echoline_outbox_pending_total` | > 1000 for > 5 min | Outbox worker is stuck or Kafka is down |
| `echoline_outbox_dlq_total` | > 0 | A message permanently failed to publish |
| `echoline_http_requests_total{status="5xx"}` | Rate > 0.1% | API error rate spike |
| `echoline_kafka_consumer_lag` | > 1000 | Consumer is falling behind |
| `echoline_ws_connections_active` | Drop by > 50% in 1 min | Gateway crash or network partition |
| `echoline_message_send_duration_ms` p99 | > 500ms | Message send P99 latency spike |

---

## ADR Cross-Reference

| Reliability Decision | ADR |
|---------------------|-----|
| Message ordering (seq) | Core design, C001-C002 |
| Idempotency (`client_msg_id`) | Core design, D001 |
| Transactional outbox | D006-D007 |
| DLQ | D008, docs/dlq-replay.md |
| Delivery state machine | D003-D004 |
| Multi-device ACK | D005, C007 |
| Sync compensation | C006, C007 |
| Cross-instance WS delivery | ADR 0004 |
| Cache consistency | ADR 0005 |
| At-least-once Kafka delivery | RS10, D007-D008 |
| Chaos engineering | docs/chaos-playbook.md |

---

## Files Involved

- `backend/internal/message/idempotency.go`
- `backend/internal/outbox/`
- `backend/internal/worker/outbox.go`
- `backend/internal/worker/dlq.go`
- `backend/internal/delivery/state.go`
- `backend/internal/sync/handler.go`
- `backend/internal/device/sync.go`
- `docs/reliability.md`
- `docs/dlq-replay.md`
- `docs/chaos-playbook.md`
- `docs/adr/0004-ws-gateway-routing.md`
- `docs/adr/0005-cache-consistency.md`
