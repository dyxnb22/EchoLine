# Next Actions

> **Context:** T001–T440 与 backlog 已关闭。以下为**可选**增强，非阻塞项。

## Recommended (local developer)

1. `make dev-up && make dev-app` — full stack with worker
2. `make smoke-full` — end-to-end API smoke
3. 确认 `RUN_INTEGRATION=1` 集成测试在 Postgres 上通过

## Code quality (optional)

1. Owner RBAC integration test for `POST /api/channels/{id}/entitlements/require`
2. Swap OTel stub for real exporter SDK（ADR 0029 → 0008 Phase 2）
3. Expand `docs/openapi.yaml` with per-endpoint request/response body schemas
4. Full-stack Playwright in CI against compose stack（见 `frontend/e2e/integration.spec.ts`）

## Environment

```bash
make dev-up
make dev-app
export ADMIN_USER_IDS=<uuid>
make smoke-full
```

Completed scope: [`FINAL_COMPLETION_MANIFEST.md`](./FINAL_COMPLETION_MANIFEST.md).  
Doc navigation: [`docs/README.md`](./docs/README.md).
