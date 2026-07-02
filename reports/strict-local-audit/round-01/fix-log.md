# Round 01 Fix Log

| ID | Priority | Fix summary | Files | Verification |
|----|----------|-------------|-------|--------------|
| SLA-R01-001 | P1 | Executable-relative migrations + MIGRATIONS_DIR | migrate.go, docker-compose.yml | migrate_test.go, API health 200 |
| SLA-R01-002 | P1 | Outbox processing_started_at + reclaim worker | 00018 migration, outbox/, worker/ | go test ./internal/outbox ./internal/worker |
| SLA-R01-003 | P0 | Attachment download UI + API client | frontend/api.ts, ChatPage.tsx | npm run build |
| SLA-R01-004 | P1 | Optimistic unread clear + refresh | ChatPage.tsx | manual + e2e |
| SLA-R01-005 | P1 | Sync re-trigger on conversations load | ChatPage.tsx | code review |
| SLA-R01-006 | P2 | ACK seq binding + mark-read cap | delivery/, realtime/, message/handler.go | go test ./... |
| SLA-R01-007 | P2 | Media membership for linked attachments | media/repository.go | go test ./internal/media |
| SLA-R01-008 | P2 | WS conv send rate limit | realtime/server.go, options.go | go test ./internal/realtime |
| SLA-R01-009 | P2 | WS origin allowlist | realtime/origin.go | origin_test.go |
| SLA-R01-010 | P2 | sync next_seq field | sync/handler.go, docs/api.md | sync handler_test.go |
| SLA-R01-011 | P2 | Smoke script aligned to API | scripts/smoke-api-full.sh | make smoke-full PASS |
| SLA-R01-012 | P1 | Disable payment self-serve in compose | docker-compose.yml | config review |
| SLA-R01-013 | P2 | Refresh JTI rotation store | auth/refresh_store.go, service.go | refresh_store_test.go |
| SLA-R01-014 | P2 | OpenSearch live membership filter | search/handler.go | go test ./internal/search |
| SLA-R01-015 | P3 | Register/refresh rate limits | server/options.go | server tests |
