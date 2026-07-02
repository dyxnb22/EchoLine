# Research: Discord and Slack Architecture Compared to EchoLine

## Purpose

This document compares the known architectures of Discord and Slack with EchoLine's design, highlighting where EchoLine aligns, diverges, and what tradeoffs each platform made.

---

## 1. Overview

| Dimension | Discord | Slack | EchoLine |
|-----------|---------|-------|----------|
| Primary audience | Gaming communities, public servers | Workplace teams | General IM (Telegram-like) |
| Message model | Server → Channel → Message | Workspace → Channel → Message | Conversation (DM/Group/Channel) |
| Real-time protocol | Custom WS (Gateway API) | Phoenix LiveView / custom WS | Custom WS JSON hub |
| Backend language | Elixir (Elixir/Erlang OTP for Gateway), Go, Rust | Erlang, Java, Go | Go (modular monolith) |
| Database | Cassandra (messages), PostgreSQL (users/guilds) | MySQL → Vitess (sharded) | PostgreSQL |
| Search | Elasticsearch | Elasticsearch | PostgreSQL FTS → OpenSearch (planned) |
| Scale (DAU) | ~200M+ | ~38M+ | Prototype / interview |

---

## 2. Message Storage

### Discord

Discord originally used PostgreSQL, hit write throughput limits at scale, and migrated to **Cassandra** for message storage. Key design:

- Messages are partitioned by `(channel_id, bucket)` where bucket is a time window (e.g., 10 days).
- This avoids hot partitions on a single channel_id.
- Cassandra's wide-row model maps naturally: `channel_id` → sorted list of messages.

**EchoLine vs Discord**: EchoLine uses PostgreSQL for all storage. For a prototype/interview project this is correct — PostgreSQL with BRIN indexes and table partitioning can handle tens of millions of messages. At Discord-scale (billions of messages/month), Cassandra or ScyllaDB would be required.

### Slack

Slack uses **MySQL** (now Vitess for horizontal sharding). Messages are stored per-workspace, and popular workspaces are sharded.

**EchoLine vs Slack**: Same tradeoff — PostgreSQL is the right starting point. Sharding would be introduced via conversation_id-based routing (see ADR 0007).

---

## 3. Real-Time Gateway

### Discord: Gateway API

Discord runs a large fleet of Erlang/Elixir OTP processes as the "Gateway" — each process handles thousands of WebSocket connections. Key features:

- **IDENTIFY payload**: Client sends token + capabilities on connect.
- **Heartbeat/ACK**: Client sends heartbeat every `heartbeat_interval` ms; Gateway responds with `HEARTBEAT_ACK`. Missed ACKs trigger reconnect.
- **Session resumption**: Client stores `sequence` number; on reconnect sends `RESUME` with session_id + seq to replay missed events (if session not expired).
- **Sharding**: Large bots connect to multiple Gateway shards (each shard handles a subset of guilds).

**EchoLine parallel**:
- `IDENTIFY` → EchoLine sends `auth` frame with JWT on connect.
- Heartbeat/ACK → implemented in `hub.go` (ping/pong).
- Session resumption → EchoLine uses per-device sync cursor (GET /sync after reconnect) — conceptually the same as Discord's RESUME + sequence.
- Sharding → not implemented; planned via conversation_id routing.

### Slack: Phoenix / Custom WS

Slack uses Erlang for its real-time message bus (Socket Mode for bots). The Elixir Phoenix framework's Channel abstraction is a close analogue to EchoLine's hub.

---

## 4. Fan-out

### Discord

For large servers (100k+ members), Discord cannot push every message to all connected members synchronously. Their approach:

1. **Direct fan-out** for small servers (< ~1000 online members): push to all connected Gateway processes.
2. **Lazy fan-out** for large servers: members subscribe to topics (PubSub); message published once, consumers pull.

Discord published their "How Discord Scaled to 14 Million Concurrent Users" post describing their move from a single Elixir monolith to a microservices architecture with dedicated fan-out services.

**EchoLine parallel**: EchoLine's hub does direct in-process fan-out. The large-group threshold is configurable (`FANOUT_THRESHOLD=500`). Above that, messages go to the outbox → Kafka → workers that fan-out to individual connections. See ADR 0003 and ADR 0004.

### Slack

Slack uses a Redis Pub/Sub layer for inter-node fan-out. Each API server subscribes to channels relevant to its connected users. Message publish to Redis is received by all subscribed nodes.

**EchoLine parallel**: EchoLine plans a Redis Pub/Sub fan-out for multi-node deployments (currently in-process for single node). See ADR 0004.

---

## 5. Search

### Discord

Discord indexes messages in Elasticsearch partitioned by channel. Search scope is always within a single server/channel. Cross-server search is not supported (privacy).

### Slack

Slack uses Elasticsearch with per-workspace index isolation. Slack's "full workspace search" is a key differentiation from Discord.

### EchoLine

Phase 1 uses PostgreSQL full-text search (tsvector, GIN index). This is sufficient for millions of messages but not billions. The migration path is to OpenSearch (managed Elasticsearch) per ADR G009.

---

## 6. Presence

### Discord

Discord tracks presence at the user level via the Gateway. When a user's last connection closes, Discord waits ~10s before broadcasting offline status (debounce for mobile reconnects).

Presence updates are rate-limited to prevent spamming (100ms debounce per user).

### Slack

Slack shows "active" (green), "away" (yellow), "DND" (red), "offline" (grey). Presence is cached in Redis with a 5-minute TTL; explicit away/DND is stored in DB.

### EchoLine

EchoLine uses Redis TTL presence (5-minute TTL, updated on heartbeat). Same debounce principle as Discord applies. See `backend/internal/presence/`.

---

## 7. Key Differences vs EchoLine

| Feature | Discord / Slack | EchoLine |
|---------|----------------|----------|
| Message storage scale | Cassandra / Vitess | PostgreSQL (sharding planned) |
| Gateway sharding | Yes (Discord: 1 shard per ~2500 guilds) | Single hub (multi-node planned) |
| Search scope | Server/workspace isolated | Cross-user full-text (PostgreSQL FTS) |
| E2EE | Not offered (server decrypts) | Planned (Signal Protocol) |
| Self-hostable | No | Yes (docker compose) |
| Voice/video | Discord (WebRTC, dedicated media servers) | Not planned in EchoLine MVP |

---

## Interview Angle

> "Discord's key scaling inflection was moving from PostgreSQL to Cassandra for message storage when they hit write throughput limits. Their wide-row model with time-bucketed partitions is a natural fit for append-heavy message data. EchoLine uses PostgreSQL with BRIN indexes and partitioning as the right tradeoff for prototype and early scale — the same migration path is viable when needed."

> "Slack's fan-out uses Redis Pub/Sub for inter-node message delivery. EchoLine's current in-process hub fan-out is equivalent for a single node. The multi-node upgrade path is a Redis Pub/Sub layer: each node subscribes to the user IDs it's hosting, and message publish hits all nodes with connected recipients."
