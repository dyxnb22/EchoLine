# Research: Telegram and WhatsApp Architecture Analysis

Reference for EchoLine design decisions. Sources: public engineering blogs, conference talks, and published reverse-engineering analyses.

---

## RS01: Telegram Architecture

### Overview

Telegram is an MTProto-based messaging platform with approximately 900M MAU (2024). Its architecture differs significantly from traditional HTTP-based platforms.

### MTProto Protocol

- Telegram uses a custom binary protocol (MTProto 2.0) over TCP/UDP.
- MTProto provides encryption at the transport layer for cloud chats, and E2EE for Secret Chats.
- The protocol combines symmetric encryption (AES-256-IGE) with RSA for key exchange.
- EchoLine uses standard HTTPS/WSS, avoiding MTProto's complexity. This is a deliberate tradeoff: less efficient but much easier to debug and implement.

### Storage Model

- **Cloud chats** (regular): Messages stored in Telegram's distributed database (custom-built, cluster of MongoDB-inspired shards).
- **Secret chats**: No server storage; keys on device only.
- Telegram does not use PostgreSQL at scale; they built a custom distributed storage system. EchoLine uses PostgreSQL with explicit acknowledgment that it will require sharding beyond ~10M DAU.

### Datacenter Architecture

- Telegram operates 5 datacenter clusters globally.
- Each cluster has a full copy of user data (users can connect to any DC).
- Cross-DC replication is eventually consistent; messages are strongly consistent within a DC.
- EchoLine's equivalent would be multi-region Postgres replication (not yet implemented).

### Key Differences from EchoLine

| Aspect | Telegram | EchoLine |
|--------|----------|---------|
| Transport | Custom MTProto (binary, UDP-capable) | HTTPS/WSS (standard) |
| Storage | Custom distributed DB | PostgreSQL |
| E2EE | Optional (Secret Chat only) | Planned extension |
| Scale | 900M MAU | Demo/prototype |
| Channels | Yes (unlimited subscribers) | Yes (implemented) |
| Bot API | Extensive | Not implemented |

### EchoLine Design Lessons from Telegram

1. **Conversation-centric data model**: Telegram organizes data by conversation (channel/group/DM). EchoLine follows this model with `conversation_id` as the sharding key.
2. **Hybrid fanout**: Telegram uses a combination of push (small groups) and polling (large channels). EchoLine's ADR 0003 mirrors this.
3. **Sync endpoint**: Telegram's clients sync via `getMessages` after reconnect. EchoLine's `/api/sync` serves the same purpose.
4. **Datacenter routing**: Telegram routes clients to the nearest DC. EchoLine's ADR 0004 (Redis Pub/Sub gateway routing) is the intra-region equivalent.

---

## RS02: WhatsApp Architecture

### Overview

WhatsApp serves ~2.5B users with a small engineering team (~50 engineers at acquisition). It is famous for efficiency: a 2012 blog post described ~1M simultaneous connections per server.

### XMPP Core

- WhatsApp originally used XMPP (Jabber) with custom extensions.
- Messages are delivered via a custom binary protocol over XMPP connections.
- Each message is assigned a `message_id` and ACK'd. If no ACK within a timeout, retry.
- EchoLine's WebSocket + ACK model mirrors this delivery pattern at a conceptual level.

### Erlang Backend

- WhatsApp's backend is written in Erlang/OTP.
- Erlang's lightweight processes (1-2KB each) allow millions of concurrent WS-equivalent connections per node.
- Erlang's supervision trees provide automatic process restart, analogous to Kubernetes pod restart.
- EchoLine uses Go with goroutines. Go goroutines are ~2KB each, comparable to Erlang processes.

### Signal Protocol (E2EE)

- WhatsApp adopted the Signal Protocol in 2016, providing E2EE for all chats.
- Pre-key bundles are uploaded to WhatsApp's servers; private keys never leave devices.
- EchoLine's ADR 0011 describes the equivalent design.

### Multi-Device (WhatsApp Web / Desktop)

- WhatsApp's multi-device model (introduced 2021) uses a device-to-device key exchange model.
- The phone is the "primary device"; other devices are linked via a QR code.
- Each linked device gets its own Signal identity key; the primary device distributes message keys.
- EchoLine's simpler model: each device registers independently with the server; the server distributes pre-key bundles (no primary device concept).

### Media Architecture

- WhatsApp stores media in a distributed object store (similar to S3).
- Media is encrypted client-side before upload; the server stores ciphertext.
- EchoLine uses MinIO (S3-compatible) with server-side stored plaintext media. E2EE media encryption is planned but not implemented.

### Key Differences from EchoLine

| Aspect | WhatsApp | EchoLine |
|--------|----------|---------|
| Protocol | Custom binary over XMPP | REST + WebSocket |
| Backend language | Erlang/OTP | Go |
| E2EE | Signal Protocol (all chats) | Planned extension |
| Multi-device | Phone-centric, QR-linked | All devices equal, server-mediated |
| Media | Client-side encrypted | Plaintext (demo) |
| Scale | 2.5B users | Demo/prototype |

### EchoLine Design Lessons from WhatsApp

1. **ACK-based delivery**: WhatsApp's single/double checkmark model is replicated in EchoLine's `delivery_state` (sent/delivered/read).
2. **Offline delivery via push + sync**: WhatsApp sends FCM/APNs push to wake up the app; the app then fetches messages via the protocol. EchoLine's sync endpoint serves the same role.
3. **Efficiency through connection reuse**: WhatsApp reuses a single persistent connection per device. EchoLine's WS connection is per-device and persistent.
4. **Media presigned URLs**: WhatsApp uploads media to blob storage and sends the URL (encrypted) in the message. EchoLine uses MinIO presigned URLs for the same decoupling.

---

## Comparison Summary

| Dimension | Telegram | WhatsApp | EchoLine |
|-----------|----------|---------|---------|
| Transport protocol | MTProto (custom) | Custom binary / XMPP | REST + WSS |
| Server language | C++ (clients), custom backend | Erlang/OTP | Go |
| Storage | Custom distributed | Custom distributed | PostgreSQL + Redis |
| E2EE | Optional (Secret Chat) | All chats | Planned |
| Fanout strategy | Hybrid push/pull | Push (ACK-based) | Hybrid push/pull |
| Sync mechanism | `getMessages` | App reconnect protocol | `/api/sync` endpoint |
| Open protocol | Partially (MTProto spec) | No | REST (OpenAPI) |

---

## Files Involved

- `docs/adr/0003-large-group-fanout.md` — fanout inspired by Telegram channels
- `docs/adr/0004-ws-gateway-routing.md` — multi-instance WS delivery
- `docs/adr/0010-e2ee-threat-model.md` — E2EE threat model (WhatsApp/Signal influenced)
- `docs/adr/0011-e2ee-key-management.md` — pre-key bundle design from WhatsApp
- `docs/reliability.md` — ACK delivery model
