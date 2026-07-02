# Code Review: Security (M006)

> **Historical note (2026-07-02):** 报告中 `backend/internal/api/*` 路径为设计期命名；当前见 `auth/service.go`、`media/handler.go` 等。

**Reviewer**: Automated review via agent
**Date**: 2026-07-01
**Scope**: Authentication, authorization, input validation, rate limiting, audit logging

---

## Summary

EchoLine's security posture is appropriate for a prototype. Core authentication (JWT, bcrypt) and authorization (membership checks, role-based group actions) are implemented correctly. The following findings identify gaps for production hardening.

---

## Finding 1: JWT Secret — No Minimum Length Enforcement

**Severity**: High
**Files**: `backend/internal/config/config.go`, `backend/internal/auth/`

**Observation**: The JWT is signed with `JWT_SECRET` from env. If the operator sets a short or weak secret (e.g., `secret`), JWT tokens can be brute-forced.

**Recommendation**: Enforce minimum secret length at startup:
```go
if len(cfg.JWTSecret) < 32 {
    log.Fatal("JWT_SECRET must be at least 32 characters")
}
```

For production, use a 256-bit random secret: `openssl rand -base64 32`. Document this in `.env.example`.

---

## Finding 2: No CSRF Protection for State-Changing Endpoints

**Severity**: Medium
**Files**: All API handlers

**Observation**: EchoLine's API uses JWT Bearer tokens, which are not automatically sent by browsers (unlike cookies). This means CSRF is not a risk for the API itself. However, if any endpoint uses cookie-based session (e.g., for legacy auth), CSRF protection is needed.

**Recommendation**: Confirm that no cookie-based auth is used. If JWT is always in the `Authorization: Bearer` header (not a cookie), CSRF is not applicable. Document this explicitly in `docs/security-checklist.md`.

**If cookies are added in the future**: Add `SameSite=Strict` cookie attribute and CSRF token middleware.

---

## Finding 3: Attachment Download URL — No Authorization Check

**Severity**: High
**Files**: `backend/internal/api/media.go`

**Observation**: MinIO presigned download URLs are time-limited but do not enforce conversation membership after generation. If Alice shares a download URL with Carol (who is not a member), Carol can download the file.

**Status (2026-07-02):** Partially addressed — `GetAccessibleByObjectKey` checks conversation membership; presign expiry reduced to 5 minutes; frontend download UI added. URL sharing outside membership window remains a residual risk (Option 2 mitigation).
1. **URL signing**: Sign the download URL with `user_id` and validate on each access (requires a proxy layer in front of MinIO).
2. **Short expiry**: Use a 5-minute presign expiry, so URLs are useless after the user's session ends. Currently implemented with 1-hour expiry — consider reducing.
3. **Membership re-check in download endpoint**: Create a proxy endpoint `GET /api/media/{attachment_id}/download` that checks membership and redirects to a freshly generated presigned URL.

Option 3 is most secure but adds latency. Option 2 is a practical mitigation for MVP.

---

## Finding 4: Password Reset — Not Implemented

**Severity**: Medium (completeness)
**Files**: `backend/internal/api/auth.go`

**Observation**: There is no password reset flow. A user who forgets their password cannot recover their account. This is a significant UX gap, but also a security concern: without a verified email-based reset, account recovery is impossible if credentials are lost.

**Recommendation**: Implement:
1. `POST /api/auth/forgot-password` — sends a time-limited reset token to the user's email.
2. `POST /api/auth/reset-password` — validates the token and sets a new password.
3. Store reset tokens in Postgres with 1-hour expiry.

**Prerequisite**: User registration must collect an email address (currently only username + password).

---

## Finding 5: Audit Log Not Covering All Critical Events

**Severity**: Medium
**Files**: `backend/internal/audit/`

**Observation**: Current audit logging covers: login attempts and message recall. Missing from audit:
- Group member removal (kick/leave)
- Role changes (promote/demote)
- Account deletion (future)
- Admin actions
- Password changes

**Recommendation**: Add audit logging to all security-relevant state changes. The audit logger should be a service called from handlers, not embedded in handlers directly.

---

## Finding 6: No Content-Security-Policy Header

**Severity**: Low (frontend)
**Files**: `backend/internal/api/` (response middleware), `frontend/`

**Observation**: The API and frontend serve responses without a `Content-Security-Policy` header. This increases XSS risk if any user content is rendered without escaping.

**Recommendation**: Add CSP header middleware:
```
Content-Security-Policy: default-src 'self'; script-src 'self'; img-src 'self' blob: data:; connect-src 'self' wss://;
```
Adjust `connect-src` to allow the WS endpoint origin.

---

## Finding 7: Refresh Token Not Invalidated on Password Change

**Severity**: Medium
**Files**: `backend/internal/api/auth.go`

**Observation**: If a user changes their password (after implementing Finding 4), their existing refresh tokens remain valid. An attacker who has stolen a refresh token can continue using it after the victim changes their password.

**Recommendation**: On password change: `DELETE FROM refresh_tokens WHERE user_id = $1` to revoke all existing refresh tokens. Force all sessions to re-authenticate.

---

## Finding 8: WS Endpoint Not Protected Against Credential Stuffing

**Severity**: Low
**Files**: `backend/internal/realtime/server.go`

**Observation**: The WS endpoint validates JWTs but has no rate limit on failed WS authentication attempts. An attacker could attempt to connect with invalid tokens at high frequency without being blocked.

**Recommendation**: Add IP-based rate limiting on the `/ws` upgrade endpoint (e.g., max 10 failed auth attempts per IP per minute). Reuse the existing Redis rate limiter.

---

## Overall Assessment

**Security score**: 7/10. JWT + bcrypt + membership checks are correctly implemented. High-priority findings: JWT secret minimum length (Finding 1), download URL authorization (Finding 3). Medium-priority: password reset (Finding 4), refresh token invalidation (Finding 7).

## Priority Fixes

1. **Finding 1** (HIGH): JWT secret minimum length enforcement.
2. **Finding 3** (HIGH): Download URL authorization.
3. **Finding 5** (MEDIUM): Expand audit logging.
4. **Finding 4** (MEDIUM): Implement password reset.
5. **Finding 7** (MEDIUM): Invalidate refresh tokens on password change.

## Files to Update

- `backend/internal/config/config.go` — JWT secret length check
- `backend/internal/api/media.go` — download URL authorization
- `backend/internal/audit/` — expand audit event coverage
- `backend/internal/api/auth.go` — password reset, token revocation
- `docs/security-checklist.md` — mark fixed items
