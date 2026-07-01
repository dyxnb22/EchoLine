# EchoLine WebSocket 协议

WebSocket 用于在线消息推送、ACK、presence heartbeat 和轻量实时事件。

## 连接

```text
GET /ws?token=<access_token>&device_id=<device_id>
```

连接建立后服务端将用户、设备和连接绑定。

## 通用消息格式

```json
{
  "type": "message.created",
  "request_id": "optional-request-id",
  "payload": {}
}
```

## 客户端到服务端

### `ping`

```json
{
  "type": "ping",
  "payload": {
    "ts": 1710000000
  }
}
```

### `message.send`

```json
{
  "type": "message.send",
  "request_id": "req-1",
  "payload": {
    "conversation_id": "c1",
    "client_msg_id": "uuid-from-client",
    "type": "text",
    "body": "hello"
  }
}
```

### `message.ack`

```json
{
  "type": "message.ack",
  "payload": {
    "conversation_id": "c1",
    "message_id": "m1",
    "seq": 10,
    "status": "delivered"
  }
}
```

## 服务端到客户端

### `pong`

### `message.created`

### `message.updated`

### `message.recalled`

### `presence.updated`

### `error`

```json
{
  "type": "error",
  "request_id": "req-1",
  "payload": {
    "code": "invalid_payload",
    "message": "missing conversation_id"
  }
}
```

## 可靠性原则

- 服务端以 DB 持久化为准。
- WS 推送失败时，客户端通过 sync endpoint 补偿。
- 客户端重试必须携带同一个 `client_msg_id`。
- ACK 是 delivery/read state，不是消息是否存在的唯一依据。

