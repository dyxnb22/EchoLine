# ADR 0024: Webhook Delivery Retry

## Status

Accepted

## Context

Outbound webhooks on `message.created` must not block the send path and must survive transient receiver failures.

## Decision

1. `PersistingDispatcher` attempts HTTP POST; on failure enqueues row in `webhook_deliveries`.
2. Worker `RetryWorker` polls pending rows every 10s, max 5 attempts, then marks `failed`.
3. Send path uses async goroutine (unchanged).

## Tradeoff

At-least-once delivery to webhook receivers; receivers must be idempotent.

## Files

- `backend/internal/webhook/repository.go`
- `backend/cmd/worker/main.go`
- `backend/migrations/00014_admin_webhook_ads.sql`
