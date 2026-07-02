# ADR 0022: GraphQL API Layer with Subscriptions

## Status

Accepted (design; implementation deferred to extension phase)

## Problem

EchoLine's current API is REST + custom WebSocket JSON. As the feature surface grows (reactions, threads, recommendations, ads), third-party developers and the mobile team have requested a **GraphQL API** because:

1. REST over-fetches for mobile (e.g., fetching a conversation list returns full message objects when only title + unread count is needed).
2. The WS protocol is bespoke and requires client-side knowledge of frame types.
3. GraphQL Subscriptions provide a standard real-time mechanism that tooling (Apollo, urql) handles automatically.

The design questions are:

- GraphQL-over-HTTP vs. replacing REST entirely?
- How do GraphQL Subscriptions integrate with the existing WS hub?
- How do we prevent N+1 queries (the classic GraphQL performance pitfall)?
- Schema design: single schema or federation?

## Decision

### Approach: GraphQL as Additional Layer (not Replace REST)

The REST API remains the primary internal API. GraphQL is an **additional facade** targeting:

- Third-party developers
- The React frontend (progressive migration)
- Mobile clients (future)

REST endpoints remain for the WebSocket gateway (which speaks its own protocol), admin endpoints, and webhook delivery.

### Transport: GraphQL over HTTP + WebSocket (graphql-ws protocol)

```
HTTP POST /graphql          — queries and mutations
WebSocket /graphql/ws       — subscriptions (graphql-ws subprotocol)
```

The `graphql-ws` protocol is the standard (supersedes the older `subscriptions-transport-ws`). Apollo Client, urql, and Relay all support it natively.

### Schema Design

```graphql
type Query {
  me: User!
  conversation(id: ID!): Conversation
  conversations(limit: Int = 20, cursor: String): ConversationPage!
  messages(conversationId: ID!, limit: Int = 50, before: String): MessagePage!
  search(query: String!, limit: Int = 20): [SearchResult!]!
  recommendations: Recommendations!
}

type Mutation {
  sendMessage(conversationId: ID!, text: String!, clientMsgId: String!): Message!
  editMessage(messageId: ID!, text: String!): Message!
  recallMessage(messageId: ID!): Message!
  addReaction(messageId: ID!, emoji: String!): Reaction!
  removeReaction(messageId: ID!, emoji: String!): Boolean!
  createConversation(peerId: ID!): Conversation!
  createGroup(name: String!): Group!
}

type Subscription {
  messageAdded(conversationId: ID!): Message!
  messageEdited(conversationId: ID!): Message!
  typingIndicator(conversationId: ID!): TypingEvent!
  reactionAdded(conversationId: ID!): ReactionEvent!
  presenceChanged(userId: ID!): PresenceEvent!
}
```

### N+1 Prevention: DataLoader

Every resolver that fetches related objects uses a **DataLoader** to batch DB calls:

```go
// Without DataLoader: N+1 (one DB query per message for sender)
for _, msg := range messages {
    msg.Sender = db.GetUser(msg.SenderID)  // N queries
}

// With DataLoader: 1 batched query
loader := userLoader.Load(ctx, senderIDs)  // 1 query for all senders
```

The Go library `graph-gophers/dataloader` provides this. Loaders are scoped per-request (one DataLoader instance per HTTP request).

### Subscription Fan-out Architecture

GraphQL Subscriptions use the existing in-process event bus (the WS hub). A subscription resolver subscribes to the in-memory channel for a conversation:

```
Client opens WebSocket /graphql/ws
  → sends `subscribe` frame for `messageAdded(conversationId: "...")`
  → GraphQL engine calls Subscription.messageAdded resolver
  → Resolver registers a listener on hub.conversationChannel("...")
  → On new message event: resolver pushes to GraphQL subscription stream
  → graphql-ws serialises and sends to client
```

This reuses the existing hub without duplication. The GraphQL layer is a thin adapter.

### Authorization

Every resolver checks the claims in the request context (same JWT middleware as REST). Subscription resolvers re-validate the JWT on connect (the `graphql-ws` `connectionInit` event carries an `Authorization` payload).

## Implementation Files

- `backend/graph/schema.graphql` — full schema
- `backend/graph/resolver.go` — root resolver wiring
- `backend/graph/query.go`, `mutation.go`, `subscription.go` — resolver implementations
- `backend/graph/loader/` — DataLoader instances
- `backend/cmd/api/main.go` — mount `/graphql` and `/graphql/ws` routes
- `docs/graphql-prototype.md` — client usage guide

## Testing

- Unit: resolver query tests with DB mock.
- Unit: DataLoader batching test (assert single DB call for N objects).
- Integration: subscription test — send message via mutation, assert subscription receives event.

## Interview Talking Points

- **Why not replace REST?** "REST is already integrated in the WS hub, admin tools, and CI smoke tests. Adding GraphQL as a facade means mobile and third-party developers get a better API without breaking existing consumers. We migrate the React frontend incrementally."
- **N+1**: "Every relationship field (message → sender, group → members) goes through a DataLoader that batches DB calls per request. Without this, a messages query for 50 messages would issue 51 DB queries."
- **Subscriptions vs. existing WS**: "We keep both. The existing bespoke WS is simpler and lower overhead for the main app. GraphQL subscriptions use the graphql-ws protocol and are better for third-party SDKs and code generation."
