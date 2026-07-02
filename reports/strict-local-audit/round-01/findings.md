# Round 01 Findings

## SLA-R01-001: Docker API fails — migrations path broken

Priority: P1  
Status: fixed  
Area: backend/reliability  
Business flow: `make dev-app` → API startup  
Expected: Migrations apply from `/app/migrations` in container  
Actual: `runtime.Caller` resolved to build-time source path; API exited  
Evidence:
- file: `backend/internal/migrate/migrate.go`
- runtime: `docker logs echoline-api-1` → `. directory does not exist`
Impact: Core stack unavailable via Docker Compose  
Root cause: migrationsDir() not executable-relative  
Fix plan: MIGRATIONS_DIR env + executable-adjacent fallback  
Required tests: `backend/internal/migrate/migrate_test.go`  
Docs to update: docker-compose.yml

---

## SLA-R01-002: Outbox events stuck in processing

Priority: P1  
Status: fixed  
Area: reliability  
Expected: Crashed worker reclaims stale processing rows  
Actual: No reclaim path existed  
Evidence: `backend/internal/outbox/repository.go` FetchPending only selects pending  
Impact: Async side effects permanently lost  
Fix: migration 00018 + ReclaimStaleProcessing + worker loop  
Required tests: `repository_test.go`, `worker_test.go`

---

## SLA-R01-003: Attachment download missing in frontend

Priority: P0  
Status: fixed  
Area: frontend  
Expected: Download via POST `/api/media/download-url`  
Actual: Upload only; no download UI  
Evidence: `frontend/src/api.ts`, `ChatPage.tsx`  
Fix: `presignDownload()` + Download button

---

## SLA-R01-004: Unread badge stale while viewing

Priority: P1  
Status: fixed  
Area: frontend  
Expected: Opening conversation clears sidebar unread  
Actual: markConversationRead called but state not updated  
Evidence: `ChatPage.tsx` lines 279-283, 656  
Fix: `clearUnread()` optimistic update + refresh after mark-read

---

## SLA-R01-005: WS reconnect sync race

Priority: P1  
Status: fixed  
Area: frontend/realtime  
Expected: Sync after reconnect even if conversations load late  
Actual: runSync skipped when cursors empty at WS onOpen  
Evidence: `ChatPage.tsx` runSync + conversations load effect  
Fix: Re-trigger sync when conversations first load while WS open

---

## SLA-R01-006: ACK read-seq forgery

Priority: P2  
Status: fixed  
Area: security/reliability  
Expected: MarkRead uses message's actual seq  
Actual: Client-supplied seq accepted  
Evidence: `delivery/handler.go`, `realtime/server.go`  
Fix: Bind MarkRead to GetByID seq; cap HandleMarkRead at latest_seq

---

## SLA-R01-007: Media owner bypass after removal

Priority: P2  
Status: fixed  
Area: security  
Expected: Conversation attachments require membership  
Actual: OwnerID match bypassed membership check  
Evidence: `media/repository.go:123-125`  
Fix: Owner bypass only for pending (unlinked) uploads

---

## SLA-R01-008: WS message.send rate-limit bypass

Priority: P2  
Status: fixed  
Area: security  
Expected: Same conv_send quota as REST  
Actual: WS send had no limiter  
Fix: SetConvSendLimiter on realtime Server

---

## SLA-R01-009: WS CheckOrigin always true

Priority: P2  
Status: fixed  
Area: security  
Expected: Restrict cross-origin WS  
Actual: CheckOrigin returned true  
Fix: `origin.go` with localhost defaults + WS_ALLOWED_ORIGINS

---

## SLA-R01-010: Sync missing next_seq cursor hint

Priority: P2  
Status: fixed  
Area: backend/reliability  
Expected: Pagination exposes page max seq  
Actual: Only latest_seq (head) returned  
Fix: Added `next_seq` field in sync response

---

## SLA-R01-011: Smoke script API drift

Priority: P2  
Status: fixed  
Area: tests  
Expected: Smoke matches live API  
Actual: Non-UUID client_msg_id, wrong group path/fields  
Evidence: `scripts/smoke-api-full.sh` failures at runtime  
Fix: UUID client_msg_ids, `/api/conversations/groups`, `title` field

---

## SLA-R01-012: Payment self-serve enabled in compose

Priority: P1 (when enabled)  
Status: fixed  
Area: security  
Expected: Self-serve off by default in compose  
Actual: PAYMENT_SELF_SERVE=true in docker-compose  
Fix: Set to false

---

## SLA-R01-013: Refresh token reuse

Priority: P2  
Status: fixed  
Area: security  
Expected: Refresh rotation / reuse detection  
Actual: Stateless refresh without JTI tracking  
Fix: JTI on refresh tokens + MemoryRefreshStore

---

## SLA-R01-014: OpenSearch fallback skips live membership

Priority: P2  
Status: fixed  
Area: security  
Expected: OS hits re-validated against DB membership  
Actual: Index-only filter  
Fix: SetMemberChecker on search handler

---

## SLA-R01-015: Register/refresh unrated

Priority: P3  
Status: fixed  
Area: security  
Fix: registerMW + refreshMW in applyRateLimits

---

## SLA-R01-016: Presigned URL shareability

Priority: P3  
Status: documented  
Area: security  
Expected: MVP accepts presigned URL sharing risk  
Actual: By design for MVP  
Impact: Low — documented in prior ADR/reports

---

## SLA-R01-017: Frontend stale messages on conversation switch

Priority: P3  
Status: open  
Area: frontend  
Expected: Clear messages while loading  
Actual: Previous conversation messages flash until load completes  
Impact: UX confusion, not data loss

---

## SLA-R01-018: Integration tests opt-in only

Priority: P3  
Status: open  
Area: tests  
Expected: CI runs integration with DATABASE_URL  
Actual: t.Skip without RUN_INTEGRATION=1  
Impact: Reduced CI coverage; unit/smoke compensate-page.tsx compensate locally
