# Strict Local Audit — Final Report

**Date:** 2026-07-02  
**Branch:** cursor/strict-local-audit-90b4

## Audit rounds

| Round | P0 | P1 | P2 | P3 | P4 | P5 |
|-------|----|----|----|----|----|-----|
| 01 (found) | 1 | 5 | 9 | 3 | 0 | 0 |
| 01 (fixed) | 1 | 5 | 9 | 1 | 0 | 0 |
| 02 (new) | 0 | 0 | 0 | 3 | 1 | 0 |

## P0/P1/P2 fix summary

All P0/P1/P2 from round 01 were fixed and re-verified in round 02:

1. **Docker migrations (P1)** — API container starts; smoke-full runs against live stack
2. **Outbox stuck processing (P1)** — processing_started_at + reclaim worker
3. **Attachment download missing (P0)** — frontend presignDownload + UI
4. **Unread while viewing (P1)** — optimistic clear + refresh after mark-read
5. **WS reconnect sync race (P1)** — sync re-trigger when conversations load
6. **Payment self-serve in compose (P1)** — disabled by default
7. **ACK seq forgery (P2)** — bind to message seq; cap mark-read at latest_seq
8. **Media owner bypass (P2)** — membership required for linked attachments
9. **WS rate limit bypass (P2)** — conv_send limiter on message.send
10. **WS CheckOrigin (P2)** — localhost allowlist + WS_ALLOWED_ORIGINS
11. **Sync next_seq (P2)** — pagination cursor field added
12. **Smoke script drift (P2)** — UUID client_msg_id, correct group API
13. **Refresh token reuse (P2)** — JTI consumption store
14. **OpenSearch authz (P2)** — live membership re-check on fallback hits
15. **Register/refresh unrated (P3→fixed)** — rate limit middleware added

## Why stop condition is met

1. ✅ At least 2 full audit rounds completed
2. ✅ Round 02 was independent full audit
3. ✅ Round 02 new findings are P3/P4 only
4. ✅ No open P0/P1/P2
5. ✅ All discovered P0/P1/P2 fixed and verified
6. ✅ `make test` — PASS
7. ✅ `go test ./...` — PASS
8. ✅ `go vet ./...` — PASS
9. ✅ `npm run lint` — PASS
10. ✅ `npm run build` — PASS
11. ✅ `make smoke-full` — PASS (product path verified)
12. ✅ This final report written

## Remaining P3/P4/P5

| ID | Priority | Summary |
|----|----------|---------|
| SLA-R02-001 | P3 | Conversation switch loading flash |
| SLA-R02-002 | P3 | Sync errors swallowed in UI |
| SLA-R02-003 | P3 | Integration tests opt-in (RUN_INTEGRATION=1) |
| SLA-R02-004 | P4 | k6 not installed locally |
| SLA-R02-005 | P3 | Presigned URL sharing accepted for MVP |

## Test commands and results

```
make test                          PASS
cd backend && go test ./...        PASS
cd backend && go vet ./...         PASS
cd frontend && npm run lint        PASS
cd frontend && npm run build       PASS
cd frontend && npm audit --omit=dev PASS (0 vulns)
make smoke-full                    PASS (16/16, 1 skip WS optional)
cd frontend && npx playwright test PASS (4/4)
k6 run --dry-run loadtests/*.js    SKIP (k6 not on PATH)
```

## Local environment

- Docker/OrbStack: available and used
- Full stack: postgres, redis, redpanda, minio, api, worker
- No blockers for local verification

## Follow-up suggestions

1. Add `setMessages([])` + loading state on conversation switch (P3)
2. Run integration tests in CI with Postgres service + RUN_INTEGRATION=1
3. Install k6 locally or document in dev setup
4. Optional: RUN_WS_SMOKE=1 in smoke-full for WS path
5. Redis-backed refresh store for multi-instance deployments

## Reports

- `reports/strict-local-audit/round-01/`
- `reports/strict-local-audit/round-02/`
