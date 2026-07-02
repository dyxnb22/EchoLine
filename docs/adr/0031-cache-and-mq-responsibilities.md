# ADR 0031：Redis 与 MQ 的职责边界

> **Note:** ADR 0003 is reserved for [Large Group Fanout](./0003-large-group-fanout.md). This document captures the cache/MQ responsibility split originally drafted under a duplicate `0003` filename.

## 状态

Draft — direction accepted; details validated in Phase 6 implementation.

## 初始方向

- PostgreSQL 是 source of truth。
- Redis 用于 presence、限流、热点会话 summary。
- MQ 用于通知、审计、搜索索引、fanout 等异步任务。

## 实现对照

| 职责 | 模块 / 文件 |
|------|-------------|
| Presence TTL | `internal/presence`, Redis |
| Rate limiting | `internal/rate_limit`, Redis |
| Transactional outbox | `outbox_events`, `internal/outbox` |
| Event publish / consume | `internal/eventbus`, `cmd/worker` |
| Fanout / search / webhook | worker handlers |

## 待验证问题

- MQ 发布失败是否影响主链路 — **否**：outbox 异步 drain，主链路只写 DB + outbox。
- 消费失败如何重试或进入死信队列 — `dead_letter_events` + admin DLQ replay。
- Redis 缓存和 DB 如何处理一致性边界 — 见 [ADR 0005](./0005-cache-consistency.md)。
