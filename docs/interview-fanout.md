# Interview Guide: EchoLine Fanout Design

This document prepares you to explain EchoLine's message fanout strategy in a system design or engineering interview.

---

## What Is Fanout?

In a messaging system, **fanout** is the process of delivering a single sent message to all intended recipients. The cost of fanout determines the system's scalability ceiling.

The fundamental tradeoff is:

| Strategy | Write cost | Read cost | Latency |
|----------|------------|-----------|---------|
| Fanout-on-write (push) | O(recipients) | O(1) | Low |
| Fanout-on-read (pull) | O(1) | O(messages × open clients) | High |
| Hybrid | Configurable | Configurable | Configurable |

---

## EchoLine's Fanout Strategy

EchoLine uses a **hybrid fanout** strategy based on group size and online status.

### Case 1: Direct Messages (2 users)

**Write-path fanout:**
1. API persists message to Postgres.
2. API looks up WS connection for recipient in the in-process hub.
3. If online on this instance: push `message.received` over WS. Done.
4. If online on another gateway instance: publish to `gw:user:{recipient_id}` Redis channel.
5. If offline: outbox worker triggers notification service (FCM/APNs).

Cost: O(1) per send. Complexity: low.

### Case 2: Small Groups (≤ 256 members)

**Synchronous hub fanout (write-path):**
1. API persists message.
2. API iterates over online members in the hub (in-process).
3. For each online member: push `message.received`.
4. For cross-instance online members: Redis Pub/Sub.
5. For offline members: outbox → notification service.

Cost: O(online_members) on the API goroutine. For a 10-member group with 5 online: 5 WS writes, each < 0.1ms. Total: < 0.5ms added to the API response time (negligible vs DB write).

**Threshold:** If group size ≤ 256 AND all online members' WS writes complete in < 5ms, do it synchronously. Otherwise, hand off to fanout worker.

### Case 3: Large Groups & Channels (> 256 members)

**Asynchronous worker fanout:**
1. API persists message.
2. API publishes `message.created` to Kafka (via outbox, or directly if latency allows).
3. API returns HTTP 200 **before** fanout completes.
4. Fanout worker consumes `message.created`:
   - Queries `conversation_members` for the channel.
   - Batches recipients into groups of 1000.
   - For each batch: publishes `gw:user:{id}` to Redis Pub/Sub for each online user.
5. Online users receive the message with up to ~500ms delay vs instant for small groups.
6. Offline users sync on next reconnect.

Cost to API: O(1) (just a Kafka publish). Cost to worker: O(members) but amortized and batched.

**Why not synchronous for large groups?** A channel with 100k subscribers at 60 messages/min = 6M WS writes/min = 100k writes/s. That would block the API goroutine for hundreds of milliseconds per message.

---

## Hot Group Problem

**Problem:** A viral channel sends 100 messages in 10 minutes. 10k members are online. The fanout worker has to push 100 × 10k = 1M WS writes.

**Solutions:**

1. **Worker horizontal scaling**: Multiple fanout workers consume from Kafka in parallel (different partition assignments). Current Kafka topic: 8 partitions → 8 parallel workers.

2. **WS gateway horizontal scaling**: With multiple gateway instances, each instance only handles its own connected clients. 10k online users across 5 gateways = 2k users/gateway/fanout = manageable.

3. **Fanout rate limiting per channel**: If a channel sends > 10 messages/second, buffer messages in the worker and batch-push (e.g., every 500ms push up to 50 messages at once). Reduces WS connection overhead.

4. **Lazy fanout for inactive clients**: Clients that haven't ACKed the last 10 messages (likely suspended tabs/apps) receive a single "you have N new messages" notification instead of N individual pushes.

---

## Unread Count Fanout

**Problem:** After fanout, each recipient's unread count must increment. For a 1000-member group, incrementing unread for all members on every message is expensive.

**Solutions considered:**

| Approach | Pros | Cons |
|----------|------|------|
| Increment `conversation_members.unread_count` per message | Always accurate | O(members) DB writes per message |
| Compute on-demand: `messages.seq - member.last_read_seq` | Zero write cost | Requires DB read on every conv list load |
| Cache computed unread in Redis (chosen) | Fast read, bounded stale | Cache invalidation complexity |

**EchoLine choice**: Compute unread count as `conversation.latest_seq - member.last_read_seq` at read time. Cache the result in the conversation list cache (30s TTL). No write on message send.

This means unread counts are eventually consistent (up to 30s stale), which is acceptable for a messaging app. Telegram uses the same approach.

---

## Cross-Instance Fanout (Multi-Gateway)

With N gateway instances, a message sent to Alice (connected to gateway-1) must reach Bob (connected to gateway-2).

**Flow:**
1. API publishes to Redis channel `gw:user:{bob_id}`.
2. Gateway-2 subscribes to `gw:user:{bob_id}` (subscribed when Bob connected).
3. Gateway-2 receives the Redis message and pushes to Bob's WS connection.

**Failure mode:** Redis is down. Redis Pub/Sub delivery fails. Bob does not receive real-time push. Bob reconnects later and calls sync endpoint. Message is in Postgres; sync delivers it.

---

## Fanout for Offline Members (Push Notifications)

When a recipient has no active WS connection:

1. Outbox worker (or fanout worker) detects recipient is offline (no Redis presence key).
2. Publishes `notification.send` event to Kafka.
3. Notification service consumes event:
   - Looks up device push token in `devices.push_token`.
   - Sends FCM (Android) or APNs (iOS) push notification.
4. Device wakes up, app opens, WS reconnects, sync endpoint delivers full message.

**Why not deliver the full message in the push notification?** Push notification payloads are size-limited (FCM: 4KB) and can be dropped by the OS. The push notification is only a "wake-up" signal. The actual message is always fetched via the sync endpoint.

---

## Files Involved

- `backend/internal/realtime/server.go` — hub, in-process fanout
- `backend/internal/push/worker.go` — async fanout worker
- `backend/internal/presence/store.go` — presence lookup for online/offline routing
- `docs/adr/0003-large-group-fanout.md` — ADR for hybrid fanout
- `docs/adr/0004-ws-gateway-routing.md` — ADR for cross-instance Redis Pub/Sub
