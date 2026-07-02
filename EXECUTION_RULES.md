# EchoLine 长时执行规则

本文件用于让 Cursor Cloud Agent 在长时执行中持续推进 EchoLine，避免空转、重复规划和低价值 token 消耗。

## 启动规则

每轮执行前：

1. 读取 `CURRENT_STATE.md`。
2. 读取 `NEXT_ACTIONS.md`。
3. 读取 `DONE.md`。
4. 读取 `BLOCKERS.md`。
5. 读取 `PROGRESS_LOG.md` 最近 30 行。
6. 读取 `TASKS.md` 当前 phase。
7. 读取 `TASK_GRAPH_DETAILED.md` 当前任务相关部分。
8. 只读取与当前任务相关的源码和文档。

默认不要全量扫描 repo。

## 主线做完后如何继续

> **2026-07-02:** Phase 0–10 与 T001–T440 已关闭。Agent 继续执行时优先消费 `NEXT_ACTIONS.md` 可选项，或 `docs/extensions-roadmap.md` 中的 future work，而非重读完整 phase 列表。

执行顺序固定为：

1. Phase 0 到 Phase 10。
2. `Secondary Backlog`。
3. `Stretch Backlog`。
4. `Research Backlog`。
5. `docs/extensions-roadmap.md` 中的 future extensions。

当一个阶段所有验收标准完成后，必须更新 `DONE.md` 和 `PROGRESS_LOG.md`，再进入下一阶段。

## 何时切换 backlog

- 当前 phase 所有验收标准满足：进入下一 phase。
- Phase 10 完成：进入 secondary backlog。
- secondary 中没有可执行工程任务：进入 stretch backlog。
- 工程实现受环境限制：转 research backlog，输出 ADR 或 report。

## 遇到阻塞时怎么办

同一问题最多连续尝试 3 次：

1. 第 1 次失败：阅读相关代码和测试，修复。
2. 第 2 次失败：缩小范围，写最小复现或降级实现。
3. 第 3 次失败：在 `PROGRESS_LOG.md` 记录 blocker、失败命令、错误摘要、建议下一步，然后跳过该子任务。

不允许因为单点阻塞停止整个项目。

## 何时必须更新文档

- 新增或修改数据模型。
- 新增或修改 API。
- 新增 WebSocket 协议消息。
- 引入 Redis、MQ、搜索、对象存储等基础设施。
- 改变可靠性、顺序性、扩散、缓存一致性策略。
- 完成一个 phase。

## 何时必须跑测试

- 修改核心业务逻辑后，至少跑相关单测。
- 修改 DB migration 后，跑 migration 测试或本地启动验证。
- 修改 WebSocket / delivery / ACK 后，跑集成或 smoke test。
- phase 完成前，跑 `make test`。
- 测试失败但决定继续时，必须在 `PROGRESS_LOG.md` 记录原因。

## 何时必须写 report

- 每完成一个 phase，写或追加 `reports/iteration-xx.md`。
- 完成压测，写 `reports/load-test-xx.md`。
- 做出重要架构权衡，写 ADR。
- 遇到三次失败的 blocker，写入 `PROGRESS_LOG.md`。

## 何时必须压缩上下文

- 每完成 3-5 个 atomic tasks。
- 每完成一个 milestone。
- 每次准备结束本轮执行前。
- 每次上下文过长、计划过多或实现跨越多个模块时。

压缩方式不是依赖平台功能，而是更新：

- `CURRENT_STATE.md`
- `NEXT_ACTIONS.md`
- `PROGRESS_LOG.md`
- `DONE.md`
- `BLOCKERS.md`
- `DECISIONS.md`

## 如何避免反复全量扫描 repo

- 先读 `PROGRESS_LOG.md`、`DONE.md`、`TASKS.md`。
- 新运行优先读 `CURRENT_STATE.md` 和 `NEXT_ACTIONS.md`。
- 使用 `rg` 定位相关文件。
- 只打开当前任务涉及的目录。
- 已完成模块只有在测试失败、依赖变更或明确任务要求时才重新读取。

## 如何避免低价值重复工作

- 不反复改文案。
- 不反复重排目录。
- 不为了抽象而抽象。
- 不引入无法立刻验证的新组件。
- 不跳过测试直接进入下一阶段。

## 最终汇报必须包含

- 本轮任务。
- 文件改动。
- 验收标准完成情况。
- 测试命令和结果。
- blocker 或风险。
- 下一步建议。
