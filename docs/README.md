# EchoLine 文档索引

本目录与根目录文档的导航地图。修改代码时按 [engineering-standards.md](./engineering-standards.md) 同步更新对应条目。

## 入门

| 文档 | 用途 |
|------|------|
| [../README.md](../README.md) | 项目定位与快速开始 |
| [../CURRENT_STATE.md](../CURRENT_STATE.md) | Agent 恢复用当前状态 |
| [../NEXT_ACTIONS.md](../NEXT_ACTIONS.md) | 可选后续任务 |
| [../AGENTS.md](../AGENTS.md) | 长期执行 Agent 说明 |
| [../FINAL_COMPLETION_MANIFEST.md](../FINAL_COMPLETION_MANIFEST.md) | T001–T440 关闭范围 |

## 架构与流程

| 文档 | 用途 |
|------|------|
| [architecture.md](./architecture.md) | 模块边界、运行时视图 |
| [business-flows.md](./business-flows.md) | 注册/发消息/付费频道等业务序列 |
| [reliability.md](./reliability.md) | 幂等、seq、ACK、离线补偿 |
| [websocket-protocol.md](./websocket-protocol.md) | WS 消息类型与 envelope |
| [data-model.md](./data-model.md) | 表结构与 migration 索引 |
| [api.md](./api.md) | REST API 契约 |
| [openapi.yaml](./openapi.yaml) | OpenAPI 路由镜像（61 paths；body schema 详见 api.md） |
| [engineering-standards.md](./engineering-standards.md) | 工程标准与 PR 检查项 |

## 决策记录

| 文档 | 用途 |
|------|------|
| [adr/README.md](./adr/README.md) | ADR 完整索引（0001–0031） |
| [../DECISIONS.md](../DECISIONS.md) | 轻量决策（未升格 ADR） |

## 运维与安全

| 文档 | 用途 |
|------|------|
| [security-checklist.md](./security-checklist.md) | 安全控制清单 |
| [../deploy/gateway/README.md](../deploy/gateway/README.md) | API gateway 原型 |
| [extensions-roadmap.md](./extensions-roadmap.md) | 长期扩展方向 |
| [scaling.md](./scaling.md) | 扩展与分片设计 |
| [chaos-playbook.md](./chaos-playbook.md) | 故障演练 |
| [dlq-replay.md](./dlq-replay.md) | DLQ 重放操作 |

## 面试与讲述

| 文档 | 用途 |
|------|------|
| [interview-mapping.md](./interview-mapping.md) | 面试题 ↔ 模块映射 |
| [interview-system-design.md](./interview-system-design.md) | 系统设计讲述提纲 |
| [interview-reliability.md](./interview-reliability.md) | 可靠性专题 |
| [interview-fanout.md](./interview-fanout.md) | 群聊扩散 |
| [interview-multi-device-sync.md](./interview-multi-device-sync.md) | 多端同步 |

## 原型与扩展设计

| 文档 | 用途 |
|------|------|
| [encryption-prototype.md](./encryption-prototype.md) | E2EE demo |
| [graphql-prototype.md](./graphql-prototype.md) | GraphQL facade |
| [payments-prototype.md](./payments-prototype.md) | 支付账本 |
| [ads-prototype.md](./ads-prototype.md) | 广告平台 |
| [recommendation-prototype.md](./recommendation-prototype.md) | 推荐 |
| [push-notifications.md](./push-notifications.md) | Push 骨架 |

## 审查报告

| 文档 | 用途 |
|------|------|
| [../reports/code-review-final.md](../reports/code-review-final.md) | 全量收尾审查 |
| [../reports/engineering-review-02.md](../reports/engineering-review-02.md) | RBAC + validate + 文档对齐 |
| [../reports/engineering-review-03.md](../reports/engineering-review-03.md) | API 层统一 + 验证深化 |
| [../reports/review-docs-consistency.md](../reports/review-docs-consistency.md) | 文档一致性审查与修复记录 |

## 验证命令

```bash
make verify          # go test + build + playwright
make test            # 后端单元测试
make frontend-build  # 前端构建
make dev-up          # 本地依赖栈（需 Docker）
make smoke-full      # 全栈 API smoke（需 dev-up + dev-app）
```
