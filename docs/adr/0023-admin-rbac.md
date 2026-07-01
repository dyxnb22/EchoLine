# ADR 0023: Admin RBAC

## Status

Accepted

## Context

Admin endpoints (`/api/admin/*`) initially accepted any authenticated JWT. Production requires role-based access control.

## Decision

- Prototype: `ADMIN_USER_IDS` env (comma-separated UUIDs) checked by `admin.RequireAdmin` middleware.
- Migration `00014` adds `users.is_admin` column for future DB-backed roles.
- DLQ list/replay and user/report/audit endpoints require admin when allowlist is configured.

## Consequences

- Simple to configure in dev/CI.
- Production should migrate to DB `is_admin` + audit on every admin action.

## Files

- `backend/internal/admin/middleware.go`
- `backend/migrations/00014_admin_webhook_ads.sql`
