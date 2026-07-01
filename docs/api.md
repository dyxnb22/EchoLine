# EchoLine API 设计

本文档记录 REST API。WebSocket 协议见 `docs/websocket-protocol.md`。

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

