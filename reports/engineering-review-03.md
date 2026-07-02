# Engineering Review #03 — API Layer Unification & Validation Depth

Date: 2026-07-01  
Branch: `cursor/engineering-review-03-758a`

## Scope

Third-pass review building on #02: complete frontend HTTP migration, deepen message validation, documentation index, local verify pipeline.

## Findings & fixes

### P0 — Consistency

| Issue | Fix |
|-------|-----|
| ~35 `api.ts` calls still used raw `fetch` + manual error handling | Full migration to `api/http.ts` helpers |
| `lib/e2ee.ts` bypassed auth refresh layer | Uses `authedRequest` + `parseResponse` |
| Message edit skipped sanitize/validate | `Service.Edit` applies `SanitizeBody` + `validate.MessageBody` |

### P1 — Documentation

| Issue | Fix |
|-------|-----|
| No single docs navigation index | `docs/README.md` with cross-links |
| Architecture lacked runtime diagram | Mermaid flowchart in `architecture.md` |
| No one-command local verify | `make verify` → `scripts/verify-all.sh` |

### P2 — Tests

| Addition | Purpose |
|----------|---------|
| `integration_validation_test.go` | Empty + oversized message body → 400 |

## Architecture notes

- **Frontend API layer** is now a thin facade over `http.ts` — aligns with single-responsibility and DRY.
- **Validation** applied symmetrically on Send and Edit paths.
- **parseResponse** handles empty JSON bodies (204) without throwing.

## Module scores (post #03)

| Module | Score | Change |
|--------|-------|--------|
| frontend api | A | B+ → A (full http.ts migration) |
| message | A | Edit path validation added |
| docs | A+ | Index + mermaid + verify script |
| validate | A | Integration test coverage |

## Remaining gaps (documented)

- GraphQL prototype without schema codegen
- Full stack smoke requires Docker
- OpenAPI per-endpoint body schemas (partial — key endpoints in `docs/api.md`)

## Verification

```bash
make verify
RUN_INTEGRATION=1 DATABASE_URL=... go test -run Integration ./tests/...
```
