# Parallel Agent Plan

EchoLine can be advanced by multiple Cloud Agents if the work is split by ownership boundaries.

For the full orchestration protocol, use `SUBAGENT_ORCHESTRATION.md`. For concrete assignments, use `SUBAGENT_TASK_PACKET.md`.

## Composer 2.5 Rule

When using Composer 2.5 for any sub Agent:

- Fast mode must be disabled.
- The task prompt must explicitly say: `Use Composer 2.5 with Fast mode disabled.`
- If Fast mode status is unclear, confirm before starting.
- Do not trade away context reading, tests, docs, or acceptance checks for speed.

## Agent Tracks

| Agent | Ownership | Primary Files |
|---|---|---|
| Backend Core Agent | auth, user, device, conversation, message API | `backend/internal/auth`, `backend/internal/user`, `backend/internal/conversation`, `backend/internal/message` |
| Realtime Agent | WebSocket, connection manager, presence | `backend/internal/realtime`, `backend/internal/presence`, `docs/websocket-protocol.md` |
| Reliability Agent | idempotency, ACK, seq, sync | `backend/internal/delivery`, `docs/reliability.md`, ADRs |
| Infra Agent | Docker, Redis, MQ, worker, CI | `docker-compose.yml`, `backend/cmd/worker`, `backend/internal/eventbus` |
| Search/Media Agent | MinIO, OpenSearch, notification | `backend/internal/media`, `backend/internal/search` |
| Frontend Agent | web UI and E2E tests | `frontend/` |
| Test/Perf Agent | integration tests, k6, chaos | `backend/tests`, `loadtests`, `reports` |
| Docs/Review Agent | ADR, interview docs, review reports | `docs`, `reports`, `REVIEW_PLAN.md` |

## Coordination Rules

- Each Agent must update `PROGRESS_LOG.md`.
- If multiple Agents may edit the same file, record intent in `CURRENT_STATE.md`.
- Avoid broad refactors while another Agent owns adjacent work.
- Prefer small, independently testable changes.
- Review Agent should not rewrite working code unless review findings are actionable.
- Orchestrator must dispatch work using `SUBAGENT_TASK_PACKET.md`.
- Sub Agents must not re-plan the whole project.
- Sub Agents should only edit files listed in their task packet.
- Orchestrator updates `CURRENT_STATE.md`, `NEXT_ACTIONS.md`, `DONE.md`, and `ACCEPTANCE_MATRIX.md` after collecting sub Agent results.
