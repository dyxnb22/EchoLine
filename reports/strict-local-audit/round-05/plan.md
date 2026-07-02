# Round 05 — Strict Local Audit Plan

**Date:** 2026-07-02
**Session:** Fresh full audit (do NOT trust rounds 01–04 or `final.md` conclusions)

## Scope

Complete zero-based audit per strict-local-audit protocol:

### Backend
- Auth (register/login/refresh/JWT/device)
- Conversations (direct/group/channel, permissions, entitlements)
- Messages (send/edit/recall/forward/reactions/threads)
- WebSocket (auth, send, ACK, typing, fanout)
- Sync/offline/cursors/unread
- Outbox/worker/DLQ
- Media upload/download
- Search, payment, ads, admin RBAC, audit, rate limit

### Frontend
- Auth flow, token refresh, WS reconnect
- Chat UX, optimistic send, pagination, sync
- Group/channel/paid flows
- API/WS field consistency

### Data & Tests
- Migrations, FK/UNIQUE, seq/idempotency
- Integration/smoke/E2E authenticity
- CI skip masking

## Verification commands (this round)

```bash
make dev-up
make test
cd backend && go test ./...
cd backend && go vet ./...
cd frontend && npm ci && npm run lint && npm run build
RUN_INTEGRATION=1 cd backend && go test ./tests/... -count=1
make smoke-full
cd frontend && npx playwright test
```

## Evidence priority

1. Runtime behavior
2. Test results
3. Source code
4. Schema/migrations
5. Product docs
6. Previous reports (untrusted)

## Output files

- `findings.md`
- `fix-log.md`
- `test-results.md`
- `round-summary.md`
