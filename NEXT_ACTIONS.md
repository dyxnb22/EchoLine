# Next Actions

## Immediate P0

1. `make dev-up && make dev-app` — run API in compose profile
2. `make smoke-full` with running stack
3. Wire react-router in `App.tsx`

## Sequential

4. Channel entitlement check on subscribe
5. Playwright login→send→verify (non continue-on-error)
6. GraphQL addReaction mutation
7. Webhook admin list/replay UI

## Environment

```bash
make dev-up
make dev-app
export ADMIN_USER_IDS=<uuid>
make smoke-full
```
