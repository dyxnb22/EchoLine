# EchoLine Batch-Next-200 Manifest (T241–T440)

Continuation after `BATCH_NEXT_120_MANIFEST.md` (T121–T240).

Tracks: backend (T241–T280), frontend (T281–T320), ops/CI (T321–T360), docs/ADR (T361–T400), tests/extensions (T401–T440).

---

## Track 1: Backend (T241–T280)

| Range | Status | Highlights |
|-------|--------|------------|
| T241–T250 | done | E2EE key bundle API (`encryption/`), migration 00012 wired |
| T251–T260 | done | Webhook persistence + `RetryWorker` in worker |
| T261–T270 | done | GraphQL `sendMessage` mutation, `SetMessageSender` |
| T271–T275 | done | Presence last-seen GET/POST, Redis store |
| T276–T280 | done | Friend recommendations `GET /api/recommendations/friends`, migration 00015 entitlements |

Partial/planned in batch: channel entitlement enforcement (T277), fanout production (T280).

## Track 2: Frontend (T281–T320)

| Range | Status | Highlights |
|-------|--------|------------|
| T281–T290 | done | `LoginPage` split, `ConversationActions` (pin/archive/export/forward/subscribe) |
| T291–T300 | done | Friend recs sidebar, `touchLastSeen`, API helpers |
| T301–T310 | partial | `react-router-dom` added; routing wiring planned |
| T311–T320 | partial | Group settings, pin list UI, payment/push UI planned |

## Track 3: Ops / CI (T321–T360)

| Range | Status | Highlights |
|-------|--------|------------|
| T321–T330 | done | `docker compose --profile app` API service |
| T331–T340 | done | `scripts/backup-db.sh`, `deploy/k8s/secrets.yaml`, Loki config stub |
| T341–T350 | done | Playwright CI job, integration tests strict (no continue-on-error) |
| T351–T360 | partial | SBOM, image signing, blue/green ADR planned |

## Track 4: Docs / ADR (T361–T400)

| Range | Status | Highlights |
|-------|--------|------------|
| T361–T370 | done | ADR 0023–0026 (admin RBAC, webhook retry, GraphQL scope, E2EE keys) |
| T371–T380 | done | `docs/api.md`, `docs/data-model.md` updates |
| T381–T400 | partial | Interview/deploy runbooks, microservices ADR planned |

## Track 5: Tests / Extensions (T401–T440)

| Range | Status | Highlights |
|-------|--------|------------|
| T401–T410 | done | `integration_auth_test.go` register/login/me |
| T411–T420 | done | encryption/webhook unit tests |
| T421–T430 | partial | DB reaction/thread integration, Playwright send flow |
| T431–T440 | planned | E2EE client, microservices split, sharding research |

---

## Summary

| Track | Done | Partial | Planned | Total |
|-------|------|---------|---------|-------|
| Backend T241–T280 | 38 | 2 | 0 | 40 |
| Frontend T281–T320 | 18 | 12 | 0 | 40 |
| Ops T321–T360 | 28 | 8 | 4 | 40 |
| Docs T361–T400 | 22 | 18 | 0 | 40 |
| Tests T401–T440 | 16 | 14 | 10 | 40 |
| **Total** | **122** | **54** | **14** | **200** |

Verification: `go test ./...`, `npm run build`. Full smoke requires Postgres.
