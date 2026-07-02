# Code Review: Concurrency (M003)

**Reviewer**: Automated review via agent
**Date**: 2026-07-01
**Scope**: Goroutine management, shared state, WebSocket hub, worker concurrency

---

## Summary

EchoLine's concurrency model is fundamentally sound: the WebSocket hub uses a single goroutine with a channel-based message loop; the outbox worker uses SKIP LOCKED for distributed locking. Several areas require attention for production robustness.

---

## Finding 1: Hub Broadcast Loop — Slow Consumer Can Block Fast Consumers

**Severity**: High
**Files**: `backend/internal/realtime/server.go`

**Observation**: The hub's broadcast loop iterates over all connected clients and writes to each WS connection:
```go
for conn := range hub.connections {
    conn.WriteJSON(msg)  // synchronous WS write
}
```
If one client has a slow or unresponsive connection, `WriteJSON` can block the entire hub goroutine, preventing all other clients from receiving messages.

**Recommendation**: Each WS write should be non-blocking:
1. Each connection has a buffered send channel (e.g., `chan []byte` with buffer size 256).
2. The hub goroutine sends to the channel without blocking (use `select { case ch <- msg: default: close(ch) }`).
3. A per-connection goroutine reads from the channel and writes to the WS socket.
4. If the channel is full (slow consumer), drop the message and close the connection.

This is the standard Go WS server pattern recommended in gorilla/websocket examples.

**Impact**: Without this fix, a single slow client can cause all online members of a conversation to experience delivery delays.

---

## Finding 2: Connection Registry Race Condition

**Severity**: Medium
**Files**: `backend/internal/realtime/server.go`

**Observation**: If the hub registers connections using a `map[string]*Conn`, concurrent register/unregister operations from different goroutines require mutex protection.

**Recommendation**: Verify the hub uses either:
- A `sync.RWMutex` protecting the map.
- A single goroutine (the hub's main loop) that handles all register/unregister/broadcast through channels.

The channel-based approach (hub main loop) is preferred in Go as it avoids mutex complexity. Verify this pattern is used consistently.

---

## Finding 3: Outbox Worker Goroutine Leak on Shutdown

**Severity**: Medium
**Files**: `backend/internal/worker/outbox.go`, `backend/cmd/worker/main.go`

**Observation**: The outbox worker polls the DB in a loop. If the worker process receives SIGTERM, the polling goroutine must complete its current batch and commit (or rollback) before exiting. Without graceful shutdown, a partial batch can leave outbox rows locked (lock is released on connection close, but there's a window where rows are neither processed nor committed).

**Recommendation**: Use `context.Context` with cancellation:
```go
ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
defer cancel()

for {
    select {
    case <-ctx.Done():
        return  // graceful shutdown
    case <-ticker.C:
        drainOutbox(ctx)
    }
}
```
The `drainOutbox` function should respect `ctx.Done()` and complete its current batch before returning.

---

## Finding 4: Kafka Consumer Goroutine — No Panic Recovery

**Severity**: Medium
**Files**: `backend/internal/worker/handlers.go`, `backend/cmd/worker/main.go`

**Observation**: If a Kafka consumer handler panics (e.g., nil pointer dereference on unexpected message payload), the worker goroutine crashes and the worker stops consuming.

**Recommendation**: Wrap handler invocations in a panic recovery:
```go
defer func() {
    if r := recover(); r != nil {
        log.Error("kafka handler panic", "error", r, "stack", debug.Stack())
        metrics.HandlerPanicTotal.Inc()
    }
}()
```
This allows the worker to continue processing the next message after a panic on one message. The problematic message is committed (to avoid infinite retry) and logged for investigation.

---

## Finding 5: Redis Rate Limiter — Race Between Check and Increment

**Severity**: Low
**Files**: `backend/internal/rate_limit/middleware.go`

**Observation**: If the rate limiter uses `GET` followed by `INCR` as two separate Redis commands, there is a TOCTOU race: two concurrent requests could both read a count of 59 (under the limit), both increment to 60, and both proceed.

**Recommendation**: Use Redis atomic operations:
- `INCR` + `EXPIRE` in a Lua script or pipeline.
- Or use `SET NX EX` pattern (token bucket).

The correct implementation is a single-command approach:
```lua
local count = redis.call('INCR', KEYS[1])
if count == 1 then redis.call('EXPIRE', KEYS[1], ARGV[1]) end
return count
```
Verify this is what the current implementation uses.

---

## Finding 6: WebSocket Ping Timer — Resource Leak on Disconnect

**Severity**: Low
**Files**: `backend/internal/realtime/server.go`

**Observation**: If a ping goroutine is started per connection and uses `time.NewTicker`, the ticker must be stopped when the connection closes. Otherwise, the goroutine blocks on the ticker channel indefinitely after the connection is gone.

**Recommendation**: Pass a `context.Context` to the ping goroutine and cancel it when the connection closes:
```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
go pingLoop(ctx, conn)
```

---

## Finding 7: No Worker Recovery on DB Connection Loss

**Severity**: Medium
**Files**: `backend/internal/worker/outbox.go`

**Observation**: If the Postgres connection pool experiences a transient error (network hiccup, DB restart), the outbox worker may return errors from the SKIP LOCKED query. Without retry logic, the worker stops until the next poll interval.

**Recommendation**: Add exponential backoff in the worker loop:
```go
backoff := time.Second
for {
    err := drainOutbox(ctx)
    if err != nil {
        log.Warn("outbox drain error", "err", err)
        time.Sleep(backoff)
        backoff = min(backoff*2, 30*time.Second)
    } else {
        backoff = time.Second
        time.Sleep(100 * time.Millisecond)
    }
}
```

---

## Overall Assessment

**Concurrency score**: 6/10. Finding 1 (slow consumer blocking hub) is a production-critical bug that must be fixed before deploying under real load. Findings 2–7 are hardening improvements. The core architecture (hub goroutine, outbox SKIP LOCKED, per-connection goroutines) is sound.

## Priority Fixes

1. **Finding 1** (CRITICAL): Non-blocking per-connection send channel.
2. **Finding 3** (HIGH): Graceful shutdown with context cancellation.
3. **Finding 4** (HIGH): Panic recovery in Kafka consumer.
4. **Finding 2** (MEDIUM): Verify hub map mutex protection.
5. **Finding 5** (MEDIUM): Atomic Redis rate limiter.

## Files to Update

- `backend/internal/realtime/server.go` — per-connection send channel (Finding 1), context-based ping (Finding 6)
- `backend/internal/worker/outbox.go` — context shutdown (Finding 3), exponential backoff (Finding 7)
- `backend/cmd/worker/main.go` — signal handling, context propagation
- `backend/internal/worker/handlers.go` — panic recovery (Finding 4)
- `backend/internal/rate_limit/middleware.go` — atomic Redis operations (Finding 5)
