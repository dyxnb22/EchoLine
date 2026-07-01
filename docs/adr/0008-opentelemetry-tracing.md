# ADR 0008: Distributed Tracing with OpenTelemetry

## Status

Accepted (current: X-Trace-ID correlation header; OTel SDK integration planned)

## Context

EchoLine currently implements basic observability via:
- **Structured JSON logs** with `trace_id` field (from `X-Trace-ID` request header or auto-generated UUID).
- **Prometheus metrics** at `/metrics` (request count, WS gauge, message send latency histogram).

This is sufficient for single-service debugging. However, as EchoLine evolves toward a multi-component architecture (API gateway → worker → fanout → notification), tracing a single message send across these components becomes difficult with log correlation alone.

Specific pain points observed during chaos testing:
1. A message send that times out — is the delay in the API handler, the outbox worker, the Kafka consumer, or the WS push?
2. A DLQ entry — which exact component first encountered the error, and what was the input state at that point?
3. P99 latency spikes on `/api/conversations/:id/messages` — are they from Postgres, Redis, or Kafka publish?

## Decision

Adopt **OpenTelemetry (OTel)** as the tracing standard with a **Jaeger-compatible OTLP exporter**.

### Implementation Plan

**Phase 1 (current):** Propagate `trace_id` via `X-Trace-ID` HTTP header through all log lines. Already implemented in `backend/internal/middleware/trace.go`.

**Phase 2 (next):** Instrument Go backend with OTel SDK:
- Add `go.opentelemetry.io/otel` and `go.opentelemetry.io/otel/sdk/trace`.
- Create a `tracer` provider in `backend/internal/telemetry/tracer.go` pointing to OTLP endpoint (`OTEL_EXPORTER_OTLP_ENDPOINT`).
- Instrument: HTTP handler middleware (creates root span), Postgres queries (via `otelpgx`), Redis commands (via `otelredis`), Kafka produce/consume (manual span creation).
- WS events: create child spans from the HTTP upgrade trace context.

**Phase 3 (stretch):** Add Grafana Tempo or Jaeger UI; configure Grafana trace-to-log correlation using `trace_id`.

### Trace Context Propagation

- HTTP: W3C `traceparent` header (OTel default).
- WS: trace context passed in the initial upgrade request; subsequent WS events reference the session root span.
- Kafka: inject trace context into message headers using OTel Kafka instrumentation.
- Inter-worker: pass `trace_id` in the outbox `metadata` JSONB column.

### Sampling Strategy

- Development: 100% sampling.
- Production: head-based probabilistic sampling at 1% for high-volume paths (message send), 100% for error paths.

## Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| OpenTelemetry (chosen) | Vendor-neutral, OTel is the standard | SDK overhead, OTLP endpoint required |
| Zipkin only | Simple | Narrower ecosystem, less Go library support |
| Datadog APM | Excellent UI | Vendor lock-in, cost |
| Log correlation only | Zero infra | Cannot visualize cross-service call graphs |
| No tracing | Zero cost | Blind during incidents |

## Implementation Files

- `backend/internal/middleware/trace.go` — X-Trace-ID middleware (Phase 1, implemented)
- `backend/internal/telemetry/tracer.go` _(planned)_ — OTel SDK provider init
- `backend/internal/db/db.go` — add `otelpgx` instrumentation
- `backend/internal/cache/redis.go` — add `otelredis` instrumentation
- `backend/internal/mq/producer.go` — inject trace headers into Kafka messages
- `backend/cmd/api/main.go` — initialize tracer on startup
- `docker-compose.yml` _(planned)_ — add Jaeger all-in-one container
- `grafana/echoline-dashboard.json` — trace exemplars linked to Prometheus metrics

## Consequences

**Positive:**
- End-to-end latency attribution for every message send.
- Error root cause visible in a single trace view.
- P99 latency spikes attributed to specific components.
- OTel is vendor-neutral; backend can be switched to any OTLP-compatible collector (Jaeger, Tempo, Honeycomb).

**Negative:**
- OTel SDK adds ~1–3% CPU overhead for high-cardinality spans.
- Requires OTLP collector deployment (Jaeger or OpenTelemetry Collector).
- Developer experience during local testing requires running Jaeger.

## Verification

- Unit test: `TestTraceMiddleware_PropagatesTraceID` — assert `X-Trace-ID` is set on response and logged.
- Integration test (planned): send a message; assert OTel exporter receives a span tree with spans for `HTTP /api/…/messages`, `db.query`, `redis.set`, `kafka.produce`.

## Interview Talking Points

- **Why OTel over proprietary APM?** "OTel is the CNCF standard. Instrumenting once with OTel gives us the freedom to ship traces to Jaeger in development, Grafana Tempo in staging, and Honeycomb or Datadog in production without re-instrumenting."
- **Trace context across Kafka**: "We inject `traceparent` into Kafka message headers. The consumer reads the header and creates a child span, giving us a complete trace graph across async boundaries."
- **Sampling at scale**: "At 100k messages/s, 100% sampling would generate ~100k spans/s. We use head-based 1% sampling for the happy path and tail-based 100% for error paths, so we never miss an error trace."
