# Next Actions

## Immediate P0

1. Integration smoke with docker compose when available (`RUN_API_SMOKE=1 RUN_WS_SMOKE=1 make smoke`).
2. Outbox integration test with Postgres (`FOR UPDATE SKIP LOCKED` hardening optional).

## Sequential

3. G005: presigned download URL for attachments.
4. I001: request_id in structured logs (partially done via middleware).
5. J007: optimistic send + error toast polish.

## Environment

```bash
make dev-up
export DATABASE_URL=postgres://echoline:echoline@localhost:5432/echoline?sslmode=disable
export JWT_SECRET=change-me
export REDIS_ADDR=localhost:6379
export KAFKA_BROKERS=localhost:9092
export S3_ENDPOINT=http://localhost:9000
export S3_ACCESS_KEY=minio
export S3_SECRET_KEY=minio123
export S3_BUCKET=echoline
make api-run
make worker-run
make seed
make frontend-dev
```
