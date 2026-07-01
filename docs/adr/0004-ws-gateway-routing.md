# ADR 0004: WebSocket Gateway Multi-Instance Routing

## Status

Accepted

## Context

EchoLine's WebSocket gateway (`backend/internal/realtime/server.go`) holds all active connections in a single in-process `Hub`. This works correctly for a single-instance deployment but breaks delivery correctness when multiple gateway replicas run behind a load balancer:

- **Fanout gap**: A message arrived at replica A cannot reach a recipient connected to replica B without a cross-process forwarding mechanism.
- **Presence skew**: Each replica tracks only its own connections; no instance has a global view.
- **Horizontal scaling blocker**: Without cross-instance routing, horizontal scaling is purely for connection capacity, not delivery completeness.

The key question is: what is the minimum viable cross-instance routing layer that preserves correctness without abandoning the modular-monolith architecture?

## Decision

Adopt a **Redis Pub/Sub fan-forward** model for cross-instance gateway routing:

1. Each gateway instance subscribes to a Redis channel named `gw:user:{user_id}` when a client connects.
2. When the message publisher (outbox worker or API hot-path) needs to push to a user, it publishes the payload to `gw:user:{user_id}` via Redis.
3. Every instance that has a live connection for that user receives the Redis message and forwards it over the WebSocket to the client.
4. Instances with no connection for that user receive the Redis message and discard it (no side effect).

Presence tracking is stored in Redis as a `SET gateway:presence:{user_id}` with the gateway instance ID and connection count, with a 10-second TTL refreshed on heartbeat.

## Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| Redis Pub/Sub (chosen) | Stateless gateways, O(1) publish, simple eviction via TTL | Redis becomes a SPOF; message delivery requires Redis availability |
| Consistent-hash routing (sticky sessions) | No Redis dependency for routing | LB coupling, sticky affinity complicates deploys |
| gRPC gossip mesh | Peer-to-peer, no central broker | Operational complexity, service discovery overhead |
| Shared DB polling | Zero infra | Polling latency unacceptable for realtime |

## Implementation Files

- `backend/internal/realtime/server.go` — Hub, connection register/unregister
- `backend/internal/realtime/gateway_router.go` _(planned)_ — Redis Pub/Sub publisher/subscriber
- `backend/internal/presence/redis.go` — Presence TTL in Redis
- `backend/internal/cache/redis.go` — Redis client shared by presence and routing

## Consequences

**Positive:**
- Gateway instances are stateless with respect to routing; new replicas start immediately.
- Presence data is globally consistent across instances within TTL bounds.
- Single-instance deployments require no change (Redis Pub/Sub with one subscriber is a no-op).

**Negative:**
- Redis must be highly available; a Redis outage causes WS push to fail (messages are still persisted in Postgres and recoverable via sync).
- Pub/Sub does not guarantee exactly-once; duplicate delivery to multiple gateway instances for the same user is possible and must be deduplicated at the client level by `server_seq`.

## Verification

- Unit test: mock Redis Pub/Sub; assert cross-instance forward calls `conn.WriteJSON` on the correct connection.
- Integration smoke: `RUN_WS_SMOKE=1 make smoke` with two gateway processes — both clients receive the message.

## Interview Talking Points

- **Problem framing**: "Without cross-instance routing, WS horizontal scaling only increases connection capacity, not delivery coverage."
- **Redis Pub/Sub choice**: "It gives us a broadcast primitive with O(1) publish cost and no persistent state — which is exactly what we need for ephemeral WS delivery."
- **Failure mode**: "If Redis is down, WS push fails gracefully; clients notice missed messages on next reconnect and call `/api/sync` to recover. Postgres is the source of truth."
- **Scalability ceiling**: "Redis Pub/Sub throughput is ~1M msgs/s on a single node; at EchoLine's scale we would shard by `user_id % N_redis_nodes` using a Redis Cluster and a consistent-hash key mapping."
