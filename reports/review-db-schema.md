# Code Review: Database Schema (M002)

**Reviewer**: Automated review via agent
**Date**: 2026-07-01
**Scope**: All migrations in `backend/migrations/` and schema referenced in `docs/data-model.md`

---

## Summary

The EchoLine database schema is well-structured with correct use of UUIDs, appropriate foreign keys, and good index placement on hot query paths. Several improvements are identified for correctness, performance, and future-proofing.

---

## Finding 1: Missing Index on `messages.created_at` for Archival

**Severity**: Medium
**Files**: `backend/migrations/`

**Observation**: The `messages` table is indexed on `(conversation_id, seq)` for read queries. However, the planned archival worker (ADR 0006) will query `WHERE created_at < threshold`, which lacks an index.

**Recommendation**: Add `CREATE INDEX ON messages (created_at)` or use a composite index `(created_at, conversation_id)` to support the archival query efficiently.

---

## Finding 2: `outbox.payload` Column as JSONB — No Schema Enforcement

**Severity**: Low
**Files**: `backend/migrations/` (outbox table), `backend/internal/worker/outbox.go`

**Observation**: The `outbox.payload` is JSONB with no schema validation. A bug in the producer could enqueue malformed payloads that the consumer cannot parse, causing repeated DLQ entries.

**Recommendation**: Add a Postgres domain or check constraint validating minimum required fields:
```sql
ALTER TABLE outbox ADD CONSTRAINT payload_has_event_id
  CHECK (payload ? 'event_id' AND payload ? 'conversation_id');
```
Alternatively, validate at the application layer before insert and add a unit test for payload structure.

---

## Finding 3: `conversation_members.role` Stored as Text — No Enum

**Severity**: Low
**Files**: `backend/migrations/`, `backend/internal/conversation/roles.go`

**Observation**: `role TEXT CHECK (role IN ('owner', 'admin', 'member'))` — using a check constraint on TEXT rather than a Postgres enum.

**Recommendation**: Either is acceptable. If the role set is fixed, a Postgres enum (`CREATE TYPE member_role AS ENUM ('owner', 'admin', 'member')`) provides stronger type checking at the DB level and prevents typos from application code reaching the DB.

**Tradeoff**: Postgres enums are hard to extend (requires `ALTER TYPE`); TEXT with check constraint is more flexible. Given `owner/admin/member` is unlikely to change, enum is appropriate.

---

## Finding 4: No `deleted_at` for Soft Deletes

**Severity**: Medium
**Files**: `backend/migrations/` (users, conversations, messages tables)

**Observation**: Message recall (`POST /api/messages/:id/recall`) updates `messages.status = 'recalled'`. User account deletion is not yet implemented. No tables use soft-delete (`deleted_at TIMESTAMPTZ`).

**Recommendation**: For GDPR compliance (right to erasure), plan for soft-delete on `users` and hard-delete job for old soft-deleted records. For `messages`, the current `status = 'recalled'` approach is correct (recalled messages should remain for legal/audit purposes with `body = null`).

Add `deleted_at TIMESTAMPTZ` to `users` table now, before the table has production data. Adding it later requires a migration on a live large table.

---

## Finding 5: `messages.seq` Potential Race on High-Concurrency Send

**Severity**: Medium
**Files**: `backend/internal/message/repo.go`, `backend/migrations/`

**Observation**: Seq allocation uses:
```sql
UPDATE conversations SET latest_seq = latest_seq + 1 RETURNING latest_seq
```
This is correct for serializable isolation. However, if the API uses `READ COMMITTED` (Postgres default), two concurrent sends could assign the same seq.

**Recommendation**: Verify the seq allocation transaction uses `BEGIN ISOLATION LEVEL SERIALIZABLE` or that the query is run within a transaction that prevents the race. Alternatively, use a Postgres sequence object:
```sql
CREATE SEQUENCE conversation_{id}_seq START 1;
SELECT nextval('conversation_{id}_seq');
```
Note: per-conversation sequences are not practical; the UPDATE approach with a row-level lock (the UPDATE locks the row) is correct and sufficient.

**Verification needed**: Add a concurrent test that sends 100 messages simultaneously and asserts all seqs are unique and contiguous.

---

## Finding 6: No Partial Index for `outbox` Pending Rows

**Severity**: Low (performance)
**Files**: `backend/migrations/`

**Observation**: The outbox worker queries `WHERE status = 'pending'`. Without a partial index, this scans all outbox rows including old `published` rows.

**Current state**: `CREATE INDEX ON outbox (status, created_at)` exists (or should exist).

**Recommendation**: Use a partial index for better performance:
```sql
CREATE INDEX idx_outbox_pending ON outbox (created_at)
  WHERE status = 'pending';
```
This index only contains pending rows; the index size stays small as published rows are not included.

---

## Finding 7: `device_sync_cursors` Lacks Updated-At Trigger

**Severity**: Low
**Files**: `backend/migrations/`

**Observation**: `device_sync_cursors.updated_at` exists but must be manually updated by the application. If the application code forgets to set it, the column becomes stale.

**Recommendation**: Add a trigger:
```sql
CREATE TRIGGER set_updated_at
  BEFORE UPDATE ON device_sync_cursors
  FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();
```
Where `trigger_set_timestamp()` is a reusable function that sets `updated_at = now()`.

---

## Finding 8: Missing Composite Index on `message_deliveries`

**Severity**: Medium (performance)
**Files**: `backend/migrations/`

**Observation**: The delivery state query filters by `(message_id, user_id)`. If only individual indexes exist on each column, Postgres may use a bitmap scan instead of a single index scan.

**Recommendation**: 
```sql
CREATE UNIQUE INDEX ON message_deliveries (message_id, user_id);
```
This also enforces the constraint that each (message, user) pair has exactly one delivery record.

---

## Overall Assessment

**Schema quality score**: 8/10. The schema correctly implements the domain model. Key improvements: partial index on outbox, soft-delete column on users, composite index on deliveries, and concurrent seq allocation test.

## Files to Update

- `backend/migrations/` — add missing indexes, partial index, soft-delete column, triggers
- `docs/data-model.md` — document indexing strategy and soft-delete plan
- `backend/internal/message/repo.go` — verify seq allocation transaction isolation
