# Current State

Current phase: **Documentation alignment — pass 2 complete**.

Milestone: `FINAL_COMPLETION_MANIFEST.md` (closed) + full doc scan aligned with code.

Last session highlights:

- Living docs: fixed ghost `backend/internal/api/` paths in ADRs + interview/security docs
- `architecture.md` module table expanded; `data-model.md` columns completed
- `openapi.yaml` expanded to 61 paths with Error schema + standard error responses
- `extensions-roadmap.md`, `RESEARCH_PLAN.md`, `CLOUD_AGENT_PROMPT.md` closure alignment
- Manifests + review reports: historical path disclaimers
- `conversation.read` WS event documented as proposed (not implemented)

Tests:

- `make verify` — unchanged (doc-only passes)

Blocker:

- Docker/Postgres unavailable in cloud VM for `make smoke-full` — see `BLOCKERS.md`

Next (optional):

1. Local `make dev-up && make smoke-full`
2. OpenAPI per-endpoint request/response schemas
3. Migrate `conversation/handler` to `apierror` envelope
