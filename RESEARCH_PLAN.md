# Research Plan

> **Status (2026-07-02):** 下列研究主题已完成。下表「Original plan filename」为最初计划文件名（未使用单独子目录）；「Actual output」为仓库中的真实文件。

Research tasks should produce concrete docs, ADRs, or prototypes. They are used when engineering tasks are blocked or when main/stretched backlogs are complete.

## Research Topics

| ID | Topic | Original plan filename | Actual output (done) |
|---|---|---|---|
| R001 | Telegram-like system architecture | `telegram-like-architecture.md` | [`docs/research-telegram-whatsapp.md`](./docs/research-telegram-whatsapp.md) |
| R002 | Read fanout vs write fanout | `fanout-strategies.md` | [`docs/research-fanout-unread.md`](./docs/research-fanout-unread.md), [`docs/interview-fanout.md`](./docs/interview-fanout.md) |
| R003 | Unread count at scale | `unread-count.md` | [`docs/research-fanout-unread.md`](./docs/research-fanout-unread.md) |
| R004 | Multi-device sync models | `multi-device-sync.md` | [`docs/interview-multi-device-sync.md`](./docs/interview-multi-device-sync.md) |
| R005 | Kafka partitioning for chat | `kafka-partitioning.md` | [`docs/research-kafka-sharding.md`](./docs/research-kafka-sharding.md) |
| R006 | Presence system scaling | `presence-scaling.md` | [`docs/research-presence-search-outbox.md`](./docs/research-presence-search-outbox.md) |
| R007 | Message search consistency | `search-consistency.md` | [`docs/research-presence-search-outbox.md`](./docs/research-presence-search-outbox.md) |
| R008 | E2EE tradeoffs | `e2ee.md` | [`research-e2ee-tradeoffs.md`](./research-e2ee-tradeoffs.md) |
| R009 | Modular monolith to microservices | `microservices-split.md` | [`docs/adr/0028-microservices-boundary.md`](./docs/adr/0028-microservices-boundary.md) |
| R010 | Ads/payments/recommendations in messaging apps | `business-extensions.md` | [`docs/payments-prototype.md`](./docs/payments-prototype.md), [`docs/ads-prototype.md`](./docs/ads-prototype.md), [`docs/recommendation-prototype.md`](./docs/recommendation-prototype.md) |

## Research Quality Bar

Each research output must include:

- problem statement,
- candidate approaches,
- selected recommendation for EchoLine,
- engineering impact,
- interview talking points,
- follow-up tasks.
