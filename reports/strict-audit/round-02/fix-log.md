# Round 02 Fix Log

| ID | Priority | Fix | Files |
|---|---|---|---|
| R02-001 | P1 | IPKey uses RemoteAddr unless TRUSTED_PROXY | rate_limit/middleware.go, memory_test.go |
| R02-002 | P1 | In-memory limiter fallback | rate_limit/memory.go, server/options.go |
| R02-003 | P2 | GraphQL POST rate limit | server/options.go |
| R02-004 | P2 | Transactional impression cap | ads/handler.go |
