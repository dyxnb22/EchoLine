# EchoLine Virus Scan Mock Design (ST06 / H008)

This document describes the virus scanning architecture for EchoLine's media uploads, including the mock implementation for development and the production integration design.

---

## Problem

EchoLine allows users to upload arbitrary files (images, documents, videos) via MinIO presigned URLs. Without virus scanning:

1. A user could upload malware-infected files that are then served to other conversation members.
2. An adversary could use EchoLine as a malware distribution vector.
3. EchoLine could face legal liability for hosting and distributing known malware or CSAM.

## Threat Model

- **Scope**: Binary executables, Office macros, PDF exploits, archive bombs, known malware signatures.
- **Out of scope**: Zero-day exploits not in AV signature databases; encrypted files (impossible to scan without decryption).
- **CSAM**: Handled separately via perceptual hash matching (PhotoDNA or equivalent). Out of scope for this document.

## Architecture

### Upload Flow (with Virus Scan)

```
Client                     API                MinIO              Kafka             Scanner
  │                         │                   │                   │                 │
  ├─ POST /media/upload-url ►│                   │                   │                 │
  │  ◄─ presigned PUT URL ──┤                   │                   │                 │
  │                         │                   │                   │                 │
  ├─ PUT {presigned URL} ───────────────────────►│                   │                 │
  │  ◄─ 200 (upload done) ──────────────────────┤                   │                 │
  │                         │                   │                   │                 │
  ├─ POST /media/confirm ───►│                   │                   │                 │
  │                         ├──[media.uploaded]──────────────────────►│                 │
  │  ◄─ 202 (scanning) ─────┤                   │                   │                 │
  │                         │                   │                   ├─[consume]──────►│
  │                         │                   │                   │                 ├─ ClamAV scan
  │                         │                   │                   │                 │
  │                         │                   │ [if clean]        ◄─[media.scanned]─┤
  │                         ├──[update status]──►│                   │                 │
  │                         │                   │                   │                 │
  │ (push scan result)      ◄──[WS: media.ready]─┤                   │                 │
```

### Key Properties

1. **Non-blocking upload**: File upload completes immediately. Scanning happens asynchronously.
2. **Status-aware delivery**: Media attachments have a `scan_status` field: `pending | clean | infected | error`.
3. **Pre-delivery hold**: Until `scan_status = clean`, the download URL is not served to other users.
4. **Time limit**: If scan doesn't complete within 5 minutes, status transitions to `error` and the file is treated as undeliverable.

---

## Database Schema

```sql
ALTER TABLE attachments ADD COLUMN scan_status TEXT NOT NULL DEFAULT 'pending'
  CHECK (scan_status IN ('pending', 'scanning', 'clean', 'infected', 'error'));
ALTER TABLE attachments ADD COLUMN scanned_at TIMESTAMPTZ;
ALTER TABLE attachments ADD COLUMN scan_engine TEXT;     -- 'clamav', 'mock', etc.
ALTER TABLE attachments ADD COLUMN scan_result TEXT;     -- human-readable result
```

---

## Mock Implementation (Development)

In development, a mock scanner processes every file and marks it as `clean` after a 2-second simulated delay. This allows the full upload → scan → deliver flow to be tested without a real AV engine.

```go
// backend/internal/scanner/mock.go

type MockScanner struct {
    // InfectedPatterns: file names containing these strings will be marked infected
    // Useful for testing the infected flow: upload a file named "test-malware.exe"
    InfectedPatterns []string
    ScanDelay        time.Duration
}

func (s *MockScanner) Scan(ctx context.Context, objectKey string) (ScanResult, error) {
    time.Sleep(s.ScanDelay) // simulate scan time

    for _, pattern := range s.InfectedPatterns {
        if strings.Contains(objectKey, pattern) {
            return ScanResult{
                Status:   "infected",
                Engine:   "mock",
                Details:  fmt.Sprintf("Mock: file matches infected pattern '%s'", pattern),
            }, nil
        }
    }

    return ScanResult{
        Status: "clean",
        Engine: "mock",
    }, nil
}
```

