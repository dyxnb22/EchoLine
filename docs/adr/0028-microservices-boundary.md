# ADR 0028: Microservices Boundary (Prototype)

## Status

Accepted (design-only prototype)

## Context

Long-term scaling may require splitting auth, message, and realtime paths.

## Decision

Document service boundaries without splitting the monolith yet:

| Service | Owns |
|---------|------|
| auth | users, sessions, JWT |
| conversation | memberships, channels, groups |
| message | seq, body, edit/recall |
| realtime | WebSocket fanout |
| media | S3 presign |
| search | index worker |
| notification | push, in-app |
| audit | admin, reports, DLQ |

Gateway routes remain path-based until gRPC internal APIs exist.

## Consequences

- Interview-ready boundary story without operational complexity today.
- Actual split deferred until load or team size warrants it.

## Files

- `docs/extensions-roadmap.md`
- `deploy/gateway/`
