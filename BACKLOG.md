# EchoLine Backlog

> **Status (2026-07-02): closed.** 下列条目已在 T001–T440 与 final completion 中实现、文档化或显式 stub。本文件保留为**历史任务清单**与 interview 对照；新工作见 [`NEXT_ACTIONS.md`](./NEXT_ACTIONS.md) 与 [`docs/extensions-roadmap.md`](./docs/extensions-roadmap.md)。

完成范围摘要见 [`FINAL_COMPLETION_MANIFEST.md`](./FINAL_COMPLETION_MANIFEST.md)。

## Secondary Backlog — done

- [x] 前端会话列表、聊天窗口、频道页基础版。
- [x] 管理员踢人、禁言、频道 owner 权限。
- [x] 消息编辑 / 撤回。
- [x] pinned message。
- [x] 用户拉黑。
- [x] 简单 notification center。
- [x] API 错误码统一（`apierror` envelope）。
- [x] OpenAPI spec（`docs/openapi.yaml`）。
- [x] 种子数据脚本（`cmd/seed`）。
- [x] 本地开发脚本优化（`Makefile`, compose profiles）。
- [x] 邮件或 webhook 通知 mock（webhook worker）。
- [x] 管理后台 skeleton（`AdminPanel`, admin APIs）。

## Stretch Backlog — done / stubbed

- [x] 大群在线用户分批推送（fanout worker, ADR 0003）。
- [x] 热点 conversation 缓存（设计 + ADR 0005）。
- [x] 消息冷热分层（ADR 0006）。
- [x] 死信队列和重放工具（DLQ + `cmd/replay`）。
- [x] WebSocket gateway 多实例路由设计（ADR 0004, `deploy/gateway`）。
- [x] 按 `conversation_id` 分片的 schema proposal（ADR 0007）。
- [x] 附件病毒扫描 mock（`docs/virus-scan-mock.md`）。
- [x] chaos test：模拟 Redis / MQ / DB 故障（`docs/chaos-playbook.md`）。
- [x] Grafana dashboard（compose + metrics endpoint）。
- [x] OpenTelemetry tracing（stub ADR 0029）。
- [x] SQL 慢查询分析报告（`reports/review-performance.md`）。
- [x] API gateway prototype（`deploy/gateway`, ADR 0027）。

## Research Backlog — documented

研究结论已写入 ADR、`docs/research-*.md` 与 `docs/reliability.md`：

- Telegram / WhatsApp / Discord / Slack 的消息模型对比 → `docs/research-telegram-whatsapp.md`, `research-discord-slack.md`
- 读扩散、写扩散、混合扩散 → ADR 0003, `docs/interview-fanout.md`
- 未读数在大群场景的近似与精确方案 → `docs/research-fanout-unread.md`
- 多端同步模型 → `docs/interview-multi-device-sync.md`
- Kafka 顺序性与 partition key → `docs/research-kafka-sharding.md`
- IM 分库分表 → ADR 0007, `docs/scaling.md`
- Presence 心跳风暴 → `docs/research-presence-search-outbox.md`
- 搜索索引一致性 → outbox + worker
- Outbox pattern → `docs/reliability.md`, migration 00004
- WebSocket 网关水平扩展 → ADR 0004

## Future Extension Backlog

长期扩展方向见 [`docs/extensions-roadmap.md`](./docs/extensions-roadmap.md)（E2EE、微服务、广告、支付、推荐原型已 stub）。

### Encryption — prototype

- [x] 威胁模型 ADR 0010
- [x] Key management ADR 0011/0026
- [x] Client demo（`frontend/src/lib/e2ee.ts`）
- [ ] Production Signal/MLS — future

### Microservices Split — design

- [x] 拆分边界 ADR 0009/0028
- [x] Gateway prototype
- [ ] 独立服务部署 — future

### Ads / Payments / Recommendation — skeleton

- [x] 数据模型 ADR 0012/0020
- [x] Payment ledger ADR 0019 + settle/grant
- [x] Recommendation API + ADR 0021
- [ ] 生产级计费与风控 — future
