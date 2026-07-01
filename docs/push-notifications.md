# Push Notifications — EchoLine

## Problem

When a user's device is offline or the app is in the background, EchoLine cannot deliver messages over WebSocket. The user needs an OS-level push notification to know a message arrived.

## Tradeoff

| Option | Pros | Cons |
|--------|------|------|
| APNs + FCM (native) | Works when app is killed; OS-native | Requires separate gateway, token management |
| Web Push (VAPID) | No SDK; works in browser service worker | Blocked on iOS < 16.4; push permission UX is poor |
| Email/SMS fallback | Always works | High latency; wrong channel for chat |

**Decision**: APNs for iOS, FCM for Android, Web Push (VAPID) for PWA. All delivered asynchronously from a Kafka consumer. See ADR 0017.

## Device Token Registration

After login, the client registers its push token:

```bash
curl -X POST http://localhost:8080/devices/push-token \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "platform": "ios",
    "push_token": "abc123...",
    "device_id": "dev-uuid"
  }'
```

Token is stored in the `devices` table alongside the sync cursor.

## Notification Worker Flow

```
Kafka: message.created event consumed
  ↓
Check Redis presence key: presence:<recipient_user_id>
  ├─ TTL > 0 → user online → skip push
  └─ TTL = 0 → user offline
       ↓
       Fetch push tokens from devices table
       ↓
       iOS devices  → POST to APNs (HTTP/2 + p8 JWT auth)
       Android      → POST to FCM v1 (OAuth2 service account)
       Web          → POST to Web Push endpoint (VAPID)
       ↓
       Log result in push_log table
       ↓
       On APNs 410 → mark device push_enabled=false
```

## APNs Payload

```json
{
  "aps": {
    "alert": { "title": "Alice", "body": "Hey Bob!" },
    "badge": 3,
    "sound": "default",
    "thread-id": "<conversation_id>",
    "category": "MESSAGE"
  },
  "conversation_id": "...",
  "message_id": "..."
}
```

`thread-id` groups notifications per conversation in iOS Notification Center. `category` enables quick-reply action buttons (reply without opening the app).

## FCM Payload

```json
{
  "message": {
    "token": "<device_fcm_token>",
    "notification": {
      "title": "Alice",
      "body": "Hey Bob!"
    },
    "data": {
      "conversation_id": "...",
      "message_id": "..."
    },
    "android": {
      "notification": {
        "channel_id": "messages",
        "notification_count": 3
      }
    }
  }
}
```

## Mute and Opt-out

Users can mute a conversation via `POST /conversations/:id/mute`. Muted conversations are excluded from push delivery in the notification worker (checked before sending). Users can also disable all push notifications per device (`PUT /devices/:id/push-settings {"enabled": false}`).

## Environment Variables

```
APNS_KEY_FILE      path to .p8 private key
APNS_KEY_ID        10-char key ID from Apple Developer
APNS_TEAM_ID       Apple Developer Team ID
APNS_BUNDLE_ID     iOS app bundle ID
FCM_SERVICE_ACCOUNT_JSON  path to Firebase service account JSON
VAPID_PRIVATE_KEY  Base64url-encoded VAPID private key
VAPID_PUBLIC_KEY   Base64url-encoded VAPID public key
VAPID_SUBJECT      mailto:push@echoline.dev
```

## Testing

```bash
# Unit: mock APNs/FCM; verify push skipped for online users
go test ./internal/push/... -v

# Integration with mock APNs server
APNS_ENDPOINT=http://localhost:2197 go test ./internal/push/... -run TestAPNsIntegration

# Verify push_log table after test
psql $DATABASE_URL -c "SELECT * FROM push_log ORDER BY created_at DESC LIMIT 10;"
```

## Interview Angle

> "Push delivery is fully async via Kafka. We check Redis presence before sending — if the user's WS is alive, we skip the push. This prevents a common annoyance where the user gets an OS notification while actively reading the app. The key failure mode we handle is APNs 410 (stale token): we immediately disable the token so we don't accumulate send errors."
