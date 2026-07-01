# Review Plan

EchoLine should go through multiple review and fix rounds. Review is not a final polish step; it is part of the long-running engineering loop.

## Review Rounds

### Round 1：Backend API Review

Scope:

- auth,
- user,
- device,
- conversation,
- message REST API.

Checklist:

- consistent error format,
- input validation,
- permission checks,
- DB constraints,
- tests for failure cases.

### Round 2：Realtime and Reliability Review

Scope:

- WebSocket,
- ACK,
- retry,
- idempotency,
- seq ordering,
- offline sync.

Checklist:

- no message loss after DB success,
- duplicate client requests dedupe,
- ACK state only moves forward,
- reconnect compensation works.

### Round 3：Cache, MQ, Worker Review

Scope:

- Redis,
- event bus,
- outbox,
- workers,
- DLQ.

Checklist:

- DB remains source of truth,
- event publish failure has recovery path,
- consumers are idempotent,
- cache invalidation is documented.

### Round 4：Security and Governance Review

Scope:

- auth,
- permissions,
- rate limiting,
- risk,
- audit.

Checklist:

- unauthorized users cannot access conversations,
- admin operations are audited,
- abusive actions are limited,
- sensitive errors do not leak internals.

### Round 5：Performance and Scaling Review

Scope:

- message write path,
- history pagination,
- fanout,
- search,
- WebSocket gateway.

Checklist:

- indexes match query patterns,
- hot paths have metrics,
- load tests exist,
- bottlenecks are documented.

