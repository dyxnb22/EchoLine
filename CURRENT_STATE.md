# Current State

Current phase: **Deep review iteration 05 — stop condition met**.

Last session highlights:

- **Iter 05 P2:** forward attachment clone (`CloneUnlinkedForForward` + S3 copy), thread reply `client_msg_id` idempotency
- **Iter 05 P3:** JWT secret min 32 chars, register rate limit 10/min/IP
- **Iter 05 P4:** conversation list + message loading/empty states
- **Wontfix P2 (MVP):** WS rate limit, CheckOrigin, notifications producer, client ACK

Tests: `go test ./...` + `npm run build` pass. `make smoke-full` still blocked (no Docker in cloud VM).

Reports: `reports/deep-review-iteration-05.md`, `reports/deep-review-final.md`
