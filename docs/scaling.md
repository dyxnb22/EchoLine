# EchoLine 扩展性设计

## 缓存策略

Redis 适合：

- presence TTL。
- 会话 summary 热点缓存。
- 限流计数。
- WebSocket 网关连接路由。

PostgreSQL 仍然是 source of truth。

## MQ 策略

消息写入后发布事件：

- `message.created`
- `message.updated`
- `message.recalled`
- `user.presence.updated`
- `audit.event.created`

消费者：

- 通知 worker。
- 搜索索引 worker。
- 审计 worker。
- fanout worker。

## 群聊扩散

小群：

- 可以写扩散或直接在线推送。
- 未读数容易维护。

大群：

- 写扩散成本高。
- 更适合读扩散或混合扩散。
- 在线用户优先推送，离线用户重连后拉取。

## 分库分表讨论

消息表增长最快，候选分片键：

- `conversation_id`
- `conversation_id + time bucket`

优点：

- 同一会话历史查询局部性好。
- conversation 内顺序性容易维护。

风险：

- 超大热点 conversation 可能形成单分片热点。
- 跨会话全局搜索需要依赖搜索索引。

## WebSocket 网关扩展

多实例下需要：

- user/device 到 gateway 的路由表。
- gateway 间事件转发。
- 在线状态 TTL。
- 连接迁移后幂等 ACK。

