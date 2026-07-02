# Round 05 Findings

Fresh audit — prior rounds 01–04 and `final.md` not trusted.

## SLA-R05-001: Unread badge reappears while viewing active conversation

Priority: P1 | Status: fixed  
Area: frontend  
Business flow: User reads active conversation; inbound message arrives  
Expected: Sidebar unread stays 0 for active thread  
Actual: `markActiveRead` called `refreshConversations()` which overwrote optimistic `unread: 0`  
Evidence:
- file: `frontend/src/pages/ChatPage.tsx`
- line: 130–136 (before fix)

## SLA-R05-002: WebSocket torn down on every access-token rotation

Priority: P1 | Status: fixed  
Area: frontend/realtime  
Expected: WS stays connected when token refreshed via localStorage  
Actual: `refreshAccessToken` called `setToken`, re-running WS `useEffect`  
Evidence: `ChatPage.tsx:235`, `useEffect` deps included `token`

## SLA-R05-003: Partial sync on reconnect (only opened conversations)

Priority: P1 | Status: fixed  
Area: frontend/reliability  
Expected: All conversations sync on reconnect  
Actual: `runSync` only bootstrapped cursors when `seqCursorsRef` was empty  
Evidence: `ChatPage.tsx:169–178`

## SLA-R05-004: WS reconnect after refresh failure still opens socket

Priority: P1 | Status: fixed  
Area: frontend  
Evidence: `api.ts:492–494` — ignored null from `refreshAccessToken`

## SLA-R05-005: Integration tests fail without JWT_SECRET locally

Priority: P2 | Status: fixed  
Area: tests  
Evidence: `integration_auth_test.go:46` — `config.Load()` without JWT_SECRET

## SLA-R05-006: PAYMENT_SELF_SERVE can enable in production

Priority: P1 | Status: fixed  
Area: security/payment  
Evidence: `payment/handler.go:212–227`, no production guard in `config.Load()`

## SLA-R05-007: GraphQL prototype always mounted

Priority: P2 | Status: fixed  
Area: security/architecture  
Evidence: `server/options.go:137` — always registered

## SLA-R05-008: Sensitive endpoints lack rate limits

Priority: P2 | Status: fixed  
Area: security  
Evidence: `server/server.go:46–51` sync/ack/search/export/media unbounded

## SLA-R05-009: Settings vs Chat device_id mismatch

Priority: P2 | Status: fixed  
Area: frontend  
Evidence: `SettingsPage.tsx:11` used `"web"` fallback vs Chat UUID

## SLA-R05-010: Stale messages when WS dead (no REST fallback)

Priority: P1 | Status: fixed  
Area: frontend/reliability  
Fix: periodic sync + sync on WS close when not open

## Open P3/P4 (documented)

- SLA-R05-011: Sync errors swallowed (P3)
- SLA-R05-012: WS JWT in query string MVP tradeoff (P3)
- SLA-R05-013: k6 not installed locally (P4)

**Round 05 new P0/P1/P2 found: 9 (0 P0, 5 P1, 4 P2)**
