# Performance and Load Test Plan

## Goals

- Understand EchoLine's bottlenecks through measurement.
- Produce interview-ready performance reports.
- Avoid premature optimization before baseline measurement.

## Baseline Scenarios

1. Register/login throughput.
2. Send direct message throughput.
3. Send group message throughput.
4. WebSocket concurrent connections.
5. Offline sync latency.
6. History pagination latency.
7. Search query latency.
8. Attachment upload metadata latency.

## Required Reports

- `reports/load-test-01.md`: API baseline.
- `reports/load-test-02.md`: WebSocket baseline.
- `reports/load-test-03.md`: group fanout baseline.
- `reports/load-test-04.md`: sync/search baseline.

## Metrics

- p50/p95/p99 latency.
- error rate.
- throughput.
- DB CPU and slow queries.
- Redis latency.
- MQ lag.
- WebSocket active connections.
- message delivery success rate.

## Optimization Rules

1. Measure before optimizing.
2. Record bottleneck.
3. Make one change.
4. Re-run the same test.
5. Record before/after.

