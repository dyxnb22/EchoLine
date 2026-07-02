# Research Plan

> **Status (2026-07-02):** 下列研究主题已完成，输出路径见下表「实际产出」列（非 `docs/research/` 子目录）。

Research tasks should produce concrete docs, ADRs, or prototypes. They are used when engineering tasks are blocked or when main/stretched backlogs are complete.

## Research Topics

| ID | Topic | Planned output | Actual output (done) |
|---|---|---|---|
| R001 | Telegram-like system architecture | `docs/research/telegram-like-architecture.md` | [`docs/research-telegram-whatsapp.md`](./docs/research-telegram-whatsapp.md) |
| R002 | Read fanout vs write fanout | `docs/research/fanout-strategies.md` | [`docs/research-fanout-unread.md`](./docs/research-fanout-unread.md), [`docs/interview-fanout.md`](./docs/interview-fanout.md) |
| R003 | Unread count at scale | `docs/research/unread-count.md` | [`docs/research-fanout-unread.md`](./docs/research-fanout-unread.md) |
| R004 | Multi-device sync models | `docs/research/multi-device-sync.md` | [`docs/interview-multi-device-sync.md`](./docs/interview-multi-device-sync.md) |
| R005 | Kafka partitioning for chat | `docs/research/kafka-partitioning.md` | [`docs/research-kafka-sharding.md`](./docs/research-kafka-sharding.md) |
| R006 | Presence system scaling | `docs/research/presence-scaling.md` | [`docs/research-presence-search-outbox.md`](./docs/research-presence-search-outbox.md) |
| R007 | Message search consistency | `docs/research/search-consistency.md` | [`docs/research-presence-search-outbox.md`](./docs/research-presence-search-outbox.md) |
| R008 | E2EE tradeoffs | `docs/research/e2ee.md` | [`research-e2ee-tradeoffs.md`](./research-e2ee-tradeoffs.md) |
| R009 | Modular monolith to microservices | `docs/research/microservices-split.md` | [`docs/adr/0028-microservices-boundary.md`](./docs/adr/0028-microservices-boundary.md) |
| R010 | Ads/payments/recommendations in messaging apps | `docs/research/business-extensions.md` | [`docs/payments-prototype.md`](./docs/payments-prototype.md), [`docs/ads-prototype.md`](./docs/ads-prototype.md), [`docs/recommendation-prototype.md`](./docs/recommendation-prototype.md) |

## Research Quality Bar

Each research output must include:

- problem statement,
- candidate approaches,
- selected recommendation for EchoLine,
- engineering impact,
- interview talking points,
- follow-up tasks.
