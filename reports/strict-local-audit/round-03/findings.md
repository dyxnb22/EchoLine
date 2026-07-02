# Round 03 Findings

## SLA-R03-001: Refresh JTI store process-local only

Priority: P1 | Status: fixed  
Fix: `RedisRefreshStore` when Redis configured; memory fallback for single-node dev.

## SLA-R03-002: Outbox republish after publish+failed mark

Priority: P1 | Status: fixed  
Fix: `markPublishedWithRetry` (5 attempts); FanoutWorker idempotency by message ID.

## SLA-R03-003: Payment self-serve entitlement bypass

Priority: P1 (when enabled) | Status: fixed  
Fix: Startup warning when `PAYMENT_SELF_SERVE=true`; compose default remains false.

## SLA-R03-004: Sync path skips mark-read while viewing

Priority: P1 | Status: fixed  
Fix: `markActiveRead` after sync merge for active conversation.

## SLA-R03-005: History pagination regresses sync cursor

Priority: P1 | Status: fixed  
Fix: Only advance `seqCursorsRef` on initial load (`beforeSeq == null`).

## SLA-R03-006: Concurrent refresh causes logout

Priority: P2 | Status: fixed  
Fix: `refreshTokenOnce()` deduplicates in-flight refresh calls.

## SLA-R03-007: Sync payloads missing attachment metadata

Priority: P2 | Status: fixed  
Fix: Sync handler batch-loads attachments via `SetAttachments`.

## SLA-R03-008: OpenSearch search limit uncapped

Priority: P2 | Status: fixed  
Fix: Cap `limit` at 100 in search handler before both engines.

## SLA-R03-009: WS origin empty always allowed

Priority: P2 | Status: fixed  
Fix: Deny empty Origin in production unless `WS_ALLOW_EMPTY_ORIGIN=true`.

## SLA-R03-010: Metrics exposed without token

Priority: P2 | Status: fixed  
Fix: `/metrics` returns 403 in production when `METRICS_TOKEN` unset.

## SLA-R03-011: WS upgrade unrated

Priority: P2 | Status: fixed  
Fix: `ws_upgrade` rate limit (30/min/IP) in `applyRateLimits`; removed duplicate `/ws` route.

## SLA-R03-012: Presence probes arbitrary user IDs

Priority: P2 | Status: fixed  
Fix: `ShareAnyConversation` gate on online/last-seen lookups.

## SLA-R03-013: Mark-read failure leaves optimistic unread clear

Priority: P2 | Status: fixed  
Fix: On mark-read failure, call `refreshConversations()` to resync.

## SLA-R03-014: WS send silently drops when not OPEN

Priority: P2 | Status: fixed  
Fix: `connectWS().send()` returns boolean success indicator.

## SLA-R03-015: JWT in WS query string (MVP)

Priority: P3 | Status: documented  
Impact: Token may appear in proxy logs; server logs path only. Future: subprotocol auth.

## SLA-R03-016: In-memory rate limits without Redis

Priority: P3 | Status: documented  
Impact: Per-process limits in dev; Redis required for multi-instance production.
