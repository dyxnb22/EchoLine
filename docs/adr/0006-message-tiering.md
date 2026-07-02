# ADR 0006: Message Storage Tiering (Hot / Warm / Cold)

## Status

Accepted (design; implementation deferred to stretch backlog)

## Context

EchoLine stores every message in a single Postgres `messages` table. For a real-time messaging platform, message access patterns are strongly time-skewed:

- **Hot (0–7 days)**: Actively read. Loaded on conversation open, sync on reconnect, search. Requires < 10 ms P99 read latency.
- **Warm (7–90 days)**: Occasionally read for scroll-back. Acceptable latency 50–200 ms.
- **Cold (> 90 days)**: Rarely accessed. Acceptable latency in seconds. Storage cost dominates.

Without tiering, a large Postgres table has growing index bloat, slower `seq`-range scans, and escalating storage costs. At 1B messages (1 KB avg), the table is ~1 TB. At 10B messages it becomes operationally expensive to vacuum, index, and backup.

## Decision

Implement a **three-tier architecture**:

### Tier 1 — Hot: PostgreSQL (current)

- All new messages land in Postgres as today.
- Postgres `messages` table retains the most recent 7 days (configurable via `MESSAGE_HOT_DAYS`).
- Partitioned by `created_at` month (Postgres declarative partitioning).
- Index: `(conversation_id, seq)` on each partition.

### Tier 2 — Warm: Postgres read replica or columnar extension (pg_mooncake / Citus)

- Archival job moves rows older than 7 days to a warm table/shard.
- OR: pg_partman automates partition management; older partitions are detached and stored on cheaper disk.
- Queries for scroll-back transparently hit warm partitions via partition pruning.

### Tier 3 — Cold: Object Storage (S3 / MinIO)

- Archival job exports conversation segments to compressed NDJSON in S3/MinIO, keyed by `s3://echoline-archive/{conversation_id}/{year}/{month}.ndjson.zst`.
- A "retrieve archive" API endpoint fetches from S3 and caches in Redis for 1 hour.
- Search in cold tier: deferred (could use S3 Select or a separate index).

### Transition Triggers

| Transition | Trigger | Mechanism |
|------------|---------|-----------|
| Hot → Warm | `created_at < now() - 7d` | Cron job detaches old Postgres partition |
| Warm → Cold | `created_at < now() - 90d` | Cron job exports to S3 and truncates Postgres partition |

## Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| Postgres-only, no tiering | Simple | Cost and perf degrade at scale |
| Cassandra for all messages | Write-scalable | Complex ops, no ACID for outbox |
| Time-series DB (InfluxDB) | Efficient time queries | Not suitable for arbitrary search |
| Postgres partitioning only (chosen interim) | Native, no new infra | Cold tier still in Postgres |
| Full S3 cold tier (chosen design) | Cost-optimal at scale | Retrieval latency for old messages |

## Implementation Files

- `backend/migrations/` — add Postgres declarative partition DDL
- `backend/internal/worker/archiver.go` _(planned)_ — hot-to-warm and warm-to-cold archival cron
- `backend/internal/api/message.go` — "retrieve archive" endpoint placeholder
- `backend/internal/cache/redis.go` — 1-hour cache for cold-tier retrieval results
- `docs/data-model.md` — update partitioning notes

## Consequences

**Positive:**
- Hot-tier Postgres table stays small; queries remain fast.
- Storage cost decreases as messages age out of Postgres.
- Cold-tier cost per GB is ~20× cheaper than Postgres EBS.

**Negative:**
- Partition management adds operational complexity.
- Cross-tier queries (e.g., search spanning hot+cold) require federation logic.
- Cold-tier retrieval has high latency (S3 GET + decompress).

## Verification

- Unit test: `TestArchiverWorker_MovesOldRows` — mock DB; assert rows older than threshold are exported and deleted from hot table.
- Load test: `loadtests/k6-api-send.js` with read ratio — measure P99 read latency stays < 10 ms on hot tier after archival.

## Interview Talking Points

- **Problem framing**: "Every production messaging system eventually faces the 'long tail of cold messages' cost problem. Tiering is the standard answer."
- **Why Postgres partitioning first?** "It's the simplest change with no new infrastructure. We get the benefits of partition pruning and detachment before introducing S3 complexity."
- **Why S3 for cold?** "S3 costs ~$0.023/GB/month vs ~$0.10/GB for EBS. For billions of messages, this is a 5× cost reduction."
- **Retrieval UX**: "For cold messages, we show a 'Loading older messages…' spinner and stream them from the archive endpoint. Most users never scroll this far."
