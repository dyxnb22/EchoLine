# Deep Review Fix Log

## Iteration 01

| Issue | Fix | Files |
|-------|-----|-------|
| ISSUE-001 | Sync cursor: client `last_seq` authoritative; cursor only when `!has_more`; `pageSize+1` pagination | `sync/handler.go`, `sync/pagination.go` |
| ISSUE-002 | Frontend sync loops on `has_more` | `ChatPage.tsx` |
| ISSUE-003/004 | Idempotent send: pre-lookup, no seq bump on duplicate, no cross-conv reuse | `message/repository.go`, `message/service.go` |
| ISSUE-005 | Attachment download for conversation members | `media/repository.go`, `media/handler.go`, `server/options.go` |
| ISSUE-006 | ACK validates message in conversation | `delivery/handler.go` |
| ISSUE-007 | Outbox claim as `processing` | `migrations/00017`, `outbox/repository.go` |
| ISSUE-008 | `ApiError` + `isPaymentRequired` | `api/http.ts`, `ChatPage.tsx`, `ConversationActions.tsx` |
| ISSUE-009 | WS reconnect uses fresh token | `api.ts` `connectWS(getToken)` |
| ISSUE-010 | Cache includes `role`/`can_publish` | `cache/conversation.go`, `conversation/handler.go` |
| ISSUE-011/012 | Send/upload adopt REST response; dedupe by `client_msg_id` | `api.ts`, `ChatPage.tsx` |
| ISSUE-013 | WS `ErrCannotPublish` → forbidden | `realtime/server.go` |
| ISSUE-015 | Admin health requires admin RBAC | `server/server.go` |
| ISSUE-016/017/018 | WS edit/recall; read on inbound; search navigation ref | `ChatPage.tsx` |

## Iteration 02

| Issue | Fix | Files |
|-------|-----|-------|
| ISSUE-019 | WS ACK message validation | `message/service.go`, `realtime/server.go` |

## Iteration 03 (full-scope)

| Issue | Fix | Files |
|-------|-----|-------|
| ISSUE-023 | Payment settle requires `requires_entitlement` + amount | `payment/handler.go` |
| ISSUE-024/020 | Outbox stale `processing` reaper | `migrations/00018`, `outbox/repository.go`, `cmd/worker/main.go` |
| ISSUE-025 | Conversation list cache invalidation | `conversation/handler.go`, `message/handler.go` |
| ISSUE-026 | Sync includes attachment metadata | `sync/handler.go` |
| ISSUE-027/033 | MarkRead capped; ACK uses message seq | `conversation/repository.go`, `delivery/handler.go`, `realtime/server.go` |
| ISSUE-028 | Pin/report message-in-conversation | `pin/handler.go`, `report/handler.go` |
| ISSUE-031 | Search index on edit/recall | `worker/handlers.go`, `search/repository.go`, `cmd/worker/main.go` |
| ISSUE-032 | GraphQL reaction RBAC | `graph/reaction.go` |
| ISSUE-023 FE | WS edit/recall `message_id` field | `ChatPage.tsx` |
| ISSUE-024 FE | Search navigation replace mode | `ChatPage.tsx` |
| ISSUE-025 FE | Archived API response parse | `api.ts` |
| ISSUE-026 FE | Logout remount clears state | `App.tsx` |
| ISSUE-029 FE | Clear messages on conv switch | `ChatPage.tsx` |
| ISSUE-021 | Attachment download UI | `api.ts`, `ChatPage.tsx` |

## Iteration 04 (full-scope verification)

No new code fixes; verified iteration 03 fixes. Documented wontfix for WS rate limit, CheckOrigin, notifications, client ACK.

## Iteration 05

| Issue | Fix | Files |
|-------|-----|-------|
| ISSUE-044 | Forward clones attachment via `CloneUnlinkedForForward` + S3 copy | `media/repository.go`, `media/client.go`, `message/service.go`, `server/options.go` |
| ISSUE-045 | Thread reply `client_msg_id` + idempotent `SendReply` | `thread/handler.go`, `message/service.go`, `api.ts`, `ThreadPanel.tsx` |
| ISSUE-047 | JWT secret min 32 chars | `config/config.go`, `config_test.go` |
| ISSUE-048 | Register rate limit 10/min/IP | `server/options.go` |
| ISSUE-049 | Loading/empty UI states | `ChatPage.tsx` |

## Wontfix (documented)

- ISSUE-014: Presigned URL sharing within expiry — MVP accepted per `review-security.md`
- ISSUE-041–043, ISSUE-046: WS rate limit, CheckOrigin, notifications producer, client ACK — MVP documented in `deep-review-iteration-04.md`
