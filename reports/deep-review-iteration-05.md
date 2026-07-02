# Deep Review — Iteration 05 (Remaining P2/P3/P4)

Date: 2026-07-02  
**Policy:** Close open items from iteration 04 stop-condition gap.

## Fixes

| ID | Priority | Status | Fix |
|----|----------|--------|-----|
| ISSUE-044 | P2 | fixed | `CloneUnlinkedForForward` + S3 `CopyObject`; `Forward` passes `object_key` into `Send` |
| ISSUE-045 | P2 | fixed | Thread reply requires client `client_msg_id`; `SendReply` idempotent pre-lookup |
| ISSUE-047 | P3 | fixed | `JWT_SECRET` must be ≥ 32 chars (`config.Load`) + `TestLoadWeakJWTSecret` |
| ISSUE-048 | P3 | fixed | Register rate limit 10/min per IP (`applyRateLimits`) |
| ISSUE-049 | P4 | fixed | Conversation list + message pane loading/empty hints (`ChatPage.tsx`) |

## Files

- `backend/internal/media/repository.go`, `client.go`
- `backend/internal/message/service.go`
- `backend/internal/thread/handler.go`
- `backend/internal/config/config.go`, `config_test.go`
- `backend/internal/server/options.go`
- `backend/internal/graph/handler.go`
- `frontend/src/api.ts`, `ThreadPanel.tsx`, `ChatPage.tsx`
- `docs/api.md`

## Tests

- `go test ./...` — pass
- `npm run build` — pass

## Counts (iteration 05)

| Priority | Found | Fixed | Wontfix | Remaining |
|----------|-------|-------|---------|-----------|
| P0 | 0 | 0 | 0 | 0 |
| P1 | 0 | 0 | 0 | 0 |
| P2 | 0 | 2 | 0 | 0 |
| P3 | 0 | 2 | 0 | 0 |
| P4 | 0 | 1 | 0 | 0 |

**Stop condition:** Met for actionable items — no open P0/P1/P2; P3 hardening done; P4 polish done. Wontfix P2 from iter 04 remain documented (WS rate limit, CheckOrigin, notifications, client ACK).
