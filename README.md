# EchoLine

EchoLine 是一个受 Telegram 启发的实时消息平台，用来系统性训练长连接、消息可靠性、群聊扩散、多端同步、缓存、MQ、搜索、限流、审计和可观测性等真实互联网工程能力。

它不是 Telegram 的 1:1 复刻，也不是只有 WebSocket 收发消息的 IM demo。EchoLine 的目标是成为一个可持续演进、可测试、可讲述的长期作品集项目。

## 项目目标

- 构建 Telegram-like messaging platform，而不是普通聊天 UI。
- 用真实工程问题覆盖后端 / 全栈 / 系统设计面试高频场景。
- 保持主流互联网技术栈和可解释架构取舍。
- 支持 Cursor Cloud Agent 根据任务图长期执行，持续推进 10 小时以上。
- 每个阶段都产出代码、测试、文档、ADR 或报告，避免低价值空转。

## 非目标

- 不 1:1 复刻 Telegram。
- 不追求复杂 UI 动效。
- MVP 阶段不做真正生产级分布式系统，只保留可演进接口和设计文档。
- 不一开始微服务化，优先模块化单体，后续通过 MQ、worker、cache 演进。
- MVP 和 Phase 0-10 不做端到端加密、语音视频通话、支付、推荐系统、生产级风控；加密、微服务拆分、广告、支付、推荐作为 future extensions 记录在 `docs/extensions-roadmap.md`。

## 建议技术栈

| 层 | 技术 |
|---|---|
| Backend | Go, Chi/Gin/Fiber, WebSocket, REST |
| Database | PostgreSQL |
| Cache | Redis |
| MQ | Kafka 或 Redpanda |
| Search | OpenSearch |
| Object Storage | MinIO / S3 compatible |
| Frontend | React / Next.js |
| Auth | JWT + refresh token |
| Observability | OpenTelemetry, Prometheus, Grafana, structured logs |
| DevOps | Docker Compose, Makefile, GitHub Actions |
| Test | unit test, integration test, k6 load test |

## MVP 范围

- 用户注册 / 登录。
- 私聊会话。
- 基础群聊。
- 消息发送、保存、分页拉取。
- WebSocket 在线推送。
- 会话列表。
- 未读数基础版。
- 单设备 ACK。
- Docker Compose 本地启动。
- README、架构文档、任务图、进度日志。

## Growth 范围

- 多端登录。
- 离线消息同步。
- 消息编辑 / 撤回。
- 群成员角色。
- 频道 / broadcast channel。
- Redis 缓存会话和在线状态。
- MQ 异步通知。
- 附件上传下载。
- OpenSearch 搜索。
- 限流、审计日志、监控指标。

## Stretch 范围

- 大群扩散策略。
- 消息顺序性优化。
- 多端一致性策略。
- 死信队列。
- 压测报告。
- 分库分表设计文档。
- 热点群缓存策略。
- WebSocket 网关水平扩展设计。
- 端到端链路 tracing。

## 快速开始

```bash
make verify          # go test + frontend build + Playwright (local CI)
make dev-up          # Postgres, Redis, Redpanda, MinIO
make dev-app         # API + worker (compose profile app)
make smoke-full      # API smoke (needs running stack)

cd backend && go test ./...
cd frontend && npm run build && npx playwright test
```

当前阶段：**工程审查 #03 完成** — 见 `reports/engineering-review-03.md`、`docs/README.md`。

## 核心文档

- [CLOUD_AGENT_PROMPT.md](./CLOUD_AGENT_PROMPT.md)：给 Cursor Cloud Agent 的 10h 长跑启动 prompt。
- [AGENTS.md](./AGENTS.md)：给长期执行 Agent 的工作说明。
- [TASKS.md](./TASKS.md)：可持续消费任务图。
- [TASK_GRAPH_DETAILED.md](./TASK_GRAPH_DETAILED.md)：细粒度 atomic task graph。
- [ACCEPTANCE_MATRIX.md](./ACCEPTANCE_MATRIX.md)：阶段、模块和能力验收矩阵。
- [BACKLOG.md](./BACKLOG.md)：secondary、stretch、research 和 future extension backlog。
- [EXECUTION_RULES.md](./EXECUTION_RULES.md)：阻塞、测试、报告和继续执行规则。
- [CONTEXT_COMPACTION.md](./CONTEXT_COMPACTION.md)：repo-based memory compaction 规则。
- [CURRENT_STATE.md](./CURRENT_STATE.md)：当前状态，供新一轮 Agent 快速恢复。
- [NEXT_ACTIONS.md](./NEXT_ACTIONS.md)：下一批最应该执行的任务。
- [SUBAGENT_ORCHESTRATION.md](./SUBAGENT_ORCHESTRATION.md)：子 Agent 分派、总控、冲突规避和状态回写计划。
- [SUBAGENT_TASK_PACKET.md](./SUBAGENT_TASK_PACKET.md)：子 Agent 任务包模板。
- [PROGRESS_LOG.md](./PROGRESS_LOG.md)：追加式进度记录。
- [DONE.md](./DONE.md)：已完成能力索引。
- [docs/README.md](./docs/README.md)：文档导航索引。
- [docs/architecture.md](./docs/architecture.md)：总体架构。
- [docs/business-flows.md](./docs/business-flows.md)：核心业务流程。
- [docs/engineering-standards.md](./docs/engineering-standards.md)：工程标准与约定。
- [docs/interview-mapping.md](./docs/interview-mapping.md)：面试题映射。
- [docs/extensions-roadmap.md](./docs/extensions-roadmap.md)：加密、微服务、广告、支付、推荐等未来扩展。

## Cursor Cloud Agent 长跑方式

将 [CLOUD_AGENT_PROMPT.md](./CLOUD_AGENT_PROMPT.md) 的内容作为 Cursor Cloud Agent 的启动 prompt。Agent 应先读取 `CURRENT_STATE.md`、`NEXT_ACTIONS.md`、`DONE.md`、`BLOCKERS.md`，再按 `TASK_GRAPH_DETAILED.md` 执行。

本仓库同时提供 `.cursor/rules/` 和 `.cursor/skills/`：

- `.cursor/rules/`：项目级长期执行规则。
- `.cursor/skills/`：长跑、后端核心、可靠性 review、文档 ADR 的执行技能说明。

子 Agent 分派时必须使用 [SUBAGENT_TASK_PACKET.md](./SUBAGENT_TASK_PACKET.md)。如果使用 Composer 2.5，必须关闭 Fast mode。
