# Architecture Decision Records

ADR 用于记录 EchoLine 的关键架构取舍。每个 ADR 应回答：

- 背景是什么？
- 有哪些候选方案？
- 为什么选择当前方案？
- 带来什么收益？
- 有什么代价和后续风险？

命名格式：

```text
0001-short-title.md
```

## Index

| ADR | Title | Status |
|-----|-------|--------|
| [0001](./0001-architecture-style.md) | Modular monolith first | Accepted |
| [0002](./0002-message-sequence.md) | Per-conversation seq ordering | Accepted |
| [0003](./0003-large-group-fanout.md) | Large group fanout (hub vs worker) | Accepted |
| [0004](./0004-ws-gateway-routing.md) | WebSocket gateway routing | Accepted |
| [0005](./0005-cache-consistency.md) | Cache consistency boundaries | Accepted |
| [0006](./0006-message-tiering.md) | Message hot/cold tiering | Proposed |
| [0007](./0007-conversation-sharding.md) | Conversation sharding | Proposed |
| [0008](./0008-opentelemetry-tracing.md) | OpenTelemetry tracing plan | Accepted |
| [0009](./0009-microservices-split.md) | Microservices split overview | Proposed |
| [0010](./0010-e2ee-threat-model.md) | E2EE threat model | Design |
| [0011](./0011-e2ee-key-management.md) | E2EE key management | Design |
| [0012](./0012-ads-data-model.md) | Ads data model | Design |
| [0013](./0013-payments-ledger.md) | Payments ledger (v1) | **Superseded by 0019** |
| [0014](./0014-mobile-prototype.md) | Mobile prototype scope | Design |
| [0015](./0015-desktop-prototype.md) | Desktop prototype scope | Design |
| [0016](./0016-reactions-threads.md) | Reactions and threads | Accepted |
| [0017](./0017-push-notifications.md) | Push notifications | Accepted (skeleton) |
| [0018](./0018-webhook-delivery.md) | Webhook delivery | Accepted (skeleton) |
| [0019](./0019-payment-ledger.md) | Payment ledger (canonical) | Accepted (skeleton) |
| [0020](./0020-ads-platform.md) | Ads platform | Accepted (skeleton) |
| [0021](./0021-recommendation-engine.md) | Recommendation engine | Accepted (skeleton) |
| [0022](./0022-graphql-subscriptions.md) | GraphQL subscriptions | Proposed |
| [0023](./0023-admin-rbac.md) | Admin RBAC | Accepted |
| [0024](./0024-webhook-retry.md) | Webhook retry worker | Accepted |
| [0025](./0025-graphql-facade-scope.md) | GraphQL facade scope | Accepted (prototype) |
| [0026](./0026-e2ee-key-bundle.md) | E2EE key bundle API | Accepted (demo) |
| [0027](./0027-api-gateway-prototype.md) | API gateway prototype | Accepted |
| [0028](./0028-microservices-boundary.md) | Microservices boundary | Design |
| [0029](./0029-opentelemetry-stub.md) | OpenTelemetry stub | Accepted |
| [0030](./0030-entitlement-authorization.md) | Paid channel entitlements RBAC | Accepted |
| [0031](./0031-cache-and-mq-responsibilities.md) | Redis vs MQ responsibilities | Draft |

Lightweight decisions that do not warrant a full ADR live in [`../../DECISIONS.md`](../../DECISIONS.md).
