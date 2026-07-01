# ADR 0005: Cache Consistency Strategy

## Status

Accepted

## Context

EchoLine uses Redis as an application-level cache for two primary datasets:

1. **Conversation list** — the sidebar view of a user's recent conversations, ordered by `latest_message_at`, including unread counts (`F004`).
2. **Presence** — whether a user is currently online, used for the online indicator and to skip push-notification delivery (`F003`).

Both datasets are read on every page load and every WebSocket reconnect. Without caching, each request hits Postgres with a multi-join query touching `conversations`, `conversation_members`, `messages`, and `deliveries`. At scale this is the dominant Postgres load.

The risk is **cache inconsistency**: stale unread counts or invisible new messages confuse users and erode trust.

## Decision

### Conversation List Cache (F004)

Use a **write-through + short TTL** strategy:

- TTL: **30 seconds** per cache entry keyed by `convlist:{user_id}`.
- On cache miss: fetch from Postgres, populate cache, return.
- On write (new message, mark-read, member change): call `cache.InvalidateConvList(user_id)` synchronously in the same API handler **before** returning the response.
- Rationale: 30 s TTL bounds worst-case staleness for background updates (e.g., MQ-driven fanout events that don't go through the API handler). Explicit invalidation handles foreground writes immediately.

### Presence Cache (F003)

Use a **write-through TTL-only** strategy:

- TTL: **10 seconds** per key `presence:{user_id}:{device_id}`.
- Each WebSocket heartbeat (30 s client interval, 10 s server check) refreshes the TTL.
- On disconnect: `DEL presence:{user_id}:{device_id}`.
- No active invalidation needed — expiry is the eviction mechanism.
- Rationale: Presence is inherently approximate. A 10-second stale window is acceptable UX and is consistent with WhatsApp/Telegram behavior.

### Read-Your-Writes Guarantee

After a user sends a message, the response includes the authoritative message from Postgres, not from cache. The conversation list cache is invalidated immediately. The next GET of the conversation list will miss cache and return fresh data.

## Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| TTL-only (chosen for presence) | Zero invalidation complexity | Bounded staleness acceptable for presence |
| Write-through + invalidation (chosen for conv list) | Strong read-your-writes | Invalidation logic in every write path |
| Cache-aside only | Simple | Worst-case stale until TTL |
| Event-driven invalidation (Kafka) | Decoupled | Latency adds complexity, overkill for this scale |
| No cache | Always fresh | Unacceptable Postgres load at scale |

## Implementation Files

- `backend/internal/cache/convlist.go` — conversation list cache read/write/invalidate
- `backend/internal/cache/redis.go` — Redis client, `Get`/`Set`/`Del` wrappers
- `backend/internal/presence/redis.go` — presence TTL set/refresh/delete
- `backend/internal/api/message.go` — invalidates conv list on send/edit/recall
- `backend/internal/api/conversation.go` — invalidates conv list on member changes

## Consequences

**Positive:**
- Postgres read load reduced by ~80% for conversation list queries in active sessions.
- Presence reads never touch Postgres.
- Cache correctness for foreground writes is strong; background writes have bounded (30 s) staleness.

**Negative:**
- If cache invalidation is missed (bug, code path omission), users see stale data until TTL expires.
- Redis outage degrades to full Postgres read load but does not cause data loss.

## Verification

- Unit test: `TestConvListCache_InvalidateOnSend` — send a message, assert cache is cleared, next GET fetches from DB.
- Unit test: `TestPresenceTTL_Eviction` — set presence, mock time +11s, assert key missing.

## Interview Talking Points

- **Why 30 s TTL for conv list?** "It bounds the worst case for async paths (MQ workers) and is imperceptible in normal use. Foreground writes invalidate immediately so you always see your own changes."
- **Why not event-driven invalidation?** "The Kafka consumer could invalidate the cache, but it adds a round trip and couples cache invalidation to MQ availability. For this scale, synchronous invalidation in the handler is simpler and more reliable."
- **Cache stampede mitigation**: "At high scale we would add a distributed lock (Redis `SET NX EX`) to serialize cache rebuilds. Current scale does not require it."
