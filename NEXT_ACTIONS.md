# Next Actions

## Immediate P0

1. F008 + D007: outbox publisher worker (replace direct Kafka publish in hot path).
2. E006: fanout to online group members (already via hub; add tests).
3. Integration smoke with docker compose when available.

## Sequential

4. H003: conversation-scoped rate limit.
5. G001-G004: MinIO attachment skeleton.
6. J001-J003: frontend login + conversation list.

## Environment

```bash
make dev-up
export DATABASE_URL=postgres://echoline:echoline@localhost:5432/echoline?sslmode=disable
export JWT_SECRET=change-me
export REDIS_ADDR=localhost:6379
export KAFKA_BROKERS=localhost:9092
make api-run
make worker-run
```
