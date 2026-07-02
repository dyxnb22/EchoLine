# Round 03 Fix Log

| ID | Fix | Files |
|----|-----|-------|
| R03-001 | Redis refresh JTI store | auth/refresh_store_redis.go, options.go |
| R03-002 | Outbox mark retry + fanout dedup | outbox/publisher.go, worker/handlers.go |
| R03-003 | Payment self-serve startup warn | options.go |
| R03-004 | Sync mark-read on active conv | ChatPage.tsx |
| R03-005 | Cursor monotonic on initial load | ChatPage.tsx |
| R03-006 | refreshTokenOnce dedup | api.ts, AuthContext.tsx, ChatPage.tsx |
| R03-007 | Sync attachment metadata | sync/handler.go, options.go |
| R03-008 | Search limit cap | search/handler.go |
| R03-009 | Production empty WS origin | realtime/origin.go |
| R03-010 | Production metrics gate | metrics/metrics.go, server.go |
| R03-011 | WS upgrade rate limit | options.go, server.go |
| R03-012 | Presence contact gate | conversation/repository.go, presence/*.go |
| R03-013 | Mark-read failure resync | ChatPage.tsx |
| R03-014 | WS send boolean return | api.ts |
