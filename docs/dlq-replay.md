# EchoLine Dead Letter Queue (DLQ) Replay Runbook (ST03)

This document describes the DLQ design, how entries arrive in the DLQ, and the step-by-step procedure for replaying or discarding DLQ entries.

---

## What Is the DLQ?

The **Dead Letter Queue** (`dead_letter` table in Postgres) holds outbox events that the worker failed to publish to Kafka after `OUTBOX_MAX_ATTEMPTS` (default: 5) consecutive failures.

Entries in the DLQ represent events that:
1. The message was successfully persisted to Postgres.
2. The outbox worker tried to publish the event to Kafka but failed every time.
3. Downstream consumers (search indexer, fanout worker, notification service) **have not** processed the event.

**DLQ ≠ Message Loss**: The original message is still in Postgres and accessible to users via the sync endpoint. The DLQ represents a failure of async event delivery, not of message persistence.

---

## DLQ Table Schema

```sql
CREATE TABLE dead_letter (
  id             BIGSERIAL PRIMARY KEY,
  outbox_id      BIGINT NOT NULL,          -- original outbox row id
  event_type     TEXT NOT NULL,            -- e.g., 'message.created'
  payload        JSONB NOT NULL,           -- original event payload
  error          TEXT,                     -- last error message
  attempts       INT NOT NULL DEFAULT 0,  -- how many times outbox worker tried
  created_at     TIMESTAMPTZ DEFAULT now(),
  replayed_at    TIMESTAMPTZ,              -- set when replayed
  status         TEXT NOT NULL DEFAULT 'pending'
                   CHECK (status IN ('pending', 'replaying', 'replayed', 'discarded'))
);

CREATE INDEX ON dead_letter (status, created_at);
```

---

## Alert: DLQ Entry Created

When a DLQ entry is created, the worker logs:

```json
{
  "level": "error",
  "event": "outbox_dlq",
  "outbox_id": 1234,
  "event_type": "message.created",
  "attempts": 5,
  "last_error": "kafka: leader not available",
  "message": "outbox entry moved to DLQ after max attempts"
}
```

A Prometheus metric `echoline_outbox_dlq_total` increments. If this counter is > 0, investigate immediately.

---

## Investigation Procedure

### Step 1: Identify DLQ Entries

```sql
SELECT id, event_type, error, attempts, created_at
FROM dead_letter
WHERE status = 'pending'
ORDER BY created_at ASC;
```

### Step 2: Classify the Error

| Error Pattern | Root Cause | Action |
|--------------|-----------|--------|
| `kafka: leader not available` | Kafka broker down/restarting | Wait for Kafka recovery, then replay |
| `kafka: topic does not exist` | Topic not created | Create topic, then replay |
| `dial tcp: connection refused` | Network partition | Investigate network, then replay |
| `json: cannot unmarshal` | Malformed payload (code bug) | Fix the bug, then replay or discard |
| `context deadline exceeded` | Kafka timeout | Check Kafka health, then replay |

### Step 3: Verify Kafka Is Healthy

Before replaying, confirm Kafka is accepting produce requests:

```bash
# List topics (should not error)
kafka-topics.sh --bootstrap-server localhost:9092 --list

# Check if required topic exists
kafka-topics.sh --bootstrap-server localhost:9092 --describe --topic message.created
```

### Step 4: Check Downstream State

Before replaying, check if the downstream consumers already processed the event (they might have received it through another path):

```sql
-- For a message.created DLQ entry, check if the message is in the search index
-- (If search index is ElasticSearch, check via ES API)
-- If using pg tsvector:
SELECT id, body_tsv IS NOT NULL AS indexed
FROM messages
WHERE id = '<message_id_from_dlq_payload>';
```

---

## Replay Procedure

### Option A: Automated Replay (Preferred)

If Kafka is now healthy, mark DLQ entries for replay:

```sql
-- Mark specific entries for replay
UPDATE dead_letter
SET status = 'replaying'
WHERE status = 'pending'
  AND created_at > now() - interval '24 hours';
```

The DLQ replay worker (if implemented) picks up `status = 'replaying'` entries and re-inserts them into the outbox:

```sql
INSERT INTO outbox (event_type, payload, status, attempts)
SELECT event_type, payload, 'pending', 0
FROM dead_letter
WHERE status = 'replaying';

UPDATE dead_letter SET status = 'replayed', replayed_at = now()
WHERE status = 'replaying';
```

### Option B: Manual Kafka Produce

For urgent single-event replay, produce the event directly:

```bash
# Get the payload from DLQ
psql "${DATABASE_URL}" -t -c "SELECT payload FROM dead_letter WHERE id = 42;" \
  | kafka-console-producer.sh --bootstrap-server localhost:9092 --topic message.created
```

Then mark as replayed:

```sql
UPDATE dead_letter SET status = 'replayed', replayed_at = now() WHERE id = 42;
```

### Option C: Discard (No Replay Needed)

If the event is stale (e.g., the message was deleted, the conversation no longer exists), discard:

```sql
UPDATE dead_letter SET status = 'discarded' WHERE id = 42;
```

Document the reason in `PROGRESS_LOG.md` or an incident report.

---

## Replay Idempotency

Replaying a `message.created` event is safe because all consumers are idempotent:

- **Search indexer**: Checks if the message is already indexed before inserting.
- **Fanout worker**: Checks if the event was already processed using a Redis key `processed:{event_id}` (or by checking if the message delivery records already exist).
- **Notification service**: Checks if push was already sent by looking at delivery state.

Replaying the same event twice will not cause duplicate messages visible to users.

---

## Prevention

To minimize DLQ accumulation:

1. **Monitor Kafka health** before each deployment. Do not deploy API changes when Kafka lag is > 100.
2. **Graceful worker shutdown**: The worker's context cancellation ensures in-flight batches commit before shutdown. This prevents outbox rows from accumulating unnecessarily.
3. **Exponential backoff on Kafka errors**: The worker backs off 1s → 2s → 4s → ... → 30s on consecutive Kafka errors. This prevents rapid exhaustion of attempts.
4. **Alert on Kafka lag > 1000**: Investigate before lag grows to the point of disruption.

---

## DLQ Metrics and Alerts

```
# Prometheus alert (alert fires immediately on any DLQ entry)
- alert: DLQEntryCreated
  expr: increase(echoline_outbox_dlq_total[5m]) > 0
  for: 0m
  labels:
    severity: critical
  annotations:
    summary: "DLQ entry created - event delivery failed after max attempts"
    description: "Check dead_letter table and replay after Kafka recovery"

# Alert on DLQ growing
- alert: DLQGrowing
  expr: count(dead_letter{status="pending"}) > 50
  for: 10m
  labels:
    severity: critical
  annotations:
    summary: "DLQ has > 50 unprocessed entries for > 10 minutes"
```

---

## Files Involved

- `backend/internal/worker/dlq.go` — DLQ writer skeleton
- `backend/internal/worker/outbox.go` — outbox drainer that writes to DLQ on max attempts
- `backend/migrations/` — `dead_letter` table DDL
- `backend/internal/metrics/` — `echoline_outbox_dlq_total` metric
- `docs/reliability-adr-suite.md` — reliability ADR index
- `docs/chaos-playbook.md` — CHAOS-002 (Kafka down experiment)
