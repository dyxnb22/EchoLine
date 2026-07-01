# ADR 0017: Push Notification Gateway (APNs + FCM)

## Status

Accepted (design; implementation deferred to extension phase)

## Problem

EchoLine delivers messages over WebSocket to connected clients. When a user's device is offline or the app is in the background, the WS connection is closed or suspended. The user should still receive a push notification so they know a message arrived.

Push notifications require integration with two platform-specific gateways:

- **Apple Push Notification service (APNs)** — HTTP/2 + JWT-based auth (p8 key)
- **Firebase Cloud Messaging (FCM HTTP v2 API)** — OAuth2 service account + HTTP POST

The questions are:

- How do devices register their push tokens with EchoLine?
- How does the backend decide when to send a push (not every message — only when offline)?
- How do we handle push delivery failures and token expiry?
- Do we build this inline in the API server or as a separate worker?

## Decision

### Device Token Registration

Devices register their push token after login:

```
POST /devices/push-token
{ "platform": "ios" | "android" | "web",
  "push_token": "...",
  "device_id": "..." }
```

Tokens are stored in the `devices` table alongside the existing device sync cursor:

```sql
ALTER TABLE devices ADD COLUMN push_token  TEXT;
ALTER TABLE devices ADD COLUMN push_platform TEXT CHECK (push_platform IN ('ios','android','web'));
ALTER TABLE devices ADD COLUMN push_enabled  BOOL NOT NULL DEFAULT true;
```

### Fan-out Trigger

Push is sent by the **notification worker** (a Kafka consumer), not inline in the API request path. Flow:

```
message.send API
  → write to messages table
  → publish message.created to Kafka (outbox pattern, already implemented)
  
Kafka consumer (notification worker)
  → check online presence (Redis): is any device of the recipient online?
  → if all devices offline → fetch push tokens from devices table
  → send APNs / FCM → log result in push_log table
```

This keeps the push delivery fully async and prevents APNs/FCM latency from affecting message send latency.

### Push Payload

```json
{
  "aps": {
    "alert": { "title": "Alice", "body": "Hey Bob!" },
    "badge": 3,
    "sound": "default",
    "thread-id": "<conversation_id>"
  },
  "conversation_id": "...",
  "message_id": "..."
}
```

We include `conversation_id` and `message_id` so the app can deep-link directly to the conversation on tap.

### Failure Handling

| Failure | Action |
|---------|--------|
| APNs 410 (token invalid) | Mark token as `push_enabled=false` |
| FCM 404 (token not registered) | Same |
| APNs 429 / FCM 500 | Retry with exponential back-off (max 3 attempts) |
| All retries exhausted | Write to `push_log` with `status=failed`; do not block message delivery |

### Architecture Diagram

```
[API Server]          [Kafka]           [Notification Worker]
message.send  ──────► message.created ──► check presence (Redis)
                                          │
                                          ├─ user online → skip push
                                          │
                                          └─ user offline
                                              ├─ iOS device → APNs HTTP/2
                                              └─ Android    → FCM HTTP v2
```

### Avoiding Double Push

The worker reads the presence key `presence:<user_id>` from Redis. If the key exists (TTL > 0), the user is considered online and no push is sent. This is a best-effort heuristic — the WS connection may have dropped in the last few seconds without the key expiring. A slight double-push is acceptable; a missed push is worse. Users can suppress push in-app per conversation (mute).

## Implementation Files

- `backend/internal/push/apns.go` — APNs HTTP/2 client (golang.org/x/net/http2)
- `backend/internal/push/fcm.go` — FCM v1 HTTP client
- `backend/internal/push/worker.go` — Kafka consumer, presence check, dispatch
- `backend/migrations/00012_devices_push.sql` — alter devices table
- `backend/migrations/00013_push_log.sql` — push delivery log table
- `docs/push-notifications.md` — operational guide

## Testing

- Unit: mock APNs/FCM client; verify correct payload construction.
- Unit: presence check skips push when Redis key is present.
- Integration: end-to-end with mock APNs server (e.g., `apns-mock`).
- CI: `PUSH_TEST_MODE=mock` env var activates mock transport.

## Interview Talking Points

- **Why async worker?** "APNs and FCM can add 100–500 ms of latency. If we block the send API on push delivery, every message send is slow. The Kafka consumer decouples latency."
- **Presence check before push**: "We check Redis presence before sending push. It's a best-effort heuristic — we might occasionally push when the WS just reconnected, but that's benign (app ignores it). The alternative — missing pushes — silently drops notifications."
- **Token lifecycle**: "APNs returns HTTP 410 when a token is stale. We handle this by marking the device as push_disabled. Without this, we'd keep sending to dead tokens, which APNs rate-limits against."
