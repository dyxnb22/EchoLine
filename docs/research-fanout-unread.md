# Research: Fanout Strategies and Unread Count at Scale

Reference for EchoLine fanout and unread count design decisions.

---

## RS03: Fanout-on-Write vs Fanout-on-Read

### Definitions

**Fanout-on-write (push model)**: When a message is sent, the system immediately "fans out" — creates a copy or notification for every recipient. Recipients read from their own "inbox", which is pre-populated.

**Fanout-on-read (pull model)**: When a message is sent, it is stored once. When recipients open a conversation, they read from the shared message log.

### Tradeoff Analysis

#### Fanout-on-Write

**Pros:**
- Read is O(1): just read from personal inbox.
- Low read latency: no joins needed on the hot read path.
- Can deliver personalized content (e.g., custom notification text per recipient).

**Cons:**
- Write amplification: for a group of 1000 members, one message send = 1000 writes.
- Storage amplification: 1000 copies of the same message stored (or 1000 pointers).
- Stale recipients: if a member leaves the group, their inbox still has the old messages.
- Hot user problem: if a "superstar" user has 1M followers and sends a message, 1M inbox writes are triggered. This creates a write throughput spike.

**Used by:** Twitter's timeline (historically), news feed systems, email.

#### Fanout-on-Read

**Pros:**
- Write is O(1): one message write regardless of group size.
- No storage amplification: messages stored once.
- No hot user problem: the write cost is flat.

**Cons:**
- Read is O(messages × conversations): to build a conversation list, must scan all conversations.
- Higher read latency: join-heavy queries.
- Cache invalidation is harder (must invalidate on any write).

**Used by:** Slack (per-channel log), Discord (per-channel log), most database-backed messaging.

#### Hybrid (EchoLine's Choice)

For messaging, messages are naturally partitioned by conversation. The inbox view is "my conversation list with unread counts", not a global timeline. This allows:

- Store messages **once** per conversation (fanout-on-read for message storage).
- Maintain per-user unread counts as `latest_seq - last_read_seq` (computed on read, cached).
- Push delivery notifications via WS (fanout-on-write for the push event, not for message storage).

This is the same approach used by Slack, Discord, and most modern IM platforms.

### EchoLine Implementation

```
Write path:
  INSERT INTO messages (single row per message)
  UPDATE conversations SET latest_seq = latest_seq + 1 (single row)
  Publish message.created to Kafka (single event)
  [async] Push message.received to online WS connections

Read path:
  GET /api/conversations → JOIN conversation_members ON user_id
                         → latest_seq - last_read_seq = unread
                         → Latest message: JOIN messages WHERE seq = latest_seq
  Cache result (30s TTL)
```

Write amplification: O(1) DB writes + O(online_members) WS pushes.
Read complexity: single query with index scans.

---

## RS04: Unread Count at Scale

### The Problem

Unread count is one of the most-read pieces of data in a messaging app. Every time a user opens the app, every conversation must show an accurate unread count. At 50M DAU with an average of 20 conversations each: 1B unread count reads per day.

### Approaches

#### Approach 1: Stored Counter (Increment on Message, Decrement on Read)

```sql
ALTER TABLE conversation_members ADD COLUMN unread_count INT DEFAULT 0;

-- On message send:
UPDATE conversation_members SET unread_count = unread_count + 1
WHERE conversation_id = $1 AND user_id != $2

-- On read:
UPDATE conversation_members SET unread_count = 0
WHERE conversation_id = $1 AND user_id = $2
```

**Pros:** O(1) read (just read the column).
**Cons:** O(members) writes per message send. For a 1000-member group: 999 UPDATE rows per message. Hot row contention if multiple messages arrive simultaneously.

**Used when:** Group size is small (< 50). WhatsApp uses this approach.

#### Approach 2: Computed on Read (`latest_seq - last_read_seq`)

```sql
-- No stored counter
-- On read:
SELECT c.latest_seq - cm.last_read_seq AS unread_count
FROM conversation_members cm
JOIN conversations c ON c.id = cm.conversation_id
WHERE cm.user_id = $1
```

**Pros:** Zero write cost for unread count. Correct by construction.
**Cons:** Read requires join; must be cached.

**Used when:** Large groups or channels. Telegram channels use this.

#### Approach 3: Redis Counter (Hybrid)

```
Key: unread:{user_id}:{conversation_id}
On message send: INCR unread:{recipient_id}:{conversation_id} (per recipient)
On read: SET unread:{user_id}:{conversation_id} 0
On sync: if key missing, compute from DB and cache

TTL: 24 hours (re-computed if expired)
```

**Pros:** O(1) reads and writes; Redis is in-memory.
**Cons:** Redis is not the source of truth; must reconcile with DB. Write amplification: O(members) Redis INCRs per message, same as Approach 1 but cheaper (Redis vs Postgres).

**Used when:** Very high read rate, high accuracy requirements.

### EchoLine Choice

**Approach 2 (computed on read) with conversation list cache.**

- No unread counter to maintain on writes.
- `latest_seq - last_read_seq` is always accurate.
- Conversation list (including unread counts) cached in Redis for 30 seconds.
- Cache invalidated on new message send (write-through).

**Tradeoff accepted:** Up to 30 seconds of staleness for background updates (another device marks a conversation read). Foreground writes (user sends or reads in this session) invalidate immediately.

### Unread Count at Scale Analysis

At 50M DAU × 20 conversations each:

- **Without cache**: 1B unread reads/day = ~12k reads/s → Postgres handles ~50k simple reads/s → OK but tight.
- **With 30s TTL cache**: 30s window means each count is read from Postgres at most once per 30s per user × 20 convs. For an active user opening the app: 1 cache miss (first load), then cache hits. Postgres load: ~1% of naive approach.

### The "Badge Count" Problem

The app icon badge count (total unread across all conversations) requires summing unread counts. Options:

1. Sum from Postgres on each load: expensive.
2. Sum from Redis unread keys: still O(conversations) Redis reads.
3. Maintain a separate `total_unread` counter in Redis: O(1) read, requires careful increment/decrement on every conversation read.

EchoLine defers badge count to a future iteration.

---

## Files Involved

- `backend/internal/conversation/unread.go` — unread count computation
- `backend/internal/cache/convlist.go` — conversation list cache
- `docs/adr/0003-large-group-fanout.md` — fanout strategy
- `docs/adr/0005-cache-consistency.md` — cache consistency for unread counts
