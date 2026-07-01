# EchoLine Progress Log

本文件采用追加式记录。每轮执行结束时在顶部或底部追加均可，但必须包含任务、文件、测试、阻塞和下一步。

## 2026-07-01 Phase 4/6 group/channel + kafka + rate limit (checkpoint 3)

任务：E001-E005, F005-F008, H001-H002, H004-H005, D006 migration.

测试：`go test ./...` 通过。

下一步：outbox publisher, frontend J001, integration smoke.

## 2026-07-01 Phase 2/3 realtime + sync + ACK (checkpoint 2)

任务：

- B005-B010, C004-C006, D003-D004, A019-A022, F001/F003/F007 skeleton.

文件：

- `backend/internal/realtime/protocol.go`, updated `server.go`
- `backend/internal/message/service.go`, updated handlers
- `backend/internal/sync/`, `delivery/`, `apierror/`
- `backend/internal/redisx/`, `presence/`, `eventbus/`
- `backend/cmd/seed`, `backend/cmd/worker`
- `backend/migrations/00003_deliveries.sql`
- `docs/openapi.yaml`, `docs/api.md`, `reports/iteration-02.md`

测试：

- `go test ./...` 通过
- `make test` 通过
- `RUN_WS_SMOKE=1 make smoke` 通过

阻塞：

- 仍无 Docker/Postgres，integration smoke 未跑

下一步：

- E001-E005, F005-F008, integration smoke

## 2026-07-01 Phase 1 backend foundation (checkpoint 1)

任务：

- A001-A010: Go API skeleton, config, health, DB pool, goose migrations, users repo, Argon2id, register/login, JWT middleware.
- A012-A018: devices/conversations/messages schema and REST APIs.
- A011, C003, B001-B004: refresh token, conversation list API, WebSocket endpoint with auth/ping/connection hub.
- Updated docs, Makefile, smoke script, repo state files.

文件：

- `backend/go.mod`, `backend/go.sum`
- `backend/cmd/api/main.go`
- `backend/internal/config/`, `db/`, `migrate/`, `server/`, `auth/`, `user/`, `device/`, `conversation/`, `message/`
- `backend/migrations/00001_users.sql`, `00002_conversations_messages.sql`
- `Makefile`, `scripts/smoke-test.sh`
- `docs/api.md`, `docs/data-model.md`
- `CURRENT_STATE.md`, `NEXT_ACTIONS.md`, `DONE.md`, `ACCEPTANCE_MATRIX.md`, `BLOCKERS.md`

测试：

- `cd backend && go test ./...` 通过（无 DATABASE_URL 时 integration tests skip）。
- `make test` 通过。
- `make smoke` 通过。

阻塞：

- Docker/PostgreSQL 不可用，未跑 full API integration smoke。

不要重复：

- 不要重做 A001-A018 骨架。
- 不要重写 Phase 0 文档。

## 2026-07-01 Phase 0 长时执行系统补强

任务：

- 添加 Cursor Cloud Agent 10h 长跑启动 prompt。
- 添加细粒度 atomic task graph。
- 添加验收矩阵、质量门禁、review plan、performance plan、research plan。
- 添加 repo-based context compaction 文件。
- 添加 parallel agents 计划。
- 添加 `.cursor/rules/` 和 `.cursor/skills/`。
- 添加 future extension roadmap：加密、微服务拆分、广告、支付、推荐。
- 添加子 Agent orchestration plan 和 task packet 模板。
- 明确规定子 Agent 使用 Composer 2.5 时必须关闭 Fast mode。

文件：

- `CLOUD_AGENT_PROMPT.md`
- `TASK_GRAPH_DETAILED.md`
- `ACCEPTANCE_MATRIX.md`
- `BACKLOG.md`
- `CONTEXT_COMPACTION.md`
- `CURRENT_STATE.md`
- `NEXT_ACTIONS.md`
- `BLOCKERS.md`
- `DECISIONS.md`
- `QUALITY_GATES.md`
- `REVIEW_PLAN.md`
- `PERFORMANCE_PLAN.md`
- `RESEARCH_PLAN.md`
- `PARALLEL_AGENTS.md`
- `SUBAGENT_ORCHESTRATION.md`
- `SUBAGENT_TASK_PACKET.md`
- `docs/extensions-roadmap.md`
- `.cursor/rules/*`
- `.cursor/skills/*`

测试：

- `make help` 通过。
- `make test` 通过，占位输出：Phase 1 将添加后端测试。
- `make smoke` 通过，占位输出：Phase 1 将实现 smoke tests。
- `.cursor/rules` 和 `.cursor/skills` 文件清单验证通过。

阻塞：

- 暂无。

下一步：

- 开始 Phase 1：初始化 Go backend、config、health、PostgreSQL、migration、users/auth。
