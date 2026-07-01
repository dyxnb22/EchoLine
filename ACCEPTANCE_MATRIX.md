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
| Phase 1 | todo | auth、user、device、conversation、message REST API | unit + API smoke | data-model、api、iteration report |
| Phase 2 | todo | WebSocket 连接、心跳、在线推送 | WS smoke | websocket-protocol |
| Phase 3 | todo | 会话列表、未读、历史、离线 sync | sync/unread tests | reliability、api |
| Phase 4 | todo | 群聊、频道、presence、多端 | role/presence tests | architecture、data-model |
| Phase 5 | todo | 幂等、ACK、重试、去重、顺序性 | reliability tests | reliability ADR |
| Phase 6 | todo | Redis、MQ、worker、异步化 | eventbus tests | cache/MQ ADR |
| Phase 7 | todo | 附件、搜索、通知 | media/search tests | api、scaling |
| Phase 8 | todo | 限流、审计、监控、风控 | limiter/audit/metrics tests | observability notes |
| Phase 9 | todo | 测试、压测、报告、面试讲稿 | CI + k6 | load test reports |
| Phase 10 | todo | 增强项和探索项 | prototype tests | scaling/research reports |

## Core Capability Matrix

| 能力 | 状态 | 验收标准 |
|---|---|---|
| 用户注册/登录 | todo | 可注册、登录、鉴权；密码 hash；错误码稳定 |
| 多设备登录 | todo | device 记录和 session 管理可用 |
| 私聊 | todo | 双方会话唯一，消息可写可读 |
| 群聊 | todo | 成员角色和权限正确 |
| 频道 | todo | owner/admin 可发布，subscriber 可接收 |
| 会话列表 | todo | 最近消息排序，未读数展示 |
| 历史消息 | todo | cursor 分页不漏不重 |
| WebSocket | todo | 心跳、断线清理、在线推送 |
| 离线同步 | todo | 重连后可补拉缺失消息 |
| 多端同步 | todo | 多设备在线接收，读状态一致 |
| ACK | todo | delivered/read 状态可记录 |
| 幂等去重 | todo | 同一 client_msg_id 不重复入库 |
| 顺序性 | todo | conversation 内 seq 单调 |
| Redis presence | todo | TTL 过期和 heartbeat 正常 |
| MQ worker | todo | message.created 可异步消费 |
| 附件 | todo | 预签名上传、元数据、权限校验 |
| 搜索 | todo | 消息搜索和权限过滤 |
| 通知 | todo | 异步通知事件不阻塞主链路 |
| 限流 | todo | 用户/IP/会话维度限流 |
| 风控 | todo | 高频、重复内容基础规则 |
| 审计 | todo | 登录、撤回、管理操作可追溯 |
| 可观测性 | todo | logs、metrics、trace_id |
| 压测 | todo | k6 脚本和报告 |
| chaos | todo | Redis/MQ 故障演练 |
| 前端 | todo | 登录、会话、聊天、实时消息 |
| 移动/桌面原型 | todo | PWA 或原型 ADR |

## Definition of Done

一个模块只有同时满足以下条件，才能标记为 `done`：

1. 代码实现完成。
2. 有相关单测、集成测试或 smoke test。
3. 相关文档更新。
4. `PROGRESS_LOG.md` 有记录。
5. 如涉及架构取舍，已有 ADR 或 `DECISIONS.md` 记录。

