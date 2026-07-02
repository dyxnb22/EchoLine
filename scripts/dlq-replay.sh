#!/usr/bin/env bash
# scripts/dlq-replay.sh
# Replay one or all failed events from the EchoLine dead-letter queue.
#
# Usage:
#   # Replay a specific event by ID
#   scripts/dlq-replay.sh --id <dlq-event-uuid>
#
#   # Replay all pending events (admin token required)
#   scripts/dlq-replay.sh --all
#
#   # Dry-run: list events without replaying
#   scripts/dlq-replay.sh --list
#
# Environment variables:
#   BASE_URL      EchoLine API base (default: http://localhost:8080)
#   ADMIN_TOKEN   JWT of an admin user (required for --all and --id)
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
ADMIN_TOKEN="${ADMIN_TOKEN:-}"

log()  { echo "[dlq-replay] $*"; }
die()  { echo "[dlq-replay] ERROR: $*" >&2; exit 1; }
usage() {
  cat >&2 <<EOF
Usage: $0 [--id <uuid>] [--all] [--list] [--base-url <url>] [--token <jwt>]

  --id <uuid>      Replay a single DLQ event by its UUID
  --all            Replay all events with status 'failed'
  --list           List DLQ events (no replay)
  --base-url <url> API base URL (default: http://localhost:8080)
  --token <jwt>    Admin JWT token

Environment variables override defaults:
  BASE_URL, ADMIN_TOKEN
EOF
  exit 1
}

# ── arg parse ─────────────────────────────────────────────────────────────────
MODE=""
TARGET_ID=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --id)        MODE="single"; TARGET_ID="$2"; shift 2 ;;
    --all)       MODE="all";    shift ;;
    --list)      MODE="list";   shift ;;
    --base-url)  BASE_URL="$2"; shift 2 ;;
    --token)     ADMIN_TOKEN="$2"; shift 2 ;;
    -h|--help)   usage ;;
    *)           die "Unknown argument: $1" ;;
  esac
done

[[ -z "$MODE" ]] && usage

# ── auth check ────────────────────────────────────────────────────────────────
if [[ -z "$ADMIN_TOKEN" ]] && [[ "$MODE" != "list" ]]; then
  die "ADMIN_TOKEN is required. Set it with --token or export ADMIN_TOKEN=<jwt>"
fi

AUTH_HEADER="Authorization: Bearer ${ADMIN_TOKEN:-}"

# ── health check ──────────────────────────────────────────────────────────────
log "Checking API health at $BASE_URL/health …"
curl -sf "$BASE_URL/health" > /dev/null \
  || die "API not reachable at $BASE_URL"

# ── list ──────────────────────────────────────────────────────────────────────
fetch_dlq_events() {
  curl -sf "$BASE_URL/admin/dlq" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    || echo '{"events":[]}'
}

if [[ "$MODE" == "list" ]]; then
  log "Fetching DLQ events from $BASE_URL/admin/dlq …"
  RESP=$(fetch_dlq_events)
  COUNT=$(echo "$RESP" | python3 -c \
    "import sys,json; d=json.load(sys.stdin); evts=d.get('events',[]); print(len(evts))" 2>/dev/null || echo "?")
  log "Found $COUNT event(s):"
  echo "$RESP" | python3 -c "
import sys, json
d = json.load(sys.stdin)
for e in d.get('events', []):
    print(f\"  id={e.get('id','?')} type={e.get('event_type','?')} attempts={e.get('attempts','?')} error={e.get('last_error','?')!r}\")
" 2>/dev/null || echo "$RESP"
  exit 0
fi

# ── replay single event ───────────────────────────────────────────────────────
replay_event() {
  local event_id="$1"
  log "Replaying event $event_id …"
  local resp
  resp=$(curl -sf -X POST "$BASE_URL/admin/dlq/$event_id/replay" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    || echo '{"error":"request failed"}')
  local status
  status=$(echo "$resp" | python3 -c \
    "import sys,json; d=json.load(sys.stdin); print(d.get('status','unknown'))" 2>/dev/null || echo "unknown")
  if [[ "$status" == "replayed" || "$status" == "ok" ]]; then
    log "  ✓ Event $event_id replayed successfully"
    return 0
  else
    log "  ✗ Event $event_id replay failed: $resp"
    return 1
  fi
}

if [[ "$MODE" == "single" ]]; then
  [[ -z "$TARGET_ID" ]] && die "--id requires a UUID argument"
  replay_event "$TARGET_ID"
  exit $?
fi

# ── replay all ────────────────────────────────────────────────────────────────
if [[ "$MODE" == "all" ]]; then
  log "Fetching all failed DLQ events …"
  RESP=$(fetch_dlq_events)
  IDS=$(echo "$RESP" | python3 -c "
import sys, json
d = json.load(sys.stdin)
for e in d.get('events', []):
    if e.get('status','') in ('failed', 'pending', ''):
        print(e.get('id',''))
" 2>/dev/null || echo "")

  if [[ -z "$IDS" ]]; then
    log "No failed events found in DLQ."
    exit 0
  fi

  SUCCEEDED=0
  FAILED=0
  while IFS= read -r eid; do
    [[ -z "$eid" ]] && continue
    if replay_event "$eid"; then
      ((SUCCEEDED++))
    else
      ((FAILED++))
    fi
    sleep 0.2   # gentle pacing
  done <<< "$IDS"

  log ""
  log "Replay complete: $SUCCEEDED succeeded, $FAILED failed"
  [[ $FAILED -gt 0 ]] && exit 1 || exit 0
fi
