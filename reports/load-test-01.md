# Load Test 01

## 状态

**Done (scripts)** — k6 脚本位于 `loadtests/`；CI 以 dry-run 方式执行。完整压测报告需在本地 compose 栈上跑满负载后更新本节数值。

## 脚本

| 脚本 | 覆盖 |
|------|------|
| `loadtests/k6-ws-connect.js` | WebSocket 连接数 |
| `loadtests/k6-send-message.js` | 单会话发送吞吐 |
| `loadtests/k6-large-group.js` | 群聊 fanout 延迟 |
| `loadtests/k6-api-send.js` | REST 发送路径 |
| `loadtests/k6-mixed-workload.js` | 混合负载 |

## 计划覆盖

- WebSocket 连接数。
- 单会话发送消息吞吐。
- 群聊 fanout 延迟。
- 历史消息分页延迟。

## 运行

```bash
make dev-up && make dev-app
# 需安装 k6
k6 run loadtests/k6-send-message.js
```

详细方法论见 `reports/load-test-01.md` 与 `PERFORMANCE_PLAN.md`。
