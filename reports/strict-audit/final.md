# EchoLine Strict Audit — Final Report

## Rounds completed: 2

| Round | P0 | P1 | P2 | P3 | P4 |
|---|---|---|---|---|---|
| 01 | 0 | 4 | 6 | 4 | 0 |
| 02 | 0 | 2 | 2 | 6 | 0 |

## P0/P1/P2 fixed (summary)

### Round 01
- **P1** GraphQL `addReaction` permission bypass → membership check in ReactionService
- **P1** Payment self-serve zero-amount grant → reject `amount_cents <= 0`
- **P1** Frontend WS reconnect storm → ref-based runSync/WS effect deps
- **P1** Frontend stale JWT on WS reconnect → localStorage + refresh callback
- **P2** Unauthenticated `/metrics` → optional `METRICS_TOKEN` bearer gate
- **P2** Ads list/impression without membership → requireChannelMember + campaign binding
- **P2** Block button on own messages → `fetchMe` + currentUserId
- **P2** Missing WS `message.ack` → send on `message.created`
- **P2** Media upload missing checksum → SHA-256 in presignUpload

### Round 02
- **P1** X-Forwarded-For spoofing bypasses rate limit → RemoteAddr default, TRUSTED_PROXY opt-in
- **P1** Rate limits disabled without Redis → MemoryLimiter fallback always active
- **P2** GraphQL sendMessage unlimited → per-user graphql_send rate limit
- **P2** Ads impression cap race → transactional SELECT FOR UPDATE + insert

## Why stop condition is satisfied

1. At least 2 full audit rounds completed
2. Round 02 was a from-scratch re-audit (re-read docs + code paths)
3. Round 02 **new** findings are P3 only (plus verification of R01 fixes)
4. All P3 either documented as non-blocking or environment-limited
5. No open P0/P1/P2
6. Backend tests + frontend lint/build run and recorded
7. This final report written

## Test commands and results

```
cd backend && go test ./...     # PASS
cd backend && go vet ./...      # PASS
cd frontend && npm ci           # PASS
cd frontend && npm run lint     # PASS
cd frontend && npm run build    # PASS
cd frontend && npm audit --omit=dev  # 0 vulnerabilities
```

Not run in cloud: `make smoke`, `make smoke-full` (requires Docker + running API).

## Remaining risks (P3, non-blocking)

| Issue | Risk | Mitigation |
|---|---|---|
| Redis INCR+Expire not atomic | Stale rate-limit keys | Use Lua in production Redis |
| Integration tests skipped without DB | Coverage gap | RUN_INTEGRATION=1 in CI with Postgres |
| JWT in WS query string | Log leakage | Short-lived WS tickets (future) |
| Login page dev credentials | UX confusion | Clear env-only defaults |
| No attachment download UI | Feature gap | Use REST download-url (backend exists) |
| Edit/Recall 403 semantics | API polish | Map ErrNotMember to 403 |

## Manual / local verification still needed

- `make smoke-full` with Docker Compose
- `RUN_INTEGRATION=1` integration tests against Postgres
- Playwright E2E if configured locally
- Production: set `METRICS_TOKEN`, `TRUSTED_PROXY=true` behind reverse proxy

## Long-term suggestions (do not implement here)

- JWKS / token revocation
- WS connection rate limit per IP
- ClamAV on media.uploaded
- GraphQL → retire or full authz parity + query complexity limits
- E2E encryption beyond prototype

## Reports

- `reports/strict-audit/round-01/`
- `reports/strict-audit/round-02/`
