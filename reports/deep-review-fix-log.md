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
| WS ACK message validation | `GetByID` on service + WS handler check | `message/service.go`, `realtime/server.go` |

## Wontfix (documented)

- ISSUE-014: Presigned URL sharing within expiry — MVP accepted per `review-security.md`
