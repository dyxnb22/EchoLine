# Future Extensions Roadmap

> **Status (2026-07-02):** 下列方向中 **骨架 / 原型 / ADR** 已在 T001–T440 与 final completion 中落地；**生产级**能力仍为 future work。实现对照见 [`FINAL_COMPLETION_MANIFEST.md`](../FINAL_COMPLETION_MANIFEST.md) 与 [`BACKLOG.md`](../BACKLOG.md)。

本文件记录 EchoLine 主线之外的长期扩展方向。它们不是 MVP 优先任务，但用于支撑更长周期工程投入。

## 加密

**已落地（demo）：** ADR 0010/0011/0026，`POST/GET /api/encryption/keys`，前端 XOR demo（`frontend/src/lib/e2ee.ts`）。

**仍为 future：** Signal/MLS、生产密钥轮换、加密与搜索/审核冲突的完整方案。

目标：探索 Telegram-like 平台在安全和可用性之间的取舍。

候选任务：

- 端到端加密威胁模型。 ✅ ADR 0010
- secret chat 与普通 cloud chat 的产品边界。
- per-chat key management。 ✅ ADR 0011
- device key registration。 ✅ key bundle API
- key rotation。
- revoked device 后的历史消息访问策略。
- 加密消息与搜索、审核、举报之间的冲突。
- at-rest encryption prototype。

面试讲述点：

- 为什么端到端加密会影响搜索、风控和多端同步。
- 如何区分传输加密、静态加密和端到端加密。

## 微服务拆分

**已落地（设计 + 原型）：** ADR 0009/0028 边界，`deploy/gateway/` + ADR 0027，OTel stub ADR 0029。

**仍为 future：** 独立服务部署、service mesh、gRPC 内网 API。

目标：从模块化单体演进到服务化架构，训练服务边界、通信、数据一致性和可观测性。

候选服务：

- API gateway。 ✅ nginx prototype
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

- 服务拆分 ADR。 ✅ 0028
- service-to-service auth。
- gRPC 或 HTTP internal API。
- distributed tracing。 ✅ stub
- outbox/event-driven consistency。 ✅ outbox + worker
- migration plan：从单体拆到服务。

## 广告

**已落地（skeleton）：** ADR 0012/0020，`internal/ads/`，campaign + impression API，频控 migration 00014。

**仍为 future：** 生产投放引擎、报表、隐私合规。

目标：探索消息平台商业化能力，但不破坏核心聊天体验。

候选任务：

- sponsored channel message。
- ad campaign 数据模型。 ✅
- creative / placement / impression / click。 ✅ impression
- 频控和去重。 ✅ frequency_cap
- 广告审计。
- 广告与隐私边界。
- 广告报表 mock。

面试讲述点：

- 广告投放如何异步化。
- 如何做曝光去重和频控。
- 如何避免影响消息主链路。

## 支付

**已落地（skeleton）：** ADR 0019，`internal/payment/`，ledger + settle；settle 可 grant 频道 entitlement（00016）。

**仍为 future：** Stripe 集成、退款、生产账本对账。

目标：支持频道订阅、付费内容或创作者收入模型。

候选任务：

- paid channel subscription。 ✅ entitlement gate
- payment ledger。 ✅
- invoice / payment / refund 状态机。
- payment webhook 幂等。
- entitlement 校验。 ✅ ADR 0030
- paid media message。
- 支付审计。

面试讲述点：

- 为什么支付系统需要 ledger。
- webhook 为什么必须幂等。
- 支付状态机如何设计。

## 推荐

**已落地（skeleton）：** ADR 0021，`internal/recommendation/`，channels + friends API。

**仍为 future：** 个性化排序、特征日志、冷启动策略。

目标：探索频道、群组、联系人推荐，但保持隐私和风控边界。

候选任务：

- channel recommendation research。 ✅ API
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
