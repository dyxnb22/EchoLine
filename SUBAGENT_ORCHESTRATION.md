# Subagent Orchestration Plan

本文件定义 EchoLine 如何使用子 Agent 分派任务，以缓解上下文压力、提高并行度，并保持全局工程状态一致。

## 核心原则

子 Agent 不是自由探索者，而是接收明确 task packet 的执行者。主 Agent / Orchestrator 负责全局状态、任务拆分、冲突控制、验收和合并。

最强组合：

```text
Orchestrator Agent 负责全局状态
Subagents 负责局部实现
Repo files 负责长期记忆
Acceptance Matrix 负责质量门禁
Review Agent 负责找问题和修复任务生成
```

## Composer 2.5 规则

如果使用 Composer 2.5 创建或驱动子 Agent：

- 必须关闭 Fast mode。
- 不允许为了速度牺牲上下文读取、测试、文档更新和验收。
- 子 Agent prompt 必须显式包含：`Use Composer 2.5 with Fast mode disabled.`
- 如果无法确认 Fast mode 状态，子 Agent 必须在开始前提醒操作者确认。

原因：

- EchoLine 是长时工程项目，不是快速补全任务。
- 子 Agent 需要读任务包、相关文档、局部代码和验收标准。
- Fast mode 容易跳过必要上下文，增加重复实现和接口不一致风险。

## Orchestrator Responsibilities

Orchestrator 每轮负责：

1. 读取 `CURRENT_STATE.md`、`NEXT_ACTIONS.md`、`DONE.md`、`BLOCKERS.md`。
2. 选择 1-5 个互不冲突的任务。
3. 为每个子 Agent 生成 task packet。
4. 明确允许读写文件范围。
5. 明确验收标准和测试命令。
6. 收集子 Agent summary。
7. 运行必要的整体验证。
8. 更新 `PROGRESS_LOG.md`、`CURRENT_STATE.md`、`NEXT_ACTIONS.md`、`DONE.md`、`ACCEPTANCE_MATRIX.md`。

Orchestrator 不应该：

- 同时让多个子 Agent 改同一个核心模块。
- 把模糊目标丢给子 Agent。
- 让子 Agent 自行决定架构大改。
- 忽略子 Agent 的 blocker。

## Recommended Subagents

| Agent | Primary Ownership | Good First Tasks |
|---|---|---|
| Backend Core Agent | auth, user, device, conversation, message API | A001-A022 |
| Realtime Agent | WebSocket, connection manager, presence | B001-B012 |
| Reliability Agent | idempotency, ACK, seq, sync | C001-C007, D001-D010 |
| Group Channel Agent | group, channel, fanout | E001-E010 |
| Infra Agent | Docker, Redis, MQ, worker, CI | F001-F010 |
| Media Search Agent | MinIO, OpenSearch, notification | G001-G010 |
| Security Audit Agent | rate limit, risk, audit | H001-H010 |
| Observability Perf Agent | metrics, tracing, k6, chaos | I001-I010 |
| Frontend Agent | React/Next.js UI | J001-J010 |
| Prototype Agent | PWA/mobile/desktop prototype | K001-K005 |
| Docs ADR Agent | docs, ADR, interview writing | L001-L010 |
| Review Refactor Agent | review, findings, targeted fixes | M001-M010 |

## Dispatch Strategy

### Phase 1 Dispatch

Start sequentially until the backend base is stable:

1. Backend Core Agent: A001-A005.
2. Backend Core Agent: A006-A010.
3. Docs ADR Agent: update `docs/api.md` and `docs/data-model.md`.
4. Review Refactor Agent: review auth/API consistency.

Do not parallelize heavily before A005, because repo structure and backend conventions are still forming.

### Phase 2-5 Dispatch

After core API exists:

- Realtime Agent can work on B tasks.
- Reliability Agent can work on C/D tasks.
- Docs ADR Agent can update protocol and reliability docs.
- Review Agent can review message write path.

Avoid conflicts:

- Realtime Agent should not change DB schema without coordinating with Reliability Agent.
- Reliability Agent owns seq/idempotency semantics.
- Backend Core Agent owns REST envelope and auth middleware.

### Phase 6-8 Dispatch

After reliable message flow exists:

- Infra Agent works on Redis/MQ/worker.
- Media Search Agent works on attachments/search.
- Security Audit Agent works on limiter/audit/risk.
- Observability Perf Agent works on metrics/load tests.

Avoid conflicts:

- Infra Agent owns eventbus interfaces.
- Media Search Agent consumes events but should not redesign eventbus.
- Security Audit Agent can add middleware but must preserve API envelope.

### Phase 9-10 Dispatch

Use parallel review and hardening:

- Test/Perf work can run alongside Docs/ADR.
- Review Agent creates findings and targeted fix tasks.
- Prototype Agent can work on PWA/mobile/desktop after web UI is stable.

## Conflict Control

Before dispatch, Orchestrator must check:

- Does another task touch the same directory?
- Does this task change shared schema/API/protocol?
- Does it require an ADR?
- Does it require updating docs used by another sub Agent?

If yes, dispatch sequentially or create a coordination note in `CURRENT_STATE.md`.

## Subagent Completion Contract

Every sub Agent must return or write:

- Task ID(s).
- Files changed.
- Tests run.
- Result.
- Docs updated.
- Blockers.
- Follow-up tasks.

If the sub Agent cannot edit repo files directly, Orchestrator must copy its summary into:

- `PROGRESS_LOG.md`
- `BLOCKERS.md` if blocked
- `NEXT_ACTIONS.md` if follow-ups exist

## Recovery Rules

If a sub Agent stops early:

1. Orchestrator checks git diff.
2. Runs relevant tests.
3. Records partial state in `CURRENT_STATE.md`.
4. Creates a continuation task packet.

If a sub Agent makes broad unrelated changes:

1. Do not blindly continue.
2. Review diff.
3. Keep useful scoped changes.
4. Record risk in `PROGRESS_LOG.md`.
5. Create Review Agent task if needed.

