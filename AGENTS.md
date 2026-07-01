# EchoLine Agent 执行说明

你是 EchoLine 项目的长期执行 Agent。EchoLine 是 Telegram-like messaging platform，不是普通 IM demo。你的目标是持续推进 repo 中的任务图，优先交付可运行、可测试、可讲述的工程能力。

## 工作入口

每轮开始时只读取必要上下文：

1. `CURRENT_STATE.md`。
2. `NEXT_ACTIONS.md`。
3. `DONE.md`。
4. `BLOCKERS.md`。
5. `PROGRESS_LOG.md` 最近记录。
6. `TASKS.md` 当前 phase。
7. `TASK_GRAPH_DETAILED.md` 中当前任务相关部分。
8. 如使用子 Agent，读取 `SUBAGENT_ORCHESTRATION.md` 和 `SUBAGENT_TASK_PACKET.md`。
9. `.cursor/rules/` 和 `.cursor/skills/` 中相关文件。
10. 与当前任务直接相关的源码或文档。

不要每轮全量扫描仓库。需要定位文件时优先使用 `rg`。

## 任务选择顺序

1. 优先完成当前 phase 未完成任务。
2. 当前 phase 验收通过后，更新 `DONE.md`、`PROGRESS_LOG.md`，再进入下一 phase。
3. 主线 Phase 0-10 完成后，消费 `secondary backlog`。
4. secondary 清空后，消费 `stretch backlog`。
5. stretch 清空后，消费 `research backlog`。
6. research 清空后，消费 `docs/extensions-roadmap.md` 中的 future extensions。

## 每轮必须产出

每轮执行至少产生一种高价值产出：

- 可运行代码。
- 测试。
- 文档更新。
- ADR。
- 迭代报告。
- 可复现的 bug / blocker 记录。

## 长时执行约束

- 使用 `CLOUD_AGENT_PROMPT.md` 启动长跑。
- 每完成 3-5 个 atomic tasks，执行 `CONTEXT_COMPACTION.md` 中的 repo-based memory compaction。
- 不要把一次回复视为项目完成；一次回复只是 checkpoint。
- 主线完成后继续 `BACKLOG.md`，再继续 `docs/extensions-roadmap.md` 中的 future extensions。
- 子 Agent 必须通过 `SUBAGENT_TASK_PACKET.md` 分派，不允许自由重规划整个项目。
- 使用 Composer 2.5 的子 Agent 必须关闭 Fast mode。

## 禁止事项

- 不反复重写 README。
- 不反复重构目录结构。
- 不为了“更优雅”重写可工作的模块。
- 不在没有需求的情况下引入新框架。
- 不跳过验收标准去做炫技功能。
- 不重复实现 `DONE.md` 中已经完成的模块。

## 修改文档的规则

- 新增或修改数据模型：更新 `docs/data-model.md`。
- 新增或修改 API：更新 `docs/api.md`。
- 新增 WebSocket 消息类型：更新 `docs/websocket-protocol.md`。
- 改变架构取舍：新增 ADR。
- 完成 phase：更新 `DONE.md`、`PROGRESS_LOG.md` 和对应 report。
- 发现重要限制：更新 `PROGRESS_LOG.md` 或 `reports/iteration-xx.md`。

## 最终汇报格式

每轮结束时汇报：

- 本轮选择的任务。
- 修改了哪些文件。
- 完成了哪些验收标准。
- 跑了哪些测试，结果如何。
- 遇到的 blocker。
- 下一步建议任务。
