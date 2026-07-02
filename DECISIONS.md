# Decisions

This file records lightweight decisions that do not yet require a full ADR. Promote an item to `docs/adr/` when it affects architecture, reliability, scaling, security, or long-term maintainability.

## 2026-07-01

- EchoLine starts as a modular monolith with worker processes and event interfaces.
- PostgreSQL is the source of truth.
- Redis is used for cache, presence, and rate limiting.
- Redpanda/Kafka is used for async events after the core path exists.
- Future extensions are documented but should not distract from the main phase sequence.

## 2026-07-01 (review #02)

- Channel entitlement grant is admin-only; require-paid is owner-only (ADR 0030).
- Input validation centralized in `internal/validate` (username, display_name, message body).
- Frontend HTTP layer: `api/http.ts` with `authedRequest` + `parseResponse`; incremental migration from raw fetch.
- Business flows documented in `docs/business-flows.md`; engineering standards in `docs/engineering-standards.md`.

