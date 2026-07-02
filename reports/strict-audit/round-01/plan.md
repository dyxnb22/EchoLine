# Round 01 — Strict Audit Plan

## Audit scope

Fresh full audit of EchoLine as if first-time takeover. No trust in DONE/CURRENT_STATE/completion manifests.

### Backend
- Auth (register, login, refresh, JWT)
- Conversations (direct, group, channel, subscribe, roles)
- Messages (send, edit, recall, sync, ACK, delivery)
- WebSocket (auth, heartbeat, message.send, ACK, fanout)
- Security (membership checks, admin RBAC, prototype routes)
- Payment/entitlement, ads, GraphQL, search, media, export

### Frontend
- Auth flows, ChatPage WS lifecycle, paid channel flow
- API/WS protocol alignment, error states

### Data model & tests
- Migrations, constraints, test coverage vs real behavior

## Documents read this round

- README.md, docs/architecture.md, docs/business-flows.md
- docs/api.md, docs/websocket-protocol.md, docs/data-model.md
- docs/reliability.md, docs/security-checklist.md
- ACCEPTANCE_MATRIX.md, QUALITY_GATES.md

## Code modules to inspect

- `backend/internal/graph/` — GraphQL prototype mutations
- `backend/internal/reaction/` — REST reaction membership
- `backend/internal/payment/` — self-serve ledger
- `backend/internal/ads/` — campaign list/impression
- `backend/internal/server/` — route wiring, metrics
- `frontend/src/pages/ChatPage.tsx` — WS reconnect, sync
- `frontend/src/api.ts` — WS client, media upload

## Test commands

```bash
cd backend && go test ./...
cd backend && go vet ./...
cd frontend && npm ci
cd frontend && npm run lint
cd frontend && npm run build
cd frontend && npm audit --omit=dev
make smoke  # if available
```

## Top 20 high-risk checkpoints

1. GraphQL mutations bypass REST permission checks
2. Message-scoped ops without membership validation
3. Payment self-serve free entitlement grant
4. Admin routes reachable without admin role
5. Search/export/media cross-conversation access
6. WebSocket auth bypass or missing membership on send
7. ACK/delivery state forward-only constraints
8. client_msg_id idempotency scope
9. Channel paid gate bypass via subscribe without entitlement
10. Ads prototype routes without channel membership
11. Metrics endpoint information disclosure
12. Rate limiting disabled without Redis
13. Frontend WS reconnect storm
14. Frontend stale JWT on WS reconnect
15. Frontend/backend WS field mismatch
16. Integration tests skipped vs claimed coverage
17. Prototype GraphQL/ads/payment exposed in production paths
18. seq allocation transaction safety
19. Outbox/MQ consistency on failure
20. Frontend optimistic send dedup correctness

## Historical conclusions NOT trusted

- FINAL_COMPLETION_MANIFEST.md "complete"
- ACCEPTANCE_MATRIX Phase done markers
- reports/deep-review-final.md
- DONE.md / CURRENT_STATE.md

## Execution

Plan written → immediate full audit → fix P0/P1/P2 → verify tests → round summary.
