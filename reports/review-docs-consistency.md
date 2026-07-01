# Code Review: Documentation Consistency (M008)

**Reviewer**: Automated review via agent
**Date**: 2026-07-01
**Scope**: All docs in `docs/`, `reports/`, ADRs, OpenAPI spec, code comments

---

## Summary

EchoLine's documentation covers the key architectural areas with ADRs, API reference, data model, and websocket protocol docs. The following findings identify gaps where docs are inconsistent with the implemented code or with each other.

---

## Finding 1: `docs/api.md` Missing New Endpoints

**Severity**: Medium
**Files**: `docs/api.md`

**Observation**: Several endpoints implemented in recent iterations are not documented in `docs/api.md`:
- `GET /api/search/messages` (G007)
- `PATCH /api/messages/{id}` (C008)
- `POST /api/messages/{id}/recall` (C009)
- `POST /api/media/upload-url` (G001)
- `POST /api/media/download-url` (G002/G005)
- `GET /api/sync` (C006)

**Recommendation**: Add entries for each endpoint with:
- Method + path
- Auth requirement
- Request body schema (with field descriptions)
- Response schema
- Error codes

Also update `docs/openapi.yaml` to include all endpoints.

---

## Finding 2: `docs/websocket-protocol.md` Missing Recent Events

**Severity**: Medium
**Files**: `docs/websocket-protocol.md`

**Observation**: The WS protocol doc was written during Phase 2. Since then, the following events were added but not documented:
- `message.edited` (C008 — triggered when a message is edited)
- `message.recalled` (C009)
- `conversation.read` (C004 — when another device marks a conversation read)
- `typing.started` / `typing.stopped` (B011 — typing indicator)

**Recommendation**: Add entries for each event with: `type`, `payload` schema, directionality (client→server or server→client), and expected client behavior.

---

## Finding 3: `docs/data-model.md` Does Not Reflect Outbox and Audit Tables

**Severity**: Medium
**Files**: `docs/data-model.md`

**Observation**: The data model doc describes core tables (users, conversations, messages) but does not include:
- `outbox` table (transactional outbox)
- `audit_logs` table
- `device_sync_cursors` table
- `message_deliveries` table details

**Recommendation**: Add a section for "Infrastructure Tables" in `docs/data-model.md` covering these tables with their purpose, key columns, and relationships.

---

## Finding 4: ADR README Missing New ADRs (0004–0015)

**Severity**: Low
**Files**: `docs/adr/README.md`

**Observation**: The ADR README lists ADRs 0001–0003. ADRs 0004–0015 were added in Batch-100 but are not listed in the README.

**Recommendation**: Update `docs/adr/README.md` with entries for all new ADRs, including a one-line summary and status.

---

## Finding 5: `docs/architecture.md` Does Not Reference Worker Service

**Severity**: Low
**Files**: `docs/architecture.md`

**Observation**: The architecture doc describes the API server and may not mention the worker service (`cmd/worker`) as a separate process, or its role in consuming Kafka events and driving outbox draining.

**Recommendation**: Update the architecture diagram and narrative to include:
- Worker service (outbox drainer, fanout worker, search indexer)
- Its Kafka consumer relationship
- That it shares the same Postgres and Redis instances as the API

---

## Finding 6: `docs/openapi.yaml` — Auth Endpoints Have No Error Examples

**Severity**: Low
**Files**: `docs/openapi.yaml`

**Observation**: The OpenAPI spec documents happy paths but does not include examples of error responses (e.g., 401, 422, 429) with the error envelope schema.

**Recommendation**: Add `responses` sections for `400`, `401`, `422`, and `429` on all protected endpoints, with example error payloads:
```yaml
'401':
  description: Unauthorized
  content:
    application/json:
      schema:
        $ref: '#/components/schemas/Error'
      example:
        error:
          code: "unauthorized"
          message: "invalid or expired token"
```

---

## Finding 7: Interview Docs Not Cross-Referenced from README

**Severity**: Low
**Files**: `README.md`, `docs/interview-*.md`

**Observation**: The interview preparation guides (`docs/interview-*.md`) are high-value assets for the project but are not referenced from the main README or `docs/` index.

**Recommendation**: Add an "Interview Preparation" section in `README.md` linking to the interview guides. This makes them discoverable for reviewers and future contributors.

---

## Finding 8: `docs/reliability.md` Is in Chinese

**Severity**: Low (consistency)
**Files**: `docs/reliability.md`

**Observation**: `docs/reliability.md` is written entirely in Chinese, while all other docs in `docs/` are in English. This creates an inconsistency for English-speaking reviewers.

**Recommendation**: Add an English summary section at the top of `docs/reliability.md`, or translate it to English and note the original was in Chinese in a comment.

---

## Documentation Coverage by Area

| Area | Status | Gaps |
|------|--------|------|
| REST API | Partial | Missing 6+ endpoints |
| WebSocket Protocol | Partial | Missing 4+ events |
| Data Model | Partial | Missing infra tables |
| Architecture | Done | Worker not shown |
| ADRs | Done (0001–0015) | README not updated |
| Reliability | Done (Chinese) | English translation needed |
| Scaling | Done | — |
| Security Checklist | Done | — |
| Interview Guides | Done | Not cross-referenced |
| Load Tests | Done | k6 scripts have inline comments |

---

## Overall Assessment

**Documentation consistency score**: 7/10. Core architectural decisions are well-documented. The primary gaps are: API doc for recent endpoints, WS protocol for recent events, and data model for infrastructure tables. These are medium-priority for external reviewers.

## Priority Fixes

1. **Finding 1** (MEDIUM): Update `docs/api.md` with missing endpoints.
2. **Finding 2** (MEDIUM): Update `docs/websocket-protocol.md` with missing events.
3. **Finding 3** (MEDIUM): Add infrastructure tables to `docs/data-model.md`.
4. **Finding 4** (LOW): Update ADR README.
5. **Finding 8** (LOW): Translate `docs/reliability.md` to English.

## Files to Update

- `docs/api.md` — add 6+ missing endpoints
- `docs/websocket-protocol.md` — add 4+ missing events
- `docs/data-model.md` — add outbox, audit_logs, deliveries, sync_cursors
- `docs/adr/README.md` — add ADR 0004–0015 entries
- `docs/architecture.md` — add worker service to diagram
- `README.md` — link to interview guides
