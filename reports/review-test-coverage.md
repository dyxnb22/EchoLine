# Code Review: Test Coverage (M007)

**Reviewer**: Automated review via agent
**Date**: 2026-07-01
**Scope**: Unit tests, integration tests, smoke tests, load tests

---

## Summary

EchoLine's test suite covers the core happy paths and critical idempotency behavior. Integration coverage is limited by the Docker/Postgres unavailability constraint in the cloud VM. The following findings identify gaps for production readiness.

---

## Coverage Assessment by Component

| Component | Unit Tests | Integration | Smoke | Notes |
|-----------|-----------|-------------|-------|-------|
| Auth (register/login/JWT) | ✓ | Blocked (no DB) | Smoke script | bcrypt, JWT marshal/unmarshal |
| Message send (REST) | ✓ | Blocked | Yes | idempotency, seq allocation |
| WebSocket hub | ✓ | Blocked | Partial | Unit: mock conn; Integration: requires live WS |
| Outbox worker | Partial | Blocked | — | Unit: mock Kafka; Integration: blocked |
| Delivery state machine | ✓ | Blocked | — | State transition tests |
| Rate limiter | ✓ | — | — | Mock Redis; sliding window |
| Full-text search | Partial | Blocked | — | tsvector parsing; DB required for GIN |
| Group/channel auth | ✓ | Blocked | — | Role check unit tests |
| Presence | Partial | — | — | TTL logic; Redis mock |
| Conversation list cache | Partial | — | — | Cache invalidation unit tests |

---

## Finding 1: No Concurrent Seq Allocation Test

**Severity**: High
**Files**: `backend/internal/message/repo.go`

**Observation**: The seq allocation (`UPDATE conversations SET latest_seq = latest_seq + 1`) is critical for message ordering. There is no test that verifies correctness under concurrent sends.

**Recommendation**: Add a test that launches 100 goroutines simultaneously, each sending one message to the same conversation, and asserts:
1. All assigned seqs are unique.
2. Seqs form a contiguous range `[1, 100]`.
3. No duplicates.

```go
func TestConcurrentSeqAllocation(t *testing.T) {
    // requires live DB; skip if DATABASE_URL unset
    if os.Getenv("DATABASE_URL") == "" {
        t.Skip("requires DATABASE_URL")
    }
    // ... goroutine pool, send 100 messages, assert uniqueness
}
```

---

## Finding 2: No Chaos/Fault Injection Tests

**Severity**: Medium
**Files**: `backend/internal/reliability/`

**Observation**: The chaos playbook (`docs/chaos-playbook.md`) defines fault scenarios but no automated tests simulate these faults. Manual chaos tests are not repeatable.

**Recommendation**: Add fault injection tests using interfaces and mock implementations:
- `TestOutboxWorker_KafkaDown`: Mock Kafka producer to return error; assert outbox rows remain pending; assert attempt count increments.
- `TestMessageSend_DBWriteFailure`: Mock DB to fail INSERT; assert HTTP 500 returned; assert no outbox row.
- `TestSyncCompensation`: Send 3 messages while client "offline"; reconnect; call sync; assert all 3 received.

---

## Finding 3: WS Integration Smoke Not Automated

**Severity**: Medium
**Files**: `scripts/smoke-test.sh`

**Observation**: The WS smoke test is a shell script that requires manual invocation and visual inspection of output. It is not run in CI.

**Recommendation**: Convert the WS smoke test to a Go test:
```go
func TestWSSmoke(t *testing.T) {
    if os.Getenv("RUN_WS_SMOKE") == "" {
        t.Skip()
    }
    // Connect WS, send a message via REST, assert WS receives message.received event
}
```
Run as `RUN_WS_SMOKE=1 go test ./integration/...` in CI after `make dev-up`.

---

## Finding 4: No Load Test Baseline in CI

**Severity**: Low
**Files**: `loadtests/k6-api-send.js`

**Observation**: Load tests exist but are not run in CI. Performance regressions are not detected automatically.

**Recommendation**: Run a lightweight k6 smoke load test in CI (10 VUs × 30 seconds) to detect catastrophic performance regressions:
```yaml
# .github/workflows/perf.yml
- name: k6 smoke load test
  run: k6 run --vus 10 --duration 30s loadtests/k6-api-send.js
  env:
    API_BASE_URL: http://localhost:8080
```

Set a threshold: `http_req_duration: p(95) < 500`. Fail CI if exceeded.

---

## Finding 5: Test Data Cleanup Between Tests

**Severity**: Medium
**Files**: All integration tests

**Observation**: Integration tests that share a database (when available) may leave test data that affects subsequent tests. Test isolation requires either per-test DB transactions (rolled back after) or per-test DB schemas.

**Recommendation**: Use `pgx` transaction rollback for unit-level DB tests:
```go
tx, _ := db.Begin(ctx)
defer tx.Rollback(ctx)
// run test using tx
```
For integration tests that must commit: use a unique random conversation per test, and clean up in `t.Cleanup`.

---

## Finding 6: No Snapshot Test for API Response Shape

**Severity**: Low
**Files**: `backend/internal/api/`

**Observation**: API handlers return JSON responses without asserting the exact shape in tests. A handler refactor could inadvertently change the field names without breaking any test.

**Recommendation**: Add golden file tests for key API responses:
```go
// golden/message_send_response.json
{
  "id": "<uuid>",
  "seq": 1,
  "body": "hello",
  "sender_id": "<uuid>",
  "created_at": "<timestamp>"
}
```
Use `testify/assert.JSONEq` with placeholder-replaced golden files. This catches unintended API shape changes.

---

## Coverage Metrics Target

| Layer | Current Estimate | Target |
|-------|-----------------|--------|
| Unit test coverage (`go test -cover`) | ~40% | 80% |
| Integration test scenarios | ~10 | 50 |
| API endpoints with smoke test | ~8/20 | 20/20 |
| Chaos scenarios automated | 0 | 3 |

---

## Overall Assessment

**Test quality score**: 5/10. Unit tests cover the critical idempotency and state machine logic. The main gaps are: concurrent seq test (critical), automated WS smoke, and fault injection tests. The Docker/Postgres constraint explains the integration coverage gap.

## Priority Additions

1. **Finding 1** (HIGH): Concurrent seq allocation test.
2. **Finding 2** (MEDIUM): Fault injection tests (at minimum, Kafka down scenario).
3. **Finding 3** (MEDIUM): Automated WS smoke in Go test.
4. **Finding 5** (MEDIUM): Test isolation with transaction rollback.
5. **Finding 4** (LOW): k6 smoke in CI.

## Files to Add/Update

- `backend/internal/message/repo_test.go` — concurrent seq test
- `backend/internal/reliability/chaos_test.go` — fault injection tests
- `backend/integration/ws_smoke_test.go` — WS smoke as Go test
