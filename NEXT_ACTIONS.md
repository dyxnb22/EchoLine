# Next Actions

> **Context:** T001–T440 与 backlog 已关闭。local full-stack smoke、WS smoke 与 Postgres integration 已在 2026-07-02 通过。以下为**可选**增强，非阻塞项。

## Verified locally

- `make dev-up && make dev-app`
- `make smoke-full`
- `RUN_WS_SMOKE=1 make smoke-full`
- `RUN_INTEGRATION=1 DATABASE_URL=postgres://echoline:echoline@localhost:5432/echoline?sslmode=disable go test ./tests`

## Code quality (optional)

1. Owner RBAC hardening on `POST /api/channels/{id}/entitlements/require`（已实现 owner-only，可补集成测试）
2. Migrate `conversation/handler` legacy `writeError` → `apierror` envelope
3. Swap OTel stub for real exporter SDK（ADR 0029 → 0008 Phase 2）
4. Expand `docs/openapi.yaml` with per-endpoint request/response body schemas

## Environment

```bash
make dev-up
make dev-app
export ADMIN_USER_IDS=<uuid>
make smoke-full
RUN_WS_SMOKE=1 make smoke-full
```

Completed scope: [`FINAL_COMPLETION_MANIFEST.md`](./FINAL_COMPLETION_MANIFEST.md).
Doc navigation: [`docs/README.md`](./docs/README.md).
