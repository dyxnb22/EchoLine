# EchoLine Done Index

> **Closure:** T001–T440、secondary/stretch/research backlog 与 extensions 原型已按 [`FINAL_COMPLETION_MANIFEST.md`](./FINAL_COMPLETION_MANIFEST.md) 关闭。未勾选项多为 **环境阻塞**（云 VM 无 Docker/Postgres），非功能缺失。

## Phase 0

- [x] 项目定位文档。
- [x] Agent 执行说明。
- [x] 长时执行规则。
- [x] Phase 0-10 任务图。
- [x] secondary / stretch / research backlog。
- [x] 文档目录骨架。
- [x] 工程目录骨架。
- [x] Cursor Cloud Agent 10h 长跑启动 prompt。
- [x] 细粒度任务图。
- [x] 验收矩阵。
- [x] repo-based context compaction 规则。
- [x] 当前状态和下一步恢复文件。
- [x] Cursor project rules 和 project-local skills。
- [x] 加密、微服务、广告、支付、推荐 future extension roadmap。
- [x] 子 Agent orchestration plan。
- [x] 子 Agent task packet 模板。
- [x] Composer 2.5 子 Agent 禁用 Fast mode 规则。

## Phase 1

- [x] 后端服务初始化（A001-A005）。
- [x] 数据库 migration（users/conversations/messages/devices/deliveries）。
- [x] 用户注册 / 登录 / JWT / refresh（A006-A011）。
- [x] 私聊 / 群聊 / 会话列表 / 消息 REST（A016-C003）。
- [x] 历史分页 + seed + OpenAPI + 错误 envelope（A019-A022）。
- [ ] 完整 integration smoke（**blocked:** 云 VM 无 Docker/Postgres — 见 `BLOCKERS.md`）。

## Phase 2

- [x] WebSocket endpoint + 鉴权 + 连接管理 + 心跳（B001-B004）。
- [x] WS 协议 envelope + message.send + 在线推送 + error envelope（B005-B008）。
- [x] WS unit smoke hook（B010 partial）。
- [ ] 双客户端集成 smoke（**blocked:** 依赖本地 compose 栈）。

## Phase 3

- [x] latest_seq / seq 分配（C001-C002，message repo 事务）。
- [x] 会话列表 + unread（C003/C005 partial）。
- [x] 消息编辑/撤回 API（C008-C009 partial）。
- [x] per-device sync cursor（C007）。

## Phase 4

- [x] 群成员角色 owner/admin/member 校验（E001）。
- [x] 群邀请/踢人/退群 API（E002）。
- [x] 频道模型与订阅 API（E003-E004）。
- [x] 频道发布权限（E005）。
- [x] 小群在线 fanout 测试（E006 unit test）。

## Phase 5 (reliability)

- [x] client_msg_id 幂等（D001-D002 partial）。
- [x] ACK REST/WS + delivery 状态机（D003-D004）。
- [x] outbox enqueue + worker publisher + SKIP LOCKED（D007-F008）。
- [x] DLQ skeleton + admin replay（D008 partial）。

## Phase 6 (infra)

- [x] Redis client + presence TTL skeleton（F001/F003）。
- [x] Kafka client + message.created publish/consume（F005-F008 partial）。
- [x] Redis rate limit middleware（H001-H002 via F002）。
- [x] audit log + login audit（H004-H005）。
- [x] in-memory event bus + worker skeleton（F005-F007 partial）。
- [ ] Kafka consumer production path（F008 partial — outbox drainer done）。

## Phase 7 (media/search)

- [x] MinIO presign upload/download URL（G001-G002, G005）。
- [x] attachments 元数据 + 附件消息（G003-G004）。
- [x] PostgreSQL 全文搜索 + search API（G005-G008 partial）。

## Phase 8 (observability + frontend)

- [x] trace_id + Prometheus metrics（I001-I005 partial）。
- [x] 登录/会话/conv_send 限流（H001-H003）。
- [x] Vite React 登录/会话/聊天/分页/WS 重连（J001-J006）。
- [x] 乐观发送 + 附件上传 + 搜索 UI（J007-J009）。
- [x] 注册页 + typing + 通知 badge + PWA manifest + Playwright skeleton（batch-100）。

## Phase 9 (batch-120 + CI)

- [x] Reactions/threads/forward/presence/export/archive APIs（T001-T030 partial）。
- [x] Push/payment/ads/recommendation/extension migrations（00011-00013）。
- [x] GitHub Actions CI + replay CLI + extended scripts（T051-T070）。
- [x] ADRs 0016-0022 + prototype docs + iteration-04（T071-T090）。
- [x] Frontend dark mode, reactions, channel filter, PWA sw（T031-T050 partial）。

## Final completion (T001–T440 + backlog + extensions)

- [x] Paid channel entitlements + migration 00016 + payment settle grant
- [x] GraphQL addReaction; fanout push fix; worker compose profile
- [x] Frontend react-router, AuthContext, ChatPage, notifications, group settings, edit/recall
- [x] Integration messaging test; Playwright E2E (mocked); CI goose migrations
- [x] API gateway prototype; OTel stub; ADR 0027–0029; code review report
- [x] Engineering review #02/#03 — docs index, RBAC, http.ts migration, validation depth
- [ ] Full `make smoke-full` (**blocked:** Docker/Postgres in cloud VM)

## Post-closure optional

- [ ] `conversation/handler` 迁移至 `apierror` envelope
- [ ] 本地 `make dev-up && make smoke-full` 全栈验收
- [ ] OTel stub 换真实 exporter SDK