**Test scenario: upload an infected file**:
```bash
# Upload a file with "malware" in the name → mock scanner marks as infected
curl -X PUT "${PRESIGNED_URL}" \
  --data-binary @test-malware.exe \
  -H "Content-Type: application/octet-stream"

# Confirm upload
curl -X POST "${API_BASE}/api/media/confirm" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{"attachment_id": "..."}'

# After 2 seconds, the attachment scan_status = 'infected'
# Download URL is NOT served; message shows "[File removed: failed safety check]"
```

---

## Production Implementation (ClamAV)

In production, the scanner worker consumes `media.uploaded` events from Kafka and calls ClamAV:

```go
// backend/internal/scanner/clamav.go

type ClamAVScanner struct {
    socket string // e.g., "/var/run/clamav/clamd.sock" or "tcp://clamav:3310"
}

func (s *ClamAVScanner) Scan(ctx context.Context, objectKey string) (ScanResult, error) {
    // 1. Download file from MinIO to a temp file
    tmpFile, err := downloadToTemp(ctx, objectKey)
    if err != nil { return ScanResult{}, err }
    defer os.Remove(tmpFile)

    // 2. Send to ClamAV via clamd protocol
    result, err := clamdScan(s.socket, tmpFile)
    if err != nil { return ScanResult{}, err }

    // 3. Return result
    if result.Infected {
        return ScanResult{
            Status:  "infected",
            Engine:  "clamav",
            Details: result.VirusName,
        }, nil
    }
    return ScanResult{Status: "clean", Engine: "clamav"}, nil
}
```

### ClamAV Deployment

```yaml
# docker-compose.yml addition
clamav:
  image: clamav/clamav:latest
  ports:
    - "3310:3310"
  volumes:
    - clamav-db:/var/lib/clamav  # AV signature database
  environment:
    - CLAMD_CONF_FILE=/etc/clamav/clamd.conf
```

ClamAV requires:
- Signature database update: `freshclam` runs daily as a cron job.
- Memory: ~1 GB for the signature database in RAM.
- CPU: Scan throughput ~50 MB/s on 1 CPU core.

### Cloud AV Alternative

For higher throughput or to avoid self-hosting ClamAV:
- **AWS Macie** (S3-native, but limited to Macie patterns)
- **VirusTotal API** (multi-engine, 1000 file/day on free tier; $0.0001/file on paid)
- **OPSWAT MetaDefender** (on-premise or cloud, multi-engine)

---

## Handling Infected Files

When a file is marked `infected`:

1. `attachments.scan_status = 'infected'`.
2. The MinIO object is moved to a quarantine bucket (not deleted, for forensics).
3. The message is delivered with `attachment.status = 'removed'` and body replaced by: `"[File removed: failed safety check]"`.
4. An alert is fired to the security team with the file hash and uploader's `user_id`.
5. The uploader's account is flagged for review (threshold: 3 infected uploads → account review).

---

## Testing

| Test | Expected Result |
|------|-----------------|
| Upload clean file (e.g., `hello.txt`) | `scan_status = 'clean'` after 2s (mock) |
| Upload "infected" file (name contains `malware`) | `scan_status = 'infected'` after 2s (mock) |
| Send message with attached clean file | Recipient can download after scan completes |
| Send message with attached infected file | Recipient sees `[File removed]` |
| ClamAV service down | `scan_status = 'error'` after 5-minute timeout; file not delivered |

---

## Files Involved

- `backend/internal/scanner/mock.go` _(planned)_ — mock scanner
- `backend/internal/scanner/clamav.go` _(planned)_ — ClamAV integration
- `backend/internal/scanner/interface.go` _(planned)_ — `Scanner` interface
- `backend/internal/worker/scanner_worker.go` _(planned)_ — Kafka consumer for `media.uploaded`
- `backend/migrations/` — add `scan_status`, `scanned_at` to `attachments`
- `docs/security-checklist.md` — virus scan checklist items
- `docker-compose.yml` — add ClamAV container
