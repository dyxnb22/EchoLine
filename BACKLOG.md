# EchoLine Backlog

本文件承接 `TASK_GRAPH_DETAILED.md` 之外的长期任务。Agent 完成主线后继续消费本文件，不要提前停止。

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
- 邮件或 webhook 通知 mock。
- 管理后台 skeleton。

## Stretch Backlog

- 大群在线用户分批推送。
- 热点 conversation 缓存。
- 消息冷热分层。
- 死信队列和重放工具。
- WebSocket gateway 多实例路由设计。
- 按 `conversation_id` 分片的 schema proposal。
- 附件病毒扫描 mock。
- chaos test：模拟 Redis / MQ / DB 故障。
- Grafana dashboard。
- OpenTelemetry tracing。
- SQL 慢查询分析报告。
- API gateway prototype。

## Research Backlog

- Telegram / WhatsApp / Discord / Slack 的消息模型对比。
- 读扩散、写扩散、混合扩散优缺点。
- 未读数在大群场景的近似与精确方案。
- 多端同步模型：account-level read vs device-level delivery。
- Kafka 顺序性与 partition key 选择。
- IM 系统分库分表策略。
- Presence 系统如何避免心跳风暴。
- 消息搜索索引一致性方案。
- Outbox pattern 与 MQ 事务一致性。
- WebSocket 网关水平扩展和连接路由。

## Future Extension Backlog

这些内容不属于 EchoLine MVP，但用于防止长跑任务耗尽，也用于把 EchoLine 扩成更大的作品集项目。

### Encryption

- 端到端加密威胁模型。
- per-chat key management ADR。
- message payload encryption prototype。
- key rotation 设计。
- lost device / revoked device 处理。
- 搜索与加密的冲突分析。

### Microservices Split

- 服务拆分边界 ADR：auth、user、conversation、message、realtime、media、search、notification、audit。
- API gateway prototype。
- auth service split prototype。
- message service split prototype。
- realtime gateway split prototype。
- worker service split prototype。
- service-to-service auth。
- distributed tracing across services。
- migration plan：modular monolith 到 microservices。

### Ads

- sponsored channel message 数据模型。
- ad campaign / creative / placement 设计。
- 频控、去重和曝光审计。
- 广告投放事件流。
- 广告与用户隐私边界。
- 广告报表 mock。

### Payments

- paid channel subscription ADR。
- payment ledger 数据模型。
- invoice / payment / refund 状态机。
- webhook 幂等处理。
- entitlement 校验。
- paid media message prototype。
- 支付审计日志。

### Recommendation

- channel recommendation research。
- contact/friend recommendation research。
- ranking feature log schema。
- candidate generation prototype。
- privacy-safe recommendation constraints。
- anti-spam recommendation controls。

