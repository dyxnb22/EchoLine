# Round 01 Summary

## Counts

| Priority | Found | Fixed | Open |
|---|---|---|---|
| P0 | 0 | 0 | 0 |
| P1 | 4 | 4 | 0 |
| P2 | 6 | 6 | 0 |
| P3 | 4 | 0 | 4 |
| P4 | 0 | 0 | 0 |

## Highlights

- GraphQL prototype had a real permission bypass on `addReaction` — fixed with same membership path as REST.
- Frontend WS lifecycle caused reconnect storms and stale-token loops — fixed with ref indirection and refresh-on-reconnect.
- Prototype ads/payment routes had missing guards — fixed with membership, campaign binding, and positive amount validation.
- Metrics endpoint now supports optional `METRICS_TOKEN` bearer protection.

## Next step

Round 02 full re-audit required (P1/P2 found this round).
