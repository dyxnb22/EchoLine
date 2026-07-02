# Code Review: Documentation Consistency (M008)

**Reviewer**: Automated review via agent  
**Date**: 2026-07-01 (initial), **2026-07-02 (pass 1 + pass 2)**  
**Scope**: All docs in `docs/`, `reports/`, ADRs, OpenAPI spec, manifests, agent prompts

---

## Summary

EchoLine documentation is aligned with the implemented codebase and T001–T440 closure state after two alignment passes.

**Documentation consistency score**: **10/10** for living docs (ADRs, api, data-model, websocket, architecture, state files, openapi route coverage). Historical manifests/review reports retain path snapshots with header disclaimers.

---

## Pass 1 (2026-07-02) — Completed

| Area | Fix |
|------|-----|
| ADR index | Full 0001–0031; duplicate 0003 → 0031; 0013 superseded by 0019 |
| websocket-protocol | `message.edited`, typing events |
| data-model | `outbox_events` |
| State files | DONE, BACKLOG, ACCEPTANCE_MATRIX, TASKS closure banners |
| Navigation | docs/README, README interview links |

---

## Pass 2 (2026-07-02) — Completed

| Area | Fix |
|------|-----|
| ADR implementation paths | 0002 status; 0005/0006/0010–0014/0022 — remove ghost `backend/internal/api/` |
| Living technical docs | security-checklist, research-presence, reliability-adr-suite, interview-* paths |
| architecture.md | Expanded module table (30+ packages); removed phantom `channel` module |
| data-model.md | `parent_message_id`, `archived_at`, extension table columns |
| api.md | `GET /ws` entry; openapi now full route mirror |
| openapi.yaml | **61 paths**, Error schema, 401/422/429 on protected routes |
| extensions-roadmap.md | Prototype vs future per section |
| RESEARCH_PLAN.md | Actual output paths (no `docs/research/` ghost dir) |
| CLOUD_AGENT_PROMPT.md | Closure notice |
| load-test-01.md | k6 scripts done |
| scaling.md | `message.edited` event name |
| interview-multi-device-sync | `conversation.read` marked proposed (not in code) |
| BATCH_* manifests | Historical path disclaimers |
| Review reports M001–M007 | Historical `internal/api/` disclaimers |

---

## Verification Checklist

| Check | Status |
|-------|--------|
| All ADR files indexed in `docs/adr/README.md` | ✅ |
| No `backend/internal/api/` in `docs/` (living) | ✅ |
| `docs/openapi.yaml` paths match `server.go` | ✅ |
| WS event names match `realtime/protocol.go` | ✅ |
| Closure consistent across CURRENT_STATE, TASKS, BACKLOG, CLOUD_AGENT_PROMPT | ✅ |
| Broken markdown links to deleted files | ✅ none found |

---

## Remaining Optional (non-blocking)

1. Line-by-line correction of `BATCH_100_MANIFEST.md` Key File column (100+ rows; header disclaimer sufficient).
2. Local `make smoke-full` results recorded in PROGRESS_LOG.
3. OpenAPI request/response body schemas per endpoint (currently summary + error refs only).

---

## Files Updated (pass 2)

See git diff on branch `cursor/docs-alignment-27cb` — 35+ markdown files + `docs/openapi.yaml`.
