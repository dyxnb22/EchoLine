# Code Review: API Consistency (M001)

> **Historical note (2026-07-02):** 本报告引用 `backend/internal/api/` 为 Phase 1 设计路径。当前 handler 位于 `backend/internal/*/handler.go` 与 `auth/service.go`。

**Reviewer**: Automated review via agent
**Date**: 2026-07-01
**Scope**: All REST API endpoints (originally `backend/internal/api/`)

---

## Summary

The EchoLine REST API is generally consistent in structure, authentication, and error handling. Several minor inconsistencies are identified below with severity ratings and recommended fixes.

---

## Finding 1: Mixed Pagination Parameters

**Severity**: Medium
**Files**: `backend/internal/api/message.go`, `backend/internal/api/conversation.go`

**Observation**: Message history pagination uses `before_id` (cursor-based), while conversation sync uses `after_seq`. These are both valid pagination strategies but use different parameter names and semantics.

**Recommendation**: Standardize on cursor-based pagination across all list endpoints. Use `cursor` as the parameter name and `next_cursor` in the response. Document this in `docs/api.md`.

**Current state**: Functional; inconsistency only matters when building SDK clients.

---

## Finding 2: Inconsistent Response Envelope Fields

**Severity**: Low
**Files**: `backend/internal/api/errors.go`, `backend/internal/api/message.go`

**Observation**: Success responses sometimes return data directly (`{"id": "...", "body": "..."}`) and sometimes wrap in `{"data": {...}}`. Error responses consistently use `{"error": {"code": "...", "message": "..."}}`.

**Recommendation**: Adopt a consistent success envelope: `{"data": {...}, "meta": {...}}` for all list responses, and `{"data": {...}}` for single-resource responses. Update `docs/api.md`.

---

## Finding 3: Missing `Location` Header on 201 Created

**Severity**: Low
**Files**: `backend/internal/api/auth.go`, `backend/internal/api/conversation.go`

**Observation**: `POST /api/auth/register` and `POST /api/conversations/group` return 201 Created but do not include a `Location` header pointing to the newly created resource.

**Recommendation**: Add `Location: /api/users/{id}` and `Location: /api/conversations/{id}` headers on 201 responses. This follows REST convention and allows clients to discover the resource URI.

---

## Finding 4: No `X-Request-ID` in Error Responses

**Severity**: Medium
**Files**: `backend/internal/api/errors.go`

**Observation**: Error responses include `error.code` and `error.message` but do not include the request ID. Users cannot correlate a client-side error message with a server log entry.

**Recommendation**: Include `"request_id": "<trace_id>"` in all error response bodies. The `X-Trace-ID` value is already available in the request context (from `backend/internal/middleware/trace.go`).

---

## Finding 5: Inconsistent HTTP Methods for State Transitions

**Severity**: Medium
**Files**: `backend/internal/api/message.go`, `backend/internal/api/ack.go`

**Observation**:
- Message recall uses `POST /api/messages/{id}/recall` (action endpoint).
- Message edit uses `PATCH /api/messages/{id}` (REST resource update).
- Message ACK uses `POST /api/messages/{id}/ack` (action endpoint).
- Mark conversation read uses `POST /api/conversations/{id}/read` (action endpoint).

**Recommendation**: For idempotent state transitions (mark read, ACK), consider using `PUT` with the target state in the body. For non-idempotent actions (recall), `POST` to an action URL is acceptable. Document the convention in `docs/api.md`.

---

## Finding 6: Missing Rate Limit Headers

**Severity**: Low
**Files**: `backend/internal/middleware/ratelimit.go`

**Observation**: The rate limit middleware returns 429 Too Many Requests when the limit is hit, but does not include `X-RateLimit-Limit`, `X-RateLimit-Remaining`, or `Retry-After` headers.

**Recommendation**: Add rate limit headers on all responses:
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 59
X-RateLimit-Reset: 1717200000
```
This allows clients to implement proactive backoff.

---

## Finding 7: No API Versioning

**Severity**: Medium (future risk)
**Files**: All API handlers

**Observation**: All endpoints are under `/api/` with no version prefix (e.g., `/api/v1/`). Any breaking change will require a flag day or complex compatibility shims.

**Recommendation**: Prefix all routes with `/api/v1/`. When a breaking change is needed, add `/api/v2/` routes alongside `/api/v1/`. This is a low-cost change now; expensive to retrofit later.

---

## Finding 8: Auth Token Not Refreshed on 401

**Severity**: Low (DX)
**Files**: `backend/internal/api/auth.go`

**Observation**: When a token expires, clients receive a 401 response with `error.code = "unauthorized"`. The response does not suggest refreshing the token or provide a `WWW-Authenticate` header.

**Recommendation**: Add `WWW-Authenticate: Bearer realm="echoline", error="token_expired"` on 401s caused by expired tokens (distinct from invalid tokens). This allows clients to distinguish "refresh needed" from "login required".

---

## Overall Assessment

**API consistency score**: 7/10. Core functionality is correct and consistent. The above findings are refinements that improve developer experience and REST compliance without requiring breaking changes. Priority order: Finding 4 (request ID in errors) → Finding 6 (rate limit headers) → Finding 3 (Location headers) → Finding 2 (response envelope) → Finding 1 (pagination) → Finding 7 (versioning).

## Files to Update

- `backend/internal/api/errors.go` — add request_id, rate limit headers
- `backend/internal/middleware/ratelimit.go` — add rate limit response headers
- `backend/internal/api/auth.go` — WWW-Authenticate header, Location on 201
- `backend/internal/api/conversation.go` — Location on 201
- `docs/api.md` — document conventions, pagination, envelope format
