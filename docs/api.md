# EchoLine API 设计

本文档记录 REST API 与 WebSocket 入口。WS 帧协议见 `docs/websocket-protocol.md`。机器可读契约见 `docs/openapi.yaml`（与 `backend/internal/server/server.go` 路由对齐；请求/响应细节以本文为准）。

## Health

### GET `/health`

返回服务与数据库连通状态。

响应示例：

```json
{
  "status": "ok",
  "env": "development"
}
```

## WebSocket

### GET `/ws`

WebSocket 实时网关。查询参数：`token`（access JWT）、`device_id`。

连接、心跳、消息类型与重连策略见 [`docs/websocket-protocol.md`](./websocket-protocol.md)。

## Auth

### POST `/api/auth/register`

请求：

```json
{
  "username": "alice",
  "password": "secret",
  "display_name": "Alice"
}
```

成功响应 `201`：

```json
{
  "id": "uuid",
  "username": "alice",
  "display_name": "Alice",
  "created_at": "2026-07-01T00:00:00Z"
}
```

错误码：

- `username_taken`：用户名已存在。

### POST `/api/auth/login`

请求：

```json
{
  "username": "alice",
  "password": "secret"
}
```

返回：

```json
{
  "access_token": "...",
  "refresh_token": "...",
  "token_type": "Bearer",
  "expires_in": 86400
}
```

### POST `/api/auth/refresh`

请求：

```json
{
  "refresh_token": "..."
}
```

返回新的 access/refresh token 对。

### GET `/api/me`

需要 `Authorization: Bearer <access_token>`。

## Conversations

### GET `/api/conversations`

返回当前用户的会话列表，按 `updated_at` 降序，包含 `unread` 字段。需要 Bearer token。

### POST `/api/conversations/direct`

创建或获取私聊会话。需要 Bearer token。

请求：

```json
{
  "user_id": "peer-user-uuid"
}
```

### POST `/api/conversations/groups`

创建群聊。需要 Bearer token。

请求：

```json
{
  "title": "Engineering",
  "member_ids": ["uuid-1", "uuid-2"]
}
```

### POST `/api/conversations/channels`

创建频道。创建者为 owner，可发布消息。

### POST `/api/conversations/{conversation_id}/subscribe`

订阅频道（subscriber 角色，只读）。

### DELETE `/api/conversations/{conversation_id}/subscribe`

退订频道。

### POST `/api/conversations/{conversation_id}/members`

邀请用户加入群聊（需要 owner/admin）。

### DELETE `/api/conversations/{conversation_id}/members/{user_id}`

踢出成员或自己退群（owner 不可被踢）。

## Messages

### POST `/api/conversations/{conversation_id}/messages`

发送消息。需要 Bearer token。

请求：

```json
{
  "client_msg_id": "uuid-from-client",
  "type": "text",
  "body": "hello"
}
```

### GET `/api/conversations/{conversation_id}/messages`

历史消息分页。需要 Bearer token。

查询参数：

- `before_seq`
- `limit`

响应包含 `messages` 和可选 `next_before` cursor。

### POST `/api/conversations/{conversation_id}/read`

更新 `last_read_seq`。需要 Bearer token。

请求：

```json
{
  "last_read_seq": 42
}
```

### POST `/api/messages/ack`

记录 delivered/read ACK。需要 Bearer token。

请求：

```json
{
  "message_id": "uuid",
  "conversation_id": "uuid",
  "seq": 42,
  "status": "read",
  "device_id": "device-1"
}
```

## Sync

### POST `/api/sync`

离线同步。

请求：

```json
{
  "device_id": "device-id",
  "cursors": [
    {
      "conversation_id": "conversation-id",
      "last_seq": 100
    }
  ]
}
```

## Media

### POST `/api/media/upload-url`

获取直传对象存储的预签名 PUT URL。需要 Bearer token，且服务端已配置 `S3_ENDPOINT`。

请求：

```json
{
  "mime_type": "image/png",
  "size_bytes": 1024,
  "checksum": "sha256:..."
}
```

响应：

```json
{
  "upload_url": "https://...",
  "object_key": "uploads/<user-id>/<uuid>",
  "bucket": "echoline",
  "expires_in": 900
}
```

上传完成后，发送消息时引用 `attachment.object_key`：

```json
{
  "client_msg_id": "uuid",
  "type": "image",
  "body": "optional caption",
  "attachment": {
    "object_key": "uploads/<user-id>/<uuid>"
  }
}
```

### POST `/api/media/download-url`

获取已上传附件的预签名 GET URL。需要 Bearer token，且用户拥有该 `object_key`。

请求：

```json
{
  "object_key": "uploads/<user-id>/<uuid>"
}
```

## Search

