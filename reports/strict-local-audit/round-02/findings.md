# Round 02 Findings

Fresh audit — no reliance on round-01 closure claims.

## SLA-R02-001: Conversation switch shows stale messages briefly

Priority: P3
Status: open
Area: frontend
Business flow: Switch active conversation
Expected: Loading state or immediate clear
Actual: Previous messages remain until fetch completes
Evidence:
- file: `frontend/src/pages/ChatPage.tsx`
- line: ~251-255 (no clear on activeId change before load)
Impact: UX confusion only
Root cause: Missing loading/clear on conversation switch
Fix plan: setMessages([]) + loading flag (future polish)
Required tests: E2E optional
Docs to update: none

---

## SLA-R02-002: Sync errors swallowed silently

Priority: P3
Status: open
Area: frontend
Expected: User feedback on sync failure
Actual: Empty catch in runSync
Evidence: `ChatPage.tsx` ~202-204
Impact: User may miss offline gap without knowing
Fix plan: Surface non-fatal banner (future)

---

## SLA-R02-003: Integration tests opt-in in CI

Priority: P3
Status: documented
Area: tests
Expected: Documented opt-in for Postgres integration
Actual: `t.Skip` without RUN_INTEGRATION=1
Evidence: `backend/tests/integration_*.go`
Impact: CI relies on unit + smoke; acceptable with local smoke-full PASS

---

## SLA-R02-004: k6 not installed locally

Priority: P4
Status: documented
Area: tests/ops
Expected: k6 dry-run in local audit
Actual: k6 binary not on PATH
Impact: CI dry-run covers syntax; local gap only

---

## SLA-R02-005: Presigned URL sharing (MVP accepted)

Priority: P3
Status: documented
Area: security
Expected: MVP accepts time-limited URL sharing risk
Actual: Presigned URLs work as designed
Impact: Low for MVP; documented in extensions/security notes

---

## Re-verified areas — no new P0/P1/P2

| Area | Result |
|------|--------|
| Docker API startup + migrations | PASS — health 200, migration v18 |
| Auth refresh rotation | PASS — JTI store rejects reuse |
| ACK seq binding | PASS — uses message seq |
| Media member download | PASS — owner bypass removed for linked msgs |
| WS origin + rate limit | PASS — origin_test + limiter wired |
| Outbox reclaim | PASS — worker + migration |
| OpenSearch membership | PASS — live IsMember filter |
| smoke-full | PASS 16/16 |
| Frontend download/unread/sync | PASS — code + build + e2e |

**Round 02 new P0/P1/P2 count: 0**
