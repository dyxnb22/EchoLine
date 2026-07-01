# EchoLine Acceptance Matrix

本文件定义可验收标准。Agent 不应仅凭“代码写完”判断完成，必须对照本矩阵更新状态。

## 状态标记

- `todo`：未开始。
- `doing`：正在实现。
- `blocked`：阻塞，必须在 `BLOCKERS.md` 有记录。
- `partial`：有降级实现。
- `done`：代码、测试、文档均满足。

## Phase Acceptance

| Phase | 状态 | 必须能力 | 必须测试 | 必须文档 |
|---|---|---|---|---|
| Phase 0 | done | repo skeleton、长时执行文档 | `make help`, `make test`, `make smoke` | README、AGENTS、TASKS、规则文档 |
| Phase 1 | partial | auth、user、device、conversation、message REST API | unit + API smoke | data-model、api、iteration report |
| Phase 2 | partial | WebSocket 连接、心跳、在线推送 | WS smoke | websocket-protocol |
| Phase 3 | partial | 会话列表、未读、历史、离线 sync | sync/unread tests | reliability、api |
| Phase 4 | partial | 群聊、频道、presence、多端 | role/presence tests | architecture、data-model |
| Phase 5 | partial | 幂等、ACK、重试、去重、顺序性 | reliability tests | reliability ADR |
| Phase 6 | partial | Redis、MQ、worker、异步化 | eventbus tests | cache/MQ ADR |
| Phase 8 | partial | 限流、审计、监控、风控 | limiter/audit/metrics tests | observability notes |
| Phase 9 | todo | 测试、压测、报告、面试讲稿 | CI + k6 | load test reports |
| Phase 10 | todo | 增强项和探索项 | prototype tests | scaling/research reports |

## Core Capability Matrix

| 能力 | 状态 | 验收标准 |
|---|---|---|
| 用户注册/登录 | partial | 可注册、登录、鉴权；密码 hash；refresh token |
| 多设备登录 | partial | device 表/repo；WS device_id 绑定 |
| 私聊 | partial | 去重 direct API + 消息读写（待 DB integration） |
| 群聊 | partial | 创建/邀请/踢人/退群 + owner/admin/member 校验 |
| 频道 | partial | 创建/订阅/退订 + owner/admin 发布权限 |
| 会话列表 | partial | 列表 + unread 字段 |
| 历史消息 | partial | cursor 分页 + next_before |
| WebSocket | partial | 连接、鉴权、ping/pong、message.send、push、ACK |
| 离线同步 | partial | sync endpoint 已实现（待 integration） |
| ACK | partial | REST/WS ACK + forward-only 状态机 |
| 幂等去重 | partial | client_msg_id 唯一约束 + 重复返回原消息 |
| 顺序性 | partial | conversation seq 事务分配 |
| Redis presence | partial | Redis TTL presence on WS（可选 REDIS_ADDR） |
| MQ worker | partial | outbox drainer + Kafka/memory publish |
| 附件 | partial | 预签名上传、元数据入库、附件消息发送 |
| 搜索 | partial | PostgreSQL tsvector + 成员权限过滤 |
| 通知 | todo | 异步通知事件不阻塞主链路 |
| 限流 | partial | 登录、全局发消息、会话级 conv_send 限流（需 REDIS_ADDR） |
| 风控 | todo | 高频、重复内容基础规则 |
| 审计 | partial | audit_logs + 登录/撤回审计 |
| 可观测性 | partial | trace_id、Prometheus /metrics、WS/发送延迟指标 |
| 压测 | todo | k6 脚本和报告 |
| chaos | todo | Redis/MQ 故障演练 |
| 前端 | partial | 登录、聊天、分页、WS、乐观发送、附件、搜索 |
| 移动/桌面原型 | todo | PWA 或原型 ADR |

## Definition of Done

一个模块只有同时满足以下条件，才能标记为 `done`：

1. 代码实现完成。
2. 有相关单测、集成测试或 smoke test。
3. 相关文档更新。
4. `PROGRESS_LOG.md` 有记录。
5. 如涉及架构取舍，已有 ADR 或 `DECISIONS.md` 记录。

