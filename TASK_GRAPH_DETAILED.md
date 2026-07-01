# EchoLine Detailed Task Graph

本文件把 `TASKS.md` 的 phase 拆成可长期消费的 atomic tasks。Agent 执行时应优先从当前 phase 的 P0/P1 任务开始，每完成 3-5 个任务做一次 repo-based context compaction。

## Track A：Backend Core

| ID | 优先级 | 任务 | 输入依赖 | 产出 | 验收 |
|---|---|---|---|---|---|
| A001 | P0 | 初始化 Go module 和 API server skeleton | Phase 0 | `backend/go.mod`, `cmd/api` | `go test ./...` 可运行 |
| A002 | P0 | 添加配置加载和 env 校验 | A001 | config package | 缺少必要 env 时有明确错误 |
| A003 | P0 | 添加 health endpoint | A001 | `/health` | smoke test 通过 |
| A004 | P0 | 接入 PostgreSQL 连接池 | A002 | db package | 可 ping DB |
| A005 | P0 | 选择 migration 工具并添加初始 migration | A004 | migrations | migration 可执行 |
| A006 | P0 | 实现 users 表和 repository | A005 | user repo | unit tests |
| A007 | P0 | 实现 password hash | A006 | auth helper | hash/verify tests |
| A008 | P0 | 实现 register API | A006 | auth handler | 注册成功和重复用户名测试 |
| A009 | P0 | 实现 login API | A007 | token response | 正确密码成功、错误密码失败 |
| A010 | P0 | 实现 JWT middleware | A009 | auth middleware | 保护路由可鉴权 |
| A011 | P1 | 实现 refresh token skeleton | A009 | token service | refresh 成功测试 |
| A012 | P0 | 实现 devices 表 | A005 | device repo | migration + repo tests |
| A013 | P0 | 实现 conversations 表 | A005 | conversation repo | 创建私聊/群聊测试 |
| A014 | P0 | 实现 conversation_members 表 | A013 | member repo | 成员权限测试 |
| A015 | P0 | 实现 messages 表 | A013 | message repo | 写入/查询测试 |
| A016 | P0 | 实现创建私聊 API | A013 | REST endpoint | 同一 pair 去重 |
| A017 | P0 | 实现创建群聊 API | A014 | REST endpoint | owner/member 正确 |
| A018 | P0 | 实现发送消息 REST API | A015 | REST endpoint | DB 写入成功 |
| A019 | P0 | 实现历史消息分页 API | A015 | REST endpoint | cursor 不重不漏 |
| A020 | P1 | 统一错误码和 response envelope | A008-A019 | error package | API 错误格式一致 |
| A021 | P1 | 添加 OpenAPI skeleton | A016-A020 | `docs/openapi.yaml` | 文档覆盖核心 API |
| A022 | P1 | 添加 seed script | A008-A018 | seed command | 可创建测试用户和会话 |

## Track B：Realtime Gateway

| ID | 优先级 | 任务 | 输入依赖 | 产出 | 验收 |
|---|---|---|---|---|---|
| B001 | P0 | 实现 WebSocket endpoint | A010 | `realtime` | 可连接 |
| B002 | P0 | 实现 WS 鉴权握手 | B001 | auth handshake | 无 token 拒绝 |
| B003 | P0 | 实现 connection manager | B002 | conn registry | 连接注册/注销测试 |
| B004 | P0 | 实现 ping/pong heartbeat | B003 | heartbeat | idle 连接清理 |
| B005 | P0 | 实现 server event envelope | B003 | protocol structs | marshal/unmarshal tests |
| B006 | P0 | 实现 `message.send` over WS | A018,B005 | WS send handler | 在线发送成功 |
| B007 | P0 | 实现在线接收方推送 | B006 | push path | 双客户端实时接收 |
| B008 | P1 | 实现 WS error envelope | B005 | error handling | 错误有 request_id |
| B009 | P1 | 实现 reconnect fallback 指南 | B007 | docs | 文档更新 |
| B010 | P1 | 添加 WS smoke test | B007 | script/test | smoke 可复现 |
| B011 | P2 | 实现 typing indicator | B003 | realtime event | 不落 DB |
| B012 | P2 | 实现 gateway 多实例路由 ADR | B003 | ADR | 说明 Redis registry |

## Track C：Conversation, Sync, Unread

