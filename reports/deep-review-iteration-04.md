# Deep Review — Iteration 04 (Full-Scope Verification)

Date: 2026-07-02  
**Policy:** Full-project audit (all 15 backend domains + 14 frontend flows + architecture/docs).

## Verification of iteration 03 fixes

| Fix | Verified |
|-----|----------|
| Conversation list cache invalidation | `InvalidateConversationListCache` wired on send/read/subscribe |
| Outbox stale `processing` reaper | `RequeueStaleProcessing` + migration `00018` |
| Payment settle gate | Requires `amount_cents >= 1` + `requires_entitlement` |
| Sync attachment metadata | `ToCreatedPayloadWithAttachment` in sync handler |
| MarkRead cap | `LEAST(..., latest_seq)` in SQL |
| ACK read uses message seq | REST + WS |
| Pin/report message-in-conversation | `GetByID` validation |
| WS edit/recall `message_id` | Frontend field fix |
| Archived API parse | `archived[].conversation_id` |
| Search navigation merge | `loadMessages(..., "replace")` |
| Logout state leak | `ChatPage key={token}` remount |
| Search index lifecycle | Worker handles `message.edited` / `message.recalled` |
| GraphQL reaction RBAC | Membership check in `graph.ReactionService` |
| Attachment download UI | `presignDownload` + Download button |

## New findings (iteration 04)

| ID | Priority | Status | Summary |
|----|----------|--------|---------|
| ISSUE-041 | P2 | wontfix | WS `message.send` bypasses `conv_send` rate limit — MVP accepts; REST path limited |
| ISSUE-042 | P2 | wontfix | `CheckOrigin` allows all origins — dev default; production should set reverse-proxy origin policy |
| ISSUE-043 | P2 | wontfix | Notification API has no producer — skeleton per extensions roadmap |
| ISSUE-044 | P2 | open | Forward drops attachment metadata |
| ISSUE-045 | P2 | open | Thread replies use server-generated `client_msg_id` |
| ISSUE-046 | P2 | wontfix | Web client does not send delivery ACK — uses `markConversationRead` instead |
| ISSUE-047 | P3 | open | JWT secret min length not enforced |
| ISSUE-048 | P3 | open | Register endpoint not rate-limited |
| ISSUE-049 | P4 | open | Conversation list loading/empty polish |

## Counts (iteration 04 only)

| Priority | Found | Fixed | Wontfix | Remaining |
|----------|-------|-------|---------|-----------|
| P0 | 0 | 0 | 0 | 0 |
| P1 | 0 | 0 | 0 | 0 |
| P2 | 6 | 0 | 4 | 2 |
| P3 | 2 | 0 | 0 | 2 |
| P4 | 1 | 0 | 0 | 1 |

**Stop condition:** Not fully met — 2 open P2 remain (forward attachment, thread idempotency). Documented for follow-up; all security-critical P1 resolved.
