# EchoLine Batch-Next-120 Manifest (T121–T240)

> **Status (2026-07-02): closure snapshot updated.** Rows marked `planned` below were completed in Batch-Next-200 / final session unless noted.

> **Note (2026-07-02):** T222–T228（微服务 ADR、gateway、分片、冷热分层、病毒扫描 mock、OTel）已在 Batch-Next-200 与 final completion 中完成；表中若仍显示 `planned`/`partial` 视为历史快照。

Continuation after `BATCH_120_MANIFEST.md`. Focus: close batch-120 partials, ops maturity, tests, extensions.

---

## Track 1: Backend Completion (T121–T140)

| ID | Status | Description | Key Files |
|----|--------|-------------|-----------|
| T121 | done | Wire webhook on message send | `message/webhook.go`, `server/options.go` |
| T122 | done | Admin role middleware (`ADMIN_USER_IDS`) | `admin/middleware.go` |
| T123 | done | `GET /api/admin/users` | `admin/repository.go`, `admin/handler.go` |
| T124 | done | `GET /api/admin/reports` | `admin/handler.go` |
| T125 | done | `GET /api/admin/audit-logs` | `admin/handler.go` |
| T126 | done | Push worker mock provider | `push/worker.go`, `cmd/worker/main.go` |
| T127 | done | GraphQL prototype `POST /graphql` | `graph/handler.go` |
| T128 | done | Payment settle endpoint (idempotent) | `payment/handler.go` |
| T129 | done | Ads impression + frequency cap | `ads/handler.go`, `00014_admin_webhook_ads.sql` |
| T130 | done | Recommendation ranking by member count | `recommendation/handler.go` |
| T131 | done | OpenSearch fallback in search handler | `search/handler.go` |
| T132 | done | Migration 00014 admin/webhook/ads | `migrations/00014_admin_webhook_ads.sql` |
| T133 | done | Sentry stub init | `telemetry/sentry.go`, `cmd/api/main.go` |
| T134 | done | Admin-guarded DLQ routes | `server/server.go` |
| T135 | partial | Webhook delivery persistence | `00014` table; worker retry planned |
| T136 | partial | E2EE key bundle API | migration `00012`; REST planned |
| T137 | partial | Paid channel entitlement check | ADR 0019; planned |
| T138 | partial | GraphQL mutations/subscriptions | prototype query only |
| T139 | partial | Presence last-seen broadcast | Redis store exists |
| T140 | partial | Fanout worker production path | log-only skeleton |

## Track 2: Frontend (T141–T160)

| ID | Status | Description | Key Files |
|----|--------|-------------|-----------|
| T141 | done | Thread reply panel UI | `components/ThreadPanel.tsx` |
| T142 | done | Admin panel (users + reports) | `components/AdminPanel.tsx` |
| T143 | done | Reactions display + remove | `App.tsx`, `api.ts` |
| T144 | done | Multi-emoji reactions (👍❤️) | `App.tsx` |
| T145 | done | `typing.stop` WS handling | `App.tsx` |
| T146 | done | Thread/sendReply API helpers | `api.ts` |
| T147 | done | Admin API client helpers | `api.ts` |
| T148 | done | Dark mode thread/admin panels | `styles.css` |
| T149 | partial | Component split (Login/Chat pages) | `App.tsx` still primary |
| T150 | partial | Channel subscribe/browse dedicated page | filter tabs only |
| T151 | done | `npm run lint` (tsc --noEmit) | `package.json` |
| T152 | partial | Playwright CI job | scaffold exists |
| T153 | partial | Frontend routing (react-router) | planned |
| T154 | partial | Group settings UI | planned |
| T155 | partial | Pin message UI | API exists |
| T156 | partial | Archive conversation UI | API exists |
| T157 | partial | Export conversation UI | API exists |
| T158 | partial | Forward message UI | API exists |
| T159 | partial | Push token register UI | API exists |
| T160 | partial | Payment ledger UI | API exists |

## Track 3: Ops / CI / Deploy (T161–T180)

