# ADR 0020: Ads Platform Data Model and Targeting

## Status

Accepted (design; implementation deferred to extension phase)

## Problem

EchoLine's revenue model includes advertising in public channels and the explore feed. The ads platform must:

1. Allow advertisers to create campaigns with a budget and targeting criteria.
2. Serve ads to eligible users at impression time.
3. Track impressions and clicks for billing.
4. Ensure privacy: targeting cannot reveal individual PII to advertisers.
5. Enforce frequency caps to avoid spamming users.

The key design questions:

- CPM (per impression) or CPC (per click) billing model?
- How do we target users without leaking PII?
- How do we enforce budgets in real time without a global lock?

## Decision

### Billing Model

Support both **CPM (cost per mille impressions)** and **CPC (cost per click)**. Each campaign specifies its model. For MVP, CPM is the default — simpler to implement and audit.

### Data Model

```sql
CREATE TABLE ad_campaigns (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    advertiser_id   UUID NOT NULL REFERENCES users(id),
    name            TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'draft',   -- draft, active, paused, exhausted
    billing_model   TEXT NOT NULL DEFAULT 'cpm',     -- cpm, cpc
    bid_micros      BIGINT NOT NULL,                 -- bid in micro-dollars (1000000 = $1.00)
    daily_budget_micros BIGINT NOT NULL,
    total_budget_micros BIGINT NOT NULL,
    start_at        TIMESTAMPTZ,
    end_at          TIMESTAMPTZ,
    targeting       JSONB NOT NULL DEFAULT '{}',     -- see targeting spec below
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ad_creatives (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES ad_campaigns(id) ON DELETE CASCADE,
    type        TEXT NOT NULL,                -- text, image, sponsored_message
    title       TEXT,
    body        TEXT,
    media_url   TEXT,
    cta_label   TEXT,
    cta_url     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ad_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES ad_campaigns(id),
    creative_id UUID NOT NULL REFERENCES ad_creatives(id),
    user_id     UUID NOT NULL REFERENCES users(id),
    event_type  TEXT NOT NULL,               -- impression, click
    cost_micros BIGINT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ad_events_campaign ON ad_events(campaign_id, created_at DESC);
CREATE INDEX idx_ad_events_user     ON ad_events(user_id, created_at DESC);
```

### Targeting Specification (JSONB)

```json
{
  "geo": ["US", "CA"],
  "language": ["en"],
  "interests": ["technology", "gaming"],
  "min_age": 18,
  "max_age": 35,
  "device_platform": ["ios", "android"]
}
```

Targeting evaluates against anonymous user profile attributes — **no PII is exposed to the advertiser**. Advertisers only see aggregated impression/click counts, not which specific users saw the ad.

### Budget Enforcement (Approximate)

Global real-time budget enforcement requires a distributed counter. We use **Redis** for approximate per-campaign spend:

```
INCRBY campaign:spend:<campaign_id> <bid_micros>
EXPIRE campaign:spend:<campaign_id> 86400   // daily reset
```

Before serving an ad, the ad server checks:

1. Redis spend < daily_budget_micros
2. Redis spend < (total_budget_micros - total_spend from DB)

This is approximate: Redis and the DB can drift by a small amount during failures. We prefer slight over-delivery (within 5%) over under-delivery. The DB `ad_events` table is the source of truth for billing.

### Frequency Cap

```
frequency:<user_id>:<campaign_id>:<day>  INCR + EXPIRE 86400
```

If the count exceeds the campaign's `max_impressions_per_day_per_user` (default: 3), the ad is skipped for that user that day.

### Ad Serving Flow

```
GET /feed?page=1
  → Ad server middleware queries eligible campaigns:
      SELECT * FROM ad_campaigns
      WHERE status = 'active'
        AND (end_at IS NULL OR end_at > NOW())
        AND targeting @> user_profile_attributes   -- JSONB containment
      ORDER BY bid_micros DESC
      LIMIT 10
  → Check Redis frequency cap per campaign
  → Check Redis daily budget per campaign
  → Select highest-bid eligible campaign
  → Inject ad creative into feed response
  → Publish ad_event (impression) async via Kafka
```

## Implementation Files

- `backend/migrations/00016_ads.sql` — ad_campaigns, ad_creatives, ad_events
- `backend/internal/ads/handler.go` — campaign CRUD, creative upload
- `backend/internal/ads/server.go` — ad selection logic, JSONB targeting
- `backend/internal/ads/budget.go` — Redis budget + frequency cap
- `docs/ads-prototype.md` — operational guide

## Testing

- Unit: JSONB targeting match logic.
- Unit: frequency cap Redis key construction and enforcement.
- Unit: budget check (Redis approximate + DB exact).
- Integration: create campaign → serve ad → verify ad_events row.

## Interview Talking Points

- **Why JSONB for targeting?** "Targeting criteria change frequently — new attributes, new geo regions. JSONB lets us add new criteria without schema migrations. The `@>` containment operator enables indexed matching."
- **Budget enforcement tradeoff**: "Global atomic budget enforcement requires distributed locking — expensive at high QPS. Redis INCRBY gives us approximate enforcement. We tolerate up to 5% over-delivery, which is standard in ad tech. The DB remains the billing source of truth."
- **Privacy**: "Advertisers set targeting criteria; they never see which users matched. Only aggregated impression/click counts are returned in campaign reports."
