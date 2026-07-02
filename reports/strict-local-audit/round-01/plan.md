# Strict Local Audit — Round 01 Plan

**Date:** 2026-07-02
**Mode:** Zero-trust full audit (treat as first audit; ignore prior completion claims)

## Stop Condition (global)

Latest full audit finds only P3/P4/P5 (or no issues). No open P0/P1/P2.

## Round 01 Objectives

1. Read all boundary docs fresh (no trust in DONE/manifest/reports).
2. Bring up local stack (`make dev-up`, `make dev-app` if needed).
3. Run full verification matrix.
4. Audit backend, frontend, data consistency, test authenticity.
5. Record all findings with evidence.
6. Fix every P0/P1/P2 before round ends.
7. Re-run verification after fixes.

## Verification Commands (this round)

| Command | Purpose |
|---------|---------|
| `make dev-up` | Postgres, Redis, Redpanda, MinIO |
| `make dev-app` | API + worker containers |
| `make test` | Backend unit tests |
| `cd backend && go vet ./...` | Static analysis |
| `cd frontend && npm ci && npm run lint` | Frontend lint |
| `cd frontend && npm run build` | Frontend build |
| `cd frontend && npm audit --omit=dev` | Dependency audit |
| `make smoke-full` | Full API smoke |
| `cd frontend && npx playwright test` | E2E (if stack up) |
| `k6 run --dry-run loadtests/k6-api-send.js` | Load test syntax |
| `k6 run --dry-run loadtests/k6-ws-connect.js` | WS load test syntax |

## Audit Scope Checklist

### Backend
- [ ] Auth: register, login, refresh, JWT validation
- [ ] User profile, device/session
- [ ] Direct chat dedup, group CRUD + roles, channel + paid entitlement
- [ ] Payment ledger, ads, recommendations, encryption key prototype
- [ ] Conversations, unread, history pagination, search
- [ ] Message send/edit/recall/delete/forward/reactions/threads
- [ ] Attachments upload/download
- [ ] WebSocket auth, send, ACK, typing, edit/recall fanout
- [ ] Offline sync, device cursor, ACK state machine
- [ ] Outbox, worker, DLQ
- [ ] Admin RBAC, audit log, rate limit, risk/spam, metrics

### Frontend
- [ ] Auth flow + token refresh + WS reconnect
- [ ] Conversation list, unread, messaging flows
- [ ] Optimistic send, failure states, history, offline sync
- [ ] Search, attachments, edit/recall realtime, reactions, threads, forward
- [ ] Group settings, notification panel, admin panel
- [ ] API/WS field consistency with backend

### Data & Consistency
- [ ] Migrations, FK, UNIQUE, indexes
- [ ] message_id + conversation_id binding
- [ ] client_msg_id idempotency scope, seq allocation
- [ ] unread calculation, sync cursor gaps
- [ ] ACK forgery, outbox duplicate/stuck
- [ ] Search authorization, cache completeness
- [ ] Payment/entitlement forgery, presigned URL risks

### Test Authenticity
- [ ] Integration tests run real flows
- [ ] Smoke paths match live API
- [ ] E2E uses real services (not mocks-only)
- [ ] Skipped tests not masking P1/P2

## Evidence Priority

1. Runtime behavior
2. Test results
3. Source code
4. DB schema/migrations
5. Product docs
6. ADR/reports (reference only)

## Deliverables (round-01)

- `findings.md` — all issues with SLA-R01-NNN IDs
- `fix-log.md` — every P0/P1/P2 fix
- `test-results.md` — command outputs
- `round-summary.md` — counts and next steps

## Decision Gate

- If any P0/P1/P2 found → fix → verify → **Round 02** (fresh full audit)
- If only P3/P4/P5 → need Round 02 anyway (minimum 2 rounds per stop condition)
