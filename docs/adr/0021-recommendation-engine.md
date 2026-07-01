# ADR 0021: Recommendation Engine (Contacts and Channel Discovery)

## Status

Accepted (design; implementation deferred to extension phase)

## Problem

EchoLine needs two recommendation surfaces:

1. **Contact suggestions** — "People you may know" based on mutual group membership and address book overlaps (hashed phone numbers).
2. **Channel discovery** — "Channels for you" ranked by subscriber overlap with channels the user already follows.

The design questions:

- Graph-based (e.g., collaborative filtering) vs. rule-based (mutual-group count)?
- Where to compute recommendations — request-time query vs. pre-computed batch?
- How to prevent recommendation of blocked or muted users?

## Decision

### Algorithm: Mutual-Group Count (Phase 1) → Collaborative Filtering (Phase 2)

For Phase 1, use a simple **mutual-group count** heuristic:

```sql
-- "Users who share at least 2 groups with me, whom I haven't messaged yet"
SELECT other.user_id,
       COUNT(DISTINCT gm.group_id) AS mutual_groups
FROM group_members gm
JOIN group_members other ON other.group_id = gm.group_id
                         AND other.user_id <> $1   -- exclude self
LEFT JOIN conversations c ON (
    (c.user_a = $1 AND c.user_b = other.user_id) OR
    (c.user_a = other.user_id AND c.user_b = $1)
)
LEFT JOIN blocks b ON b.blocker_id = $1 AND b.blocked_id = other.user_id
WHERE gm.user_id = $1
  AND c.id IS NULL                                 -- not already in a conversation
  AND b.blocked_id IS NULL                         -- not blocked
GROUP BY other.user_id
HAVING COUNT(DISTINCT gm.group_id) >= 2
ORDER BY mutual_groups DESC
LIMIT 20;
```

For Phase 2, precompute a user-user similarity matrix using **item-based collaborative filtering** on group co-membership, stored as a materialized `user_similarity` table refreshed nightly by a batch job.

### Channel Discovery

Similar approach for channels:

```sql
SELECT c.id, c.name, COUNT(DISTINCT s2.user_id) AS overlap
FROM channel_subscriptions s1
JOIN channel_subscriptions s2 ON s2.channel_id <> s1.channel_id
                               AND s2.user_id = $1
JOIN channels c ON c.id = s2.channel_id
LEFT JOIN channel_subscriptions me ON me.channel_id = c.id AND me.user_id = $1
WHERE s1.user_id = $1
  AND me.channel_id IS NULL    -- not already subscribed
GROUP BY c.id, c.name
ORDER BY overlap DESC
LIMIT 10;
```

### Pre-computation vs. Request-Time

| Approach | Latency | Freshness | Complexity |
|----------|---------|-----------|------------|
| Request-time SQL | 10–200 ms | Real-time | Low |
| Pre-computed batch (nightly) | < 5 ms | Up to 24h stale | Medium |
| Real-time graph (GraphX/DGL) | < 10 ms | Minutes | High |

**Decision**: Start with request-time SQL for Phase 1 (simpler, no infra). Pre-compute nightly batch for Phase 2 when user count exceeds 100k.

### Privacy and Safety

- Blocked users are **never** recommended (LEFT JOIN blocks, IS NULL check).
- Muted users are de-ranked (muted flag multiplies score by 0.1).
- Recommendations use only in-app graph data — no external graph enrichment.
- Address book matching (if implemented) uses **hashed phone numbers** only; raw phone numbers are never stored server-side.

### API

```
GET /recommendations/contacts?limit=20
GET /recommendations/channels?limit=10
```

Responses include a `reason` field: `"2 mutual groups"`, `"3 mutual subscribers"`.

## Implementation Files

- `backend/internal/recommendation/handler.go` — REST endpoints
- `backend/internal/recommendation/contacts.go` — mutual-group SQL query
- `backend/internal/recommendation/channels.go` — channel overlap SQL query
- `backend/internal/recommendation/batch_job.go` — nightly pre-compute (Phase 2)
- `docs/recommendation-prototype.md` — product and interview guide

## Testing

- Unit: mutual-group SQL returns correct ranked list given fixture data.
- Unit: blocked users excluded from results.
- Unit: channel overlap calculation.
- Integration: seed 5 users in 3 groups → verify recommendations.

## Interview Talking Points

- **Why mutual-group count first?** "It's explainable and computable with a single SQL query. We don't need a graph database or ML model for the first 500k users. When we need scale, we precompute nightly with a batch job."
- **Privacy**: "We never use external social graphs. Recommendations are based purely on in-app behavior — groups and channels the user explicitly joined. Blocked users are hard-excluded at the query level."
- **Scaling path**: "Above 1M users, the mutual-group query becomes O(group_size²). We'd move to a pre-computed `user_similarity` table (SVD or ALS on the co-membership matrix) refreshed nightly by a Spark or Flink job."
