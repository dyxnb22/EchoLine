# Current State

Current phase: **Deep review iteration 04 — full-scope; P0/P1 clear, 2 P2 open**.

Last session highlights:

- **Policy:** Each iteration = full-project audit (not spot-check)
- **Iter 03 P1:** cache invalidation, outbox reaper, payment settle gate, archived API, WS `message_id`
- **Iter 03–04 P2:** sync attachments, MarkRead cap, pin/report, search lifecycle, GraphQL RBAC, download UI
- **Open P2:** forward attachment metadata, thread `client_msg_id`
- **Wontfix P2:** WS rate limit, CheckOrigin, notifications producer, client ACK (MVP)

Tests: `go test ./...` + `npm run build` pass.

Reports: `reports/deep-review-final.md`, `reports/deep-review-iteration-03.md`, `04.md`
