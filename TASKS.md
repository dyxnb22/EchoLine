# EchoLine 可持续消费任务图

> **Status (2026-07-02):** Phase 0–10 主线与 T001–T440 已按 [`FINAL_COMPLETION_MANIFEST.md`](./FINAL_COMPLETION_MANIFEST.md) 关闭。下文保留各 phase 的历史定义与验收标准，供回顾与 interview 对照。**当前可选工作**见 [`NEXT_ACTIONS.md`](./NEXT_ACTIONS.md) 与 [`DONE.md`](./DONE.md) 的 Post-closure 节。

本文件是 Cursor Cloud Agent 的长期任务入口。执行时优先推进当前 phase，完成验收后再进入下一 phase。

细粒度任务见 `TASK_GRAPH_DETAILED.md`。验收状态见 `ACCEPTANCE_MATRIX.md`。主线之外的任务见 `BACKLOG.md` 和 `docs/extensions-roadmap.md`。

## Phase 0：仓库初始化与架构约束

目标：建立可长期执行的 repo skeleton 和项目约束。

子任务：

- 创建核心文档：`README.md`、`AGENTS.md`、`TASKS.md`、`EXECUTION_RULES.md`、`PROGRESS_LOG.md`、`DONE.md`。
- 创建文档目录：`docs/architecture.md`、`docs/interview-mapping.md`、`docs/data-model.md`、`docs/api.md`、`docs/websocket-protocol.md`、`docs/reliability.md`、`docs/scaling.md`、`docs/adr/README.md`。
- 创建工程目录骨架：`backend/`、`frontend/`、`scripts/`、`loadtests/`、`reports/`。
- 创建基础 `Makefile`、`.env.example`、`docker-compose.yml`、CI skeleton。

输入依赖：无。

产出物：可被 Agent 理解和持续推进的仓库骨架。

验收标准：

- 所有必须文档存在。
- 目录结构清晰。
- `make help` 可运行。
- `PROGRESS_LOG.md` 有初始化记录。

失败降级：

- 如暂不能创建完整 CI，则保留 `.github/workflows/ci.yml` skeleton。

更新文件：

- `README.md`
- `AGENTS.md`
- `TASKS.md`
- `EXECUTION_RULES.md`
- `PROGRESS_LOG.md`
- `DONE.md`
- `reports/iteration-01.md`

## Phase 1：核心数据模型与基础 API

目标：建立用户、设备、会话、成员、消息的核心数据模型和基础 REST API。

子任务：

- 初始化后端服务。
- 实现 PostgreSQL migration。
- 实现用户注册、登录、refresh token 基础流。
- 实现创建私聊、创建群聊、查询会话。
- 实现发送消息的 REST API。
- 补充数据模型文档和 API 文档。

输入依赖：Phase 0。

产出物：

- `backend/cmd/api`
- `backend/internal/auth`
- `backend/internal/user`
- `backend/internal/conversation`
- `backend/internal/message`
- `backend/migrations`

验收标准：

- 能创建用户、登录、创建私聊 / 群聊、写入消息。
- 核心 API 有测试或 smoke script。
- `docs/data-model.md` 和 `docs/api.md` 已更新。

失败降级：

- 暂不接 frontend。
- 如 auth 阻塞，先用 dev token 中间件推进核心链路。

更新文件：

- `docs/data-model.md`
- `docs/api.md`
- `DONE.md`
- `PROGRESS_LOG.md`

## Phase 2：实时消息主链路

目标：实现 WebSocket 连接管理和在线消息推送。

子任务：

- 实现 WebSocket endpoint。
- 实现连接注册、注销、心跳。
- 实现发送消息后推送在线接收方。
- 定义 WS 消息类型和错误格式。
- 写最小 WS smoke test。

输入依赖：Phase 1。

产出物：

- `backend/internal/realtime`
- `docs/websocket-protocol.md`
- `scripts/smoke-test.sh`

验收标准：

- 在线用户能实时收到消息。
- 断线后连接状态能清理。
- WS 协议文档已更新。

失败降级：

- 若 WS 不稳定，保留 REST send + polling fallback，并记录 blocker。

更新文件：

- `docs/websocket-protocol.md`
- `PROGRESS_LOG.md`
- `DONE.md`

