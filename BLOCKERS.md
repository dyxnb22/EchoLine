# Blockers

## 2026-07-01 - Docker/PostgreSQL unavailable in cloud VM

Task: A004-A018 integration verification, `make dev-up`, full API smoke.

Attempts:

1. `make dev-up` → `docker: No such file or directory`.
2. `apt-get install postgresql` → permission denied.

Commands:

- `make dev-up`
- `apt-get install postgresql`

Error summary:

Cloud agent environment has no Docker and no permission to install PostgreSQL.

Impact:

DB-dependent integration tests skip when `DATABASE_URL` unset. Full end-to-end API smoke not executed in this session.

Decision:

Proceed with unit tests and code implementation. Re-run integration smoke when Postgres is available via `docker compose` or external `DATABASE_URL`.

Next unblocked task:

A019, A011, C003, B001 (code + unit tests).

## Blocker Template

```md
## YYYY-MM-DD - Short title

Task:

Attempts:

Commands:

Error summary:

Impact:

Decision:

Next unblocked task:
```

