.PHONY: help test lint dev-up dev-down smoke api-run worker-run seed frontend-dev frontend-build smoke-full chaos-redis chaos-mq loadtest-api loadtest-ws verify

help:
	@echo "EchoLine commands:"
	@echo "  make verify         Run local CI-equivalent checks (test + build + e2e)"
	@echo "  make test           Run backend unit tests"
	@echo "  make smoke          Run unit + optional WS/API smoke"
	@echo "  make smoke-full     Full API smoke (needs running server)"
	@echo "  make dev-up         Start docker compose deps"
	@echo "  make api-run        Run API server"
	@echo "  make worker-run     Run background worker"
	@echo "  make frontend-build Build React frontend"
	@echo "  make chaos-redis    Redis failure drill"
	@echo "  make chaos-mq       Kafka failure drill"
	@echo "  make loadtest-api   k6 API send load test"
	@echo "  make loadtest-ws    k6 WS connect load test"

test:
	cd backend && go test ./...

verify:
	chmod +x scripts/verify-all.sh && ./scripts/verify-all.sh

api-run:
	cd backend && go run ./cmd/api

seed:
	cd backend && go run ./cmd/seed

worker-run:
	cd backend && go run ./cmd/worker

frontend-dev:
	cd frontend && npm install && npm run dev

frontend-build:
	cd frontend && npm install && npm run build

lint:
	cd backend && golangci-lint run ./... 2>/dev/null || go vet ./...

loadtest-ws:
	k6 run loadtests/k6-ws-connect.js

dev-up:
	docker compose up -d

dev-down:
	docker compose down

smoke:
	./scripts/smoke-test.sh

smoke-full:
	RUN_API_SMOKE=1 ./scripts/smoke-api-full.sh

chaos-redis:
	./scripts/chaos-redis-down.sh

chaos-mq:
	./scripts/chaos-mq-down.sh

loadtest-api:
	k6 run loadtests/k6-api-send.js

replay:
	cd backend && go run ./cmd/replay

backup-db:
	chmod +x scripts/backup-db.sh && ./scripts/backup-db.sh

dev-app:
	docker compose --profile app up -d

seed-extended:
	./scripts/seed-extended.sh

bootstrap-minio:
	./scripts/bootstrap-minio.sh

loadtest-mixed:
	k6 run loadtests/k6-mixed-workload.js
