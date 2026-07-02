# Round 02 Findings

## STRICT-R02-001: IPKey trusts raw X-Forwarded-For

Priority: P1
Status: fixed
Area: security/backend
Expected: Client IP derived from connection unless trusted proxy configured
Actual: Any client could spoof X-Forwarded-For to bypass login rate limit
Evidence: backend/internal/rate_limit/middleware.go IPKey
Fix: Use RemoteAddr by default; honor XFF only when TRUSTED_PROXY=true

## STRICT-R02-002: Rate limiting disabled without Redis

Priority: P1
Status: fixed
Area: security/backend
Expected: Rate limits always active
Actual: limiter nil when Redis absent — middleware passed all requests
Evidence: server/options.go, rate_limit/middleware.go:42
Fix: Default MemoryLimiter; Redis overrides when configured

## STRICT-R02-003: GraphQL sendMessage bypasses rate limit

Priority: P2
Status: fixed
Area: security/backend
Expected: Same send rate limit as REST
Actual: POST /graphql had auth only, no conv_send limit
Evidence: server/options.go applyRateLimits
Fix: graphql_send per-user middleware (60/min)

## STRICT-R02-004: Ads impression frequency cap TOCTOU

Priority: P2
Status: fixed
Area: backend/reliability
Expected: Cap enforced under concurrency
Actual: COUNT + INSERT not transactional
Evidence: ads/handler.go RecordImpression
Fix: Transaction with SELECT FOR UPDATE

## STRICT-R02-005: Redis Allow INCR+Expire non-atomic

Priority: P3
Status: open (wontfix this round)
Area: reliability
Impact: Edge case key without TTL if Expire fails
Reason: Prototype limiter; production should use Lua script

## STRICT-R02-006: Edit/Recall return 400 not 403 for membership errors

Priority: P3
Status: open
Area: backend/API semantics
Impact: HTTP status imprecision, not security bypass

## STRICT-R02-007: Server integration tests cover only health

Priority: P3
Status: open
Area: tests
Impact: Regression gap for middleware wiring

## STRICT-R02-008: Login page pre-filled dev credentials

Priority: P3
Status: open
Area: frontend

## STRICT-R02-009: JWT in WS query string

Priority: P3
Status: open (documented browser limitation)

## STRICT-R02-010: No attachment download UI

Priority: P3
Status: open
Area: frontend/UX

## Round 01 fix verification (re-audited)

- GraphQL addReaction membership: VERIFIED fixed
- Payment zero amount: VERIFIED fixed
- Metrics METRICS_TOKEN: VERIFIED fixed
- Ads membership: VERIFIED fixed
- Frontend WS reconnect/token/ack/block/checksum: VERIFIED fixed
