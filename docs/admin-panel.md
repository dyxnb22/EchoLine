# Admin Panel — EchoLine

## Problem

Operations and trust-and-safety teams need visibility into the system without direct DB access:

1. **Health monitoring** — API, DB, Redis, Kafka status at a glance.
2. **DLQ management** — view and replay failed events without a CLI.
3. **User management** — suspend, unsuspend, view user profile.
4. **Abuse review** — view reported messages and take action (warn, remove, ban).
5. **Audit log viewer** — who did what, when.

## Tradeoff

| Approach | Pros | Cons |
|----------|------|------|
| Separate admin SPA | Full UI flexibility | More to build and maintain |
| React tabs in main frontend | Reuse existing components | Admin routes must be guarded carefully |
| Third-party (Retool, Metabase) | Fast to set up | Vendor dependency, cost |

**Decision**: Admin panel as a guarded section of the main React frontend (route `/admin`, role-gated). This reuses existing auth, API client, and Tailwind components.

## Implementation Files

- `frontend/src/pages/AdminPanel.tsx` — main admin layout with tab navigation
- `frontend/src/pages/admin/HealthDashboard.tsx` — service health grid
- `frontend/src/pages/admin/DlqViewer.tsx` — DLQ list + replay button
- `frontend/src/pages/admin/UserList.tsx` — user search, suspend/unsuspend
- `frontend/src/pages/admin/ReportQueue.tsx` — reported content queue
- `frontend/src/pages/admin/AuditLog.tsx` — audit log table with filters
- `backend/internal/admin/handler.go` — existing health endpoint
- `backend/internal/admin/user_mgmt.go` — suspend/unsuspend (planned)
- `backend/internal/admin/report_handler.go` — report review actions (planned)

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/admin/health` | Deep health check (DB, Redis, Kafka) |
| `GET` | `/admin/dlq` | List DLQ events (paginated) |
| `POST` | `/admin/dlq/:id/replay` | Replay a DLQ event |
| `GET` | `/admin/users` | List users with search/filter |
| `POST` | `/admin/users/:id/suspend` | Suspend a user account |
| `POST` | `/admin/users/:id/unsuspend` | Unsuspend a user account |
| `GET` | `/admin/reports` | List open content reports |
| `POST` | `/admin/reports/:id/resolve` | Resolve a report (action: warn/remove/ban/dismiss) |
| `GET` | `/admin/audit-log` | Query audit log with filters |

All admin endpoints require `role: admin` in the JWT claims. The middleware checks this claim before routing.

## Health Dashboard

```
┌─────────────────────────────────────────────────────────────┐
│  EchoLine System Health                        2026-07-01   │
│                                                             │
│  ● API Server    OK    p95=12ms                             │
│  ● PostgreSQL    OK    connections=18/100                   │
│  ● Redis         OK    memory=42MB                          │
│  ● Kafka         OK    consumer_lag=0                       │
│  ● MinIO         OK    bucket=echoline                      │
│                                                             │
│  DLQ Events: 3 pending  [View DLQ]                          │
└─────────────────────────────────────────────────────────────┘
```

## DLQ Viewer

Shows each dead-letter event with:
- Event ID, type, timestamp
- Last error message
- Attempt count
- Replay button → `POST /admin/dlq/:id/replay`

## Role Guard

The admin panel is hidden from non-admin users at the frontend route level, and all admin API endpoints enforce server-side role checks. A stolen non-admin JWT cannot access any admin functionality.

## Testing

```bash
# Unit: admin role middleware rejects non-admin
go test ./internal/admin/... -run TestAdminRoleGuard

# E2E (Playwright)
npx playwright test tests/admin.spec.ts
```

## Interview Angle

> "We chose to embed the admin panel in the main frontend rather than a separate app. The key risk is accidental exposure of admin routes to regular users. We mitigate this with a server-side role claim check on every admin API endpoint — the frontend guard is only UX, not a security boundary."
