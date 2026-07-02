# ADR 0002：使用 conversation 内递增 seq 保证消息顺序

## 状态

Accepted — implemented in `backend/internal/message/repository.go` (transactional seq allocation).

## 初始方向

每个 conversation 维护 `latest_seq`，消息写入时在事务中分配递增 `seq`。读侧按 `(conversation_id, seq)` 排序。

## 待验证问题

- 高并发写入同一个 conversation 时如何避免锁竞争。
- 大群热点下 seq 分配是否成为瓶颈。
- 是否需要按 conversation 分片或使用单独 sequence allocator。

