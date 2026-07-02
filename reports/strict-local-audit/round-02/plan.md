# Strict Local Audit — Round 02 Plan

**Date:** 2026-07-02
**Mode:** Fresh full audit (ignore round-01 "fixed" assumptions; re-verify from code + runtime)

## Objectives

1. Re-read boundary docs (architecture, api, websocket-protocol, data-model, reliability, security-checklist)
2. Re-run full verification matrix on current branch
3. Re-audit backend auth/messaging/sync/ws, frontend flows, data consistency, test authenticity
4. Record any new findings; fix any P0/P1/P2 immediately

## Verification (re-run all)

Same command matrix as round-01 — all must pass for stop condition.

## Audit focus (independent pass)

- Auth: register, login, refresh rotation, JWT, rate limits
- Messaging: send, idempotency, edit/recall, sync pagination
- WebSocket: auth, origin, rate limit, attachment in push
- Media: download auth for members vs removed users
- Outbox: reclaim worker present
- Frontend: download, unread, sync race, optimistic send
- Smoke/E2E authenticity vs live API
- Docker compose startup path

## Stop gate for this round

If only P3/P4/P5 found → proceed to final.md
If any P0/P1/P2 → fix and start round-03
