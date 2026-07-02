# Code Review: Reliability (M004)

> **Historical note (2026-07-02):** 报告中 `backend/internal/api/*` 路径为设计期命名；当前见 `sync/handler.go`、`delivery/handler.go`、`outbox/`。

**Reviewer**: Automated review via agent
**Date**: 2026-07-01
**Scope**: Outbox pattern, idempotency, delivery state machine, DLQ, sync endpoint

---

## Summary

EchoLine's reliability mechanisms implement the core patterns correctly: transactional outbox with SKIP LOCKED, `client_msg_id` idempotency, and delivery state machine. The following findings identify gaps between the design intent and current implementation completeness.

---

## Finding 1: Outbox `max_attempts` Not Enforced

**Severity**: High
**Files**: `backend/internal/worker/outbox.go`, `backend/internal/worker/dlq.go`

**Observation**: The DLQ skeleton exists but the outbox worker may not enforce a maximum attempt count before moving rows to DLQ. Without this, a permanently-failing message (e.g., malformed payload) retries indefinitely, blocking the worker's throughput.

**Recommendation**: In the outbox worker batch loop:
```go
if row.Attempts >= maxAttempts {
    moveToDLQ(ctx, row)
    continue
}
```
Default `maxAttempts = 5` with configuration via `OUTBOX_MAX_ATTEMPTS` env var. After moving to DLQ, the worker should still commit the outbox status as `failed` to prevent re-queuing.

---

## Finding 2: No Outbox Cleanup Job

**Severity**: Medium
**Files**: `backend/internal/worker/outbox.go`

**Observation**: Published outbox rows (`status = 'published'`) accumulate indefinitely. Over time, this table grows without bound and degrades the SKIP LOCKED query performance.

**Recommendation**: Add a cleanup job that runs daily:
```sql
DELETE FROM outbox
WHERE status = 'published' AND created_at < now() - interval '24 hours';
```
Can be implemented as a cron worker or a Postgres scheduled job (via `pg_cron` extension).

---

## Finding 3: Idempotency Key Scope Too Narrow

**Severity**: Medium
**Files**: `backend/internal/message/idempotency.go`

**Observation**: The idempotency check uses `(sender_id, client_msg_id)`. However, `client_msg_id` is client-generated (a UUID). A compromised or buggy client could send the same `client_msg_id` for different messages, causing message loss.

**Recommendation**: This is by design — the client is trusted to generate unique UUIDs per message. The server cannot validate uniqueness of intent without seeing the message content. Document this trust boundary in the API docs.

For additional protection: validate that `client_msg_id` is a valid UUID format before inserting (prevents malicious short strings that could cause index collisions).

---

## Finding 4: Delivery State Machine — No Constraint on Regression

**Severity**: Medium
**Files**: `backend/internal/delivery/state.go`

**Observation**: The delivery state machine uses a conditional update to prevent regression:
```sql
UPDATE message_deliveries SET status = $1
WHERE message_id = $2 AND user_id = $3 AND status < $1
```
This requires `status` to be a Postgres enum with the correct ordering (`sent < delivered < read`). If `status` is stored as TEXT (which lacks ordering), the comparison `status < $1` may not work as intended.

**Recommendation**: Verify the `status` column is a Postgres enum or add an explicit ordering check:
```sql
WHERE message_id = $2 AND user_id = $3 AND (
  (status = 'sent' AND $1 IN ('delivered', 'read')) OR
  (status = 'delivered' AND $1 = 'read')
)
```

---

## Finding 5: Sync Endpoint Does Not Include Recalled Messages

**Severity**: Medium
**Files**: `backend/internal/api/sync.go`

**Observation**: The sync endpoint returns `messages WHERE seq > ? AND status = 'active'`. Clients that reconnect after a message was recalled will not see the recall event; the message may still appear in their local cache.

**Recommendation**: Include recalled messages in sync responses with `status = 'recalled'` and `body = null`. The client should update its local store to show "This message was recalled." for any recalled messages.

**Query change**:
```sql
-- Before
WHERE conversation_id = $1 AND seq > $2 AND status = 'active'
-- After
WHERE conversation_id = $1 AND seq > $2
-- (return all, client handles status)
```

---

## Finding 6: No Circuit Breaker on Kafka Produce

**Severity**: Low (design gap)
**Files**: `backend/internal/worker/outbox.go`, `backend/internal/mq/producer.go`

**Observation**: The outbox worker retries Kafka produce indefinitely (with the attempt count). If Kafka is down for an extended period, the worker grinds through retries, filling logs and consuming CPU.

**Recommendation**: Add a circuit breaker pattern:
- After 3 consecutive Kafka failures, open the circuit (pause all Kafka produce).
- While circuit is open, stop polling outbox (let rows accumulate).
- Every 30 seconds, attempt one Kafka produce (half-open state).
- On success, close the circuit and resume.

Can use `github.com/sony/gobreaker` or a simple custom implementation.

---

## Finding 7: WS ACK Processing Not Idempotent

**Severity**: Low
**Files**: `backend/internal/api/ack.go`

**Observation**: If a client sends the same ACK event twice (e.g., network retry), the delivery state update is idempotent (`status < $1` check prevents regression). However, if the ACK creates a new delivery record on first receipt and the second receipt tries to create it again, there could be a unique constraint violation.

**Recommendation**: Use `INSERT ... ON CONFLICT DO UPDATE SET status = EXCLUDED.status WHERE ...` for the initial delivery record creation, ensuring upsert semantics.

---

## Overall Assessment

**Reliability score**: 7/10. The core reliability model (outbox, idempotency, state machine, sync) is correct. The key production gaps are: max attempts enforcement (Finding 1, HIGH), sync including recalled messages (Finding 5, MEDIUM), and delivery state ordering verification (Finding 4, MEDIUM).

## Priority Fixes

1. **Finding 1** (HIGH): Enforce `max_attempts` before DLQ.
2. **Finding 5** (MEDIUM): Include recalled messages in sync.
3. **Finding 4** (MEDIUM): Verify delivery state ordering.
4. **Finding 2** (MEDIUM): Add outbox cleanup job.
5. **Finding 6** (LOW): Add circuit breaker on Kafka produce.

## Files to Update

- `backend/internal/worker/outbox.go` — max attempts, cleanup job
- `backend/internal/delivery/state.go` — ordering verification
- `backend/internal/api/sync.go` — include recalled messages
- `backend/internal/api/ack.go` — upsert delivery record
