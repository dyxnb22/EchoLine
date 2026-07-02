# GraphQL Prototype — EchoLine

## Problem

The REST + bespoke-WS API works well for the first-party React frontend but is verbose for third-party integrations and mobile clients. GraphQL solves:

- **Over-fetching**: REST conversation list returns full message objects; GraphQL returns only requested fields.
- **Subscriptions as first-class**: `graphql-ws` protocol is supported natively by Apollo Client and urql, removing the need for clients to learn EchoLine's bespoke WS protocol.
- **Code generation**: `graphql-codegen` can generate typed TypeScript hooks from the schema automatically.

## Tradeoff

| Factor | REST | GraphQL |
|--------|------|---------|
| Simplicity | High | Medium (schema + resolver boilerplate) |
| Over-fetching | Possible | Mitigated by field selection |
| N+1 risk | Low (hand-tuned SQL) | High (requires DataLoaders) |
| Tooling | curl, Postman | Apollo Studio, GraphiQL, codegen |
| Subscriptions | Custom WS | graphql-ws standard |

**Decision**: GraphQL added as a facade layer; REST remains for internal/admin use.

## Implementation Files

- `backend/graph/schema.graphql` — SDL schema
- `backend/graph/resolver.go` — resolver wiring
- `backend/graph/query.go` — Query resolvers
- `backend/graph/mutation.go` — Mutation resolvers
- `backend/graph/subscription.go` — Subscription resolvers (hub adapter)
- `backend/graph/loader/` — DataLoaders for users, messages, groups

## Key Endpoints

| Endpoint | Protocol | Purpose |
|----------|----------|---------|
| `POST /graphql` | HTTP | Queries and mutations |
| `GET /graphql` | HTTP | GraphiQL UI (development only) |
| `WS /graphql/ws` | WebSocket (graphql-ws) | Subscriptions |

## Example Queries

```graphql
# Fetch conversation list (no over-fetching)
query ConversationList {
  conversations(limit: 20) {
    edges {
      node {
        id
        title
        unreadCount
        lastMessage { text createdAt }
      }
    }
    pageInfo { endCursor hasNextPage }
  }
}

# Subscribe to new messages
subscription OnMessage($conversationId: ID!) {
  messageAdded(conversationId: $conversationId) {
    id text createdAt
    sender { id username avatarUrl }
    reactions { emoji count }
  }
}
```

## Development Setup

```bash
# Generate resolvers from schema (gqlgen)
cd backend
go run github.com/99designs/gqlgen generate

# Start with GraphiQL (opens at http://localhost:8080/graphql)
GRAPHIQL=true make api-run

# Run subscription test
wscat -c ws://localhost:8080/graphql/ws \
  --subprotocol graphql-ws \
  -x '{"type":"connection_init","payload":{"Authorization":"Bearer <token>"}}'
```

## Testing

```bash
# Unit resolver tests
go test ./graph/... -v

# Integration: subscription round-trip
RUN_INTEGRATION=1 go test ./tests/... -run TestGraphQLSubscription
```

## Interview Angle

> "We added GraphQL as an opt-in facade over our existing REST services. The key engineering decision was N+1 prevention via DataLoaders — without them, a `messages` query for 50 messages with sender details would issue 51 DB queries. DataLoaders batch all sender lookups into a single `SELECT WHERE id = ANY($1)` per request."
