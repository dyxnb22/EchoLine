# Subagent Task Packet Template

Use this template when assigning work to a sub Agent. Keep the task packet narrow, explicit, and testable.

## Required Header

```md
You are a sub Agent for EchoLine.

Use Composer 2.5 with Fast mode disabled.

You are not the global orchestrator. Do not re-plan the whole project. Complete only the task packet below, update required files, run required tests, and return a concise completion summary.
```

## Task Packet

```md
# EchoLine Subagent Task Packet

## Agent Role

<Backend Core Agent / Realtime Agent / Reliability Agent / etc.>

## Task IDs

- <A001>
- <A002>

## Goal

<Concrete goal in 1-3 sentences.>

## Allowed Read Files

- `CURRENT_STATE.md`
- `DONE.md`
- `BLOCKERS.md`
- `TASK_GRAPH_DETAILED.md`
- `<task-specific docs/code>`

## Allowed Write Files

- `<exact files or directories>`

Do not edit files outside this list unless required by tests or explicitly justified in the summary.

## Implementation Requirements

- <Requirement 1>
- <Requirement 2>
- <Requirement 3>

## Required Tests

- `<command>`
- `<command>`

If a test cannot run, explain why and record the fallback verification.

## Required Docs Updates

- `<docs/api.md if API changed>`
- `<docs/data-model.md if schema changed>`
- `<docs/websocket-protocol.md if WS changed>`
- `<docs/reliability.md if delivery semantics changed>`

## Acceptance Criteria

- <Criterion 1>
- <Criterion 2>
- <Criterion 3>

## Stop Conditions

Stop and report if:

- You hit the same blocker 3 times.
- Required dependency is unavailable.
- The task requires changing architecture beyond the packet.
- You need to edit files outside the allowed write list and cannot justify it safely.

## Completion Summary Format

Return:

- Tasks completed:
- Files changed:
- Tests run:
- Docs updated:
- Blockers:
- Follow-up tasks:
```

## Example: Backend Core A001-A003

```md
You are a sub Agent for EchoLine.

Use Composer 2.5 with Fast mode disabled.

You are not the global orchestrator. Do not re-plan the whole project. Complete only the task packet below, update required files, run required tests, and return a concise completion summary.

# EchoLine Subagent Task Packet

## Agent Role

Backend Core Agent

## Task IDs

- A001
- A002
- A003

## Goal

Initialize the Go backend skeleton, add configuration loading, and expose a `/health` endpoint.

## Allowed Read Files

- `CURRENT_STATE.md`
- `DONE.md`
- `BLOCKERS.md`
- `TASK_GRAPH_DETAILED.md`
- `.cursor/skills/backend-core.md`
- `README.md`
- `.env.example`

## Allowed Write Files

- `backend/`
- `Makefile`
- `docs/api.md`
- `PROGRESS_LOG.md`
- `CURRENT_STATE.md`
- `NEXT_ACTIONS.md`
- `DONE.md`

## Implementation Requirements

- Create a Go module under `backend/`.
- Add an API entrypoint under `backend/cmd/api`.
- Add config loading for at least `HTTP_ADDR`.
- Add `GET /health`.
- Keep implementation minimal and testable.

## Required Tests

- `cd backend && go test ./...`
- `make test`

## Required Docs Updates

- Update `docs/api.md` with `/health`.
- Update `PROGRESS_LOG.md`.

## Acceptance Criteria

- Go tests run.
- API server can be started.
- `/health` returns success.
- No unrelated repo-wide rewrite.

## Completion Summary Format

Return:

- Tasks completed:
- Files changed:
- Tests run:
- Docs updated:
- Blockers:
- Follow-up tasks:
```

