# Recommendation Engine Prototype — EchoLine

## Problem

Users discover new contacts and channels primarily through search and invitations. A recommendation system improves organic growth by surfacing:

1. **People you may know** — based on mutual group membership.
2. **Channels for you** — based on subscriber overlap with channels you already follow.
3. (Future) **Trending content** — popular messages in public channels.

See ADR 0021 for the full algorithm design.

## Tradeoff

| Algorithm | Freshness | Latency | Scale limit | Explainability |
|-----------|-----------|---------|-------------|----------------|
| Request-time SQL (mutual-group) | Real-time | 10–200 ms | ~500k users | High ("2 mutual groups") |
| Pre-computed batch (nightly ALS) | Hours | < 5 ms | Millions | Medium |
| Real-time graph ML | Minutes | < 10 ms | Unlimited | Low |

**Decision**: Phase 1 = request-time SQL. Phase 2 = nightly batch pre-computation. Phase 3 = graph ML if user count justifies it. See ADR 0021.

## API

```
GET /recommendations/contacts?limit=20
GET /recommendations/channels?limit=10
```

Response:

```json
{
  "contacts": [
    {
      "user_id": "...",
      "username": "carol",
      "avatar_url": "...",
      "reason": "2 mutual groups",
      "mutual_group_count": 2
    }
  ]
}
```

The `reason` field is shown in the UI: "2 mutual groups" — this increases click-through rate vs. opaque "Suggested for you".

## Contact Recommendation SQL

```sql
SELECT other.user_id,
       u.username,
       COUNT(DISTINCT gm.group_id) AS mutual_groups
FROM group_members gm
JOIN group_members other ON other.group_id = gm.group_id
                         AND other.user_id != $1
JOIN users u ON u.id = other.user_id
LEFT JOIN conversations c ON (
    (c.user_a = $1 AND c.user_b = other.user_id) OR
    (c.user_a = other.user_id AND c.user_b = $1)
)
LEFT JOIN blocks b ON b.blocker_id = $1 AND b.blocked_id = other.user_id
WHERE gm.user_id = $1
  AND c.id IS NULL
  AND b.blocked_id IS NULL
GROUP BY other.user_id, u.username
HAVING COUNT(DISTINCT gm.group_id) >= 1
ORDER BY mutual_groups DESC
LIMIT $2;
```

Performance: this query does two joins on `group_members`. With an index on `group_members(user_id)` and `group_members(group_id)`, it runs in < 50 ms for groups up to 10k members.

## Channel Recommendation SQL

```sql
SELECT c.id, c.name, c.description,
       COUNT(DISTINCT s2.user_id) AS overlap
FROM channel_subscriptions s1
JOIN channel_subscriptions s2 ON s2.channel_id != s1.channel_id
JOIN channels c ON c.id = s2.channel_id
LEFT JOIN channel_subscriptions me ON me.channel_id = c.id AND me.user_id = $1
WHERE s1.user_id = $1
  AND me.channel_id IS NULL
GROUP BY c.id, c.name, c.description
ORDER BY overlap DESC
LIMIT $2;
```

## Privacy

- Blocked users are hard-excluded at the query level (LEFT JOIN + IS NULL).
- No external social graph data is used.
- Recommendations are based only on explicit in-app actions (group join, channel subscribe).
- The `reason` field is generic ("2 mutual groups") — it does not reveal which groups.

## Caching

Recommendation results are cached in Redis per user for 15 minutes:

```
recommendation:contacts:<user_id>  → JSON blob, TTL=900s
recommendation:channels:<user_id>  → JSON blob, TTL=900s
```

Cache is invalidated on group join/leave and channel subscribe/unsubscribe events.

## Testing

```bash
# Unit: contact recommendation with fixture data
go test ./internal/recommendation/... -run TestContactRecommendation

# Unit: blocked user excluded
go test ./internal/recommendation/... -run TestBlockedExcluded

# Unit: channel overlap calculation
go test ./internal/recommendation/... -run TestChannelRecommendation

# Integration: seed 5 users in 2 groups → verify recommendations
RUN_INTEGRATION=1 go test ./tests/... -run TestRecommendationIntegration
```

## Interview Angle

> "We started with a mutual-group SQL query because it's explainable and requires no ML infrastructure. The query does two self-joins on group_members, which is fast with proper indexing for groups under 10k members. For scale beyond 500k users, we'd precompute a user-similarity matrix nightly using ALS (alternating least squares) on the co-membership matrix — the standard approach for collaborative filtering at scale."
