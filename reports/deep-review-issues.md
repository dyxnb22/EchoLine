# Deep Review Issues Log

## ISSUE-001: Sync server cursor skips messages on partial-page retry

Priority: P0  
Status: fixed  
Area: backend/reliability  
Evidence:
- file: `backend/internal/sync/handler.go`
- behavior: stored `device_sync_cursors` overrides client `last_seq` upward; cursor persisted while `has_more=true`
Expected: Client `last_seq` authoritative; cursor advances only after full catch-up  
Actual: Messages between client state and stored cursor lost on retry  
Why it matters: Offline sync data loss  
Fix plan: Remove server override; persist cursor only when `!has_more`; use `pageSize+1` for `has_more`  
Tests: `backend/internal/sync/handler_test.go`

---

## ISSUE-002: Frontend sync ignores `has_more`

Priority: P0  
Status: fixed  
Area: frontend/reliability  
Evidence:
- file: `frontend/src/pages/ChatPage.tsx`
- behavior: single sync pass; cursor set to `latest_seq` not last returned seq
Expected: Paginate until `has_more=false`  
Actual: Gap when >200 messages pending  
Why it matters: UI message loss after reconnect  
Fix plan: Loop sync per conversation; advance cursor by max returned seq  
Tests: manual + existing Playwright build

---

## ISSUE-003: Idempotent send bumps `latest_seq` and re-broadcasts

Priority: P1  
Status: fixed  
Area: backend/reliability  
Evidence:
- file: `backend/internal/message/repository.go` (duplicate path committed seq increment)
- file: `backend/internal/message/service.go` (always broadcasts)
Expected: Duplicate `client_msg_id` returns existing message without side effects  
Actual: Seq gap + duplicate `message.created` fanout  
Why it matters: Unread inflation, duplicate push  
Fix plan: Lookup before create; rollback duplicate tx; skip broadcast on idempotent hit  
Tests: `backend/internal/message/idempotency_test.go`

---

## ISSUE-004: `client_msg_id` global per sender (cross-conversation)

Priority: P1  
Status: fixed  
Area: backend/reliability  
Evidence:
- file: `backend/internal/message/repository.go` — unique on `(sender_id, client_msg_id)` only
Expected: Idempotency scoped per conversation or reject reuse  
Actual: Wrong message returned from another conversation  
Why it matters: Message mis-association  
Fix plan: Reject when existing message `conversation_id` differs  
Tests: idempotency unit test

---

## ISSUE-005: Group members cannot download received attachments

Priority: P1  
Status: fixed  
Area: backend/RBAC  
Evidence:
- file: `backend/internal/media/handler.go` — owner-only `GetByObjectKey`
Expected: Conversation members can download linked attachments  
Actual: 403 for non-owner recipients  
Why it matters: Broken attachment flow in groups  
Fix plan: Membership check via message→conversation join  
Tests: `backend/internal/media/access_test.go`

---

## ISSUE-006: ACK does not verify message in conversation

Priority: P2  
Status: fixed  
Area: backend/reliability  
Evidence:
- file: `backend/internal/delivery/handler.go`
Expected: Reject ACK when `message_id` ∉ `conversation_id`  
Actual: Orphan delivery rows possible  
Fix plan: Load message by conv+id before upsert  
Tests: delivery handler test

---

## ISSUE-007: Outbox duplicate publish (multi-worker)

Priority: P1  
Status: fixed  
Area: backend/architecture  
Evidence:
- file: `backend/internal/outbox/repository.go` — lock released before publish
Expected: Claim rows as `processing` before commit  
Actual: Duplicate MQ events with multiple workers  
Fix plan: Migration `00017` + claim pattern  
Tests: outbox claim unit test

---

## ISSUE-008: Paid channel auto-pay never triggers

Priority: P1  
Status: fixed  
Area: frontend  
Evidence:
- file: `frontend/src/pages/ChatPage.tsx` — checks error string for "402"
- file: `frontend/src/api/http.ts` — throws message only, drops `error.code`
Expected: Branch on `payment_required` / HTTP 402  
Actual: Generic error shown  
Fix plan: `ApiError` with code + status; shared `isPaymentRequired` helper  
Tests: http parse unit test

---

## ISSUE-009: WS reconnect uses stale JWT

Priority: P1  
Status: fixed  
Area: frontend/realtime  
Evidence:
- file: `frontend/src/api.ts` `connectWS` — URL built once at connect
Expected: Fresh token on each reconnect  
Actual: Reconnect loop with expired JWT  
Fix plan: `getToken()` callback per reconnect  
Tests: build verification

