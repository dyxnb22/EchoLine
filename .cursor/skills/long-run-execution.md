# Skill: Long-Run Execution

Use this skill when running EchoLine for a long Cursor Cloud Agent session.

## Procedure

1. Read `CURRENT_STATE.md`, `NEXT_ACTIONS.md`, `DONE.md`, `BLOCKERS.md`.
2. Pick the first unblocked task from `NEXT_ACTIONS.md`.
3. Cross-check details in `TASK_GRAPH_DETAILED.md`.
4. Implement the task.
5. Run relevant tests.
6. Update docs if needed.
7. After 3-5 tasks, compact context into repo files.
8. Continue with the next task.

## Stop Conditions

Only stop if:

- all main, secondary, stretch, research, and future extension tasks are complete,
- the environment is blocked and no unblocked task remains,
- the user explicitly asks to stop.

