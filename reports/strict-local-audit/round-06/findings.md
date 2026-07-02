# Round 06 Findings

Fresh zero-based audit after round-05 fixes.

## SLA-R06-001: GraphQL sendMessage missing client_msg_id

Priority: P1 | Status: fixed
Area: backend
Evidence: `graph/handler.go:114` — empty `ClientMsgID` → 400 from validator

## SLA-R06-002: Sync unbounded cursor array

Priority: P1 | Status: fixed
Area: backend/reliability
Evidence: `sync/handler.go:75` — no cap on `req.Cursors`

## SLA-R06-003: Audit login IP spoofable via X-Forwarded-For

Priority: P2 | Status: fixed
Area: security
Evidence: `auth/service.go:193` — trusted XFF without TRUSTED_PROXY gate

## SLA-R06-004: WS typing events unlimited

Priority: P2 | Status: fixed
Area: realtime
Evidence: `realtime/server.go:326` — no limiter on typing.start/stop

## SLA-R06-005: Notifications polled on every message

Priority: P2 | Status: fixed
Area: frontend
Evidence: `ChatPage.tsx:346` — `messages.length` dependency

## SLA-R06-006: Reactions N+1 parallel fetch

Priority: P2 | Status: fixed (capped)
Area: frontend
Evidence: `ChatPage.tsx:270` — 50 parallel reaction requests; capped to 20 recent

## Open P3/P4

- SLA-R06-007: Conversation switch loading flash (P3)
- SLA-R06-008: Sync error UI silent (P3)
- SLA-R06-009: WS smoke skipped — wscat/timeout unavailable locally (P4)

**Round 06 new P0/P1/P2 found: 6 (0 P0, 2 P1, 4 P2)**
