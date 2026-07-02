# Round 01 Findings

## STRICT-R01-001: GraphQL addReaction bypasses conversation membership

Priority: P1
Status: fixed
Area: security/backend
Business flow: User reacts to message in conversation they don't belong to
Expected: Membership check like REST `/api/messages/{id}/reactions`
Actual: GraphQL `addReaction` called `reaction.Repository.Add` directly without membership check
Evidence:
- file: backend/internal/graph/reaction.go (before fix)
- file: backend/internal/reaction/handler.go:30-45 (REST has requireMessageMember)
Impact: Any authenticated user could react to messages in arbitrary conversations
Root cause: GraphQL adapter skipped authz layer
Fix plan: Add membership resolution in ReactionService
Required tests: graph/reaction_test.go
Docs to update: none (behavior now matches REST)

## STRICT-R01-002: Payment self-serve allows zero-cost entitlement grant

Priority: P1
Status: fixed
Area: security/backend
Business flow: Paid channel subscription via self-serve ledger
Expected: Positive payment amount required when PAYMENT_SELF_SERVE=true
Actual: amount_cents=0 accepted, settle grants entitlement
Evidence:
- file: backend/internal/payment/handler.go:160 (no amount validation)
Impact: Free paid channel access when self-serve enabled
Root cause: Missing amount validation
Fix plan: Reject amount_cents <= 0
Required tests: payment/handler_test.go TestHandleCreateRejectsZeroAmount

## STRICT-R01-003: GET /metrics unauthenticated

Priority: P2
Status: fixed
Area: security/backend
Expected: Metrics scrape endpoint protected in production
Actual: Public Prometheus handler
Evidence: backend/internal/server/server.go:31
Impact: Operational metrics disclosure
Fix plan: Optional METRICS_TOKEN bearer gate via metrics.ProtectedHandler
Required tests: metrics/metrics_test.go

## STRICT-R01-004: Ads list without channel membership

Priority: P2
Status: fixed
Area: security/backend
Expected: Only channel members can list campaigns
Actual: Any authenticated user could list campaigns for any channel ID
Evidence: backend/internal/ads/handler.go HandleList
Fix plan: requireChannelMember check

## STRICT-R01-005: Ads impression without membership / campaign binding

Priority: P2
Status: fixed
Area: security/backend
Expected: Subscriber records impression for campaigns in their channel
Actual: Any user could record impressions for any campaign_id
Fix plan: Membership + EnsureCampaignChannel

## STRICT-R01-006: Frontend WS reconnect storm

Priority: P1
Status: fixed
Area: frontend/realtime
Expected: Stable WS connection during conversation list refresh
Actual: runSync depended on conversations state; refresh triggered WS effect teardown
Evidence: frontend/src/pages/ChatPage.tsx runSync deps + WS effect deps
Fix plan: conversationsRef + runSyncRef pattern

## STRICT-R01-007: Frontend WS stale JWT on reconnect

Priority: P1
Status: fixed
Area: frontend/realtime
Expected: WS reconnect uses refreshed access token
Actual: Closed-over token state could be expired; no refresh on WS failure
Fix plan: Read token from localStorage; refreshAccessToken callback before connect

## STRICT-R01-008: Block button on own messages

Priority: P2
Status: fixed
Area: frontend
Expected: Cannot block self
Actual: sender_id !== "me" fails after REST confirms real UUID
Fix plan: fetchMe + currentUserId comparison

## STRICT-R01-009: Frontend never sends WS message.ack

Priority: P2
Status: fixed
Area: frontend/reliability
Expected: Delivery ACK via WS per protocol
Actual: Only REST markConversationRead used
Fix plan: Send message.ack on message.created for active conversation

## STRICT-R01-010: Media upload missing checksum

Priority: P2
Status: fixed
Area: frontend/docs
Expected: checksum field per docs/api.md
Actual: Frontend omitted checksum
Fix plan: SHA-256 checksum in presignUpload

## STRICT-R01-011: Hardcoded login credentials (P3)

Priority: P3
Status: open
Area: frontend/security
Evidence: frontend/src/pages/LoginPage.tsx default state alice/secret123
Impact: Dev convenience; low risk in production builds
Fix plan: Defer — empty defaults

## STRICT-R01-012: JWT in WS query string (P3)

Priority: P3
Status: open (documented limitation)
Area: security
Impact: Token visible in logs/history; browser WS API constraint

## STRICT-R01-013: No attachment download UI (P3)

Priority: P3
Status: open
Area: frontend/UX

## STRICT-R01-014: Integration tests skipped without DATABASE_URL (P3)

Priority: P3
Status: open (environment)
Area: tests
Evidence: backend/tests/integration_*.go t.Skip guards