| ID | 优先级 | 任务 | 输入依赖 | 产出 | 验收 |
|---|---|---|---|---|---|
| C001 | P0 | 在 conversation 中维护 `latest_seq` | A015 | migration/repo | seq 单调递增 |
| C002 | P0 | 实现 conversation 内 seq 分配 | C001 | transaction | 并发测试 |
| C003 | P0 | 实现 conversation list API | A013-A015 | endpoint | 最近消息排序 |
| C004 | P0 | 实现 `last_read_seq` | A014,C002 | read state | mark read 测试 |
| C005 | P0 | 实现未读数计算 | C004 | unread service | unread 正确 |
| C006 | P0 | 实现 sync endpoint | C002 | `/api/sync` | 离线补拉 |
| C007 | P1 | 实现 per-device sync cursor | A012,C006 | sync state | 多端 cursor 测试 |
| C008 | P1 | 实现消息编辑 | A015 | API + events | 版本/状态正确 |
| C009 | P1 | 实现消息撤回 | A015 | API + events | 审计预留 |
| C010 | P2 | 实现 pinned message | A013,A015 | API | pin/unpin 测试 |

## Track D：Reliability

| ID | 优先级 | 任务 | 输入依赖 | 产出 | 验收 |
|---|---|---|---|---|---|
| D001 | P0 | 引入 `client_msg_id` 幂等唯一约束 | A015 | migration | 重复发送返回同一消息 |
| D002 | P0 | 实现 idempotency repository | D001 | service | race tests |
| D003 | P0 | 实现 message ACK API/WS event | B007 | delivery module | ACK 状态可查 |
| D004 | P0 | 实现 delivered/read 状态机 | D003 | delivery state | 状态只前进 |
| D005 | P1 | 实现多设备 ACK 聚合策略 | C007,D004 | docs + code | account read 正确 |
| D006 | P1 | 实现 outbox table | A005 | migration/service | message + event 同事务 |
| D007 | P1 | 实现 outbox publisher | D006 | worker | 失败可重试 |
| D008 | P1 | 实现 dead letter skeleton | D007 | worker | DLQ 可记录 |
| D009 | P2 | 编写可靠性故障注入测试 | D001-D008 | tests | DB/WS/MQ 部分失败可解释 |
| D010 | P2 | 完成 reliability ADR 套件 | D001-D008 | ADRs | 能讲 at-least-once |

## Track E：Group, Channel, Fanout

| ID | 优先级 | 任务 | 输入依赖 | 产出 | 验收 |
|---|---|---|---|---|---|
| E001 | P0 | 群成员角色 owner/admin/member | A014 | role checks | 权限测试 |
| E002 | P0 | 群成员邀请/退出/踢人 | E001 | APIs | 边界权限正确 |
| E003 | P0 | 频道数据模型 | A013 | migration | channel conversation |
| E004 | P0 | 频道订阅/退订 | E003 | APIs | subscriber 正确 |
| E005 | P0 | 频道发布权限 | E003 | authz | 只有 owner/admin 可发 |
| E006 | P1 | 小群在线 fanout | B007,E001 | fanout service | 在线成员收到 |
| E007 | P1 | 大群读扩散 ADR | E006 | ADR | 说明阈值和取舍 |
| E008 | P1 | fanout worker skeleton | D007 | worker | 可批量投递 |
| E009 | P2 | 热点群识别 | E006 | metrics/rules | 热点 conversation 标记 |
| E010 | P2 | 大群压测脚本 | E006 | k6 | 报告记录瓶颈 |

## Track F：Cache, MQ, Workers

| ID | 优先级 | 任务 | 输入依赖 | 产出 | 验收 |
|---|---|---|---|---|---|
| F001 | P0 | 接入 Redis client | A002 | redis package | ping 成功 |
| F002 | P0 | 实现 Redis rate limiter 基础接口 | F001 | rate limiter | 单测 |
| F003 | P0 | 实现 Redis presence TTL | F001,B004 | presence | TTL 过期 |
| F004 | P1 | conversation summary cache | C003,F001 | cache service | cache hit/miss 测试 |
| F005 | P1 | 接入 Redpanda/Kafka client | A002 | eventbus | publish/consume smoke |
| F006 | P1 | 定义事件 schema | F005 | event package | schema tests |
| F007 | P1 | worker process skeleton | F005 | `cmd/worker` | worker 可启动 |
| F008 | P1 | message.created consumer | D007,F007 | worker handler | 消费幂等 |
| F009 | P2 | MQ lag metrics | F007 | metrics | 可观测 |
| F010 | P2 | 缓存一致性 ADR | F004 | ADR | 说明 DB/Redis 边界 |

