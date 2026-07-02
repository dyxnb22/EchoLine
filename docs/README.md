# EchoLine 文档索引

本目录与根目录文档的导航地图。修改代码时按 [engineering-standards.md](./engineering-standards.md) 同步更新对应条目。

## 入门

| 文档 | 用途 |
|------|------|
| [../README.md](../README.md) | 项目定位与快速开始 |
| [../CURRENT_STATE.md](../CURRENT_STATE.md) | Agent 恢复用当前状态 |
| [../AGENTS.md](../AGENTS.md) | 长期执行 Agent 说明 |

## 架构与流程

| 文档 | 用途 |
|------|------|
| [architecture.md](./architecture.md) | 模块边界、运行时视图 |
| [business-flows.md](./business-flows.md) | 注册/发消息/付费频道等业务序列 |
| [reliability.md](./reliability.md) | 幂等、seq、ACK、离线补偿 |
| [websocket-protocol.md](./websocket-protocol.md) | WS 消息类型与 envelope |
| [data-model.md](./data-model.md) | 表结构与 migration 索引 |
| [api.md](./api.md) | REST API 契约 |
| [engineering-standards.md](./engineering-standards.md) | 工程标准与 PR 检查项 |

## 决策记录

| 文档 | 用途 |
|------|------|
| [adr/README.md](./adr/README.md) | ADR 格式说明 |
| [adr/0001-architecture-style.md](./adr/0001-architecture-style.md) | 模块化单体 |
| [adr/0030-entitlement-authorization.md](./adr/0030-entitlement-authorization.md) | 付费频道 RBAC |

完整 ADR 列表见 `docs/adr/` 目录。

## 运维与安全

| 文档 | 用途 |
|------|------|
| [security-checklist.md](./security-checklist.md) | 安全控制清单 |
| [../deploy/gateway/README.md](../deploy/gateway/README.md) | API gateway 原型 |
| [extensions-roadmap.md](./extensions-roadmap.md) | 长期扩展方向 |

## 审查报告

| 文档 | 用途 |
|------|------|
| [../reports/code-review-final.md](../reports/code-review-final.md) | 全量收尾审查 |
| [../reports/engineering-review-02.md](../reports/engineering-review-02.md) | RBAC + validate + 文档对齐 |
| [../reports/engineering-review-03.md](../reports/engineering-review-03.md) | API 层统一 + 验证深化 |

## 验证命令

```bash
make verify          # 本地 CI 等价检查
make test            # 后端单元测试
make frontend-build  # 前端构建
```
