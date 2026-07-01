# ADR 0003: Large Group Fanout Strategy

## Status

Accepted (skeleton)

## Context

EchoLine currently fanouts `message.created` via the realtime hub to online members in-process. This works for small groups but does not scale to large channels with thousands of subscribers.

## Decision

Use a hybrid strategy:

1. **Small groups (<= 256 members):** synchronous hub fanout on the API/WS hot path (current behavior).
2. **Large groups (> 256 members):** enqueue fanout work to the worker (`FanoutWorker`) which batches push notifications. Online WS clients on the same gateway instance still receive immediate hub push; cross-instance delivery uses Redis pub/sub registry (future).

Threshold `256` is configurable via env in a later iteration.

## Tradeoffs

| Approach | Pros | Cons |
|---|---|---|
| Hot-path hub fanout | Low latency, simple | O(members) per send on API |
| Worker batch fanout | Protects API latency | Slightly higher delivery delay |
| Read fanout (pull) | Scales to huge channels | Worse realtime UX |

## Implementation files

- `backend/internal/realtime/server.go` — hub fanout
- `backend/internal/worker/handlers.go` — `FanoutWorker` skeleton
- `backend/cmd/worker/main.go` — consumes `message.created`

## Verification

- E006 unit test: sender excluded from push
- Future: k6 large-group scenario (E010)

## Interview talking points

- Start with in-process fanout; move heavy fanout async when member count exceeds threshold.
- DB remains source of truth; fanout is best-effort with sync compensation via `/api/sync`.
