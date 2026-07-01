# Future Extensions Roadmap

本文件记录 EchoLine 主线之外的长期扩展方向。它们不是 MVP，也不是 Phase 1-10 的优先任务，但用于支撑更长周期工程投入，并防止长时 Agent 在主线完成后无任务可做。

## 加密

目标：探索 Telegram-like 平台在安全和可用性之间的取舍。

候选任务：

- 端到端加密威胁模型。
- secret chat 与普通 cloud chat 的产品边界。
- per-chat key management。
- device key registration。
- key rotation。
- revoked device 后的历史消息访问策略。
- 加密消息与搜索、审核、举报之间的冲突。
- at-rest encryption prototype。

面试讲述点：

- 为什么端到端加密会影响搜索、风控和多端同步。
- 如何区分传输加密、静态加密和端到端加密。

## 微服务拆分

目标：从模块化单体演进到服务化架构，训练服务边界、通信、数据一致性和可观测性。

候选服务：

- API gateway。
- auth service。
- user service。
- conversation service。
- message service。
- realtime gateway。
- media service。
- search service。
- notification service。
- audit service。

候选任务：

- 服务拆分 ADR。
- service-to-service auth。
- gRPC 或 HTTP internal API。
- distributed tracing。
- outbox/event-driven consistency。
- migration plan：从单体拆到服务。

## 广告

目标：探索消息平台商业化能力，但不破坏核心聊天体验。

候选任务：

- sponsored channel message。
- ad campaign 数据模型。
- creative / placement / impression / click。
- 频控和去重。
- 广告审计。
- 广告与隐私边界。
- 广告报表 mock。

面试讲述点：

- 广告投放如何异步化。
- 如何做曝光去重和频控。
- 如何避免影响消息主链路。

## 支付

目标：支持频道订阅、付费内容或创作者收入模型。

候选任务：

- paid channel subscription。
- payment ledger。
- invoice / payment / refund 状态机。
- payment webhook 幂等。
- entitlement 校验。
- paid media message。
- 支付审计。

面试讲述点：

- 为什么支付系统需要 ledger。
- webhook 为什么必须幂等。
- 支付状态机如何设计。

## 推荐

目标：探索频道、群组、联系人推荐，但保持隐私和风控边界。

候选任务：

- channel recommendation research。
- candidate generation。
- ranking feature log。
- 用户兴趣标签。
- 内容安全过滤。
- 推荐曝光日志。
- 反 spam 控制。

面试讲述点：

- 推荐系统如何与主业务解耦。
- 如何处理冷启动。
- 如何记录曝光和反馈。

