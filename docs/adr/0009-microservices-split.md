# ADR 0009: Microservices Split Strategy

## Status

Accepted (design; current architecture remains modular monolith)

## Context

EchoLine is built as a **modular monolith** — a single deployable binary (`cmd/api`) with cleanly separated internal packages. This is intentional for the current stage: it simplifies development, deployment, and debugging while the data model and API contracts stabilize.

As EchoLine scales beyond ~500k DAU, specific components will hit independent scaling ceilings:

1. **WebSocket gateway**: CPU and memory are dominated by connection count, not compute. Needs horizontal scaling independently of the message API.
2. **Media/storage service**: Large file uploads consume bandwidth and block API goroutines. Should be isolated.
3. **Search service**: Full-text indexing and query workloads are CPU/memory-intensive and should not share resources with the message-send hot path.
4. **Notification service**: Fan-out to FCM/APNs involves external HTTP calls with high latency variability; isolating prevents goroutine pool exhaustion in the main API.

The decision is **when and how** to split, not **whether** to split.

## Decision

Adopt a **staged extraction strategy**: split services one at a time, driven by observed scaling bottlenecks, with shared Postgres and Kafka as the integration backbone during the transition.

### Split Order and Triggers

| Phase | Service | Trigger to Split | Notes |
|-------|---------|-----------------|-------|
| 1 | WebSocket Gateway | WS connections > 100k per instance | Extract `realtime` package into `cmd/gateway`; communicate via Redis Pub/Sub (ADR 0004) |
| 2 | Media Service | Upload bandwidth > 1 Gbps per instance | Extract `media` package into `cmd/media`; keep presign logic, move validation/virus-scan here |
| 3 | Search Service | Search index rebuild blocks message index worker | Extract into `cmd/search`; consume `message.created` from Kafka |
| 4 | Notification Service | FCM/APNs calls cause P99 spikes on API | Extract into `cmd/notification`; consume `delivery.ack` events |

### Communication Pattern

- **Synchronous** (API call from one service to another): only for read operations that require fresh data. Use gRPC with protobuf.
- **Asynchronous** (events): always via Kafka topics. Services consume events they care about.
- **Shared data**: each service owns its domain tables in Postgres. Cross-service data access is via events or API calls, never direct DB queries across service boundaries.

### Service Boundaries

```
┌──────────────────────────────────────────────────────────────┐
│  clients (web, mobile, desktop)                              │
└──┬───────────────────────┬───────────────────────────────────┘
   │ REST/HTTP             │ WebSocket
   ▼                       ▼
┌─────────────┐   ┌────────────────────┐
│  API Service│   │  WS Gateway Service│
│  (cmd/api)  │   │  (cmd/gateway)     │
└──────┬──────┘   └────────┬───────────┘
       │                   │ Redis Pub/Sub
       ▼                   │
┌────────────────────────────────────────────────────────────┐
│  Kafka (message.created, delivery.ack, user.event, …)      │
└────────────┬─────────────┬──────────────┬──────────────────┘
             ▼             ▼              ▼
   ┌──────────────┐ ┌────────────┐ ┌────────────────┐
   │  Worker      │ │  Search    │ │  Notification  │
   │  (cmd/worker)│ │  Service   │ │  Service       │
   └──────────────┘ └────────────┘ └────────────────┘
```

## Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| Modular monolith forever | Simplest | Vertical scaling ceiling, single point of failure |
| Big-bang microservices upfront | Full scale from day 1 | 10× ops complexity before product-market fit |
| Staged extraction (chosen) | Scales with actual bottlenecks | Requires discipline to maintain service boundaries in monolith |
| Serverless functions | Auto-scale | Cold starts, complex state management |

## Implementation Files

- `backend/cmd/api/main.go` — current monolith entrypoint
- `backend/cmd/gateway/main.go` _(planned)_ — extracted WS gateway
- `backend/cmd/media/main.go` _(planned)_ — extracted media service
- `backend/cmd/search/main.go` _(planned)_ — extracted search service
- `backend/cmd/notification/main.go` _(planned)_ — extracted notification service
- `backend/internal/realtime/` — to become `cmd/gateway` internals
- `docs/architecture.md` — update with service diagram

## Consequences

**Positive:**
- Each service scales independently on the dimension that matters (CPU, memory, network).
- Failure isolation: a media service crash does not affect message delivery.
- Teams can own services independently.

**Negative:**
- Distributed system complexity: network partitions, service discovery, distributed tracing required.
- Kafka becomes a critical dependency; its availability affects all services.
- Schema evolution requires coordinated migration across service owners.

## Verification

- Architecture review: confirm each package in the monolith has no import cycles crossing service boundaries (enforced via `golang.org/x/tools/go/analysis` or a custom lint rule).
- Integration test (post-split): send a message via API service; assert WS gateway receives it via Redis Pub/Sub and delivers to connected client.

## Interview Talking Points

- **Monolith first**: "We deliberately started with a modular monolith. The product and data model need to be stable before you can draw correct service boundaries. Splitting too early locks you into wrong boundaries."
- **Staged extraction**: "We extract services one at a time, driven by observed bottlenecks. The WS gateway is the first candidate because its scaling dimension (connections) is fundamentally different from the API's compute/DB load."
- **Kafka as backbone**: "Kafka decouples services at write time. Services are not aware of each other — they only know about event schemas. This makes it safe to add a new service (e.g., notification) without changing existing services."
- **CAP tradeoff**: "Within a service, we have strong consistency (Postgres ACID). Across services, we accept eventual consistency mediated by Kafka. This is the same tradeoff WhatsApp and Telegram make."
