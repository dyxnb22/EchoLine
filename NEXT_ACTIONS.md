# Next Actions

## Post-closure (optional)

1. `make dev-up && make dev-app` — full stack with worker
2. `make smoke-full` — end-to-end API smoke on developer machine
3. Owner RBAC on `POST /api/channels/{id}/entitlements/require`
4. Swap OTel stub for real exporter SDK

## Environment

```bash
make dev-up
make dev-app
export ADMIN_USER_IDS=<uuid>
make smoke-full
```

See `FINAL_COMPLETION_MANIFEST.md` for completed scope.
