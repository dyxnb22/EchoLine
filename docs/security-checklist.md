# EchoLine Security Checklist

This checklist covers the security controls implemented or planned in EchoLine, categorized by layer. Use this for security review, threat modeling sessions, and deployment sign-off.

---

## Authentication and Authorization

- [x] **Password hashing**: bcrypt with cost 12. Never store or log plaintext passwords.
- [x] **JWT access tokens**: HS256, 15-minute expiry. Secret via `JWT_SECRET` env var (not hardcoded).
- [x] **Refresh tokens**: Stored in DB with expiry; revocable.
- [x] **JWT middleware**: All protected routes require `Authorization: Bearer <token>`.
- [x] **Membership check**: Every message send, read, and search verifies the requesting user is a member of the target conversation.
- [x] **Group role checks**: Group kick/invite/promote requires owner or admin role.
- [x] **Channel entitlement RBAC**: owner sets paid flag; admin grants; payment settle auto-grants (ADR 0030).
- [ ] **JWT rotation**: Implement JWKS endpoint for key rotation without service restart.
- [ ] **Token binding**: Bind refresh token to device fingerprint to prevent refresh token theft.

---

## Rate Limiting

- [x] **Login rate limit**: 5 attempts per IP per minute (Redis sliding window).
- [x] **Register rate limit**: 3 attempts per IP per 10 minutes.
- [x] **Message send rate limit**: 60 messages per user per minute.
- [ ] **WS connection rate limit**: Limit new WS connections per IP per second to prevent connection floods.
- [ ] **API global rate limit**: Per-IP rate limit on all endpoints to prevent DDoS amplification.
- [ ] **Graduated backoff**: Progressive delays after repeated failed logins.

---

## Input Validation and Sanitization

- [x] **SQL injection prevention**: All queries use parameterized statements via `pgx`.
- [x] **UUID validation**: All ID parameters validated as valid UUID format before DB query.
- [x] **Content-Type enforcement**: API requires `application/json`; rejects unexpected content types.
- [x] **Input length limits**: `username` 64, `display_name` 128, `body` 65535 via `internal/validate`.
- [ ] **XSS prevention in web client**: React JSX escapes by default; verify no `dangerouslySetInnerHTML` on user content.
- [ ] **File type validation**: Verify uploaded file MIME type matches declared type; reject executable MIME types.

---

## Media and File Security

- [x] **Presigned URLs**: File uploads use S3 presigned URLs; files never transit the API server.
- [x] **MinIO access**: MinIO bucket is private; public access disabled; all access via presigned URLs.
- [ ] **Virus scanning**: Integrate ClamAV (or cloud-based AV API) as a Kafka consumer on `media.uploaded` events. See `docs/virus-scan-mock.md`.
- [ ] **File size limits**: Enforce max 100 MB per upload at the presign URL generation step.
- [ ] **Media URL expiry**: Presigned download URLs expire in 1 hour to prevent link sharing of private media.
- [ ] **CDN signed URLs**: In production, serve media via CDN with signed tokens; revoke on conversation member removal.

---

## Transport Security

- [x] **TLS in production**: Deploy behind a TLS-terminating reverse proxy (Nginx/Caddy/ALB). All HTTP and WS connections over HTTPS/WSS.
- [ ] **HSTS header**: Set `Strict-Transport-Security: max-age=31536000; includeSubDomains` on all responses.
- [ ] **mTLS for internal services**: When microservices are extracted (ADR 0009), use mTLS between services. Use Istio or Linkerd as service mesh.
- [ ] **Certificate rotation**: Document and automate TLS certificate rotation (Let's Encrypt via cert-manager).

---

## Secrets Management

- [x] **Env-based secrets**: All secrets (`JWT_SECRET`, `DATABASE_URL`, Redis/Kafka credentials, S3 keys) provided via environment variables.
- [x] **`.env.example`**: Committed to repo with placeholder values only; actual `.env` in `.gitignore`.
- [ ] **Secrets rotation**: Document rotation procedure for `JWT_SECRET` (requires token invalidation and re-login).
- [ ] **Production secrets**: Use a secrets manager (AWS Secrets Manager, Vault, Kubernetes Secrets) instead of plain env vars in production.
- [ ] **No secrets in logs**: Verify no middleware or handler logs authorization headers, tokens, or passwords.

---

## Audit Logging

- [x] **Login audit**: Every login attempt (success/failure) logged to `audit_logs` table with IP, user agent, and timestamp.
- [x] **Message recall audit**: Recall events logged with `actor_id`, `message_id`, `reason`.
- [ ] **Admin action audit**: All admin API calls (ban user, delete message) logged.
- [ ] **Log forwarding**: Ship structured JSON logs to a SIEM (Splunk, Datadog, OpenSearch).
- [ ] **Log retention**: 90-day retention for access logs; 7-year retention for audit logs.

---

## Session and Device Security

- [x] **Per-device tokens**: Each device registration creates an independent JWT; one device's compromise does not affect others.
- [ ] **Device trust level**: High-trust operations (password change, device revocation) require re-authentication.
- [ ] **Session revocation**: API to list and revoke all active sessions (`DELETE /api/devices/{id}`).
- [ ] **Anomalous login detection**: Alert user if a new device logs in from a new country/IP.

---

## Database Security

- [x] **Parameterized queries**: All DB queries use `pgx` prepared statements.
- [ ] **DB user least-privilege**: The `echoline` Postgres user should not have `SUPERUSER`; only the permissions required for its tables.
- [ ] **Connection encryption**: Postgres connection should use `sslmode=require` in production (current default is `sslmode=disable` for dev).
- [ ] **DB backup encryption**: Backups encrypted at rest.

---

## WebSocket Security

- [x] **WS authentication**: Token validated before WS upgrade; connection rejected without valid JWT.
- [x] **Origin check**: Verify `Origin` header matches allowed origins to prevent CSRF over WS.
- [ ] **WS message size limit**: Reject WS messages larger than 1 MB to prevent memory exhaustion.
- [ ] **Connection timeout**: Disconnect WS clients that do not send a ping within 90 seconds.

---

## Dependency Security

- [x] **govulncheck**: Run `govulncheck ./...` in CI (security job).
- [x] **npm audit**: Run `npm audit` in CI for frontend dependencies.
- [ ] **Base image scanning**: Scan Docker base images with Trivy or Grype.
- [ ] **Dependency pinning**: Pin dependency versions; use Dependabot or Renovate for updates.

---

## E2EE (Extension)

- [ ] **Threat model**: See `docs/adr/0010-e2ee-threat-model.md`.
- [ ] **Key management**: See `docs/adr/0011-e2ee-key-management.md`.
- [ ] **Server-side key blindness**: Server stores only public keys. Private keys never leave devices.

---

## Incident Response

- [ ] **On-call runbook**: Document steps for: DB breach, token compromise, DDoS, data deletion request.
- [ ] **Breach notification**: GDPR requires notification within 72 hours. Document the process.
- [ ] **Chaos playbook**: See `docs/chaos-playbook.md` for planned failure injection.

---

## Files Involved

- `backend/internal/middleware/auth.go` — JWT middleware
- `backend/internal/middleware/ratelimit.go` — Redis rate limiting
- `backend/internal/audit/` — audit log service
- `backend/internal/api/auth.go` — login/register with rate limit and audit
- `docs/adr/0010-e2ee-threat-model.md` — E2EE threat model
- `docs/adr/0011-e2ee-key-management.md` — E2EE key management
- `docs/virus-scan-mock.md` — virus scan design
- `.env.example` — environment variable reference
