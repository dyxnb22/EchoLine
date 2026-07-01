# ADR 0029: OpenTelemetry Stub

## Status

Accepted

## Context

Distributed tracing is required for multi-worker and future microservice debugging.

## Decision

Add `telemetry.InitOTel` stub activated by `OTEL_EXPORTER_OTLP_ENDPOINT`. Logs initialization; no exporter dependency in MVP.

## Consequences

- Zero overhead when env unset.
- Production can swap stub for `otel` SDK without changing call sites.

## Files

- `backend/internal/telemetry/otel.go`
- `backend/internal/telemetry/sentry.go`

## Verification

- `OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 go run ./cmd/api` logs stub init.
