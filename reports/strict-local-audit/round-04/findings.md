# Round 04 Findings

Fresh audit — no reliance on round-03 closure.

## SLA-R04-001: Conversation switch loading flash

Priority: P3 | Status: open
Area: frontend UX
Evidence: `ChatPage.tsx` — messages not cleared on `activeId` change before fetch.

## SLA-R04-002: Sync errors swallowed in UI

Priority: P3 | Status: open
Evidence: `ChatPage.tsx` runSync empty catch block.

## SLA-R04-003: WS JWT in query string (MVP accepted)

Priority: P3 | Status: documented
Evidence: `realtime/server.go` — query param auth; server logs path only.

## SLA-R04-004: In-memory rate limits without Redis

Priority: P3 | Status: documented
Evidence: `options.go` — dev/single-node fallback when Redis absent.

## SLA-R04-005: Integration tests opt-in locally

Priority: P3 | Status: documented
Evidence: CI runs integration; local requires `RUN_INTEGRATION=1`.

## SLA-R04-006: k6 not installed locally

Priority: P4 | Status: documented

## Re-verified — no new P0/P1/P2

Auth refresh (Redis JTI), outbox reclaim+retry, sync attachments, presence gate, WS rate limit, metrics production gate, smoke-full, integration tests — all PASS.

**Round 04 new P0/P1/P2: 0**
