# EchoLine API 设计

本文档记录 REST API。WebSocket 协议见 `docs/websocket-protocol.md`。

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

### POST `/api/auth/login`

返回：

```json
{
  "access_token": "...",
  "refresh_token": "..."
}
```

## Conversations

### POST `/api/conversations/direct`

创建或获取私聊会话。

### POST `/api/conversations/groups`

创建群聊。

### GET `/api/conversations`

返回会话列表、最近消息、未读数。

## Messages

### POST `/api/conversations/{conversation_id}/messages`

发送消息。

请求：

```json
{
  "client_msg_id": "uuid-from-client",
  "type": "text",
  "body": "hello"
}
```

### GET `/api/conversations/{conversation_id}/messages`

历史消息分页。

查询参数：

- `before_seq`
- `after_seq`
- `limit`

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

```json
{
  "error": {
    "code": "rate_limited",
    "message": "too many messages",
    "request_id": "trace-id"
  }
}
```

