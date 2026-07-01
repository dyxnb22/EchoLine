# Next Actions

## Immediate P0

1. Integration smoke: `make dev-up` + `RUN_API_SMOKE=1 RUN_WS_SMOKE=1 make smoke-full`
2. Set `ADMIN_USER_IDS` in `.env` for admin panel testing
3. CI: remove `continue-on-error` on integration tests when stable

## Sequential

4. Playwright in GitHub Actions
5. GraphQL mutations (sendMessage, addReaction)
6. Frontend react-router split
7. Webhook delivery retry worker
8. E2EE key bundle REST API

## Environment

```bash
make dev-up
export DATABASE_URL=postgres://echoline:echoline@localhost:5432/echoline?sslmode=disable
export JWT_SECRET=change-me
export ADMIN_USER_IDS=<admin-user-uuid>
export WEBHOOK_URL=http://localhost:9999/hook
export GRAPHIQL=true
make api-run
make worker-run
make seed
make frontend-dev
```
