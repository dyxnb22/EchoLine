# EchoLine 可靠性设计

## 核心目标

- 消息不因推送失败而丢失。
- 客户端重试不会产生重复消息。
- 同一 conversation 内消息可按稳定顺序读取。
- 离线设备可以补拉缺失消息。

## 持久化优先

发送消息时先写 PostgreSQL，再进行在线推送和异步事件发布。WebSocket 不是 source of truth。

## 幂等

客户端发送消息必须提供 `client_msg_id`。服务端对 `(sender_id, client_msg_id)` 建唯一约束。

重复请求时：

- 如果原消息已创建，返回原消息。
- 如果原消息正在处理中，返回可重试错误或等待结果。

## 顺序性

每个 conversation 维护递增 `latest_seq`。消息写入时在事务中分配 `seq`。

读消息时按 `(conversation_id, seq)` 排序。

## ACK

ACK 记录消息在用户 / 设备维度的 delivery state：

- `sent`
- `delivered`
- `read`

ACK 丢失不代表消息丢失，客户端可在下一次 sync 时补偿。

## 离线补偿

客户端保存每个 conversation 的最新已读或已同步 seq。重连后调用 sync API 拉取 `seq > last_seq` 的消息。

## 失败场景

| 场景 | 策略 |
|---|---|
| DB 写入失败 | 不推送，返回失败 |
| DB 写入成功，WS 推送失败 | 用户重连后 sync 补偿 |
| 客户端重试 | `client_msg_id` 去重 |
| ACK 丢失 | 后续 sync / read receipt 修正 |
| MQ 失败 | 主链路不回滚，记录重试或 dead letter |

