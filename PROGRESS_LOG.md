# EchoLine Progress Log

本文件采用追加式记录。每轮执行结束时在顶部或底部追加均可，但必须包含任务、文件、测试、阻塞和下一步。

## 2026-07-02 Deep review quality iteration (2 rounds)

任务：深度代码审查 + 修复直至 P3 以下。

修复：sync 游标/has_more、幂等发送、附件成员下载、outbox processing claim、ACK 校验、前端 sync 分页/402/WS 刷新/去重/已读、缓存 can_publish、admin health RBAC。

测试：go test ./... + npm run build 通过。

阻塞：make smoke-full（无 Docker）。

下一步：本地 smoke-full；可选 outbox processing reaper。

## 2026-07-02 Full audit fix Phases 0–4

任务：全量审计修复计划 Phase 0–4 实施。

后端：thread/forward 主链路、RBAC、DLQ replay、outbox/sync/reliability、admin composite、device WS、client_msg_id UUID、fanout 分页、cmd/replay 对齐 schema。

前端：typing/sync/付费频道/composer/can_publish、CreateConversationModal、AdminPanel DLQ 字段。

测试/CI：integration RBAC、dual WS test、worker CI job、Playwright extended smoke。

文档：admin-panel、business-flows、websocket-protocol、CURRENT_STATE。

测试：go test ./... + npm run build 通过。

阻塞：云 VM 无 Docker 全量 smoke。

下一步：本地 make smoke-full。

## 2026-07-02 Documentation alignment pass 2: full scan

任务：全量扫描 116 个 markdown + openapi，纠正至与代码一致。

ADR/技术文档：去除 ghost `internal/api/` 路径；0002 Accepted；architecture 模块表扩展。

数据/API：data-model 补列；api.md 增 GET /ws；openapi.yaml 61 paths + Error schema。

状态/计划：extensions-roadmap、RESEARCH_PLAN、CLOUD_AGENT_PROMPT、load-test-01 对齐。

Manifest/报告：BATCH_* 历史路径说明；review M001–M007 免责声明。

测试：文档-only。

阻塞：无新增。

下一步：本地 smoke-full；可选 openapi body schemas。

## 2026-07-02 Documentation alignment: dedupe, correct, consolidate

任务：全库文档对齐 — 纠正过时内容、整合重复项、更新状态/memory 文件。

ADR：完整索引 0001–0031；0003 重复解决（cache/MQ → 0031）；0013 superseded by 0019。

技术文档：`websocket-protocol.md` 与代码对齐；`data-model.md` 补 `outbox_events`。

状态文档：`DONE.md`、`BACKLOG.md`、`ACCEPTANCE_MATRIX.md`、`TASKS.md` 关闭态 banner；`CURRENT_STATE.md`、`NEXT_ACTIONS.md` 更新。

导航：`docs/README.md` 扩展（面试/原型/报告）；`README.md` 面试链接；`review-docs-consistency.md` 刷新。

测试：文档-only，未改代码。

阻塞：无新增。

下一步：本地 `make smoke-full`；可选 openapi 错误示例补全。

## 2026-07-01 Engineering review #03: API unification, validation depth, docs index

任务：第三轮文档对齐、架构审核、业务流程校准、深度 code review。

前端：api.ts 全量迁移至 http.ts helpers；e2ee 使用 authedRequest。

后端：Edit 路径 sanitize+validate；integration_validation_test。

文档：docs/README.md 索引、architecture mermaid、make verify、engineering-review-03。

测试：go test + npm build + playwright 通过。

阻塞：云 VM 无 Docker 全量 smoke。

下一步：conversation handler apierror 统一（可选）。

## 2026-07-01 Engineering review #02: docs, architecture, RBAC, validate

任务：文档对齐、架构审核、业务流程校准、深度 code review。

安全：entitlement grant admin-only、require owner-only（ADR 0030）；validate 包统一输入长度。

测试：entitlement handler_test、integration_entitlement_test、validate limits_test。

前端：api/http.ts 集中 HTTP + AuthContext bindAuthFetch。

文档：business-flows、engineering-standards、architecture/api/reliability/security/README 对齐；engineering-review-02。

测试：go test + npm build + playwright（待跑）。

阻塞：云 VM 无 Docker 全量 smoke。

下一步：api.ts 全量迁移 authedRequest（可选）。

## 2026-07-01 Final completion: T001–T440 + backlog + extensions + code review

任务：`FINAL_COMPLETION_MANIFEST.md` — 全量编号任务收尾、backlog、extensions roadmap、全量 code review。

后端：entitlement gate + 00016，payment settle grant，GraphQL addReaction，fanout push + bounded idempotency，worker compose，integration_messaging_test，cmd/migrate，OTel stub。

前端：react-router，AuthContext，ChatPage，NotificationPanel，GroupSettings，SettingsPage，edit/recall，archived，admin DLQ，recommended subscribe fix，reaction prefetch。

运维：deploy/gateway，CI goose + strict Playwright，ADR 0027-0029。

文档：code-review-final.md，api.md，data-model.md 更新。

测试：go test ./... + npm run build 通过。

阻塞：云 VM 无 Docker 全量 smoke。

下一步：本地 make smoke-full（可选）。

## 2026-07-01 Batch-Next-200: encryption, webhook retry, graphql mutation, frontend split (checkpoint 9)

任务：200项见 `BATCH_NEXT_200_MANIFEST.md`（T241–T440）。

