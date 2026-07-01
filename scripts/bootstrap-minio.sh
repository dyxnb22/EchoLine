#!/usr/bin/env bash
# scripts/bootstrap-minio.sh
# Initialises a MinIO instance for EchoLine development:
#   1. Wait for MinIO to be reachable
#   2. Create the target bucket
#   3. Set the bucket policy to allow presigned URL downloads (private + presign OK)
#   4. Smoke-test a presigned PUT via the EchoLine API
set -euo pipefail

MINIO_ENDPOINT="${MINIO_ENDPOINT:-http://localhost:9000}"
MINIO_ACCESS_KEY="${MINIO_ACCESS_KEY:-minio}"
MINIO_SECRET_KEY="${MINIO_SECRET_KEY:-minio123}"
MINIO_BUCKET="${MINIO_BUCKET:-echoline}"
MC_ALIAS="echoline-dev"
API_BASE="${API_BASE:-http://localhost:8080}"

log() { echo "[bootstrap-minio] $*"; }
die() { echo "[bootstrap-minio] ERROR: $*" >&2; exit 1; }

# ── install mc if missing ────────────────────────────────────────────────────
if ! command -v mc &>/dev/null; then
  log "mc not found; downloading …"
  case "$(uname -s)" in
    Linux)  MC_URL="https://dl.min.io/client/mc/release/linux-amd64/mc" ;;
    Darwin) MC_URL="https://dl.min.io/client/mc/release/darwin-amd64/mc" ;;
    *)      die "Unsupported OS: $(uname -s)" ;;
  esac
  curl -fsSL "$MC_URL" -o /tmp/mc && chmod +x /tmp/mc
  export PATH="/tmp:$PATH"
  log "mc downloaded to /tmp/mc"
fi

# ── wait for MinIO ────────────────────────────────────────────────────────────
log "Waiting for MinIO at $MINIO_ENDPOINT …"
for i in $(seq 1 30); do
  if curl -sf "$MINIO_ENDPOINT/minio/health/live" > /dev/null 2>&1; then
    log "MinIO is up (attempt $i)"
    break
  fi
  if [[ $i -eq 30 ]]; then
    die "MinIO not reachable after 30 attempts. Start with: docker compose up -d minio"
  fi
  sleep 2
done

# ── configure mc alias ────────────────────────────────────────────────────────
log "Configuring mc alias '$MC_ALIAS' …"
mc alias set "$MC_ALIAS" \
  "$MINIO_ENDPOINT" \
  "$MINIO_ACCESS_KEY" \
  "$MINIO_SECRET_KEY" \
  --api "S3v4" 2>/dev/null

# ── create bucket ─────────────────────────────────────────────────────────────
if mc ls "$MC_ALIAS/$MINIO_BUCKET" &>/dev/null; then
  log "Bucket '$MINIO_BUCKET' already exists"
else
  log "Creating bucket '$MINIO_BUCKET' …"
  mc mb "$MC_ALIAS/$MINIO_BUCKET"
  log "Bucket created"
fi

# ── set bucket versioning off, no public read ──────────────────────────────────
log "Ensuring bucket is private (no anonymous access) …"
mc anonymous set none "$MC_ALIAS/$MINIO_BUCKET" 2>/dev/null || true

# ── CORS configuration (allow browser PUT from frontend origin) ───────────────
log "Writing CORS policy for presigned PUT uploads …"
CORS_JSON=$(cat <<'CORS'
{
  "CORSRules": [
    {
      "AllowedHeaders": ["*"],
      "AllowedMethods": ["GET", "PUT", "POST", "DELETE", "HEAD"],
      "AllowedOrigins": ["http://localhost:5173", "http://localhost:3000", "*"],
      "ExposeHeaders": ["ETag"],
      "MaxAgeSeconds": 3600
    }
  ]
}
CORS
)
echo "$CORS_JSON" | mc cors set "$MC_ALIAS/$MINIO_BUCKET" /dev/stdin 2>/dev/null || \
  log "  (mc cors set not supported on this mc version — configure via MinIO console at :9001)"

# ── smoke test: presigned URL from API ────────────────────────────────────────
log "Smoke-testing presigned upload URL from API at $API_BASE …"
# Requires a valid JWT; skip if API_TOKEN not set
if [[ -n "${API_TOKEN:-}" ]]; then
  PRESIGN_RESP=$(curl -sf -X POST "$API_BASE/media/upload-url" \
    -H "Authorization: Bearer $API_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"filename":"bootstrap-test.txt","content_type":"text/plain"}' || echo '{}')
  PRESIGN_URL=$(echo "$PRESIGN_RESP" | python3 -c \
    "import sys,json; d=json.load(sys.stdin); print(d.get('upload_url',''))" 2>/dev/null || echo "")
  if [[ -n "$PRESIGN_URL" ]]; then
    HTTP_STATUS=$(curl -sf -X PUT "$PRESIGN_URL" \
      -H "Content-Type: text/plain" \
      --data "bootstrap smoke test" \
      -o /dev/null -w "%{http_code}")
    if [[ "$HTTP_STATUS" == "200" ]]; then
      log "  Presigned PUT smoke test passed (HTTP 200)"
    else
      log "  Presigned PUT returned HTTP $HTTP_STATUS (expected 200)"
    fi
  else
    log "  Could not extract presigned URL from API response"
  fi
else
  log "  API_TOKEN not set; skipping presigned URL smoke test"
  log "  Run: export API_TOKEN=<your-jwt> && bash scripts/bootstrap-minio.sh"
fi

# ── list bucket contents ───────────────────────────────────────────────────────
log ""
log "Bucket contents:"
mc ls --summarize "$MC_ALIAS/$MINIO_BUCKET" 2>/dev/null || log "  (empty)"

log ""
log "Done. MinIO is configured for EchoLine."
log "  Endpoint   : $MINIO_ENDPOINT"
log "  Bucket     : $MINIO_BUCKET"
log "  Console    : ${MINIO_ENDPOINT/9000/9001} (user: $MINIO_ACCESS_KEY)"