## Track G：Media, Search, Notification

| ID | 优先级 | 任务 | 输入依赖 | 产出 | 验收 |
|---|---|---|---|---|---|
| G001 | P1 | 接入 MinIO/S3 client | A002 | media package | bucket 可访问 |
| G002 | P1 | 实现预签名上传 API | G001 | endpoint | 返回 upload URL |
| G003 | P1 | 实现附件元数据表 | A005 | migration | 元数据入库 |
| G004 | P1 | 实现附件消息发送 | A018,G003 | message type | 权限校验 |
| G005 | P1 | 接入 OpenSearch skeleton | A002 | search package | ping/search smoke |
| G006 | P1 | message indexing worker | F008,G005 | worker | message.created 建索引 |
| G007 | P1 | search API | G005 | endpoint | 关键词搜索 |
| G008 | P2 | 搜索权限过滤 | G007 | service | 不能搜非成员会话 |
| G009 | P2 | 通知事件表 | F006 | notification | 事件可记录 |
| G010 | P2 | push/email notification mock | G009 | worker | 不阻塞主链路 |

## Track H：Security, Risk, Audit

| ID | 优先级 | 任务 | 输入依赖 | 产出 | 验收 |
|---|---|---|---|---|---|
| H001 | P1 | 登录失败限流 | F002,A009 | limiter | 暴力登录受限 |
| H002 | P1 | 发消息限流 | F002,A018 | limiter | 高频发送受限 |
| H003 | P1 | 会话维度限流 | F002,A018 | limiter | 群刷屏受限 |
| H004 | P1 | audit log 表 | A005 | migration | append-only |
| H005 | P1 | 登录审计 | H004,A009 | audit event | 登录有记录 |
| H006 | P1 | 消息撤回审计 | H004,C009 | audit event | 可追溯 |
| H007 | P2 | 基础 spam 规则 | H002 | risk service | 重复内容标记 |
| H008 | P2 | 举报消息 | H004 | APIs | report 可审计 |
| H009 | P2 | 用户拉黑 | A006 | APIs | 被拉黑不可发私信 |
| H010 | P2 | 安全 checklist | H001-H009 | docs | review 可执行 |

## Track I：Observability, Load Test, Chaos

| ID | 优先级 | 任务 | 输入依赖 | 产出 | 验收 |
|---|---|---|---|---|---|
| I001 | P1 | structured logging | A001 | logger | request_id |
| I002 | P1 | trace_id middleware | A001 | middleware | API 日志串联 |
| I003 | P1 | Prometheus metrics endpoint | A001 | `/metrics` | scrape 成功 |
| I004 | P1 | WS 连接数指标 | B003,I003 | metrics | connect/disconnect 变化 |
| I005 | P1 | 消息发送延迟指标 | A018,I003 | metrics | histogram |
| I006 | P1 | k6 API send message 压测 | A018 | loadtest | 报告 |
| I007 | P1 | k6 WS connect 压测 | B001 | loadtest | 报告 |
| I008 | P2 | Redis 故障演练 | F001 | chaos script | 降级记录 |
| I009 | P2 | MQ 故障演练 | F005 | chaos script | outbox 可补偿 |
| I010 | P2 | Grafana dashboard skeleton | I003-I005 | dashboard | 指标可视化 |

## Track J：Frontend Web

| ID | 优先级 | 任务 | 输入依赖 | 产出 | 验收 |
|---|---|---|---|---|---|
| J001 | P1 | 初始化 React/Next.js | A003 | frontend app | 可启动 |
| J002 | P1 | 登录页 | A009,J001 | UI | 登录成功 |
| J003 | P1 | 会话列表 | C003,J002 | UI | 展示未读 |
| J004 | P1 | 聊天窗口 | A019,J003 | UI | 历史分页 |
| J005 | P1 | 发送消息 | A018,J004 | UI | 发送成功 |
| J006 | P1 | WebSocket 实时接收 | B007,J004 | UI | 实时显示 |
| J007 | P2 | reconnect 状态 | B009,J006 | UI | 断线提示 |
| J008 | P2 | 附件上传 UI | G002,J004 | UI | 上传成功 |
| J009 | P2 | 搜索 UI | G007,J004 | UI | 搜索结果 |
| J010 | P2 | Playwright E2E | J002-J006 | tests | 核心流通过 |