| ID | Status | Description | Key Files |
|----|--------|-------------|-----------|
| T161 | done | Multi-stage Dockerfile | `Dockerfile` |
| T162 | done | Helm chart skeleton | `deploy/helm/echoline/` |
| T163 | done | Prometheus + Grafana compose overlay | `docker-compose.monitoring.yml` |
| T164 | done | Deploy workflow (GHCR push) | `.github/workflows/deploy.yml` |
| T165 | done | golangci-lint config | `.golangci.yml` |
| T166 | done | `make loadtest-ws` target | `Makefile` |
| T167 | done | `make lint` runs go vet/golangci | `Makefile` |
| T168 | partial | Fluentd/Loki log shipping | not implemented (future) |
| T169 | partial | K8s secrets template | planned |
| T170 | partial | Docker compose app services | infra-only compose |
| T171 | done | Prometheus scrape config | `deploy/monitoring/prometheus.yml` |
| T172 | partial | Staging environment ADR | planned |
| T173 | partial | Blue/green deploy ADR | planned |
| T174 | partial | Secrets rotation runbook | planned |
| T175 | partial | Backup/restore script | planned |
| T176 | partial | DB migration CI gate (strict) | CI applies migrations |
| T177 | partial | Playwright in CI | planned |
| T178 | partial | Integration test job (strict) | `continue-on-error` today |
| T179 | partial | SBOM generation | planned |
| T180 | partial | Image signing | planned |

## Track 4: Docs / ADRs (T181–T200)

| ID | Status | Description | Key Files |
|----|--------|-------------|-----------|
| T181 | done | Update `docs/api.md` admin/graphql/payment/ads | `docs/api.md` |
| T182 | done | Update `docs/data-model.md` migration 00014 | `docs/data-model.md` |
| T183 | done | `reports/iteration-05.md` | `reports/iteration-05.md` |
| T184 | done | ADR 0023 admin RBAC | `docs/adr/0023-admin-rbac.md` |
| T185 | done | ADR 0024 webhook retry | `docs/adr/0024-webhook-retry.md` |
| T186 | done | ADR 0025 GraphQL facade scope | `docs/adr/0025-graphql-facade-scope.md` |
| T187 | done | Admin panel docs | `docs/admin-panel.md` |
| T188 | done | Payment prototype | `docs/payments-prototype.md` |
| T189 | partial | Research: K8s vs VM deploy | planned |
| T190 | done | Admin RBAC middleware | `backend/internal/admin/middleware.go` |
| T191–T200 | partial | Extension prototype doc updates | planned |

## Track 5: Tests / Quality / Extensions (T201–T240)

| ID | Status | Description | Key Files |
|----|--------|-------------|-----------|
| T201 | done | Reaction path parser tests | `reaction/handler_test.go` |
| T202 | done | Thread path parser tests | `thread/handler_test.go` |
| T203 | done | Webhook dispatcher tests | `webhook/dispatcher_test.go` |
| T204 | done | Push worker mock tests | `push/worker_test.go` |
| T205 | done | Admin middleware tests | `admin/middleware_test.go` |
| T206 | done | GraphQL handler tests | `graph/handler_test.go` |
| T207 | done | Integration test skeleton | `tests/integration_test.go` |
| T208 | done | Replay CLI flag tests | `cmd/replay/main_test.go` |
| T209 | done | `go test ./...` passes | CI |
| T210 | done | `npm run build` passes | CI |
| T211 | partial | DB-backed reaction integration test | needs Postgres |
| T212 | partial | DB-backed thread integration test | planned |
| T213 | partial | Payment settle integration test | planned |
| T214 | partial | Ads frequency cap integration test | planned |
| T215 | partial | Playwright login→send→verify | scaffold |
| T216 | partial | Property-based seq ordering test | planned |
| T217 | partial | Mutation testing | planned |
| T218 | partial | E2EE client prototype | `docs/encryption-prototype.md` |
| T219 | partial | Payment subscription flow | ADR 0019 |
| T220 | partial | Ads CPM bidding | ADR 0020 |
| T221 | partial | Friend recommendation API | ADR 0021 |
| T222 | partial | Microservice split ADR | `BACKLOG.md` |
| T223 | partial | API gateway prototype | planned |
| T224 | partial | Service mesh research | planned |
| T225 | partial | Sharding proposal doc | planned |
| T226 | partial | Cold/hot message tiering | planned |
| T227 | partial | Virus scan mock | planned |
| T228 | partial | OTel exporter wiring | trace_id exists |
| T229 | partial | SQL slow query report | planned |
| T230 | partial | Chaos test automation in CI | scripts exist |
| T231–T240 | planned | Future extensions per roadmap | `docs/extensions-roadmap.md` |

---

## Summary

| Track | Done | Partial | Planned | Total |
|-------|------|---------|---------|-------|
| Backend T121–T140 | 14 | 6 | 0 | 20 |
| Frontend T141–T160 | 8 | 12 | 0 | 20 |
| Ops T161–T180 | 7 | 13 | 0 | 20 |
| Docs T181–T200 | 3 | 17 | 0 | 20 |
| Tests/Ext T201–T240 | 10 | 20 | 10 | 40 |
| **Total** | **42** | **68** | **10** | **120** |

**Note:** Many T121–T240 items close batch-120 debt (marked done/partial). Full integration verification still requires Postgres (`BLOCKERS.md`).
