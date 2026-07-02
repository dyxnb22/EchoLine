# Strict Local Audit — Final Report

**Date:** 2026-07-02  
**Branch:** cursor/strict-local-audit-90b4  
**Session rounds:** 05, 06, 07 (fresh audit; prior rounds 01–04 not trusted)

## Round counts

| Round | P0 | P1 | P2 | P3 | P4 |
|-------|----|----|----|----|-----|
| 05 found | 0 | 5 | 4 | 3 | 0 |
| 06 found | 0 | 2 | 4 | 0 | 0 |
| 07 new | 0 | 0 | 0 | 4 | 2 |

## P0/P1/P2 fix summary

### Round 05
1. Frontend unread race while viewing active conversation
2. WS reconnect churn on token rotation
3. Multi-conversation sync on reconnect
4. WS connect after refresh failure
5. PAYMENT_SELF_SERVE blocked in production
6. GraphQL disabled by default in production
7. Rate limits on sync/ack/search/export/media
8. Integration test JWT_SECRET default
9. Shared device_id helper

### Round 06
1. GraphQL sendMessage client_msg_id generation
2. Sync cursor cap (50)
3. Audit IP uses TRUSTED_PROXY gate
4. WS typing rate limit
5. Notification poll decoupled from message stream
6. Reaction prefetch capped to 20 messages

## Stop condition

1. ≥2 full audit rounds — **yes** (05, 06, 07)
2. Latest round (07) zero new P0/P1/P2 — **yes**
3. No open P0/P1/P2 — **yes**
4. All tests pass — **yes**
5. `make smoke-full` pass — **yes**
6. `final.md` written — **yes**

## Latest test commands

```
make test                          PASS
go test ./...                      PASS
go vet ./...                       PASS
npm run lint                       PASS
npm run build                      PASS
RUN_INTEGRATION=1 go test ./tests  PASS
make smoke-full                    PASS
npx playwright test                PASS (4/4)
```

## Remaining P3/P4

- Sync error UI silent
- Conversation switch loading flash
- WS JWT in query string (MVP)
- Dev in-memory rate limits without Redis
- k6/wscat not installed locally
- Reaction batch API (prefetch capped)

## Local environment

- Docker/Postgres/Redis available — full stack verified
- k6, wscat, GNU timeout not installed — WS/k6 smoke skipped

## Follow-up suggestions

1. Add batch reactions API
2. Install wscat in dev docs for WS smoke
3. Short-lived WS ticket instead of query JWT
4. Require Redis in multi-instance deployment docs
