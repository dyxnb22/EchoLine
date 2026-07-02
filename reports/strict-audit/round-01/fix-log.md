# Round 01 Fix Log

| ID | Priority | Fix | Files |
|---|---|---|---|
| R01-001 | P1 | GraphQL reaction membership check | graph/reaction.go, server/options.go, graph/reaction_test.go |
| R01-002 | P1 | Reject zero payment amount | payment/handler.go, payment/handler_test.go |
| R01-003 | P2 | METRICS_TOKEN bearer protection | metrics/metrics.go, config/config.go, server/server.go |
| R01-004 | P2 | Ads list membership | ads/handler.go, server/options.go |
| R01-005 | P2 | Ads impression membership + campaign binding | ads/handler.go |
| R01-006 | P1 | WS reconnect storm fix | ChatPage.tsx |
| R01-007 | P1 | WS token refresh on reconnect | api.ts, ChatPage.tsx |
| R01-008 | P2 | Block self prevention | api.ts fetchMe, ChatPage.tsx |
| R01-009 | P2 | WS message.ack delivery | ChatPage.tsx |
| R01-010 | P2 | Media upload checksum | api.ts |
