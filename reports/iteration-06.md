# Iteration 06 — Batch Next 200 (T241–T440)

## Deliverables

- Encryption key bundle REST API
- Webhook delivery persistence + worker retry loop
- GraphQL sendMessage mutation
- Presence last-seen endpoints
- Friend recommendation API + migration 00015
- Frontend LoginPage, ConversationActions, friend sidebar
- Ops: compose app profile, backup-db, k8s secrets, Playwright CI
- Integration auth test (register/login/me)
- ADRs 0023–0026

## Verification

```bash
cd backend && go test ./...
cd frontend && npm run build
RUN_INTEGRATION=1 DATABASE_URL=... go test -run Integration ./tests/...
```

## Gaps

- react-router not wired yet (dependency added)
- Entitlement enforcement not on channel subscribe
- E2EE client-side prototype not started