## Track K：Mobile/Desktop Prototype

| ID | 优先级 | 任务 | 输入依赖 | 产出 | 验收 |
|---|---|---|---|---|---|
| K001 | P3 | 移动端原型技术选型 ADR | J001 | ADR | React Native/Expo 或 PWA 取舍 |
| K002 | P3 | PWA installable shell | J001 | manifest/service worker | 可安装 |
| K003 | P3 | 移动聊天布局 | J004 | responsive UI | mobile viewport 验证 |
| K004 | P3 | 桌面端原型 ADR | J001 | ADR | Tauri/Electron 取舍 |
| K005 | P3 | 桌面通知 mock | G009,J001 | prototype | 可演示 |

## Track L：Docs, ADR, Interview

| ID | 优先级 | 任务 | 输入依赖 | 产出 | 验收 |
|---|---|---|---|---|---|
| L001 | P0 | 每个 phase 更新 iteration report | 任意 phase | report | 包含测试和风险 |
| L002 | P0 | 维护 `docs/api.md` | API 变更 | docs | 与实现一致 |
| L003 | P0 | 维护 `docs/data-model.md` | migration 变更 | docs | 与 schema 一致 |
| L004 | P0 | 维护 `docs/websocket-protocol.md` | WS 变更 | docs | 协议完整 |
| L005 | P1 | 编写系统设计讲稿 | Phase 3+ | docs | 可回答面试 |
| L006 | P1 | 编写可靠性讲稿 | D tasks | docs | 覆盖 ACK/重试/去重 |
| L007 | P1 | 编写大群 fanout 讲稿 | E tasks | docs | 覆盖读写扩散 |
| L008 | P1 | 编写多端同步讲稿 | C/D tasks | docs | 覆盖设备 cursor |
| L009 | P2 | 每轮 review report | review 后 | reports | 有 findings |
| L010 | P2 | 最终作品集 README polish | 大部分完成后 | README | 不替代工程 |

## Track M：Review, Refactor, Quality

| ID | 优先级 | 任务 | 输入依赖 | 产出 | 验收 |
|---|---|---|---|---|---|
| M001 | P1 | API consistency review | A020+ | report/fixes | 错误格式一致 |
| M002 | P1 | DB schema review | migrations | report/fixes | 索引/约束合理 |
| M003 | P1 | 并发安全 review | C002,D001 | report/fixes | race 风险记录 |
| M004 | P1 | 可靠性 review | D tasks | report/fixes | 不丢/不重/可补偿 |
| M005 | P2 | 性能 review | I tasks | report/fixes | 热点路径有指标 |
| M006 | P2 | 安全 review | H tasks | report/fixes | 权限边界测试 |
| M007 | P2 | 测试覆盖 review | all | report/fixes | 核心路径有测试 |
| M008 | P2 | 文档一致性 review | docs | report/fixes | docs 不过期 |
| M009 | P2 | 重构 repository layer | A tasks | refactor | 测试通过 |
| M010 | P2 | 重构 event bus interface | F tasks | refactor | Kafka/内存可替换 |

## Future Extension Tracks

这些任务只有在主线和常规 backlog 进展顺利时执行，避免 10h 长跑提前耗尽工作。

| ID | 主题 | 任务 |
|---|---|---|
| X001 | 加密 | E2EE 需求分析和威胁模型 |
| X002 | 加密 | per-chat key management ADR |
| X003 | 加密 | at-rest encryption prototype |
| X004 | 微服务 | 服务拆分边界 ADR |
| X005 | 微服务 | API gateway + auth service split prototype |
| X006 | 微服务 | message service + realtime gateway split prototype |
| X007 | 广告 | sponsored channel message 数据模型 ADR |
| X008 | 广告 | 广告投放频控和审计设计 |
| X009 | 支付 | subscription/payment ledger ADR |
| X010 | 支付 | channel paid subscription prototype |
| X011 | 推荐 | channel recommendation research report |
| X012 | 推荐 | friend/contact recommendation prototype |

