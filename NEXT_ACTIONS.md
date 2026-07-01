# Next Actions

Agent should start here after reading `CURRENT_STATE.md`.

## 10h+ Session Plan (checkpoint 1)

### Sequential P0 (Orchestrator)

1. A019: history pagination tests + edge cases.
2. A011: refresh token endpoint skeleton.
3. C003: conversation list API.
4. C001-C002: latest_seq + seq allocation hardening tests.
5. B001-B003: WebSocket endpoint, auth handshake, connection manager.
6. B004-B007: heartbeat, event envelope, message.send, online push.
7. L001: iteration report for Phase 1 completion.
8. Phase 1 acceptance: full integration smoke with Postgres.

### Parallelizable after B003 (subagents, Fast mode disabled)

- Docs ADR Agent: L002-L004 updates alongside API/WS changes.
- Reliability Agent: C004-C006 sync/unread after seq stable.
- Review Agent: M001 API consistency after A020.

### Phase 2-3 stretch in same session if unblocked

- B010 WS smoke test.
- C006 sync endpoint.
- D001 client_msg_id idempotency (partially in message repo).

## Immediate P0 Tasks

1. A019: Implement and test history message pagination edge cases.
2. A011: Refresh token skeleton.
3. C003: Conversation list API.
4. B001: WebSocket endpoint.

## Environment Setup (when Docker available)

```bash
make dev-up
export DATABASE_URL=postgres://echoline:echoline@localhost:5432/echoline?sslmode=disable
export JWT_SECRET=change-me
make api-run
RUN_API_SMOKE=1 make smoke
```

## If Blocked

- Record blocker in `BLOCKERS.md`.
- Continue with code that does not require DB (protocol structs, WS handler skeleton, docs).
