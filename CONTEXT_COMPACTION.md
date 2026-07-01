# Context Compaction Rules

EchoLine 不依赖平台级上下文压缩。Agent 必须主动维护 repo 内的恢复上下文，保证中断、压缩、重启、换 Agent 后仍可继续。

## 每完成 3-5 个 atomic tasks

更新：

- `PROGRESS_LOG.md`
- `CURRENT_STATE.md`
- `NEXT_ACTIONS.md`

记录：

- 完成了什么。
- 改了哪些文件。
- 测试结果。
- 当前最重要的下一步。
- 不要重复做的事情。

## 每完成一个 milestone

更新：

- `DONE.md`
- `ACCEPTANCE_MATRIX.md`
- `reports/iteration-xx.md`
- 相关 `docs/`

如果产生架构取舍，新增或更新：

- `docs/adr/*.md`
- `DECISIONS.md`

## 遇到 blocker

同一问题尝试 3 次仍失败时：

1. 写入 `BLOCKERS.md`。
2. 在 `CURRENT_STATE.md` 标记影响范围。
3. 在 `NEXT_ACTIONS.md` 选择可绕开的任务。
4. 不停止整个项目。

## 上下文过长时

优先写入 repo 文件，而不是在对话里长篇总结。

必须更新：

- `CURRENT_STATE.md`
- `NEXT_ACTIONS.md`
- `PROGRESS_LOG.md`

然后继续执行下一个任务。

## 恢复流程

新 Agent 或新运行开始时只读：

1. `CURRENT_STATE.md`
2. `NEXT_ACTIONS.md`
3. `DONE.md`
4. `BLOCKERS.md`
5. 当前任务相关文档和源码

不要默认全量扫描 repo。