### GET `/api/search/messages`

关键词搜索当前用户有权限的会话消息。需要 Bearer token。

查询参数：

- `q`：搜索关键词（必填）
- `limit`：结果数量（默认 20）

## Message lifecycle

### PATCH `/api/conversations/{conversation_id}/messages/{message_id}`

编辑消息正文（仅发送者，且 `status=normal`）。

请求：

```json
{
  "body": "updated text"
}
```

### POST `/api/conversations/{conversation_id}/messages/{message_id}/recall`

撤回消息（发送者或群 admin/owner）。

## Social & Notifications

### POST/DELETE `/api/conversations/{id}/pins/{message_id}`

置顶/取消置顶消息。

### GET `/api/conversations/{id}/pins`

列出会话置顶消息。

### POST `/api/conversations/{id}/mute` / `/unmute`

静音/取消静音会话（`muted_until`）。

### POST/DELETE `/api/blocks/{user_id}`

拉黑/取消拉黑用户。

### GET `/api/blocks`

列出拉黑用户。

### POST `/api/conversations/{id}/messages/{message_id}/report`

举报消息。

### GET `/api/notifications`

通知列表。

### POST `/api/notifications/{id}/read` / `POST /api/notifications/read-all`

标记通知已读。

### PATCH `/api/me`

更新 `display_name`。

### GET `/api/devices`

当前用户设备列表。

### GET `/api/admin/health` / GET `/api/admin/dlq`

管理接口骨架（需 Bearer token）。

## Batch-120 APIs

### Reactions

- `POST /api/messages/{message_id}/reactions` body `{ "emoji": "👍" }`
- `DELETE /api/messages/{message_id}/reactions/{emoji}`
- `GET /api/messages/{message_id}/reactions`

### Threads

- `POST /api/conversations/{conv_id}/messages/{message_id}/replies`
- `GET /api/conversations/{conv_id}/messages/{message_id}/replies`

### Forward / Presence / Export / Archive

- `POST /api/messages/{message_id}/forward`
- `GET /api/presence/online?user_ids=...`
- `GET /api/conversations/{id}/export`
- `POST /api/conversations/{id}/archive` / `unarchive`
- `GET /api/conversations/archived`

### Push / Payments / Ads / Recommendations

- `POST/GET /api/push/tokens`
- `POST/GET /api/payments/ledger`
- `POST/GET /api/channels/{channel_id}/campaigns`
- `GET /api/recommendations/channels`

### Admin DLQ

- `POST /api/admin/dlq/{id}/replay`

## Batch Next-120 APIs

### Admin (requires `ADMIN_USER_IDS`)

- `GET /api/admin/users` — list users
- `GET /api/admin/reports` — list message reports
- `GET /api/admin/audit-logs` — list audit entries
- DLQ list/replay also require admin when `ADMIN_USER_IDS` is set

### GraphQL prototype

- `POST /graphql` body `{ "query": "{ conversations { id title } }" }`
- `GET /graphql` — GraphiQL HTML when `GRAPHIQL=true`

### Payments / Ads

- `POST /api/payments/ledger/settle` body `{ "reference": "..." }` — idempotent settle
- `POST /api/channels/{channel_id}/campaigns/{campaign_id}/impressions` — frequency-capped impression

### Encryption / Presence / Recommendations

- `POST/GET /api/encryption/keys` — device public key bundles
- `GET /api/presence/last-seen?user_ids=...` — last seen timestamps
- `POST /api/presence/last-seen` — touch current user last-seen
- `GET /api/recommendations/friends` — mutual-group friend suggestions

### GraphQL mutation

- `POST /graphql` mutation `sendMessage` with variables `conversationId`, `body`
- `POST /graphql` mutation `addReaction` with variables `messageId`, `emoji`

### Channel entitlements (paid channels)

- `POST /api/channels/{channel_id}/entitlements/grant` — **admin only**; body `{ "user_id": "uuid (optional)", "reference": "..." }`
- `POST /api/channels/{channel_id}/entitlements/require` — **channel owner only**; body `{ "required": true }`
- `POST /api/conversations/{id}/subscribe` returns `402 payment_required` when entitlement missing
- `POST /api/payments/ledger` + `POST /api/payments/ledger/settle` with `reference: "channel:{uuid}"` auto-grants subscriber

## Observability

### GET `/metrics`

Prometheus 指标端点（HTTP 请求计数、WS 连接数、消息发送延迟、outbox pending、MQ 消费计数）。

## 错误格式

REST 错误统一为：

```json
{
  "error": {
    "code": "rate_limited",
    "message": "too many messages",
    "request_id": "trace-id"
  }
}
```

WebSocket 错误见 `docs/websocket-protocol.md`。

