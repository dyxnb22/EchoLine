# Deep Review — Iteration 03 (Full-Scope)

Date: 2026-07-02  
**Policy:** Every iteration is a **full-project** audit — all backend domains, all frontend flows, architecture/docs/test gaps. Not a spot-check of prior fixes only.

## Full-scope checklist

### Backend (15 domains)

| # | Domain | Reviewed |
|---|--------|----------|
| 1 | auth / JWT / refresh | yes |
| 2 | user / device | yes |
| 3 | conversation / RBAC | yes |
| 4 | message / seq / idempotency | yes |
| 5 | sync / unread / cursors | yes |
| 6 | delivery / ACK | yes |
| 7 | realtime / WebSocket | yes |
| 8 | media / attachments | yes |
| 9 | search | yes |
| 10 | notification | yes |
| 11 | rate_limit / risk / audit | yes |
| 12 | worker / outbox / MQ / DLQ | yes |
| 13 | admin / entitlement / payment | yes |
| 14 | cache / presence / Redis | yes |
| 15 | config / security | yes |

### Frontend (14 flows)

| # | Flow | Reviewed |
|---|------|----------|
| 1 | login / register / refresh / logout | yes |
| 2 | conversation list / filters / archived | yes |
| 3 | direct / group / channel | yes |
| 4 | send (text / attachment / optimistic) | yes |
| 5 | WS receive (created / edited / recalled / typing) | yes |
| 6 | history pagination | yes |
| 7 | sync / reconnect | yes |
| 8 | unread badges | yes |
| 9 | search navigation | yes |
| 10 | paid channel / payment | yes |
| 11 | reactions / threads / edit / recall | yes |
| 12 | notifications / settings / admin | yes |
| 13 | error / loading / empty states | yes |
| 14 | API / WS protocol parity | yes |

### Architecture & docs

- Module boundaries, Redis vs DB, race conditions, indexes, test gaps, docs vs code — all reviewed.

## Tests planned

```bash
go test ./...
cd frontend && npm run build
```

`make smoke-full` — blocked (no Docker).

## Top 10 risk checkpoints (iteration 03)

1. Redis conversation cache invalidation (ADR 0005)
2. Outbox `processing` stuck rows (ISSUE-020 escalation)
3. Payment self-settle entitlement grant
4. WS `message.edited`/`recalled` payload field mismatch (regression)
5. Archived API response shape vs frontend
6. Search cross-conversation message merge
7. MarkRead seq inflation past `latest_seq`
8. Sync response missing attachment metadata
9. Pin/report without message-in-conversation check
10. Logout leaving in-memory chat state

## Judgment basis

`docs/reliability.md`, `docs/api.md`, `docs/websocket-protocol.md`, `docs/data-model.md`, `docs/architecture.md`, ADR 0005, prior `reports/deep-review-*.md`.

## Outcome

See `reports/deep-review-issues.md` ISSUE-023+. P1/P2 fixes applied in this branch; iteration 04 required if any P0/P1/P2 remain.
