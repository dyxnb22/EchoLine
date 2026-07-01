# Skill: Reliability Review

Use this skill when implementing or reviewing message reliability.

## Checklist

- Is the message persisted before WebSocket push?
- Is `client_msg_id` required for client-created messages?
- Is duplicate send idempotent?
- Is `seq` assigned inside a transaction?
- Can offline clients recover through sync?
- Does ACK state only move forward?
- Are workers idempotent?
- Is MQ failure recoverable through outbox or retry?

## Required Outputs

- Tests for duplicate sends.
- Tests for ordered reads.
- Documentation in `docs/reliability.md`.
- ADR for major tradeoffs.

