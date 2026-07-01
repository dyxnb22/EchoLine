# ADR 0019: Payment Ledger Design

## Status

Accepted (design; implementation deferred to extension phase)

## Problem

EchoLine's monetisation roadmap includes:

1. **Channel subscriptions** — users pay a monthly fee to access a premium channel.
2. **Pay-per-message** — tipping a creator per message received (Telegram Stars equivalent).
3. **Ad credits** — advertisers pre-load credits to fund ad campaigns.

All of these require a reliable, auditable financial ledger. The design questions are:

- Double-entry or single-entry bookkeeping?
- How to guarantee idempotency (prevent double-charges)?
- What external payment processor do we integrate?
- How to handle partial failures between payment processor and our DB?

## Decision

### Double-Entry Bookkeeping

Use a **double-entry ledger**: every financial event is two rows in a `ledger_entries` table (one debit, one credit). The sum across all entries for any account is always zero. This is the same model used by Stripe, Monzo, and every production financial system.

```sql
CREATE TABLE accounts (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id     UUID REFERENCES users(id),          -- null for system accounts
    type         TEXT NOT NULL,                       -- user, system, advertiser
    currency     TEXT NOT NULL DEFAULT 'USD',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ledger_entries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key TEXT NOT NULL UNIQUE,              -- prevents double-write
    account_id      UUID NOT NULL REFERENCES accounts(id),
    amount_cents    BIGINT NOT NULL,                   -- positive = credit, negative = debit
    currency        TEXT NOT NULL DEFAULT 'USD',
    type            TEXT NOT NULL,                     -- deposit, charge, refund, payout, ad_spend
    reference_id    TEXT,                              -- stripe payment_intent, subscription id, etc.
    description     TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ledger_account ON ledger_entries(account_id, created_at DESC);
```

**Balance = SUM(amount_cents) WHERE account_id = ?**

### Idempotency

Every charge uses a deterministic `idempotency_key`:

```
idempotency_key = sha256(user_id + ":" + event_type + ":" + reference_id + ":" + period)
```

The `UNIQUE` constraint on `idempotency_key` ensures that replaying the same event (e.g., a Kafka message re-delivery) does not double-charge. The insert either succeeds (first time) or raises a unique violation (already processed). The application distinguishes `already_exists` from other errors and treats it as a success.

### Stripe Integration

EchoLine uses **Stripe** as the payment processor:

1. **Frontend**: Stripe.js collects card details → creates PaymentIntent client secret.
2. **Backend**: receives `payment_intent.succeeded` webhook from Stripe → writes ledger entry.
3. **Idempotency**: Stripe webhook event ID is used as `reference_id`. Combined with `event_type`, this forms the idempotency key.

The ledger never trusts the frontend amount directly — it only writes entries based on verified Stripe webhook events.

### Subscription Flow

```
User clicks "Subscribe to channel"
  → POST /channels/:id/subscribe
  → Backend creates Stripe Subscription + PaymentIntent
  → Frontend completes 3D Secure if required
  → Stripe webhook: invoice.payment_succeeded
  → Backend worker writes:
      DEBIT  user account  -999 (= $9.99)
      CREDIT revenue account +999
  → Backend sets subscription.status = active in subscriptions table
```

### Failure Scenarios

| Failure | Handling |
|---------|----------|
| Stripe charge succeeds but webhook is delayed | Subscription stays pending until webhook arrives; user sees "processing" |
| DB write fails after Stripe charge | Stripe webhook re-delivers (up to 3 days); idempotency key prevents double-credit |
| DB write succeeds but Stripe refund requested | Refund entry added to ledger; no money movement needed in DB |

## Implementation Files

- `backend/migrations/00015_ledger.sql` — accounts, ledger_entries, subscriptions tables
- `backend/internal/payments/handler.go` — subscribe, top-up, balance endpoints
- `backend/internal/payments/stripe_webhook.go` — Stripe webhook handler
- `backend/internal/payments/ledger.go` — double-entry write helpers
- `docs/payments-prototype.md` — operational and product guide

## Testing

- Unit: idempotency key derivation; duplicate entry rejected without error escalation.
- Unit: balance calculation from ledger entries.
- Integration: Stripe webhook mock → verify ledger writes.
- Property: for any sequence of credits and debits, sum(all entries) == 0.

## Interview Talking Points

- **Why double-entry?** "Single-entry ledgers make bugs invisible: if you miscredit an account you have no counterpart row to catch it. Double-entry means every money movement has two rows that must sum to zero. Any imbalance is immediately detectable."
- **Idempotency key**: "Stripe can deliver the same webhook twice (network retry). Without an idempotency key with a DB UNIQUE constraint, we'd double-charge. The idempotency key is deterministic from the Stripe event ID, so replays are safe."
- **Never trust frontend amounts**: "The frontend should only send PaymentIntent IDs, never dollar amounts. The backend reads the verified amount from Stripe's webhook payload."
