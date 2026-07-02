# EchoLine Acceptance Matrix

本文件定义可验收标准。Agent 不应仅凭“代码写完”判断完成，必须对照本矩阵更新状态。

> **2026-07-02 note:** 主线 Phase 0–10 与 T001–T440 已关闭（见 `FINAL_COMPLETION_MANIFEST.md`）。矩阵保留为能力对照表；`partial` 通常表示缺 Postgres 集成 smoke 或仍为 demo/stub。

## 状态标记

- `todo`：未开始。
- `doing`：正在实现。
- `blocked`：阻塞，必须在 `BLOCKERS.md` 有记录。
- `partial`：有降级实现或缺集成验收。
- `done`：代码、测试、文档均满足。

## Phase Acceptance

| Phase | 状态 | 必须能力 | 必须测试 | 必须文档 |
|---|---|---|---|---|
| Phase 0 | done | repo skeleton、长时执行文档 | `make help`, `make test`, `make smoke` | README、AGENTS、TASKS、规则文档 |
| Phase 1 | done | auth、user、device、conversation、message REST API | unit + integration (opt-in) | data-model、api、iteration report |
| Phase 2 | done | WebSocket 连接、心跳、在线推送 | WS unit + protocol doc | websocket-protocol |
| Phase 3 | done | 会话列表、未读、历史、离线 sync | sync tests | reliability、api |
| Phase 4 | done | 群聊、频道、presence、多端 | role/presence tests | architecture、data-model |
| Phase 5 | done | 幂等、ACK、重试、去重、顺序性 | reliability tests | reliability ADR |
| Phase 6 | partial | Redis、MQ、worker、异步化 | eventbus tests | cache/MQ ADR 0031 |
| Phase 7 | done | 附件、搜索、通知 | media/search tests | api、scaling |
| Phase 8 | done | 限流、审计、监控、风控 | limiter/audit/metrics tests | observability notes |
| Phase 9 | done | CI、k6、报告、面试讲稿 | CI + k6 dry-run | load test reports |
| Phase 10 | partial | 增强项和探索项 | prototype tests | scaling/research reports |

**Environment gap:** `make smoke-full` 需本地 Docker compose；云 Agent 环境标记为 `blocked`（`BLOCKERS.md`）。

## Core Capability Matrix

| 能力 | 状态 | 验收标准 |
|---|---|---|
| 用户注册/登录 | done | 可注册、登录、鉴权；密码 hash；refresh token |
| 多设备登录 | done | device 表/repo；WS device_id 绑定 |
| 私聊 | done | 去重 direct API + 消息读写 |
| 群聊 | done | 创建/邀请/踢人/退群 + owner/admin/member 校验 |
| 频道 | done | 创建/订阅/退订 + owner/admin 发布 + 付费门控 |
| 会话列表 | done | 列表 + unread 字段 |
| 历史消息 | done | cursor 分页 + next_before |
| WebSocket | done | 连接、鉴权、ping/pong、message.send、push、ACK |
| 离线同步 | done | sync endpoint + device cursors |
| ACK | done | REST/WS ACK + forward-only 状态机 |
| 幂等去重 | done | client_msg_id 唯一约束 + 重复返回原消息 |
| 顺序性 | done | conversation seq 事务分配 |
| Redis presence | done | Redis TTL presence on WS（可选 REDIS_ADDR） |
| MQ worker | partial | outbox drainer + Kafka/memory publish |
| 附件 | done | 预签名上传、元数据入库、附件消息发送 |
| 搜索 | done | PostgreSQL tsvector + 成员权限过滤 |
| 通知 | done | in-app notifications + push skeleton |
| 限流 | done | 登录、全局发消息、会话级 conv_send 限流 |
| 风控 | partial | 重复内容 spam checker |
| 审计 | done | audit_logs + 登录/撤回审计 |
| 可观测性 | partial | trace_id、Prometheus /metrics、OTel stub |
| 压测 | done | k6 脚本（CI dry-run） |
| chaos | done | Redis/MQ 故障演练脚本 |
| 前端 | done | 登录、聊天、分页、WS、乐观发送、附件、搜索、router |
| 移动/桌面原型 | partial | PWA + ADR 0014/0015 |

## Definition of Done

一个模块只有同时满足以下条件，才能标记为 `done`：

1. 代码实现完成。
2. 有相关单测、集成测试或 smoke test。
3. 相关文档更新。
4. `PROGRESS_LOG.md` 有记录。
5. 如涉及架构取舍，已有 ADR 或 `DECISIONS.md` 记录。