后端：encryption keys API，webhook persistence+retry worker，GraphQL sendMessage，last-seen，friend recommendations，00015。

前端：LoginPage，ConversationActions，friend recs，api helpers。

运维：compose app profile，backup-db，k8s secrets，Playwright CI，integration test strict。

文档：ADR 0023-0026，iteration-06。

测试：go test + frontend build 通过。

阻塞：云 VM 无 Docker/Postgres 全量 smoke。

下一步：react-router，entitlement enforcement，Playwright send E2E。

## 2026-07-01 Batch-Next-120: admin, graphql, ops, frontend (checkpoint 8)

任务：120项见 `BATCH_NEXT_120_MANIFEST.md`（T121–T240）。

后端：admin users/reports/audit + RBAC，webhook on send，GraphQL prototype，payment settle，ads frequency cap，push worker mock，migration 00014。

前端：ThreadPanel，AdminPanel，reactions display/remove，typing.stop。

运维：Dockerfile，Helm skeleton，monitoring compose，deploy workflow，golangci-lint，make loadtest-ws。

测试：reaction/thread/webhook/push/admin/graph/replay/integration skeleton；go test + frontend build 通过。

阻塞：Docker/Postgres 云 VM 不可用。

下一步：Playwright CI，GraphQL mutations，integration smoke。

## 2026-07-01 Batch-120: CI, scripts, ADRs 0016-0022, docs, research, reports, loadtest (checkpoint 7)

任务：120项见 `BATCH_120_MANIFEST.md`（T001–T120，5 tracks）。

CI：`.github/workflows/ci.yml` — Go test (postgres+redis services) + frontend build + k6 smoke + security scans。

脚本：`scripts/seed-extended.sh`（5 users / 1 group / 1 channel）、`scripts/bootstrap-minio.sh`（bucket + CORS + presign smoke）、`scripts/dlq-replay.sh`（--id / --all / --list）。

后端：`backend/cmd/replay/main.go` — DLQ replay CLI（API mode + direct DB mode + dry-run）。

ADR：0016 reactions+threads、0017 push notifications、0018 webhook delivery、0019 payment ledger、0020 ads platform、0021 recommendation engine、0022 GraphQL subscriptions（7 新 ADR）。

文档：7 prototype docs（graphql / admin-panel / push / encryption / payments / ads / recommendation）。

研究：`research-discord-slack.md`、`research-e2ee-tradeoffs.md`。

报告：`reports/iteration-04.md`、`reports/review-fixes-batch120.md`（13 known issues, 0 critical）。

负载测试：`loadtests/k6-mixed-workload.js`（4 scenarios: auth/send/ws/search，staged ramp，threshold assertions）。

测试：`go vet ./...` 通过，`go build ./...` 通过（含 cmd/replay），`npm run build` 通过。

阻塞：Docker/Postgres 云 VM 不可用，全集成 smoke 未执行（见 `BLOCKERS.md`）。

下一步：T023 reactions migration + API，T024 thread replies，T025 DLQ replay endpoint，T026 push worker，CI push 验证。

## 2026-07-01 Batch-120: reactions, extensions, CI (checkpoint 7)

任务：120项见 `BATCH_120_MANIFEST.md`。

代码：reactions/threads/forward/presence/export/archive/push/payment/ads/recommendation。

运维：GitHub Actions, replay CLI, k6-mixed, seed-extended。

前端：dark mode, reactions, channel filter, PWA sw。

测试：go test + frontend build 通过。

## 2026-07-01 Batch-100: social, ops, docs, frontend (checkpoint 6)

任务：100项见 `BATCH_100_MANIFEST.md`（主线+backlog+stretch+research+extension）。

代码：pin/block/report/notification/mute/spam/typing/admin/dlq/reliability tests。

前端：register/typing/notifications/PWA/playwright。

运维：k6/chaos/grafana/.env.example/Makefile。

文档：12 ADR + 9 review reports + interview docs。

测试：go test + frontend build 通过。

## 2026-07-01 Batch-20: search, edit/recall, metrics, frontend (checkpoint 5)

任务（20项）：G005-G008, C007-C009, D007 SKIP LOCKED, E007-E008, F004/F008, H006, I001-I005, J007-J009.

文件：

- `backend/internal/search/`, `cache/`, `metrics/`, `worker/`
- `backend/migrations/00007_device_sync_cursors.sql`, `00008_message_search_index.sql`
- `backend/internal/message/` edit/recall handlers
- `docs/adr/0003-large-group-fanout.md`
- `frontend/` optimistic send, upload, search UI

测试：全部 unit/smoke 通过。

下一步：integration smoke, k6, typing indicator.

## 2026-07-01 Phase 4/5/7 outbox + media + frontend (checkpoint 4)

任务：D007-D008, E006, F008, G001-G004, H003, J001-J006.

文件：

- `backend/internal/outbox/`, updated `message/repository.go`
- `backend/internal/media/repository.go`, updated `handler.go`
- `backend/migrations/00005_attachments.sql`, `00006_dead_letter.sql`
- `backend/internal/rate_limit/middleware.go` (conv_send)
- `backend/internal/realtime/fanout_test.go`
- `frontend/` Vite React app
- `docs/api.md`, `docs/data-model.md`

测试：

- `go test ./...` 通过
- `make test` 通过
- `RUN_WS_SMOKE=1 make smoke` 通过
- `npm run build` 通过

阻塞：

- 仍无 Docker/Postgres，integration smoke 未跑

下一步：

- integration smoke, G005 download URL, observability I001

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
