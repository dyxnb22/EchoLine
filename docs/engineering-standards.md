# EchoLine 工程标准

本项目遵循的现代软件工程原则与实践。

## 架构原则

| 原则 | EchoLine 实践 |
|------|----------------|
| 模块化单体优先 | `internal/*` 包边界清晰，避免过早微服务 |
| Source of truth | PostgreSQL 持久化优先于 WS 推送 |
| 接口可演进 | event bus、gateway、OTel 均为可替换 stub |
| 失败可补偿 | seq + client_msg_id + sync + outbox |

## 代码组织

### 后端（Go）

- **Handler** — HTTP 解析、鉴权、状态码映射；不含 SQL。
- **Service** — 业务编排（message.Send、幂等、广播）。
- **Repository** — 参数化 SQL（pgx）。
- **validate** — 集中字段长度与格式校验。
- **apierror** — 统一错误 envelope。

新 API 必须：

1. 更新 `docs/api.md`
2. 添加 handler 或 service 测试
3. schema 变更走 goose migration + `docs/data-model.md`

### 前端（React + TypeScript）

- **pages/** — 路由级页面
- **components/** — 可复用 UI
- **context/** — 全局状态（Auth）
- **api/http.ts** — `publicJSON`、`authedJSON`、`authedVoid`、`authedBlob`、`bindAuthFetch`
- **api.ts** — 领域 API（全部经 `http.ts`  helpers，禁止裸 `fetch` 调 EchoLine API）

新 API 函数必须使用 `authedJSON` / `authedVoid` 等 helper，以便 401 自动 refresh 与统一错误解析。

## 安全基线

- 密码 bcrypt；JWT 来自环境变量
- 成员关系校验先于消息读写
- 付费频道：owner 配置、admin grant、支付 settle 自动 grant
- HTML strip 于消息体（`middleware.SanitizeBody`）
- 输入长度限制（`internal/validate`）

## 测试策略

| 层级 | 工具 | 何时跑 |
|------|------|--------|
| 本地聚合 | `make verify` | PR 前 / Agent checkpoint |
| 单元 | `go test ./...` | 每次 PR |
| 集成 | `RUN_INTEGRATION=1` + Postgres | CI + 本地 compose |
| E2E | Playwright + vite preview | CI |
| 负载 | k6 dry-run | PR |

## CI 要求

- Goose migrations（`go run ./cmd/migrate`），禁止 `psql || true`
- `go test -race`、frontend `tsc`、Playwright 不得 `continue-on-error`（关键 job）

## 文档契约

重要变更同步更新：

| 变更类型 | 文档 |
|----------|------|
| REST | `docs/api.md` |
| Schema | `docs/data-model.md` |
| WS | `docs/websocket-protocol.md` |
| 交付语义 | `docs/reliability.md` |
| 业务流 | `docs/business-flows.md` |
| 架构取舍 | `docs/adr/*.md` |

## 面试讲述

- **为什么先写 DB 再推送？** — 可靠性；WS 不可靠。
- **付费频道如何不破坏订阅 API？** — 402 + entitlement 表；支付与 admin 两条 grant 路径。
- **如何从单体演进？** — gateway 路由 + ADR 0028 服务边界 + outbox 一致性。
