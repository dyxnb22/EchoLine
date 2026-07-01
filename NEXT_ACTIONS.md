# Next Actions

## Immediate P0

1. Integration smoke: `make dev-up` + `RUN_API_SMOKE=1 RUN_WS_SMOKE=1 make smoke`
2. Extend `scripts/smoke-test.sh` with register/login/send/search flow

## Sequential

3. F009: MQ lag metrics from Kafka consumer
4. I006-I007: k6 API/WS load test scripts
5. B011: typing indicator WS event
6. G010: notification event table skeleton

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
curl localhost:8080/metrics
```
