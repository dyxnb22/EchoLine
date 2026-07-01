# Decisions

This file records lightweight decisions that do not yet require a full ADR. Promote an item to `docs/adr/` when it affects architecture, reliability, scaling, security, or long-term maintainability.

## 2026-07-01

- EchoLine starts as a modular monolith with worker processes and event interfaces.
- PostgreSQL is the source of truth.
- Redis is used for cache, presence, and rate limiting.
- Redpanda/Kafka is used for async events after the core path exists.
- Future extensions are documented but should not distract from the main phase sequence.

