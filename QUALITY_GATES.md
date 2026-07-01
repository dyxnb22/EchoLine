# Quality Gates

Quality gates prevent long-running execution from becoming low-value token consumption.

## Gate 1：Buildability

Required before marking any implementation milestone done:

- Project builds or the failing build is documented in `BLOCKERS.md`.
- `make test` runs.
- New scripts are executable.

## Gate 2：Testability

Core code changes require one of:

- unit test,
- integration test,
- smoke test,
- documented manual verification.

High-risk modules require tests:

- auth,
- message write path,
- seq allocation,
- ACK,
- sync,
- WebSocket,
- rate limiting,
- permissions.

## Gate 3：Documentation Consistency

Must update docs when changing:

- schema,
- API,
- WebSocket protocol,
- reliability behavior,
- cache/MQ behavior,
- security behavior,
- scaling assumptions.

## Gate 4：Operational Evidence

Before claiming system-level completion:

- Run local service.
- Run smoke test.
- Run relevant integration tests.
- Record commands and results in `PROGRESS_LOG.md`.

## Gate 5：No Empty Work

Do not count as progress:

- rewriting prose without new information,
- reformatting unrelated files,
- re-planning already planned phases,
- adding unused dependencies,
- adding decorative UI before core workflows.

