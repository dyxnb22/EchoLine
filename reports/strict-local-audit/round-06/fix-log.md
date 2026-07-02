# Round 06 Fix Log

| ID | Fix |
|----|-----|
| SLA-R06-001 | Auto-generate `ClientMsgID` in GraphQL sendMessage; test added |
| SLA-R06-002 | `maxSyncCursors = 50` guard in sync handler |
| SLA-R06-003 | `clientIP()` respects TRUSTED_PROXY like rate limiter |
| SLA-R06-004 | Typing rate limit 30/min per user per conversation |
| SLA-R06-005 | Notifications refresh on `activeId` change only |
| SLA-R06-006 | Reaction prefetch limited to last 20 messages |
