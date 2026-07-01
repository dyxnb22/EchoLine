# EchoLine Chaos Engineering Playbook

This playbook documents planned and executed chaos experiments for EchoLine. Each experiment defines the hypothesis, blast radius, execution steps, expected behavior, and observed results.

---

## Principles

1. **Hypothesis first**: Define what you expect before you inject. Chaos without a hypothesis is just destruction.
2. **Minimum blast radius**: Start with the smallest possible failure; escalate only if the system survives.
3. **Abort criteria**: Define a stop condition before starting. If the system degrades beyond acceptable thresholds, abort and restore.
4. **Observability prerequisite**: Before running any experiment, confirm Prometheus and logs are working.
5. **Production only if staging passes**: Run every experiment in staging first.

---

## Experiment 1: Redis Down

**ID**: CHAOS-001
**Hypothesis**: When Redis becomes unavailable, the API falls back to Postgres for conversation list reads. Message send succeeds. Rate limiting fails open (no false rejects). WS connections remain alive (hub is in-process).
**Blast radius**: All Redis-dependent paths (cache, rate limit, presence).

### Execution

```bash
# scripts/chaos-redis-down.sh
docker stop echoline-redis   # or: tc qdisc add dev eth0 root netem loss 100%

# Observe for 60s
# - Send messages via API
# - Check /metrics for error rate
# - Check logs for "redis: connection refused"

docker start echoline-redis  # restore
```

### Acceptance Criteria

| Check | Pass Threshold |
|-------|---------------|
| Message send HTTP 200 | 100% |
| Conversation list HTTP 200 (from DB) | 100% |
| Login succeeds (rate limit fails open) | 100% |
| WS push works for online users | 100% |
| Logs show cache miss, not panic | Yes |
| No data loss in Postgres | Yes |

### Known Acceptable Degradations

- Rate limiting is bypassed (fail open). A malicious user could send more requests than the limit during an outage. Acceptable for short outages.
- Presence data is unavailable; push notifications may be sent to all devices (online + offline). Acceptable.

### Observed Results (staging)

_Run date_: Pending
_Duration_: 60s
_Results_: Pending

---

## Experiment 2: Kafka / MQ Down

**ID**: CHAOS-002
**Hypothesis**: When Kafka is unavailable, the outbox worker retries with exponential backoff. The API still responds 200 for message sends (message is in Postgres; outbox is pending). Search indexing is delayed but not lost. Fanout is delayed but not lost.

### Execution

```bash
# scripts/chaos-mq-down.sh
docker stop echoline-kafka

# Observe for 120s
# - Send 10 messages
# - Verify they appear in Postgres
# - Verify they appear in outbox with status='pending'
# - Observe worker retry logs

docker start echoline-kafka
# Verify outbox drains within 30s
# Verify messages appear in search index
```

### Acceptance Criteria

| Check | Pass Threshold |
|-------|---------------|
| Message send HTTP 200 | 100% |
| Messages in Postgres immediately | 100% |
| Outbox rows created | 100% |
| Worker retry logs visible | Yes |
| No DLQ entries after Kafka restores | Yes (after 5-minute window) |
| Search index up-to-date after recovery | Yes |

### Known Acceptable Degradations

- Real-time search indexing is delayed by outage duration.
- Cross-instance WS fanout is delayed (online users on other instances miss real-time push; get it on next poll/sync).

### Observed Results (staging)

_Run date_: Pending
_Duration_: 120s
_Results_: Pending

---

## Experiment 3: Postgres Read Replica Down (Future)

**ID**: CHAOS-003
**Hypothesis**: When the read replica is unavailable, reads fall back to the primary. Write performance is unaffected. Read latency increases (primary handles all reads).
**Blast radius**: Search queries, conversation list (if read replica is used for cache misses).

### Execution

```bash
docker stop echoline-postgres-replica
# Send read-heavy traffic for 60s
# Check P99 read latency (should stay < 100ms on primary)
docker start echoline-postgres-replica
```

_Note: EchoLine currently uses a single Postgres instance. This experiment is planned for the read-replica introduction phase._

---

## Experiment 4: API Pod Crash (Kubernetes)

**ID**: CHAOS-004
**Hypothesis**: A rolling API pod restart does not cause visible errors for in-flight requests (HTTP 503 is served by load balancer during shutdown; existing requests complete with a 30s graceful shutdown window).

### Execution

```bash
kubectl rollout restart deployment/echoline-api
# Monitor during restart:
# - k6 send messages at 100 RPS
# - Assert p99 < 200ms, error rate < 0.1%
```

### Acceptance Criteria

| Check | Pass Threshold |
|-------|---------------|
| HTTP 200 during restart | > 99.9% |
| No in-flight message loss | 100% |
| WS reconnections complete within 5s | Yes |

---

## Experiment 5: WS Gateway Restart

**ID**: CHAOS-005
**Hypothesis**: When the WS gateway restarts, all connected clients disconnect. Clients detect disconnect, back off, and reconnect within 10 seconds. On reconnect, clients call sync endpoint and receive any messages missed during the outage.

### Execution

```bash
# Kill gateway process
kill -9 $(pgrep echoline-api)
# Observe: clients disconnect
# Observe: clients reconnect (check WS gauge metric in Grafana)
# Send 5 messages during the 5s downtime
# After reconnect: call sync endpoint; assert all 5 messages received
```

### Acceptance Criteria

| Check | Pass Threshold |
|-------|---------------|
| Clients reconnect within 10s | Yes |
| Messages sent during outage recoverable via sync | 100% |
| No duplicate messages on reconnect | Yes |

---

## Experiment 6: Disk Full on Postgres

**ID**: CHAOS-006
**Hypothesis**: When Postgres disk fills, writes fail with an error. API returns 500. No partial writes. On disk space recovery, system resumes normally.
**Severity**: HIGH — triggers DLQ and alerts.

### Execution (staging only, never production)

```bash
# Fill disk
dd if=/dev/zero of=/var/lib/postgresql/fill_disk bs=1M count=20000

# Observe:
# - Message send returns 500
# - Outbox worker stops (DB write fails)
# - Alerts fire (disk usage > 90%)

# Clean up
rm /var/lib/postgresql/fill_disk

# Verify:
# - DB resumes writes
# - No outbox corruption
```

---

## Runbook: Restoring After a Chaos Experiment

1. Stop the experiment script or restart the stopped container.
2. Verify Prometheus metrics return to baseline within 2 minutes.
3. Check `outbox` table for pending rows; ensure worker is draining.
4. Check `dead_letter` table for any entries; replay if needed (see `docs/dlq-replay.md`).
5. Run smoke test: `RUN_API_SMOKE=1 make smoke`.
6. Record results in this playbook under "Observed Results".

---

## Scripts

- `scripts/chaos-redis-down.sh` — stop/start Redis
- `scripts/chaos-mq-down.sh` — stop/start Kafka
- `scripts/smoke-api-full.sh` — full API smoke after restoration

## Files Involved

- `scripts/chaos-redis-down.sh`
- `scripts/chaos-mq-down.sh`
- `scripts/smoke-api-full.sh`
- `docs/dlq-replay.md`
- `grafana/echoline-dashboard.json`
- `docs/adr/0008-opentelemetry-tracing.md`
