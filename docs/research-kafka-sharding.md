# Research: Kafka Partition Strategy, Consumer Lag, and Redpanda Comparison

Reference for EchoLine message queue design decisions.

---

## RS05: Kafka Partition Strategy for Messaging

### Why Kafka?

EchoLine uses Kafka (or Redpanda) as the event backbone for:
1. `message.created` → search indexer, fanout worker, notification service
2. `delivery.ack` → delivery state updates
3. `media.uploaded` → virus scanner, thumbnail generator
4. `user.event` → audit log writer

Kafka provides:
- **Durability**: Messages are written to disk and replicated across brokers.
- **Replay**: Consumers can seek backward to reprocess events.
- **Fan-out by consumer group**: Multiple consumer groups can independently consume the same topic (search indexer and notification service both consume `message.created` independently).
- **Backpressure**: Consumers control their own pace; a slow consumer does not block producers.

### Partition Key Selection

Kafka partitions determine parallelism and ordering. A single partition is strictly ordered; messages across partitions can arrive out of order.

**For `message.created`:**

| Partition Key | Pros | Cons |
|--------------|------|------|
| `message_id` (random) | Even distribution | No ordering; fanout worker processes same conversation out of order |
| `conversation_id` (chosen) | All messages for a conversation go to same partition → ordered processing | Hot partitions for viral conversations |
| `sender_id` | Avoids hot conversations | Cross-conversation mix on same partition |
| `user_id` of recipient | Easy per-recipient processing | Each message repeated per recipient (fanout before Kafka) |

**EchoLine choice: partition by `conversation_id`.**

Reasoning:
- The fanout worker must process messages in conversation order (seq ordering).
- The search indexer benefits from processing a conversation's messages sequentially.
- Hot conversation problem is mitigated by multiple workers per partition and async processing.

### Partition Count

Start with **8 partitions** for `message.created`.

At 1000 messages/second with 8 partitions and 8 worker replicas: each worker handles ~125 messages/second. This is well within a single Kafka consumer's capacity (typical Go Kafka consumer handles ~10k msgs/sec).

Scale by increasing partition count when consumer lag exceeds 30 seconds.

### Topic Configuration

```
message.created:
  partitions: 8
  replication-factor: 3
  retention.ms: 604800000  # 7 days
  cleanup.policy: delete

delivery.ack:
  partitions: 4
  replication-factor: 3
  retention.ms: 86400000   # 1 day (delivery acks are short-lived)
```

---

## RS06: Consumer Group Lag and Backpressure

### Measuring Consumer Lag

**Consumer lag** = `latest_offset - committed_offset` per partition. It measures how far behind a consumer is from the latest message.

Acceptable lag levels:
- **0–100 messages**: Consumer is healthy.
- **100–1000**: Consumer is slightly behind; may indicate a processing spike.
- **> 1000**: Consumer is falling behind; check CPU/DB load.
- **Growing unboundedly**: Critical; consumer is stuck or too slow.

EchoLine exposes consumer lag via Prometheus metric `echoline_kafka_consumer_lag{topic,partition,consumer_group}`.

### Backpressure Strategy

When consumer lag grows, EchoLine's worker applies backpressure by:
1. **Batch processing**: Process 50 messages per batch (instead of 1) to amortize DB and network overhead.
2. **Parallel batch workers**: Each partition's batch is processed in a separate goroutine.
3. **Circuit breaker**: If DB is unavailable, the worker pauses consumption (does not commit offset) until DB recovers. Kafka lag grows but no messages are lost.
4. **DLQ**: After 5 failed attempts per message, move to dead letter queue and commit the offset. This unblocks the consumer at the cost of manual DLQ inspection.

### Consumer Offset Management

EchoLine uses **manual offset commit** (not auto-commit):
- After processing a batch successfully, commit the offset.
- If processing fails, do not commit; Kafka will re-deliver on next poll.
- At-least-once delivery semantics; consumer handlers are idempotent.

**Why not auto-commit?** Auto-commit commits offsets every 5 seconds regardless of processing success. A crash between auto-commit and successful processing loses messages.

### Kafka Consumer Lag Alerting

```yaml
# Prometheus alert rule (planned)
- alert: KafkaConsumerLagHigh
  expr: echoline_kafka_consumer_lag{topic="message.created"} > 1000
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Kafka consumer {{ $labels.consumer_group }} is lagging on {{ $labels.topic }}"
```

---

## RS07: Redpanda vs Kafka Comparison

### Why Redpanda?

Redpanda is a Kafka-API-compatible message queue written in C++ using Seastar (an async framework with bypass of the Linux kernel for I/O). It positions itself as a drop-in replacement for Kafka with:
- No ZooKeeper / KRaft dependency (simpler ops).
- Lower latency tails (C++ vs JVM GC pauses).
- Lower memory usage.

### Comparison

| Dimension | Kafka | Redpanda |
|-----------|-------|---------|
| Language | Java (JVM) | C++ (Seastar) |
| ZooKeeper | Removed in KRaft (3.x+) | Never needed |
| Latency P99 | ~5–20ms (JVM GC) | ~1–5ms |
| Throughput | ~1M msgs/s per node | ~1M msgs/s per node |
| Kafka API compatibility | Native | Yes (API-compatible) |
| Ecosystem (connectors) | Extensive (Kafka Connect) | Growing (Kafka Connect compatible) |
| Operational complexity | Medium (KRaft removes ZK) | Low |
| Storage tiering | Tiered storage (3.x+) | Built-in |
| License | Apache 2.0 | BSL (Redpanda Server); Apache 2.0 (client) |

### EchoLine Choice

Use **Redpanda in development** (simpler single-node setup, no ZooKeeper), with Kafka as the production reference implementation.

- `docker-compose.yml` uses Redpanda for local development.
- Production Kafka cluster is referenced in `docs/architecture.md` and `NEXT_ACTIONS.md`.
- Redpanda's Kafka API compatibility means no code changes are needed to switch.

### When to Choose Kafka Over Redpanda

1. **Kafka Connect ecosystem**: If you need specific Kafka connectors (Debezium, S3 sink), Kafka's ecosystem is larger.
2. **Regulatory compliance**: Some certifications/audits reference Kafka specifically.
3. **Team expertise**: Kafka expertise is more common in the market.

### When to Choose Redpanda Over Kafka

1. **Latency-sensitive**: Redpanda's C++ JVM-bypass gives lower P99 latency.
2. **Operational simplicity**: No ZooKeeper, fewer services to manage.
3. **Resource-constrained**: Redpanda uses less memory per node.

---

## Files Involved

- `backend/internal/eventbus/kafka.go` — Kafka/Redpanda client
- `backend/internal/outbox/publisher.go` — message producer with partition key
- `backend/internal/worker/handlers.go` — consumer handler (idempotent)
- `backend/internal/metrics/kafka.go` — consumer lag metrics
- `docker-compose.yml` — Redpanda container
- `docs/architecture.md` — MQ architecture overview
