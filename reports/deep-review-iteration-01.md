# Deep Review — Iteration 01

Date: 2026-07-02  
Scope: Full-stack quality audit per product docs and reliability semantics.

## Review plan

### Business flows to audit

1. Auth (register/login/refresh/session/device)
2. Direct / group / channel messaging
3. Conversation list + unread
4. History pagination + offline sync (`POST /api/sync`)
5. WebSocket realtime (send/push/ACK/typing)
6. Message edit/recall
7. Attachments upload/download
8. Paid channel subscribe + entitlements
9. Search navigation
10. Admin/DLQ/audit (RBAC)

### Code modules

| Area | Paths |
|------|-------|
| Sync | `backend/internal/sync/` |
| Messages | `backend/internal/message/` |
| Delivery/ACK | `backend/internal/delivery/` |
| Media | `backend/internal/media/` |
| Outbox | `backend/internal/outbox/`, `backend/internal/worker/` |
| Realtime | `backend/internal/realtime/` |
| Conversations | `backend/internal/conversation/` |
| Frontend chat | `frontend/src/pages/ChatPage.tsx`, `frontend/src/api.ts`, `frontend/src/api/http.ts` |

### Tests to run

```bash
make test                    # go test ./... + frontend build
cd frontend && npm run test:e2e  # Playwright (mocked)
```

Integration smoke (`make smoke-full`) blocked — no Docker/Postgres in cloud VM (`BLOCKERS.md`).

### Top 10 risk checkpoints

1. Sync cursor + `has_more` pagination (message loss on reconnect)
2. Idempotent send side effects (`latest_seq` bump + re-broadcast)
3. `client_msg_id` cross-conversation collision
4. Attachment download RBAC for group recipients
5. ACK `message_id` / `conversation_id` binding
6. Outbox duplicate publish under multi-worker
7. Frontend sync loop ignoring `has_more`
8. Paid channel 402 detection (error code vs message text)
9. WS reconnect with stale JWT
10. Cached conversation list omitting `can_publish`

### Judgment basis

- `docs/reliability.md` — persistence-first, idempotency, sync compensation
- `docs/websocket-protocol.md` — envelope types, dedup keys
- `docs/api.md` — REST shapes, payment_required
- `docs/data-model.md` — seq, deliveries, attachments
- `ACCEPTANCE_MATRIX.md` — capability expectations
- Prior reports under `reports/review-*.md` (verify fix status in code)

## Iteration 01 outcome

See `reports/deep-review-issues.md` for findings. P0/P1/P2 fixes applied in this branch; re-review in iteration 02.
