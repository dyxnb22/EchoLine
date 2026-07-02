# ADR 0016: Reactions and Threaded Replies

## Status

Accepted (design; implementation deferred to batch-120 extension phase)

## Problem

EchoLine needs two closely related engagement features used by every major messaging platform:

1. **Reactions** — users attach an emoji to a message (👍, ❤️, 😂, etc.). Multiple users can add the same emoji; the UI shows a badge with count and the set of reactors.
2. **Threaded replies** — a reply that is logically a child of a parent message, keeping the main timeline clean while allowing sub-conversations (Slack threads, Telegram replies).

The key design questions are:

- Where to store reactions: separate table vs. denormalized JSONB on the messages row.
- How to compute reaction counts efficiently under high concurrency.
- Whether threads are first-class conversations or a property of a message.
- How fan-out of reaction/thread events works over WebSocket.

## Decision

### Reactions

Use a **separate `reactions` table** (not JSONB on messages).

```sql
CREATE TABLE reactions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id  UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    emoji       TEXT NOT NULL CHECK (char_length(emoji) <= 8),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (message_id, user_id, emoji)   -- one reaction per user per emoji per message
);

CREATE INDEX idx_reactions_message ON reactions(message_id);
```

Counts are computed with `COUNT(*) GROUP BY emoji` on demand, or via a `reaction_counts` materialized view refreshed on each insert/delete (REFRESH CONCURRENTLY).

**Why not JSONB on messages?**

| Factor | JSONB column | Separate table |
|--------|-------------|----------------|
| Concurrent updates | Requires row-level lock on message | Row insert/delete, no message lock |
| Individual query | Single row read | Index scan on message_id |
| Fan-out payload | Must re-read full JSONB | Small reaction row |
| Scalability | Limited by message row contention | Independent partition |

A separate table wins because high-volume emoji reactions (hundreds of users reacting to a viral channel post) would create hot-row contention on the message row under JSONB.

### Threaded Replies

Threads use a **`parent_msg_id` foreign key** on the `messages` table (not a separate threads table).

```sql
ALTER TABLE messages ADD COLUMN parent_msg_id UUID REFERENCES messages(id);
ALTER TABLE messages ADD COLUMN thread_count  INT NOT NULL DEFAULT 0;

CREATE INDEX idx_messages_parent ON messages(parent_msg_id)
    WHERE parent_msg_id IS NOT NULL;
```

- A root message has `parent_msg_id = NULL`.
- A reply has `parent_msg_id = <root message id>`.
- `thread_count` on the root message is updated atomically via trigger or application-level increment.
- Replies live in the same `messages` table and share the same `conversation_id` as the parent. This avoids a separate "thread conversation" model and keeps queries simple.

**Why not a separate threads/sub-conversations model (Slack approach)?**

Slack creates a sub-channel per thread, which allows unlimited nesting. This is powerful but complex. Telegram uses a simpler `reply_to_msg_id` approach. We chose the Telegram-style flat model because:

- EchoLine's primary threading use case is one level deep (reply to a message, not nested trees).
- Keeping messages in a single table simplifies sequence-ordering and sync.
- If deeper nesting is needed later, we can introduce `thread_root_id` without a migration breaking change.

### WebSocket Fan-out

New frame types added to the WS protocol:

```json
{ "type": "reaction.added",   "message_id": "...", "emoji": "👍", "user_id": "...", "count": 5 }
{ "type": "reaction.removed", "message_id": "...", "emoji": "👍", "user_id": "...", "count": 4 }
{ "type": "thread.reply",     "parent_msg_id": "...", "message": { ... }, "thread_count": 3 }
```

Fan-out follows the same hub logic as `message.new`: deliver to all online conversation members, queue offline.

## Implementation Files

- `backend/migrations/00010_reactions.sql` — `reactions` table, index
- `backend/migrations/00011_threads.sql` — `parent_msg_id`, `thread_count` columns
- `backend/internal/reaction/handler.go` — POST/DELETE /messages/:id/reactions, GET list
- `backend/internal/reaction/service.go` — add/remove logic, count query
- `backend/internal/message/handler.go` — thread reply endpoint, thread fetch
- `backend/internal/realtime/server.go` — `reaction.added/removed`, `thread.reply` dispatch
- `docs/websocket-protocol.md` — new frame types
- `docs/data-model.md` — reactions table, parent_msg_id

## Testing

- Unit: reaction add/remove/deduplicate (UNIQUE constraint), count aggregation.
- Unit: thread reply seq ordering, `thread_count` increment.
- Integration: concurrent reactions from N users → count is correct under race.
- WS smoke: reaction event delivered to online peer.

## Interview Talking Points

- **Why separate table?** "Hot-row contention on a JSONB column under hundreds of concurrent reactions. A separate table with its own index lets us insert/delete rows independently."
- **Thread model choice**: "We chose Telegram-style `parent_msg_id` over Slack's sub-conversation model. One level of threading covers 95% of use cases. If we need nesting later, `thread_root_id` is a non-breaking additive column."
- **Count consistency**: "We use `COUNT(*) GROUP BY emoji` per request for simplicity. If a viral post gets 100k reactions, we'd switch to a materialized `reaction_counts` table updated by a trigger or Kafka consumer—trading immediate consistency for throughput."
