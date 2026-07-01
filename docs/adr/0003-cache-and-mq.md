# ADR 0003：Redis 与 MQ 的职责边界

## 状态

待 Phase 6 实现时补充。

## 初始方向

- PostgreSQL 是 source of truth。
- Redis 用于 presence、限流、热点会话 summary。
- MQ 用于通知、审计、搜索索引、fanout 等异步任务。

## 待验证问题

- MQ 发布失败是否影响主链路。
- 消费失败如何重试或进入死信队列。
- Redis 缓存和 DB 如何处理一致性边界。

