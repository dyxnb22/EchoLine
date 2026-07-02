# ADR 0007: Conversation-Level Sharding Strategy

## Status

Accepted (design; implementation deferred beyond current phase)

## Context

EchoLine's current architecture uses a single logical Postgres instance for all conversations. As the platform grows, a single Postgres instance becomes the vertical scaling ceiling for:

1. **Write throughput**: High-volume group conversations generate many concurrent `INSERT` + `UPDATE` operations on the same `messages` and `conversations` tables.
2. **Lock contention**: `latest_seq` update and `conversation_members` reads compete on the same rows.
3. **Backup/restore time**: A monolithic DB with hundreds of millions of messages is slow to back up and catastrophic to lose.

The fundamental question is: what is the correct unit of sharding for a messaging platform?

## Decision

Shard by **`conversation_id`** using **horizontal Postgres sharding via Citus** (or a compatible approach such as application-level shard routing).

### Sharding Key Rationale

A conversation is the natural consistency boundary:
- All messages in a conversation have `conversation_id` as a foreign key.
- Seq allocation, unread counts, and delivery state all reference `conversation_id`.
- No cross-conversation join is required for the hot read path (conversation list fetches latest message per conversation separately).

Sharding by `conversation_id` keeps all data for a single conversation on one shard node, preserving local transaction semantics for seq allocation and outbox enqueue.

### Shard Routing

- Each shard holds a contiguous range of `conversation_id` (range sharding) or a hash bucket (`conversation_id % N_shards`).
- A **shard metadata table** in a dedicated coordinator Postgres instance maps `conversation_id → shard_node`.
- The API/worker layer performs a two-phase lookup: coordinator → shard node.
- Citus automates this; in a hand-rolled approach, a `ShardRouter` component in `backend/internal/db/` handles it.

### Cross-Shard Operations

- **Conversation list**: The conversation list requires data from multiple conversations (potentially on different shards). This is served via a **scatter-gather** from all shards or via a denormalized Redis conversation list cache (already implemented in F004). Redis cache is the preferred path.
- **User-wide search**: Full-text search across all conversations spans shards. This is served via the search index (Postgres tsvector or OpenSearch), which is separate from the sharded message store.

### Migration Path

| Phase | Action |
|-------|--------|
| 0 | Current: single Postgres instance, no sharding |
| 1 | Enable Postgres declarative partitioning by `conversation_id` hash (4 partitions on same node) |
| 2 | Move partitions to separate read replicas for read scale |
| 3 | Promote to independent Postgres nodes with Citus coordinator |
| 4 | Increase shard count by re-balancing |

## Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| Shard by `user_id` | User's data is local | DMs span two users; group conversations span many |
| Shard by `conversation_id` (chosen) | Clean consistency boundary | Conversation list requires scatter-gather or cache |
| Cassandra with partition key `conversation_id` | Write-scale native | Sacrifices ACID, complex ops, no outbox transactions |
| Single DB forever | Zero complexity | Vertical limit hit at ~10M DAU |

## Implementation Files

- `backend/internal/db/shard_router.go` _(planned)_ — shard lookup and connection pool per shard
- `backend/migrations/` — partition DDL (phase 1)
- `backend/internal/cache/conversation.go` — scatter-gather mitigation via cache (already implemented)
- `docs/data-model.md` — sharding notes
- `docs/scaling.md` — update with sharding phase plan

## Consequences

**Positive:**
- Write throughput scales linearly with shard count.
- Failure isolation: one shard going down affects only its conversations.
- Each shard is independently backed up and restored.

**Negative:**
- Seq allocation must remain on a single shard (already guaranteed by `conversation_id` sharding).
- Schema migrations require coordinated rollout across all shards.
- Citus adds operational overhead; requires Citus expertise.

## Verification

- Unit test: `TestShardRouter_RoutesCorrectly` — assert `conversation_id % N_shards` maps to correct shard.
- Integration test: run two Postgres containers, route odd/even `conversation_id` to each; assert messages land on correct shard.

## Interview Talking Points

- **Why conversation_id as shard key?** "It's the smallest unit of consistency. All seq allocation, unread counts, and delivery state belong to a single conversation. Sharding on this key avoids cross-shard transactions on the hot write path."
- **Scatter-gather for conv list**: "We avoid scatter-gather on the hot path by using a Redis cache for the conversation list. The cache is invalidated on every write, so it's always fresh."
- **Migration strategy**: "We start with in-process Postgres partitioning (zero new infra), get partition pruning benefits, and migrate to Citus when vertical scaling is exhausted."
- **PACELC tradeoff**: "Within a shard, we maintain full ACID. Across shards (e.g., reading two conversations), we accept eventual consistency for the list view, bounded by the 30-second cache TTL."
