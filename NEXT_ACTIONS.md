# Next Actions

## Immediate P0

1. E001: group member role checks in APIs.
2. E003-E005: channel model, subscribe, publish permissions.
3. F005: Redpanda/Kafka client wiring.
4. Integration smoke with `make dev-up` when Docker available.

## Sequential after P0

5. F008: message.created consumer in worker.
6. H001-H002: rate limit middleware on login/send.
7. B009: reconnect fallback doc.
8. J001-J006: frontend login/chat/WS.

## Parallelizable (Composer 2.5, Fast mode disabled)

- Docs Agent: L005-L006 reliability/interview notes after E/F tasks.
- Review Agent: M001 API consistency review.

## Environment

```bash
make dev-up
export DATABASE_URL=postgres://echoline:echoline@localhost:5432/echoline?sslmode=disable
export JWT_SECRET=change-me
export REDIS_ADDR=localhost:6379
make api-run
make seed
RUN_API_SMOKE=1 RUN_WS_SMOKE=1 make smoke
```
