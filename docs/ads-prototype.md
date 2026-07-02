# Ads Platform Prototype — EchoLine

## Problem

EchoLine needs a revenue stream beyond subscriptions. Advertising in public channels and the explore feed is the primary option. The ads platform must:

1. Let advertisers create campaigns with targeting criteria and budgets.
2. Serve the highest-bid eligible ad at impression time.
3. Track impressions and clicks for billing and reporting.
4. Enforce frequency caps (don't show the same ad to the same user more than N times/day).
5. Protect user privacy: advertisers must not receive individual PII.

See ADR 0020 for the full data model and targeting design.

## Tradeoff

| Serving Model | Latency | Complexity | Accuracy |
|---------------|---------|------------|---------|
| Real-time auction (SSP/DSP) | High (external RTB round-trip) | Very high | Market price |
| Simplified in-house CPM | Low (single DB query) | Low-medium | Fixed bid |
| Pre-assigned slots (manual) | Very low | Very low | Not scalable |

**Decision**: In-house simplified CPM/CPC auction. No external RTB for MVP. This allows full control over the ad stack and avoids third-party tracking dependencies.

## Ad Serving Flow

```
User opens feed → GET /feed
                    ↓
           Ad Middleware:
           1. Build user profile (language, device, interest tags from group memberships)
           2. SELECT campaigns WHERE status='active' AND targeting matches user profile
              ORDER BY bid_micros DESC LIMIT 10
           3. For each candidate: check Redis frequency cap
           4. For each remaining: check Redis daily budget
           5. Select highest-bid eligible campaign
           6. Inject ad creative into feed response as { type: "ad", creative: {...} }
           7. Publish ad_event (impression) to Kafka (async, non-blocking)
```

## Creative Types

| Type | Fields | Display |
|------|--------|---------|
| `text` | title, body, cta_label, cta_url | Inline text banner |
| `image` | title, media_url, cta_url | Image with CTA |
| `sponsored_message` | sender_name, body, media_url | Looks like a channel post |

## Targeting

Targeting is a JSONB blob on the campaign. Matching uses PostgreSQL JSONB containment (`@>`):

```sql
-- Simplified matching (Phase 1)
SELECT id FROM ad_campaigns
WHERE status = 'active'
  AND (targeting->>'language' IS NULL
       OR targeting->'language' ? $user_language)
  AND (targeting->>'geo' IS NULL
       OR targeting->'geo' ? $user_geo)
ORDER BY bid_micros DESC;
```

For Phase 2 (interest targeting), user interest vectors are precomputed from group membership and cached in Redis.

## Reporting API

```
GET /ads/campaigns/:id/report?from=2026-07-01&to=2026-07-31
```

Response:

```json
{
  "impressions": 12483,
  "clicks": 341,
  "ctr": 0.027,
  "spend_micros": 12483000,
  "spend_usd": "12.48"
}
```

Individual user data is never exposed to advertisers. Reports contain only aggregate counts.

## Budget Enforcement

```
# Before serving ad for campaign X:
budget_used = GET campaign:spend:<campaign_id>   # Redis counter (approximate)
if budget_used >= daily_budget_micros: skip

# After impression:
INCRBY campaign:spend:<campaign_id> <bid_micros>
```

The DB `ad_events` table is the billing source of truth. Redis is approximate (tolerate ~5% over-delivery on bursts).

## Environment Variables

```
ADS_ENABLED=true
ADS_DEFAULT_FREQUENCY_CAP=3    # max impressions/user/campaign/day
ADS_REQUEST_TIMEOUT_MS=50      # max time to spend on ad selection per request
```

## Testing

```bash
# Unit: JSONB targeting match
go test ./internal/ads/... -run TestTargeting

# Unit: frequency cap enforcement
go test ./internal/ads/... -run TestFrequencyCap

# Unit: budget check
go test ./internal/ads/... -run TestBudget

# Integration: create campaign → GET /feed → assert ad in response
RUN_INTEGRATION=1 go test ./tests/... -run TestAdServing
```

## Interview Angle

> "The hardest part of ads is budget enforcement at high QPS. A global atomic counter would require a distributed lock. We use Redis INCRBY for approximate enforcement and tolerate ~5% over-delivery, which is standard in ad tech. The database is the billing source of truth — we reconcile Redis vs DB nightly and adjust budgets if needed."
