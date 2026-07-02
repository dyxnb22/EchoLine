# Payments Prototype — EchoLine

## Problem

EchoLine's monetisation requires collecting money from users for:

1. **Channel subscriptions** — monthly fee to access premium channels.
2. **Stars / tips** — micro-payments to creators per message (Telegram Stars equivalent).
3. **Ad credits** — advertisers top up a wallet to fund ad campaigns.

All flows require a reliable, auditable ledger and integration with a payment processor (Stripe). See ADR 0019 for the double-entry ledger design.

## Tradeoff

| Approach | Pros | Cons |
|----------|------|------|
| Stripe Billing (subscriptions) | Handles renewals, failed payment retries, proration | Stripe coupling; webhook complexity |
| Custom subscription loop | Full control | Build dunning, retry, SCA compliance from scratch |
| Apple/Google IAP | Required for App Store / Play Store | 30% fee; server-side validation required |

**Decision**: Stripe as primary processor for web. Apple/Google IAP planned for mobile (required by app store policies). Ledger is processor-agnostic.

## Core Flows

### 1. Channel Subscription

```
User: POST /channels/:id/subscribe
  → Backend: create Stripe Customer (if not exists)
  → Backend: create Stripe Subscription (price_id from channel config)
  → Response: { "client_secret": "...", "status": "requires_action" }
  → Frontend: stripe.confirmCardPayment(client_secret)
  → Stripe webhook: invoice.payment_succeeded
  → Backend worker: write double-entry ledger rows
  → Backend: set subscriptions.status = 'active'
```

### 2. Tips (Stars)

```
User: POST /messages/:id/tip { "amount_cents": 100 }
  → Backend: create Stripe PaymentIntent (one-off charge)
  → Frontend: stripe.confirmCardPayment(client_secret)
  → Stripe webhook: payment_intent.succeeded
  → Backend: DEBIT user wallet, CREDIT creator wallet
  → Backend: emit tip.received WS event to creator
```

### 3. Ad Credit Top-up

```
Advertiser: POST /ads/wallet/top-up { "amount_cents": 5000 }
  → Stripe PaymentIntent
  → webhook: payment_intent.succeeded
  → CREDIT advertiser ad_account
```

## Idempotency

Every ledger write uses a deterministic idempotency key:

```
idempotency_key = sha256(stripe_event_id + ":" + event_type)
```

Stripe can deliver the same webhook multiple times. The UNIQUE constraint on `idempotency_key` in the `ledger_entries` table ensures the second delivery is a no-op (pg unique violation → treat as success).

## Refunds

```
Admin: POST /admin/payments/:id/refund
  → Backend: Stripe Refund API
  → Stripe webhook: charge.refunded
  → Backend: write CREDIT entry on user account (reversal)
  → Backend: set subscription.status = 'refunded' if applicable
```

## Environment Variables

```
STRIPE_SECRET_KEY          sk_test_...
STRIPE_WEBHOOK_SECRET      whsec_...   (from Stripe dashboard)
STRIPE_SUBSCRIPTION_PRICE  price_...   (default channel subscription price ID)
```

## Testing

```bash
# Stripe CLI for local webhook forwarding
stripe listen --forward-to localhost:8080/webhooks/stripe

# Trigger test events
stripe trigger invoice.payment_succeeded
stripe trigger payment_intent.succeeded

# Unit: idempotency key uniqueness
go test ./internal/payments/... -run TestIdempotency

# Unit: double-entry ledger balance = 0
go test ./internal/payments/... -run TestLedgerBalance
```

## Interview Angle

> "The most important correctness property in payments is idempotency. Stripe can deliver the same webhook twice. Without an idempotency key with a database UNIQUE constraint, a network retry would double-charge the user. We derive the idempotency key deterministically from the Stripe event ID, so any replay is a safe no-op. The double-entry ledger is our audit trail — any imbalance is immediately detectable."
