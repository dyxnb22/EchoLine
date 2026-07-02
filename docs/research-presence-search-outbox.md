# Research: Presence at Scale, Full-Text Search, and Transactional Outbox

Reference for EchoLine design decisions on RS08–RS10.

---

## RS08: Presence at Scale

### What Is Presence?

Presence indicates whether a user is currently online (has an active WS connection). It is used for:
- The green dot in the conversation list sidebar.
- Routing decisions: skip push notification if user is online.
- "Last seen" timestamp shown in profiles.

### Naïve Approach (Single Server)

On a single server, presence is trivially tracked in the hub's in-memory connection map. A user is online if their `user_id` exists as a key in the map.

**Problem**: Does not work across multiple gateway instances. Gateway-A does not know about connections on Gateway-B.

### Distributed Presence via Redis

**EchoLine's approach:**

```
Key: presence:{user_id}:{device_id}
Type: String (value: gateway_instance_id)
TTL: 10 seconds (refreshed every 5s by heartbeat)

On connect: SET presence:{user_id}:{device_id} {gateway_id} EX 10
On heartbeat (every 5s): EXPIRE presence:{user_id}:{device_id} 10
On disconnect: DEL presence:{user_id}:{device_id}
```

**Checking if a user is online:**
```
KEYS presence:{user_id}:*  → if non-empty, user is online
```

Use `SCAN` instead of `KEYS` in production (non-blocking).

**Checking presence for a list of users** (conversation list):
```
pipeline:
  for each user_id in members:
    SCAN for keys matching presence:{user_id}:*
```

### Presence Approximation

Presence is intentionally approximate:
- TTL of 10 seconds means a user who disconnects is shown as online for up to 10 seconds.
- This matches WhatsApp and Telegram behavior (both show "online" with a multi-second lag on disconnect).
- Users find exact presence creepy (stalking concern); approximate presence is both more private and simpler to implement.

### "Last Seen" Timestamp

When a user's presence key expires (they disconnect), write `users.last_seen_at = now()` via the disconnect handler. This is a deferred write to Postgres; acceptable latency is seconds.

### Presence at Scale

At 10M online users × 2 devices each = 20M Redis keys. Each key is ~80 bytes → ~1.6 GB in Redis. Acceptable for a single Redis node (typical Redis node has 16–64 GB RAM).

At 100M online users × 2 devices = 200M keys → ~16 GB → still fits in a single large Redis node. For higher scale, shard presence across Redis Cluster by `user_id % N_shards`.

---

## RS09: Full-Text Search: PostgreSQL tsvector vs OpenSearch

### EchoLine Current State

EchoLine uses **PostgreSQL tsvector** for full-text search on `messages.body`.

```sql
ALTER TABLE messages ADD COLUMN body_tsv tsvector
  GENERATED ALWAYS AS (to_tsvector('english', coalesce(body, ''))) STORED;

CREATE INDEX ON messages USING GIN(body_tsv);

-- Search query:
SELECT * FROM messages
WHERE body_tsv @@ plainto_tsquery('english', $1)
  AND conversation_id IN (
    SELECT conversation_id FROM conversation_members WHERE user_id = $2
  )
ORDER BY created_at DESC
LIMIT 20;
```

### PostgreSQL tsvector

**Pros:**
- Zero additional infrastructure.
- Same ACID guarantees as the rest of the DB.
- GIN index is fast for typical search workloads.
- Scoped search (by conversation membership) is a standard SQL join.

**Cons:**
- Language-specific tokenization: `english` stemmer doesn't handle Chinese/Japanese/Arabic correctly.
- No relevance ranking beyond `ts_rank` (basic term frequency).
- Performance degrades on very large `messages` tables (index size becomes large).
- No fuzzy matching (typo tolerance).

**When to use**: < 10M messages per deployment, English-primary content, no multi-language requirement.

### OpenSearch / Elasticsearch

**Pros:**
- Multi-language analyzers (ICU tokenizer for CJK, Arabic, etc.).
- Relevance ranking (BM25 by default).
- Fuzzy matching, synonyms, phonetic search.
- Horizontal scaling for search index independent of message DB.
- Real-time index updates via Kafka consumer.

**Cons:**
- Additional infrastructure to operate.
- Search results are eventually consistent (index updates have latency).
- Large memory footprint per node.
- Complex query language (JSON DSL).

**When to use**: > 10M messages, multi-language requirement, fuzzy search, or dedicated search SLA.

### EchoLine Migration Path

1. **Current**: PostgreSQL tsvector (implemented).
2. **When triggered**: Table size > 100M rows OR non-English language support required.
3. **Migration**: Add OpenSearch consumer in the worker that consumes `message.created` and indexes into OpenSearch. New searches hit OpenSearch; old searches remain on Postgres (dual-read during migration). Cut over when index is fully backfilled.

