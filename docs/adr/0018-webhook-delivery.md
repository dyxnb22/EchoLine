# ADR 0018: Webhook Delivery for Bots and Integrations

## Status

Accepted (design; implementation deferred to extension phase)

## Problem

EchoLine needs an outbound webhook system so that:

1. **Bots** (automated users) can receive message events via HTTP POST to their endpoint.
2. **Third-party integrations** (CI/CD bots, ticketing systems, workflow tools) can subscribe to conversation events.
3. Developers can build event-driven integrations without polling the REST API.

This is analogous to Slack's Incoming/Outgoing webhooks and GitHub webhooks.

The key design questions:

- How do webhook endpoints get registered and authenticated?
- Which events trigger a webhook call?
- How do we handle delivery failures, retries, and timeouts?
- How do we prevent malicious payloads from impersonating EchoLine?

## Decision

### Webhook Registration

Webhooks are registered per conversation (or globally for a bot user):

```
POST /webhooks
{
  "conversation_id": "...",   // null = all conversations for this user
  "url": "https://mybot.example.com/hook",
  "events": ["message.created", "member.joined", "reaction.added"],
  "secret": "optional-hmac-secret"
}
```

Stored in a `webhooks` table:

```sql
CREATE TABLE webhooks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    conversation_id UUID REFERENCES conversations(id) ON DELETE CASCADE,
    url             TEXT NOT NULL,
    events          TEXT[] NOT NULL,
    secret          TEXT,           -- HMAC-SHA256 signing key
    active          BOOL NOT NULL DEFAULT true,
    failure_count   INT NOT NULL DEFAULT 0,
    last_delivered  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Delivery Worker

Webhook delivery is handled by a dedicated **webhook worker** (Kafka consumer):

```
Kafka topic: webhook.events
  ↓
Webhook Worker:
  1. Look up active webhooks matching (conversation_id, event_type)
  2. For each webhook, POST payload to URL with X-EchoLine-Signature header
  3. On success (HTTP 2xx): update last_delivered, reset failure_count
  4. On failure: exponential back-off (1s, 5s, 30s, 5m, 30m); max 5 attempts
  5. After 5 failures: set active=false, notify owner
```

### Payload and Signing

Every webhook POST includes:

```
POST https://mybot.example.com/hook
Content-Type: application/json
X-EchoLine-Event: message.created
X-EchoLine-Delivery: <uuid>
X-EchoLine-Signature: sha256=<hmac-hex>
```

The signature is `HMAC-SHA256(secret, raw_body)`. Recipients verify it with their registered secret, preventing replay or spoofing.

### Retry and Back-off

| Attempt | Delay   |
|---------|---------|
| 1       | 1 s     |
| 2       | 5 s     |
| 3       | 30 s    |
| 4       | 5 min   |
| 5       | 30 min  |
| > 5     | Disable webhook, notify owner |

Retries are stored in a `webhook_deliveries` table for audit and manual re-trigger:

```sql
CREATE TABLE webhook_deliveries (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_id   UUID NOT NULL REFERENCES webhooks(id),
    event_type   TEXT NOT NULL,
    payload      JSONB NOT NULL,
    attempt      INT NOT NULL,
    status       TEXT NOT NULL,   -- pending, delivered, failed
    response_code INT,
    delivered_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Timeout and Safety

- HTTP client timeout: **10 seconds** per attempt.
- Maximum payload size: **64 KB** to prevent unbounded memory under fanout.
- Rate limit per webhook URL: 100 calls/minute (Redis counter) to prevent self-DDoS on a slow endpoint.

## Implementation Files

- `backend/migrations/00014_webhooks.sql` — webhooks + webhook_deliveries tables
- `backend/internal/webhook/handler.go` — CRUD REST for webhook registration
- `backend/internal/webhook/worker.go` — Kafka consumer, delivery logic, HMAC signing
- `backend/internal/webhook/retry.go` — back-off scheduler
- `docs/api.md` — webhook registration endpoints

## Testing

- Unit: HMAC signature generation and verification.
- Unit: back-off delay calculation.
- Integration: mock HTTP server captures webhook POST; verify payload + signature.
- Integration: simulate 3 failures → verify retry queue, then 4th succeeds.

## Interview Talking Points

- **Why async via Kafka?** "Webhook delivery can be slow (recipient endpoint may take seconds). If we deliver inline, a slow bot blocks message processing. The Kafka consumer lets us retry independently."
- **HMAC signing**: "Without payload signing, any attacker can POST fake events to a bot. The HMAC-SHA256 signature lets bots verify that the payload came from EchoLine."
- **Auto-disable after failures**: "If a webhook endpoint is down for hours, we don't want to accumulate infinite retry queue. After 5 failures we disable the webhook and notify the owner. GitHub uses the same approach."
