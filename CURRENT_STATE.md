# Current State

Current phase: **Engineering review #03** — API layer unification & validation depth.

Milestone: `reports/engineering-review-03.md`

Last session highlights:

- Frontend: complete `api.ts` → `http.ts` migration (`authedJSON`, `authedVoid`, `authedBlob`)
- Backend: message Edit sanitize + validate; integration validation test
- Docs: `docs/README.md` index, architecture mermaid, `make verify`
- E2EE client uses auth-aware HTTP layer

Tests:

- `make verify` — go test + build + playwright
- `RUN_INTEGRATION=1` — validation + entitlement + messaging flows

Blocker:

- Docker/Postgres unavailable in cloud VM for `make smoke-full`

Next (optional):

1. Migrate `conversation/handler` to `apierror` envelope (legacy `writeError`)
2. Local `make dev-up && make smoke-full`
