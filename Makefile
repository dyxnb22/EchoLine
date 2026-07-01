.PHONY: help test lint dev-up dev-down smoke api-run

help:
	@echo "EchoLine commands:"
	@echo "  make test      Run tests"
	@echo "  make lint      Run linters"
	@echo "  make dev-up    Start local dependencies"
	@echo "  make dev-down  Stop local dependencies"
	@echo "  make smoke     Run smoke checks"
	@echo "  make api-run   Run API server (requires DATABASE_URL and JWT_SECRET)"

test:
	cd backend && go test ./...

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
	@echo "No linters yet. Phase 1 will add backend linting."

dev-up:
	docker compose up -d

dev-down:
	docker compose down

smoke:
	./scripts/smoke-test.sh

