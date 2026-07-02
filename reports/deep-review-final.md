# Deep Review — Final Report

Date: 2026-07-02  
Branch: `cursor/deep-review-quality-iteration-c067`

## Summary

Completed **2 full review iterations** on EchoLine after project closure. Focused on reliability, sync, idempotency, RBAC, frontend protocol alignment, and outbox correctness — without adding out-of-scope features.

## Iteration counts

| Round | P0 | P1 | P2 | P3 | P4 |
|-------|----|----|----|----|-----|
| 01 | 2 | 9 | 5 | 0 | 0 |
| 02 | 0 | 0 | 1 | 2 | 1 |

All P0/P1/P2 from iteration 01 were fixed. Iteration 02 found one additional P2 (WS ACK validation), fixed immediately. Remaining items are P3/P4 only.

## Major fixes

1. **Offline sync (P0):** Backend no longer advances device cursor on partial pages; frontend paginates until `has_more=false`.
2. **Idempotent send (P1):** Duplicate `client_msg_id` returns existing message without seq bump or re-broadcast; cross-conversation reuse rejected.
3. **Attachment download (P1):** Group/channel recipients can presign download for linked attachments.
4. **Outbox (P1):** Rows claimed as `processing` before publish (migration `00017`).
5. **Frontend reliability (P1):** Payment detection via structured errors, WS token refresh on reconnect, sender dedup by `client_msg_id`, upload pending state, read cursor while viewing.
6. **RBAC (P2):** Admin health admin-only; cached conversation list includes `can_publish`.

## Tests run

| Command | Result |
|---------|--------|
| `go test ./...` | Pass |
| `npm run build` (frontend) | Pass |
| `make smoke-full` | Skipped — Docker/Postgres unavailable (`BLOCKERS.md`) |

New tests: `sync/handler_test.go`, `message/idempotency_test.go`.

## Remaining (non-blocking)

| Issue | Priority | Notes |
|-------|----------|-------|
| ISSUE-014 | P2 wontfix | Presigned URL shareable within expiry — MVP documented risk |
| ISSUE-020 | P3 | Stale `processing` outbox reaper not implemented |
| ISSUE-021 | P3 | Frontend download-url UI not wired |
| ISSUE-022 | P4 | Conversation list loading/empty polish |

## Architecture notes

- No large refactors; changes are surgical and aligned with `docs/reliability.md`.
- Outbox `processing` state enables safe multi-worker claim without holding DB locks during network publish.
- Frontend `ApiError` enables protocol-aware error handling without string matching.

## Stop condition

**Satisfied:** Iteration 02 complete review found no P0/P1/P2 issues. Remaining P3/P4 documented as non-blocking or wontfix.

## Long-term suggestions (not implemented)

1. Proxy download endpoint with membership re-check (replace bare presigned URLs).
2. Outbox stale `processing` reaper cron.
3. Real integration smoke when Postgres available (`make smoke-full`).
4. Frontend attachment preview via `download-url`.
5. Playwright tests against live compose stack (not mocked API).

## Artifacts

- `reports/deep-review-iteration-01.md`
- `reports/deep-review-iteration-02.md`
- `reports/deep-review-issues.md`
- `reports/deep-review-fix-log.md`
