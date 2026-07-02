# Round 02 — Strict Audit Plan

Fresh full re-audit. Do not inherit Round 01 conclusions.

## Scope

Same full coverage as Round 01: auth, messaging, WS, security, prototype routes, frontend flows, tests.

## Documents re-read

Product docs (architecture, api, websocket-protocol, reliability, security-checklist, business-flows).

## Focus after Round 01 fixes

Verify fixes are correct AND hunt for new issues in:
- rate_limit (IP key, Redis fallback)
- GraphQL sendMessage rate limiting
- ads impression atomicity
- Remaining permission paths

## Test commands

Same as Round 01 quality gates.

## Historical conclusions NOT trusted

Round 01 findings/fix-log, DONE.md, completion manifests.
