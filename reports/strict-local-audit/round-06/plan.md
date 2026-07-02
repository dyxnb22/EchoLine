# Round 06 — Strict Local Audit Plan

**Date:** 2026-07-02  
**Type:** Fresh zero-based audit after round-05 P1/P2 fixes (no reliance on round-05 closure)

## Verification commands

Same full suite as round-05 plus WS smoke.

## Focus re-check

- Rate-limited routes still reachable (sync/ack/search/export/media)
- GraphQL disabled in production default
- PAYMENT_SELF_SERVE production guard
- Frontend WS/sync/unread fixes
- Integration tests with JWT_SECRET default
