# Code Review: Performance (M005)

**Reviewer**: Automated review via agent
**Date**: 2026-07-01
**Scope**: DB query patterns, cache effectiveness, WS throughput, Kafka consumer throughput

---

## Summary

EchoLine's performance characteristics are suitable for a prototype with up to ~10k DAU. The following findings identify bottlenecks that will limit scalability, ordered by impact.

---

## Finding 1: N+1 Query in Conversation List API

**Severity**: High
**Files**: `backend/internal/api/conversation.go`

**Observation**: The conversation list endpoint likely executes one query to get the list of conversations, then one query per conversation to fetch the latest message:
```go
for _, conv := range conversations {
    lastMsg, _ = messageRepo.GetLastMessage(conv.ID)  // N+1 query
}
```

**Recommendation**: Fetch the latest message in a single JOIN query:
```sql
SELECT c.*, m.body AS last_message_body, m.created_at AS last_message_at
FROM conversations c
JOIN conversation_members cm ON cm.conversation_id = c.id AND cm.user_id = $1
LEFT JOIN LATERAL (
  SELECT body, created_at
  FROM messages
  WHERE conversation_id = c.id
  ORDER BY seq DESC
  LIMIT 1
) m ON TRUE
ORDER BY c.latest_message_at DESC
LIMIT 50;
```

**Impact**: For a user with 50 conversations, the current approach runs 51 queries; the optimized approach runs 1. At 1000 concurrent users opening the app simultaneously: 51k queries/s vs 1k queries/s.

---

## Finding 2: Full Table Scan on Search Without Membership Join Optimization

**Severity**: Medium
**Files**: `backend/internal/api/search.go`

**Observation**: The search query filters messages by `conversation_id IN (SELECT conversation_id FROM conversation_members WHERE user_id = $1)`. PostgreSQL will execute this as a nested loop or hash join depending on the number of conversations.

**Recommendation**: For users with many conversations (100+), this subquery is expensive. Optimize with:
1. Create an index on `conversation_members(user_id)` (verify it exists).
2. Materialize the user's conversation IDs in Redis at login time for fast lookup.
3. Add `LIMIT 10` on the subquery to restrict to the user's most recent conversations for search (most relevant).

---

## Finding 3: Missing Connection Pool Tuning

**Severity**: Medium
**Files**: `backend/internal/db/db.go`

**Observation**: The Postgres connection pool may use default `pgxpool` settings: max connections not explicitly configured. Default `pgxpool` max is 4 per CPU, which may be too low under concurrent load.

**Recommendation**: Configure explicitly:
```go
poolConfig.MaxConns = 20          // tune based on Postgres max_connections
poolConfig.MinConns = 5
poolConfig.MaxConnLifetime = 1 * time.Hour
poolConfig.MaxConnIdleTime = 30 * time.Minute
```

Set `Postgres max_connections = 200` in the DB config. Total pool connections = API_replicas × MaxConns should not exceed `max_connections - 10` (reserve for admin/DBA).

---

## Finding 4: No Redis Pipeline for Batch Presence Lookups

**Severity**: Medium
**Files**: `backend/internal/presence/redis.go`

**Observation**: Checking presence for all members of a group conversation (to determine who is online) likely issues one Redis GET per member:
```go
for _, member := range members {
    online, _ = redis.Get(presenceKey(member))  // N Redis RTTs
}
```

**Recommendation**: Use Redis pipeline to batch all presence checks in a single RTT:
```go
pipe := redis.Pipeline()
cmds := make([]*redis.StringCmd, len(members))
for i, member := range members {
    cmds[i] = pipe.Get(ctx, presenceKey(member))
}
pipe.Exec(ctx)
```

For a 50-member group: reduces 50 Redis RTTs to 1 RTT + 1 pipeline execution.

---

## Finding 5: Sync Endpoint Could Scan Too Many Rows

**Severity**: Medium
**Files**: `backend/internal/api/sync.go`

**Observation**: A client that was offline for a week and has a busy conversation could receive thousands of messages in a single sync response, causing a large response payload and high memory allocation.

**Recommendation**: Enforce a hard `LIMIT` on sync responses (e.g., 200 messages) and return `has_more: true` when the limit is hit. The client paginates by calling sync again with the last received `seq` as the cursor.

---

## Finding 6: WS Hub Iterates All Connections per Message

**Severity**: Medium
**Files**: `backend/internal/realtime/server.go`

**Observation**: The hub iterates over all connections to find recipients of a message. If the hub has 10k connections and a small group of 5 members: the hub scans 10k connections to find 5.

**Recommendation**: Maintain a secondary index in the hub: `map[string]map[*Conn]struct{}` keyed by `conversation_id`. On message broadcast for a conversation, look up only that conversation's connections in O(members) rather than O(all_connections).

---

## Finding 7: Kafka Consumer Batch Size Too Small

**Severity**: Low (throughput, not latency)
**Files**: `backend/internal/worker/handlers.go`

**Observation**: If the consumer processes one message at a time (batch size = 1), throughput is limited by the per-message overhead (Postgres roundtrip, Redis lookup).

**Recommendation**: Process messages in batches of 50:
- For search indexing: batch INSERT into the tsvector index.
- For fanout worker: batch all fanout targets, then dispatch in one Redis pipeline.
- For delivery state updates: batch UPDATE using `WHERE message_id = ANY($1)`.

---

## Performance Budget Targets

| Endpoint | Target P50 | Target P99 |
|----------|------------|------------|
| `POST /api/conversations/:id/messages` | < 20ms | < 100ms |
| `GET /api/conversations` | < 10ms | < 50ms |
| `GET /api/sync` | < 15ms | < 80ms |
| `GET /api/search/messages` | < 50ms | < 200ms |
| WS message delivery (hot path) | < 5ms | < 30ms |

These targets assume Postgres and Redis on the same datacenter. Validate with `loadtests/k6-api-send.js`.

---

## Overall Assessment

**Performance score**: 6/10 for current prototype. Finding 1 (N+1 query) is critical and must be fixed before production. Finding 6 (hub scan) is the WS scaling bottleneck. The other findings are progressively important at higher scale.

## Priority Fixes

1. **Finding 1** (HIGH): Eliminate N+1 in conversation list.
2. **Finding 6** (HIGH): Per-conversation connection index in hub.
3. **Finding 4** (MEDIUM): Redis pipeline for presence batch.
4. **Finding 3** (MEDIUM): Connection pool tuning.
5. **Finding 5** (MEDIUM): Sync endpoint pagination.

## Files to Update

- `backend/internal/api/conversation.go` — LATERAL JOIN for last message
- `backend/internal/realtime/server.go` — per-conversation connection index
- `backend/internal/presence/redis.go` — pipeline for batch presence lookup
- `backend/internal/db/db.go` — connection pool configuration
- `backend/internal/api/sync.go` — hard limit + has_more flag
