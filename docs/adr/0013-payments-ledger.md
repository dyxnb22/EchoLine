# ADR 0013: Payments and Ledger Design

## Status

Accepted (design; implementation deferred to extension phase)

## Context

EchoLine's payments roadmap includes:
1. **In-app tipping**: Users tip creators in public channels (similar to Telegram Stars).
2. **Premium subscriptions**: Users pay for premium features (increased file size, no ads, custom themes).
3. **Advertiser billing**: Advertisers are billed for impressions/clicks (ADR 0012).

All three require a **reliable financial ledger** with the following invariants:
- No money is lost or created (double-entry accounting).
- Payment operations are idempotent (network retries must not charge twice).
- Audit trail is immutable.
- External payment processor integration (Stripe) is abstracted behind a service boundary.

## Decision

### Double-Entry Ledger Model

Use a **double-entry bookkeeping** model where every financial event creates two ledger entries that sum to zero:

```sql
-- Account types: user_wallet, escrow, revenue, stripe_receivable
CREATE TABLE accounts (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_id     UUID,             -- user_id, advertiser_id, or system
  owner_type   TEXT NOT NULL,    -- 'user', 'advertiser', 'system'
  account_type TEXT NOT NULL,    -- 'wallet', 'escrow', 'revenue', 'stripe'
  currency     TEXT NOT NULL DEFAULT 'USD',
  created_at   TIMESTAMPTZ DEFAULT now()
);

-- Immutable ledger entries (append-only)
CREATE TABLE ledger_entries (
  id              BIGSERIAL PRIMARY KEY,
  transaction_id  UUID NOT NULL,          -- groups debit+credit pair
  account_id      UUID REFERENCES accounts(id),
  amount_cents    BIGINT NOT NULL,         -- positive=credit, negative=debit
  currency        TEXT NOT NULL DEFAULT 'USD',
  entry_type      TEXT NOT NULL,           -- 'tip', 'subscription', 'ad_billing', 'stripe_charge', 'refund'
  reference_id    TEXT,                    -- e.g., message_id for tips
  idempotency_key TEXT UNIQUE NOT NULL,    -- prevents duplicate entries
  created_at      TIMESTAMPTZ DEFAULT now()
);

-- Constraint: sum of amount_cents for each transaction_id must be 0
-- Enforced at application layer via transaction service
```

### Example: User Tips Creator

| Entry | Account | Amount | Description |
|-------|---------|--------|-------------|
| 1 | `user_wallet:{alice}` | -500 | Alice sends $5.00 tip |
| 2 | `escrow:{system}` | +500 | Funds held in escrow |
| 3 | `escrow:{system}` | -425 | Platform fee (15%) deducted, net to creator |
| 4 | `user_wallet:{bob}` | +425 | Bob (creator) receives tip net |
| 5 | `revenue:{echoline}` | +75 | Platform 15% fee |

All 5 entries are in one `transaction_id`; sum = 0.

### Idempotency

Every ledger operation is keyed by an `idempotency_key` (`UUID`). The client (or API handler) generates the key and includes it in the request. The database unique constraint on `idempotency_key` ensures duplicate requests return the existing result without creating new entries.

### External Payment Processor (Stripe)

- **Stripe is the source of truth for fiat money movement.** The ledger records the intent and result, but does not move real money.
- Flow: User tops up wallet → Stripe webhook fires → ledger entry created.
- Payout: Creator requests payout → ledger debit → Stripe payout API called → webhook confirms.
- Stripe webhook events are processed idempotently using Stripe's event ID as the `idempotency_key`.

### Balance Computation

Running balance = sum of all `amount_cents` for an account. For performance:

- Cache balance in Redis with 5-second TTL.
- On balance-critical paths (tip authorization), re-read from Postgres to avoid over-spending.

### Compliance

- PCI DSS: EchoLine never stores card numbers; all card data goes to Stripe.
- GDPR: Financial data is subject to data retention rules; ledger entries are kept for 7 years (legal requirement in most jurisdictions), then archived to cold storage.
- Refunds: A refund creates a compensating ledger transaction (reverses the original entries). The original entries are never modified (immutable ledger).

## Implementation Files

- `backend/migrations/` — `accounts`, `ledger_entries` tables
- `backend/internal/payments/` _(planned)_ — ledger service, Stripe webhook handler, balance service
- `backend/internal/api/payments.go` _(planned)_ — tip API, subscription API, balance API
- `docs/data-model.md` — update with payments schema reference

## Consequences

**Positive:**
- Double-entry ensures ledger integrity; any balance discrepancy is detectable by summing entries.
- Immutable ledger provides full audit trail.
- Idempotency key prevents double-charging on network retries.

**Negative:**
- Double-entry is more complex to implement than a simple balance counter.
- High-volume tipping creates high insert rate on `ledger_entries`; requires partitioning by `created_at`.
- Cross-currency support is deferred (all amounts in USD for MVP).

## Interview Talking Points

- **Why double-entry?** "In single-entry accounting (just a balance counter), any bug that charges without crediting results in money disappearing with no audit trail. Double-entry makes imbalances impossible to hide — every transaction must sum to zero."
- **Idempotency under network failures**: "The client generates a UUID for every payment operation. If the request times out, the client retries with the same UUID. The server returns the same result, never charges twice. This is the same pattern Stripe uses for their own API."
- **Stripe as source of truth**: "EchoLine never touches raw card data. Stripe handles PCI compliance. Our ledger records the business logic; Stripe records the money movement. This separation is standard practice."