## Phase 3：会话、未读、历史与离线同步

目标：让用户能看到会话列表、未读数、历史消息，并在离线后补拉消息。

子任务：

- 实现会话列表 API。
- 实现 `last_read_seq`、`latest_seq`。
- 实现历史消息 cursor pagination。
- 实现 sync endpoint。
- 实现基础未读数计算。

输入依赖：Phase 2。

产出物：

- `backend/internal/conversation`
- `backend/internal/message`
- `backend/internal/delivery`
- `docs/reliability.md`

验收标准：

- 用户重登后能拉取离线期间消息。
- 未读数能基于 seq 计算。
- 历史消息分页稳定，不重复、不漏页。

失败降级：

- 先按 conversation cursor 实现，不做全局 inbox cursor。

更新文件：

- `docs/reliability.md`
- `docs/api.md`
- `DONE.md`
- `PROGRESS_LOG.md`

## Phase 4：群聊、频道、在线状态、多端同步

目标：扩展 Telegram-like 能力，支持基础群聊、频道、presence、多端 session。

子任务：

- 群成员角色：owner、admin、member。
- 频道模型：owner 发布，subscriber 接收。
- 设备表和多端 session。
- Redis presence TTL 设计和实现。
- 多端在线推送。

输入依赖：Phase 3。

产出物：

- `backend/internal/channel`
- `backend/internal/presence`
- `backend/internal/user/device`

验收标准：

- 多端可同时在线接收消息。
- presence 使用 TTL，断线后最终过期。
- 频道消息能被订阅者拉取或接收。

失败降级：

- 频道先只做广播，不做评论。
- ACK 仍可先维持单设备，Phase 5 再扩展。

更新文件：

- `docs/architecture.md`
- `docs/api.md`
- `docs/data-model.md`
- `PROGRESS_LOG.md`

## Phase 5：可靠性、幂等、ACK、重试、去重、顺序性

目标：补齐消息系统可靠性主线。

子任务：

- 引入 `client_msg_id` 幂等键。
- conversation 内递增 `seq`。
- delivery ACK 状态。
- 客户端重试去重。
- 失败推送后的拉取补偿。
- 编写可靠性测试。

输入依赖：Phase 4。

产出物：

- `backend/internal/delivery`
- `docs/reliability.md`
- `docs/adr/0002-message-sequence.md`

验收标准：

- 重复请求不产生重复消息。
- 同一 conversation 内消息按 seq 读取。
- ACK 可记录和查询。

失败降级：

- ACK 先做单设备，再扩展多端。

更新文件：

- `docs/reliability.md`
- `docs/adr/0002-message-sequence.md`
- `DONE.md`
- `PROGRESS_LOG.md`

## Phase 6：缓存、MQ、异步化

目标：将同步链路中的通知、审计、搜索索引等拆到异步事件。

子任务：

- Redis 缓存会话 summary 和 presence。
- 引入 Kafka/Redpanda 或本地 event bus 接口。
- 实现 worker。
- 消息写入后发布 `message.created`。
- 异步消费通知、审计、搜索索引事件。

输入依赖：Phase 5。

产出物：

- `backend/cmd/worker`
- `backend/internal/eventbus`
- `docs/adr/0031-cache-and-mq-responsibilities.md`

验收标准：

- 消息写入主链路不依赖通知成功。
- worker 可消费事件。
- Redis 和 DB 的数据边界有文档说明。

失败降级：

- 没 Kafka 时用内存 event bus 接口，保留实现替换点。

更新文件：

- `docs/scaling.md`
- `docs/adr/0031-cache-and-mq-responsibilities.md`
- `PROGRESS_LOG.md`

## Phase 7：附件、搜索、通知

目标：补齐媒体附件、全文搜索和通知事件链路。

子任务：

- MinIO/S3 预签名上传。
- 附件元数据入库。
- OpenSearch 消息索引。
- 搜索 API。
- 通知事件表或 worker handler。

输入依赖：Phase 6。

产出物：

- `backend/internal/media`
- `backend/internal/search`
- `backend/internal/notification`

验收标准：

- 可上传附件并发送附件消息。
- 可按关键词搜索消息。
- 通知事件不阻塞发送消息主链路。

失败降级：

