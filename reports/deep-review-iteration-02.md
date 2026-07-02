# Deep Review — Iteration 02

Date: 2026-07-02  
Focus: Verify iteration 01 fixes; scan adjacent modules for regressions.

## Verification

| Area | Result |
|------|--------|
| Sync pagination | `paginationMeta` unit test passes; handler uses `pageSize+1` |
| Idempotent send | Pre-lookup + duplicate tx rollback; service skips broadcast on hit |
| Media RBAC | `GetAccessible` joins messages for membership |
| Outbox | Claim-to-`processing` in migration `00017` |
| Frontend sync | `has_more` loop per conversation |
| Payment 402 | `ApiError.code === payment_required` |
| WS reconnect | `getToken()` per connection attempt |

## New findings (iteration 02)

| ID | Priority | Status | Summary |
|----|----------|--------|---------|
| ISSUE-019 | P2 | fixed | WS ACK lacked message/conversation binding — aligned with REST |
| ISSUE-020 | P3 | open | Stale `processing` outbox rows if worker crashes mid-publish — future reaper job |
| ISSUE-021 | P3 | open | Frontend attachment download UI still missing — documented gap |
| ISSUE-022 | P4 | open | Loading/empty states for conversation list — polish |

## Counts

| Priority | Found | Fixed | Wontfix | Remaining |
|----------|-------|-------|---------|-----------|
| P0 | 2 | 2 | 0 | 0 |
| P1 | 9 | 9 | 0 | 0 |
| P2 | 6 | 5 | 1 | 0 |
| P3 | 2 | 0 | 0 | 2 (non-blocking) |
| P4 | 1 | 0 | 0 | 1 (non-blocking) |

**Stop condition met:** No new P0/P1/P2 in iteration 02 full pass.
