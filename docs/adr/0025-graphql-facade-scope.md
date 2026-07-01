# ADR 0025: GraphQL Facade Scope

## Status

Accepted (prototype)

## Context

REST is verbose for mobile/third-party clients. Full gqlgen stack adds N+1 and subscription complexity.

## Decision

Phase 1 prototype: hand-rolled `POST /graphql` JSON handler supporting:
- Query: `conversations`
- Mutation: `sendMessage` via variables

REST remains source of truth; GraphQL delegates to existing services.

## Future

gqlgen + DataLoaders + graphql-ws subscriptions per ADR 0022.

## Files

- `backend/internal/graph/handler.go`
