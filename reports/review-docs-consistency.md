# Code Review: Documentation Consistency (M008)

**Reviewer**: Automated review via agent  
**Date**: 2026-07-01 (initial), **2026-07-02 (passes 1–3)**  
**Scope**: All docs, ADRs, OpenAPI, manifests, agent prompts

---

## Summary

EchoLine documentation is aligned with the implemented codebase and T001–T440 closure state after three alignment passes. Automated guard: `make validate-docs` (also runs in `make verify`).

**Documentation consistency score**: **Complete** — `scripts/validate-docs.py` exits 0.

---

## Pass 3 (2026-07-02) — Completed

| Area | Fix |
|------|-----|
| BATCH_100_MANIFEST | All Key File paths corrected; statuses updated (pin/reactions/admin/export/etc.) |
| BATCH_120 / BATCH_NEXT_120 | Paths + closure statuses; T184–T228 marked done where implemented |
| Living docs | Global path fixes (`mq`→`eventbus`, `middleware`→actual packages, sync POST body) |
| Interview docs | Correct `POST /api/sync` JSON body; WS `message.created` / `message.ack` |
| Review reports M001–M007 | Body paths corrected (not just headers) |
| openapi.yaml | v0.3.0 — 61 paths, 27 schemas, requestBody on core endpoints |
| Automation | `scripts/validate-docs.py`, `make validate-docs`, wired into `make verify` |

---

## Pass 2 — Completed

See git history (`1e82ff6`). Highlights: architecture module table, openapi route mirror, extensions-roadmap closure.

---

## Pass 1 — Completed

See git history (`102c8ca`). Highlights: ADR index, websocket-protocol, state file closure banners.

---

## Validation

```bash
make validate-docs   # ghost paths, wrong API names, openapi route parity
make verify          # includes validate-docs + tests + build + playwright
```

---

## Remaining optional (non-blocking)

1. OpenAPI request/response schemas for all 61 endpoints (core auth/message/sync covered).
2. Local `make smoke-full` results in PROGRESS_LOG.
3. Fluentd/Loki shipping (BATCH_NEXT_120 T168) — explicitly future.