- 搜索先用 DB LIKE，再替换为 OpenSearch。

更新文件：

- `docs/api.md`
- `docs/scaling.md`
- `reports/iteration-01.md`

## Phase 8：限流、审计、监控、风控基础

目标：引入平台治理能力。

子任务：

- 用户维度和 IP 维度限流。
- 会话维度发言限流。
- append-only audit log。
- structured logs。
- Prometheus metrics。
- trace_id 贯穿 API、DB、MQ、WS。
- 基础风险规则：重复内容、高频发言、异常失败率。

输入依赖：Phase 7。

产出物：

- `backend/internal/rate_limit`
- `backend/internal/audit`
- `backend/internal/observability`
- `backend/internal/risk`

验收标准：

- 高频发消息会被限流。
- 关键操作有审计记录。
- 至少有消息发送成功率、延迟、WS 连接数指标。

失败降级：

- 风控先做规则，不做复杂模型。

更新文件：

- `docs/architecture.md`
- `docs/interview-mapping.md`
- `PROGRESS_LOG.md`

## Phase 9：测试、压测、文档、报告

目标：让项目从“可运行”变成“可证明”。

子任务：

- 补充核心单元测试。
- 补充 API integration test。
- 补充 WS smoke test。
- 编写 k6 压测脚本。
- 产出 load test report。
- 整理面试讲述模板。

输入依赖：Phase 8。

产出物：

- `loadtests/k6-send-message.js`
- `loadtests/k6-ws-connect.js`
- `reports/load-test-01.md`

验收标准：

- CI 通过。
- 有压测脚本和报告。
- 文档能讲清主要架构取舍。

失败降级：

- 压测环境不足时，记录本地基准和限制。

更新文件：

- `reports/load-test-01.md`
- `docs/interview-mapping.md`
- `DONE.md`

## Phase 10：增强项与探索项

目标：围绕高阶系统设计问题做原型和 ADR。

子任务：

- 热点群 fanout 策略。
- 大群在线用户分批推送。
- 死信队列和重放工具。
- 消息冷热分层。
- WebSocket gateway 多实例路由设计。
- 分库分表设计。

输入依赖：Phase 9。

产出物：

- ADR。
- 小原型。
- research report。

验收标准：

- 至少完成 2 个增强项。
- 至少完成 1 篇研究报告。

失败降级：

- 如果工程实现成本过高，产出 ADR + 可验证小原型。

更新文件：

- `docs/scaling.md`
- `reports/`
- `DONE.md`

## Secondary Backlog

- 前端会话列表、聊天窗口、频道页基础版。
- 管理员踢人、禁言、频道 owner 权限。
- 消息编辑 / 撤回。
- pinned message。
- 用户拉黑。
- 简单 notification center。
- API 错误码统一。
- OpenAPI spec。
- 种子数据脚本。
- 本地开发脚本优化。

## Stretch Backlog

- 大群在线用户分批推送。
- 热点 conversation 缓存。
- 消息冷热分层。
- 死信队列和重放工具。
- WebSocket gateway 多实例路由设计。
- 按 `conversation_id` 分片的 schema proposal。
- 附件病毒扫描 mock。
- 管理后台。
- chaos test：模拟 Redis / MQ 故障。

## Research Backlog

- Telegram / WhatsApp / Discord / Slack 的消息模型对比。
- 读扩散、写扩散、混合扩散优缺点。
- 未读数在大群场景的近似与精确方案。
- 多端同步模型：account-level read vs device-level delivery。
- Kafka 顺序性与 partition key 选择。
- IM 系统分库分表策略。
- Presence 系统如何避免心跳风暴。
- 消息搜索索引一致性方案。

## Future Extension Backlog

当 Phase 0-10、secondary、stretch、research backlog 都完成或暂时阻塞时，继续执行 `docs/extensions-roadmap.md`：

- 加密：E2EE、key management、静态加密、加密与搜索/风控冲突。
- 微服务拆分：auth、message、realtime、media、search、notification、audit 等服务边界。
- 广告：sponsored channel message、频控、曝光审计、广告报表。
- 支付：paid channel subscription、ledger、webhook 幂等、entitlement。
- 推荐：频道推荐、联系人推荐、ranking feature log、隐私和风控边界。
