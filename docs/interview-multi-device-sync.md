# Interview Guide: EchoLine Multi-Device Sync Design

This document prepares you to explain EchoLine's multi-device synchronization in a system design or engineering interview.

---

## Problem Statement

Alice has a phone, a laptop, and a desktop. She sends a message from her phone. Her laptop and desktop must receive it in real-time and stay in sync. She reads a conversation on her laptop; her phone should show it as read.

The challenge:
1. **Real-time delivery**: New messages must arrive on all devices promptly.
2. **Read state sync**: Marking a conversation as read on one device should sync to others.
3. **Offline catch-up**: A device that was off for a week should catch up seamlessly on reconnect.
4. **Position independence**: Each device may have a different "last seen" position.

---

## Data Model for Multi-Device

### Device Registration

```sql
CREATE TABLE devices (
  id           UUID PRIMARY KEY,
  user_id      UUID REFERENCES users(id),
  device_name  TEXT,            -- "Alice's iPhone 15"
  push_token   TEXT,            -- FCM or APNs token
  push_platform TEXT,           -- 'fcm' | 'apns' | 'webpush'
  last_seen_at TIMESTAMPTZ,
  created_at   TIMESTAMPTZ DEFAULT now()
);
```

Each WS connection is associated with a `device_id`. JWT contains `user_id`; device association happens at WS handshake.

### Per-Device Sync Cursor

```sql
CREATE TABLE device_sync_cursors (
  device_id        UUID REFERENCES devices(id),
  conversation_id  UUID REFERENCES conversations(id),
  last_synced_seq  BIGINT NOT NULL DEFAULT 0,
  updated_at       TIMESTAMPTZ DEFAULT now(),
  PRIMARY KEY (device_id, conversation_id)
);
```

Each device independently tracks its sync position per conversation. This allows:
- Each device to independently paginate backward through history.
- The sync endpoint to return exactly the messages a device missed, regardless of what other devices have seen.

---

## Multi-Device Message Delivery

When Alice sends a message from her phone (device-A):

1. Message persisted to Postgres with `sender_device_id = device-A`.
2. Hot-path fanout: hub iterates over all of Alice's online connections.
3. For each of Alice's other devices (device-B, device-C) that are online: push `message.received`.
4. The phone that sent the message (device-A) receives an ACK confirming delivery, **not** a duplicate `message.received` echo.

**Sender echo suppression**: The sender's device is excluded from fanout push. The sender already has the message (it was optimistically added to the UI). Sending it back would cause a visible duplicate or require client-side deduplication.

---

## Read State Synchronization

### Account-Level Read State

```sql
-- In conversation_members
last_read_seq  BIGINT NOT NULL DEFAULT 0
```

When Alice reads on her laptop, `last_read_seq` is updated for `(conversation_id, user_id)`.

**Effect**: Alice's phone will see `unread_count = latest_seq - last_read_seq = 0` on next sync. The phone's conversation list shows the conversation as read.

### Read State Push (proposed — not yet implemented)

> **Current implementation:** multi-device read sync relies on REST `POST /api/conversations/{id}/read` updating `last_read_seq`; other devices observe the change on next sync/list refresh. The WS event below is a **proposed** enhancement.

When `last_read_seq` is updated, the server could push a `conversation.read` WS event to Alice's other online devices:

```json
{
  "type": "conversation.read",
  "payload": {
    "conversation_id": "abc",
    "last_read_seq": 42,
    "read_at": "2026-07-01T10:00:00Z"
  }
}
```

This allows Alice's phone to immediately update its unread badge without waiting for the next sync poll.

### Multi-Device Read ACK Aggregation

For delivery receipts visible to **Bob** (the message sender), the display shows:
- `✓` = sent (persisted in DB)
- `✓✓` = delivered to at least one of Bob's recipient devices
- `✓✓` (blue) = read by the recipient account (any device)

The "read" state is account-level, not device-level. This matches Telegram/WhatsApp behavior.

---

## Offline Catch-Up (Sync Endpoint)

When a device reconnects after being offline:

1. Device opens WS connection (re-authenticates with JWT).
2. Device calls `POST /api/sync` with per-conversation cursors (batch, not query params):

```http
POST /api/sync
Authorization: Bearer <jwt>
Content-Type: application/json

{
  "device_id": "phone-1",
  "cursors": [
    { "conversation_id": "abc", "last_seq": 42 }
  ]
}
```

Response:

```json
{
  "device_id": "phone-1",
  "conversations": [
    {
      "conversation_id": "abc",
      "messages": [...],
      "latest_seq": 89,
      "unread": 3
    }
  ]
}
```

3. Device updates its local `device_sync_cursors` to the returned `latest_seq`.
4. Repeat with updated `last_seq` if more than 200 messages per conversation (server cap).

**Why not a global sync endpoint?** Syncing all conversations at once is expensive (joins across many conversations). The per-conversation sync is O(delta) per conversation and can be parallelized.

---

## Conflict Resolution: Concurrent Edits

EchoLine does not support collaborative document editing; only message edit/recall.

**Message edit** creates a new version with an incremented `edit_version`:
```sql
UPDATE messages SET body=$1, edit_version=edit_version+1, edited_at=now()
WHERE id=$2 AND sender_id=$3
```

A WS `message.edited` event is pushed to all conversation members. Devices reconcile by replacing the stored message body with the latest `edit_version`.

If two edits arrive out of order (race): the higher `edit_version` wins. Client discards lower versions.

---

## Device Limit and Session Management

Alice is allowed up to `MAX_DEVICES` (default: 5) concurrent sessions. On device registration:
1. Check count of existing `devices` for `user_id`.
2. If at limit, reject new device registration or evict the oldest `last_seen_at` device.

When a device is evicted:
- Its WS connection is closed with code 4009 ("device limit reached").
- Its `device_sync_cursors` are deleted.
- Its `push_token` is cleared.

---

## Testing Multi-Device Sync

1. **Test: message delivery to all devices**: Connect device-A and device-B as the same user. Send from device-A. Assert device-B receives `message.received` (not device-A).
2. **Test: read sync**: Mark conversation read on device-A. Assert `last_read_seq` updated. Assert device-B receives `conversation.read` event.
3. **Test: offline catch-up**: Disconnect device-B. Send 3 messages. Reconnect device-B. Call sync endpoint. Assert 3 messages returned.
4. **Test: per-device cursor independence**: device-A synced to seq 20. device-B synced to seq 10. Send new message (seq=21). device-A calls sync after_seq=20 → 1 message. device-B calls sync after_seq=10 → 11 messages.

---

## Files Involved

- `backend/internal/device/sync.go` — device sync cursor service
- `backend/internal/sync/handler.go` — sync endpoint
- `backend/internal/realtime/server.go` — multi-device push (sender exclusion)
- `backend/internal/conversation/members.go` — `last_read_seq` update
- `backend/migrations/` — `device_sync_cursors` table
- `docs/websocket-protocol.md` — WS events（`conversation.read` 为 proposed，见本文）
