# ADR 0030: Channel Entitlement Authorization

## Status

Accepted

## Context

Initial entitlement endpoints allowed any authenticated user to grant themselves channel access and any member to mark a channel as paid.

## Decision

1. `POST /api/channels/{id}/entitlements/require` — **channel owner only** (via `conversation_members.role = owner`).
2. `POST /api/channels/{id}/entitlements/grant` — **admin only** (`ADMIN_USER_IDS`); optional `user_id` in body for granting another user.
3. Payment `POST /api/payments/ledger/settle` with `reference: channel:{uuid}` — grants entitlement to the settling user (existing path).

## Consequences

- Self-serve grant abuse closed.
- Manual support grants require admin configuration.
- Owners control monetization flag without admin involvement.

## Files

- `backend/internal/entitlement/handler.go`
- `backend/internal/entitlement/handler_test.go`
- `backend/tests/integration_entitlement_test.go`
- `docs/business-flows.md`

## Verification

```bash
go test ./internal/entitlement/...
RUN_INTEGRATION=1 go test -run IntegrationEntitlement ./tests/...
```
