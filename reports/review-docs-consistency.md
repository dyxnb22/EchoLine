# Code Review: Documentation Consistency (M008)

**Reviewer**: Automated review via agent  
**Date**: 2026-07-01 (initial), **2026-07-02 (resolution pass)**  
**Scope**: All docs in `docs/`, `reports/`, ADRs, OpenAPI spec, code comments

---

## Summary

EchoLine documentation covers architecture, API, data model, WebSocket protocol, ADRs, and interview materials. An initial pass (2026-07-01) identified gaps; engineering reviews #02/#03 and this alignment pass addressed most items.

**Documentation consistency score**: **9/10** (post-alignment). Remaining gap: OpenAPI error examples on all routes.

---

## Findings — Resolution Status

| # | Finding | Severity | Status | Resolution |
|---|---------|----------|--------|------------|
| 1 | `docs/api.md` missing endpoints | Medium | **Fixed** | Endpoints documented through Batch Next-120 + entitlements (review #02/#03) |
| 2 | `docs/websocket-protocol.md` missing events | Medium | **Fixed** | Added `message.edited`, `message.recalled`, `typing.stop`, `typing.indicator/stopped` (2026-07-02) |
| 3 | `docs/data-model.md` missing infra tables | Medium | **Fixed** | Added `outbox_events`; deliveries/sync/audit already present |
| 4 | ADR README incomplete | Low | **Fixed** | Full index 0001–0031 in `docs/adr/README.md` |
| 5 | Architecture missing worker | Low | **Fixed** | Mermaid + worker table in `architecture.md` (review #03) |
| 6 | OpenAPI missing error examples | Low | **Open** | `docs/openapi.yaml` — add 401/422/429 examples per route |
| 7 | Interview docs not linked |  README | Low | **Fixed** | Links in `README.md` and `docs/README.md` |
| 8 | `docs/reliability.md` language mix | Low | **Accepted** | Chinese body + English terms; consistent with project bilingual docs |

---

## ADR Hygiene (2026-07-02)

| Issue | Resolution |
|-------|------------|
| Duplicate ADR number `0003` (`cache-and-mq` vs `large-group-fanout`) | Renamed cache/MQ draft to [ADR 0031](../docs/adr/0031-cache-and-mq-responsibilities.md); **0003** = large group fanout |
| Duplicate payment ADRs 0013 / 0019 | 0013 marked **superseded by 0019** |

---

## State / Memory Doc Alignment (2026-07-02)

| File | Change |
|------|--------|
| `DONE.md` | Merged duplicate Phase 6; closure banner; post-closure optional |
| `BACKLOG.md` | Marked closed; items mapped to manifest / research docs |
| `ACCEPTANCE_MATRIX.md` | Phase 1–9 → done/partial reflecting T440 closure |
| `TASKS.md` | Closure banner; ADR 0031 reference |
| `CURRENT_STATE.md` / `NEXT_ACTIONS.md` | Optional work only; no stale phase labels |

---

## Remaining Work (optional)

1. Expand `docs/openapi.yaml` error response examples (Finding 6).
2. Run `make smoke-full` locally and note results in `PROGRESS_LOG.md`.
3. GraphQL schema codegen — tracked in engineering-review-03 gaps.

---

## Files Updated (this pass)

- `docs/adr/README.md`, `docs/adr/0031-cache-and-mq-responsibilities.md`, `docs/adr/0013-payments-ledger.md`
- `docs/websocket-protocol.md`, `docs/data-model.md`, `docs/README.md`
- `DONE.md`, `BACKLOG.md`, `ACCEPTANCE_MATRIX.md`, `TASKS.md`, `CURRENT_STATE.md`, `NEXT_ACTIONS.md`
- `README.md`, `DECISIONS.md`, `EXECUTION_RULES.md`
