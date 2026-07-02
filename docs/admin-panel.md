# Admin Panel — EchoLine

## Problem

Operations and trust-and-safety teams need visibility into the system without direct DB access: users, abuse reports, and DLQ replay.

## Current implementation (2026-07-02)

The admin UI is an **overlay panel** in the main chat app (not a separate `/admin` route):

| Component | Path | Features |
|-----------|------|----------|
| `AdminPanel` | `frontend/src/components/AdminPanel.tsx` | Users list, reports list, DLQ list + replay |
| API client | `frontend/src/api.ts` | `adminListUsers`, `adminListReports`, `adminListDLQ`, `adminReplayDLQ` |

Access: user clicks **Admin** in chat header; backend returns `403` if not in `ADMIN_USER_IDS` or `users.is_admin`.

## API Endpoints (implemented)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/admin/users` | List users (admin only) |
| `GET` | `/api/admin/reports` | List abuse reports |
| `GET` | `/api/admin/dlq` | List dead letter events (`dead_letters` key) |
| `POST` | `/api/admin/dlq/{id}/replay` | Re-queue event into outbox |
| `GET` | `/api/admin/health` | Health skeleton (any authenticated user) |
| `GET` | `/api/admin/audit-logs` | Audit log list (admin only) |

## Planned (not yet in UI)

- Health dashboard tab
- User suspend/unsuspend
- Report resolve actions
- Dedicated `/admin` route with react-router

## CLI

`backend/cmd/replay` — list/replay DLQ via API (`/api/admin/dlq`) or direct DB requeue into `outbox_events`.

## Verification

```bash
export ADMIN_USER_IDS=<your-user-uuid>
make dev-up && make dev-app
# Login → Admin button → DLQ replay
go run ./cmd/replay --list  # requires ECHOLINE_TOKEN
```