### Index Architecture in OpenSearch

```json
{
  "mappings": {
    "properties": {
      "message_id": { "type": "keyword" },
      "conversation_id": { "type": "keyword" },
      "sender_id": { "type": "keyword" },
      "body": { "type": "text", "analyzer": "icu_analyzer" },
      "created_at": { "type": "date" }
    }
  }
}
```

Search query adds a `terms` filter on `conversation_id` (values from the user's conversation membership). This ensures search is scoped to authorized conversations.

---

## RS10: Transactional Outbox Pattern

### The Problem

EchoLine needs to:
1. Persist a message to Postgres (`INSERT INTO messages`).
2. Publish an event to Kafka (`message.created`).

If we do these as two independent operations:
- If Postgres succeeds but Kafka fails → message is stored but no downstream processing occurs (search not indexed, notifications not sent).
- If Kafka succeeds but Postgres fails → Kafka consumers try to process a message that doesn't exist in the DB.

This is the **dual-write problem**: we need both operations to succeed or both to fail. But Postgres and Kafka cannot participate in a single distributed transaction (2PC) without significant complexity.

### The Transactional Outbox Solution

**Core insight**: Instead of writing to Kafka directly, write to a Postgres table (`outbox`) in the same database transaction as the message insert. A separate worker reads from the outbox and publishes to Kafka.

```
Transaction:
  INSERT INTO messages (id, seq, body, ...)
  INSERT INTO outbox (id, event_type='message.created', payload='{...}', status='pending')
COMMIT

Outbox Worker (separate process, runs every 100ms):
  BEGIN
  SELECT id, payload FROM outbox
    WHERE status = 'pending'
    ORDER BY created_at ASC
    FOR UPDATE SKIP LOCKED
    LIMIT 10;
  
  For each row:
    kafka.Produce('message.created', payload)
    UPDATE outbox SET status='published'
  
  COMMIT
```

**`FOR UPDATE SKIP LOCKED`**: Each row is locked by the worker that picks it up. Multiple worker instances can run concurrently without processing the same row twice.

### Properties of the Outbox Pattern

| Property | Guarantee |
|----------|-----------|
| Atomicity | Message in DB ↔ Event in outbox (same transaction) |
| At-least-once delivery | Worker retries on Kafka failure; message is re-published |
| Idempotency | Kafka consumers must handle duplicate events (EchoLine consumers check if already processed) |
| Ordering | Events are processed in `created_at` order within a worker instance; not globally ordered across partitions |

### Outbox Table Schema

```sql
CREATE TABLE outbox (
  id           BIGSERIAL PRIMARY KEY,
  event_type   TEXT NOT NULL,
  payload      JSONB NOT NULL,
  status       TEXT NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending', 'published', 'failed')),
  attempts     INT NOT NULL DEFAULT 0,
  last_error   TEXT,
  created_at   TIMESTAMPTZ DEFAULT now(),
  updated_at   TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX ON outbox (status, created_at) WHERE status = 'pending';
```

### Dead Letter Handling

After `max_attempts` (default: 5) failed Kafka publishes:
```sql
UPDATE outbox SET status='failed', last_error='...'
WHERE id = $1
```

Failed outbox rows are candidates for the Dead Letter Queue. See `docs/dlq-replay.md` for the replay procedure.

### Performance Considerations

- Outbox polling every 100ms with `LIMIT 10` → ~100 events/s per worker. For higher throughput, increase `LIMIT` or run multiple workers.
- Long-running outbox rows (status='pending' for > 5 minutes) indicate a stuck worker; alert on this condition.
- Outbox table should be pruned of old `published` rows periodically: `DELETE FROM outbox WHERE status='published' AND created_at < now() - interval '24 hours'`.

### Alternative: Change Data Capture (CDC) with Debezium

Instead of a worker polling the outbox, use Debezium to stream Postgres WAL changes to Kafka. This eliminates the polling overhead and reduces latency to near-zero.

**Tradeoff**: Debezium requires additional infrastructure (Kafka Connect cluster) and careful WAL retention configuration. For EchoLine's current scale, the simpler polling worker is preferred.

---

## Files Involved

- `backend/internal/presence/store.go` — presence TTL management
- `backend/internal/search/handler.go` — tsvector search endpoint
- `backend/internal/outbox/publisher.go` — SKIP LOCKED outbox drainer
- `backend/internal/outbox/dlq_handler.go` — DLQ skeleton
- `backend/migrations/` — outbox table DDL
- `docs/reliability.md` — outbox reliability design
- `docs/dlq-replay.md` — DLQ replay runbook
- `docs/adr/0005-cache-consistency.md` — presence cache design