---

## ISSUE-010: Cached conversation list missing `can_publish`

Priority: P2  
Status: fixed  
Area: backend/frontend  
Evidence:
- file: `backend/internal/conversation/handler.go` cache path
- file: `frontend/src/pages/ChatPage.tsx` — `can_publish !== false`
Expected: Consistent `role`/`can_publish` on cache hit  
Actual: Subscribers see composer until send fails  
Fix plan: Extend cache struct + fix frontend default  
Tests: cache field test

---

## ISSUE-011: Frontend sender duplicate after sync

Priority: P1  
Status: fixed  
Area: frontend  
Evidence:
- file: `frontend/src/pages/ChatPage.tsx` — dedupe by seq only; optimistic fake seq
Expected: Dedupe by `client_msg_id`; adopt REST response seq  
Actual: Duplicate rows after reconnect  
Fix plan: Return message from `sendMessage`; merge by `client_msg_id`  
Tests: build

---

## ISSUE-012: Upload optimistic message stuck pending

Priority: P1  
Status: fixed  
Area: frontend  
Evidence:
- file: `frontend/src/pages/ChatPage.tsx` `handleUpload`
Expected: Clear pending on success/failure  
Actual: Indefinite "sending..."  
Fix plan: Update state after send like `handleSend`  
Tests: build

---

## ISSUE-013: WS `ErrCannotPublish` wrong error code

Priority: P2  
Status: fixed  
Area: backend/realtime  
Evidence:
- file: `backend/internal/realtime/server.go`
Expected: `forbidden` for publish denial  
Actual: `invalid_request`  
Fix plan: Map `ErrCannotPublish` to forbidden  
Tests: existing realtime tests

---

## ISSUE-014: Presigned URL shareable without membership (MVP accepted)

Priority: P2  
Status: wontfix  
Area: security  
Evidence: `backend/internal/media/handler.go` — time-limited presign  
Expected: Proxy download with membership re-check  
Actual: URL shareable within expiry window  
Why it matters: Low risk for portfolio MVP; documented in `review-security.md`  
Fix plan: Future proxy endpoint; reduce expiry as mitigation  
Tests: n/a

---

## ISSUE-015: `/api/admin/health` any authenticated user

Priority: P2  
Status: fixed  
Area: backend/RBAC  
Evidence: `backend/internal/server/server.go:88`  
Expected: Admin-only or public  
Actual: Any logged-in user sees DB status  
Fix plan: Wrap with `RequireAdmin`  
Tests: admin middleware tests

---

## ISSUE-016: No WS handlers for edit/recall on frontend

Priority: P2  
Status: fixed  
Area: frontend/realtime  
Evidence: `ChatPage.tsx` WS handler  
Expected: Live edit/recall from peers  
Actual: Stale until reload  
Fix plan: Handle `message.edited` / `message.recalled`  
Tests: build

---

## ISSUE-017: Unread grows while viewing active conversation

Priority: P1  
Status: fixed  
Area: frontend  
Evidence: `ChatPage.tsx` — no read cursor update on inbound WS  
Expected: Mark read while focused  
Actual: Sidebar unread increases  
Fix plan: `markConversationRead` on inbound WS when active  
Tests: build

---

## ISSUE-018: Search navigation race with activeId effect

Priority: P1  
Status: fixed  
Area: frontend  
Expected: Jump to searched message  
Actual: `useEffect` reload overwrites targeted page  
Fix plan: `pendingSearchSeq` ref to coordinate load  
Tests: build

---

## ISSUE-019: WS ACK missing message/conversation binding

Priority: P2  
Status: fixed  
Area: backend/realtime  
Evidence: `backend/internal/realtime/server.go`  
Expected: Reject ACK when message not in conversation  
Actual: Same gap as REST before fix  
Fix plan: `messages.GetByID` before `UpsertACK`  
Tests: `go test ./internal/realtime/...`

---

## ISSUE-020: Stale outbox `processing` rows

Priority: P1 (iteration 03)  
Status: fixed  
Area: backend/worker  
Fix: `RequeueStaleProcessing` + `processing_at` column (`00018`)

---

## ISSUE-023–049: Iteration 03–04 full-scope findings

See `reports/deep-review-iteration-03.md` and `reports/deep-review-iteration-04.md` for complete list. Key fixes: cache invalidation, payment gate, archived API, WS payload fields, search lifecycle, GraphQL RBAC, download UI.

Remaining open: ISSUE-044 (forward attachment), ISSUE-045 (thread idempotency), P3 config hardening, P4 UI polish.

