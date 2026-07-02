# Round 05 Fix Log

| ID | Fix |
|----|-----|
| SLA-R05-001 | Removed `refreshConversations` from `markActiveRead`; preserve active conv unread=0 on list refresh |
| SLA-R05-002 | WS effect depends on `loggedIn` not token value; refresh updates localStorage only |
| SLA-R05-003 | `runSync` bootstraps all conversations into cursor map |
| SLA-R05-004 | `connectWS` aborts connect when refresh returns null |
| SLA-R05-005 | `tests/integration_env.go` sets default JWT_SECRET |
| SLA-R05-006 | `config.Load` rejects PAYMENT_SELF_SERVE in production |
| SLA-R05-007 | `GRAPHQL_ENABLED` defaults false in production |
| SLA-R05-008 | Rate limits on sync/ack/search/export/media in `applyRateLimits` |
| SLA-R05-009 | Shared `getOrCreateDeviceId()` helper |
| SLA-R05-010 | 30s sync poll when WS not open; sync on WS close transition |

Tests added: `config_test.go`, `graph/handler_test.go` (sendMessage client_msg_id)
