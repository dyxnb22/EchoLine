# ADR 0012: Advertisements Data Model

## Status

Accepted (design; implementation deferred to extension phase)

## Context

EchoLine's monetization roadmap includes an advertising tier for public channels (not for private DMs or groups, which would be a significant trust violation). The design must:

1. Preserve user trust: ads appear only in public channels, never in private conversations.
2. Comply with privacy regulations (GDPR, CCPA): no sensitive message content used for targeting.
3. Be revenue-efficient: support impression-based (CPM) and click-based (CPC) billing models.
4. Be architecturally isolated: ads data should not pollute the core messaging data model.

## Decision

### Ad Placement Scope

Ads appear **only** in:
- Public channel feeds (channel type = `PUBLIC`)
- Between messages in high-traffic channels (> 10k subscribers)

Ads **never** appear in:
- Private DMs
- Private groups
- Encrypted conversations
- System notifications

### Core Tables

```sql
-- Advertiser account
CREATE TABLE advertisers (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      UUID REFERENCES users(id),
  display_name TEXT NOT NULL,
  billing_type TEXT NOT NULL CHECK (billing_type IN ('CPM','CPC')),
  budget_cents BIGINT NOT NULL DEFAULT 0,
  created_at   TIMESTAMPTZ DEFAULT now()
);

-- Ad creative
CREATE TABLE ads (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  advertiser_id   UUID REFERENCES advertisers(id),
  headline        TEXT NOT NULL,
  body            TEXT,
  media_url       TEXT,
  cta_url         TEXT NOT NULL,
  status          TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','active','paused','completed')),
  start_at        TIMESTAMPTZ,
  end_at          TIMESTAMPTZ,
  targeting       JSONB NOT NULL DEFAULT '{}',
  created_at      TIMESTAMPTZ DEFAULT now()
);

-- Targeting: channel categories, subscriber count range, locale
-- targeting JSONB example:
-- { "channel_categories": ["tech","gaming"], "min_subscribers": 5000, "locales": ["en","zh"] }

-- Ad-to-channel assignment
CREATE TABLE ad_channel_placements (
  ad_id            UUID REFERENCES ads(id),
  conversation_id  UUID REFERENCES conversations(id),
  position_every_n INT NOT NULL DEFAULT 10,  -- show ad every N messages
  PRIMARY KEY (ad_id, conversation_id)
);

-- Impression tracking (append-only)
CREATE TABLE ad_impressions (
  id              BIGSERIAL PRIMARY KEY,
  ad_id           UUID REFERENCES ads(id),
  user_id         UUID REFERENCES users(id),
  conversation_id UUID REFERENCES conversations(id),
  device_id       UUID REFERENCES devices(id),
  shown_at        TIMESTAMPTZ DEFAULT now(),
  request_id      UUID NOT NULL  -- idempotency
);
CREATE UNIQUE INDEX ON ad_impressions (ad_id, user_id, request_id);

-- Click tracking (append-only)
CREATE TABLE ad_clicks (
  id           BIGSERIAL PRIMARY KEY,
  impression_id BIGINT REFERENCES ad_impressions(id),
  clicked_at   TIMESTAMPTZ DEFAULT now()
);
```

### Billing Model

- **CPM (cost per mille)**: billed per 1000 impressions. Budget decremented by `bid_cpm_cents / 1000` per impression.
- **CPC (cost per click)**: billed per click. Budget decremented by `bid_cpc_cents` per click.
- Budget enforcement: before serving an ad, check `budget_cents > 0`. Decrement atomically in the same transaction as the impression insert.

### Privacy Constraints

- Targeting is based on **channel metadata** (category, subscriber count, locale) — not on user message content or behavior.
- User IDs in `ad_impressions` are stored for frequency capping (same user should not see same ad > 3× per day), not for behavioral profiling.
- GDPR opt-out: users can opt out of ad targeting; `users.ad_opt_out = true` causes the ad server to skip them.

### Ad Injection in Feed

The feed API (`GET /api/conversations/:id/messages`) accepts a `?include_ads=1` query parameter. When set:
1. Fetch messages as normal.
2. For every `N`th message (per `ad_channel_placements.position_every_n`), the API inserts an ad payload of type `ad` into the response message list.
3. The client renders it as a sponsored card.

## Implementation Files

- `backend/migrations/` — ads tables DDL
- `backend/internal/ads/` _(planned)_ — ad selection, budget enforcement, impression/click tracking
- `backend/internal/message/handler.go` — inject ad into feed response (future)
- `docs/data-model.md` — update with ads schema reference

## Consequences

**Positive:**
- Ads are architecturally isolated in a separate schema area; core messaging is unaffected.
- Privacy-by-design: no behavioral targeting, channel-metadata-only targeting.
- Idempotent impression inserts prevent double-billing on retries.

**Negative:**
- Ad injection in the message feed API increases response complexity.
- Budget enforcement adds a write per impression; high-traffic channels create contention on `advertisers.budget_cents`. Mitigated by batching budget decrements.

## Interview Talking Points

- **Why only public channels?** "Ads in private messages destroy user trust and are banned by GDPR under the 'legitimate interest' basis — you need explicit consent for behavioral ads in private communication. We avoid the complexity by restricting to public channels."
- **Budget contention at scale**: "The `budget_cents` decrement is a hot row for popular advertisers. At scale, we'd use a Redis counter for real-time budget enforcement (allowing slight over-spend) and reconcile with Postgres periodically."
- **Frequency capping**: "We cap impressions per user per ad per day using a Redis sorted set keyed by `ad:{ad_id}:user:{user_id}:date:{YYYYMMDD}`. This avoids a Postgres read on every ad serve."
