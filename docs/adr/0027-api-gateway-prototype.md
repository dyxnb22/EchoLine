# ADR 0027: API Gateway Prototype

## Status

Accepted

## Context

EchoLine runs as a modular monolith. Future microservice split requires a stable edge routing layer.

## Decision

Add `deploy/gateway/nginx.conf` and optional `docker compose --profile gateway` service that proxies `/api/*` and `/ws` to the monolith API.

## Consequences

- Operators can terminate TLS and route at the edge without code changes.
- Gateway config must stay in sync with monolith route prefixes.
- Full service split still needs ADR 0028 migration plan.

## Files

- `deploy/gateway/nginx.conf`
- `deploy/gateway/README.md`
- `docker-compose.yml` (`gateway` profile)

## Verification

- `docker compose --profile app --profile gateway config` validates compose graph.
