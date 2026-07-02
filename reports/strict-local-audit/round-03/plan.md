# Strict Local Audit — Round 03 Plan

**Date:** 2026-07-02  
**Mode:** Zero-trust full audit (iteration 2 — ignore prior stop claims)

## Stop Condition (global)

Latest full audit finds only P3/P4/P5. No open P0/P1/P2.

## Round 03 Objectives

1. Re-read boundary docs; do NOT trust `final.md`, `DONE.md`, prior round summaries.
2. Full verification matrix (make test, go test, vet, lint, build, smoke-full, playwright).
3. Re-audit backend, frontend, data consistency, test authenticity from scratch.
4. Fix every P0/P1/P2 before round ends.
5. If P0/P1/P2 found → round-04 mandatory.

## Verification Commands

- `make dev-up` / `make dev-app`
- `make test`, `go vet ./...`
- `frontend`: npm ci, lint, build, audit
- `make smoke-full`
- `npx playwright test`
- `RUN_INTEGRATION=1 go test -run Integration ./tests/...` (if DB available)

## Audit Scope

Same full checklist as round-01: auth, messaging, sync, WS, media, payments, admin, outbox, frontend flows, schema consistency, smoke/E2E authenticity.

## Deliverables

- `findings.md`, `fix-log.md`, `test-results.md`, `round-summary.md`
- Update `final.md` when stop condition met (requires round-04 if this round finds P0/P1/P2)
