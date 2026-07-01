# Engineering Review #02 — Documentation, Architecture, Business Flows

Date: 2026-07-01  
Branch: `cursor/engineering-standards-758a`

## Scope

Second-pass review: document alignment, architecture audit, business flow calibration, deeper code review against modern SWE practices.

## Findings & fixes

### P0 — Security

| Issue | Fix |
|-------|-----|
| Any user could self-grant channel entitlement via `entitlements/grant` | Admin-only check (`ADMIN_USER_IDS`) |
| Any member could mark channel as paid | Owner-only check via `conversation_members.role` |
| No centralized input length validation | `internal/validate` package; auth register + message send |

### P1 — Documentation drift

| Issue | Fix |
|-------|-----|
| README still said "Phase 0" | Updated quick start + current phase |
| No single business flow doc | Added `docs/business-flows.md` with mermaid sequences |
| Architecture missing worker/entitlement | Updated `docs/architecture.md` |
| API docs incomplete on entitlement authz | Updated `docs/api.md` + ADR 0030 |
| No engineering standards doc | Added `docs/engineering-standards.md` |

### P1 — Frontend architecture

| Issue | Fix |
|-------|-----|
| Raw `fetch` duplicated; no server error surfacing | `api/http.ts`: `parseResponse`, `authedRequest`, `bindAuthFetch` |
| Auth refresh not used by API layer | `AuthContext` binds `authFetch` globally |
| Critical paths still on plain fetch | Migrated login, listConversations, sendMessage, subscribeChannel, grant |

### P2 — Test coverage

| Addition | Purpose |
|----------|---------|
| `entitlement/handler_test.go` | Owner/admin RBAC unit tests |
| `validate/limits_test.go` | Input boundary tests |
| `integration_entitlement_test.go` | Paid channel E2E with DB |

## Architecture assessment

**Strengths (unchanged):**

- Clear handler → service → repository layering
- Outbox + worker separation
- Explicit reliability docs

**Improved:**

- Authorization matrix documented for paid channels
- Validation at domain boundaries
- Frontend HTTP layer ready for full API migration

**Remaining gaps (accepted):**

- GraphQL prototype lacks schema-first codegen
- E2EE is demo-only
- Full `authedRequest` migration for all `api.ts` functions (incremental)
- Dual-client WS integration smoke needs Docker

## Module scores (post-fix)

| Module | Score | Notes |
|--------|-------|-------|
| auth + validate | A- | Length checks at register |
| entitlement | A | RBAC + tests + integration |
| message | A- | Body validation in service |
| frontend api | B+ | Partial migration to http layer |
| docs | A | business-flows + standards aligned |

## Verification commands

```bash
cd backend && go test ./...
cd frontend && npm run build && npx playwright test
RUN_INTEGRATION=1 DATABASE_URL=... go test -run Integration ./tests/...
```

## Recommended follow-ups

1. Migrate remaining `api.ts` functions to `authedRequest` + `parseResponse`
2. Add `message/service_test.go` for body length rejection
3. Local `make smoke-full` when Docker available
4. GraphQL auth hardening or deprecate in favor of REST
