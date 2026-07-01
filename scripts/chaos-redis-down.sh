#!/usr/bin/env bash
# EchoLine Chaos Script: Redis Down
# Task: I008
#
# Injects a Redis failure and observes system behavior.
# See docs/chaos-playbook.md for the full experiment design.
#
# Usage:
#   ./scripts/chaos-redis-down.sh [duration_seconds]
#
# Default duration: 60 seconds
#
# Prerequisites:
#   - Docker running with echoline-redis container
#   - EchoLine API running and serving traffic
#   - Prometheus at PROMETHEUS_URL for metrics observation
#
# What to observe during the experiment:
#   1. API /metrics endpoint: rate limit bypassed (no Redis INCR)
#   2. Logs: "redis: connection refused" warnings (not panics)
#   3. Message send: still returns 200 (DB path unaffected)
#   4. Conversation list: falls back to Postgres (slower but correct)
#   5. WS connections: maintained (hub is in-process, no Redis dependency)

set -euo pipefail

DURATION=${1:-60}
REDIS_CONTAINER=${REDIS_CONTAINER:-echoline-redis}
API_BASE=${API_BASE_URL:-http://localhost:8080}
INTERVAL=5  # seconds between health checks

# ─── Colors ───────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log()  { echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $*"; }
ok()   { echo -e "${GREEN}[$(date '+%H:%M:%S')] ✓${NC} $*"; }
warn() { echo -e "${YELLOW}[$(date '+%H:%M:%S')] ⚠${NC} $*"; }
fail() { echo -e "${RED}[$(date '+%H:%M:%S')] ✗${NC} $*"; }

# ─── Pre-flight checks ────────────────────────────────────────────────────────

log "=== CHAOS-001: Redis Down ==="
log "Duration: ${DURATION}s"
log "Redis container: ${REDIS_CONTAINER}"
log "API: ${API_BASE}"

# Check API is healthy before starting
if ! curl -sf "${API_BASE}/health" > /dev/null 2>&1; then
  fail "API health check failed. Is the API running at ${API_BASE}?"
  exit 1
fi
ok "Pre-flight: API is healthy"

# Check Redis container exists
if ! docker inspect "${REDIS_CONTAINER}" > /dev/null 2>&1; then
  fail "Redis container '${REDIS_CONTAINER}' not found. Is Docker running?"
  exit 1
fi
ok "Pre-flight: Redis container found"

# ─── Baseline metrics ─────────────────────────────────────────────────────────

log "Recording baseline metrics..."
BASELINE_ERRORS=$(curl -sf "${API_BASE}/metrics" 2>/dev/null | grep 'echoline_http_errors_total' | awk '{print $2}' || echo "0")
log "Baseline HTTP errors: ${BASELINE_ERRORS}"

# ─── Inject Failure ───────────────────────────────────────────────────────────

log "=== INJECTING FAILURE: Stopping Redis container ==="
docker stop "${REDIS_CONTAINER}"
warn "Redis is DOWN. Experiment begins now."

EXPERIMENT_START=$(date +%s)
PASS_COUNT=0
FAIL_COUNT=0

# ─── Observation Loop ─────────────────────────────────────────────────────────

while true; do
  NOW=$(date +%s)
  ELAPSED=$((NOW - EXPERIMENT_START))

  if [ "${ELAPSED}" -ge "${DURATION}" ]; then
    break
  fi

  log "--- t+${ELAPSED}s ---"

  # Check 1: API health endpoint
  HEALTH_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "${API_BASE}/health" 2>/dev/null || echo "000")
  if [ "${HEALTH_STATUS}" = "200" ]; then
    ok "Health check: 200 (API alive)"
    PASS_COUNT=$((PASS_COUNT + 1))
  else
    fail "Health check: ${HEALTH_STATUS} (API degraded)"
    FAIL_COUNT=$((FAIL_COUNT + 1))
  fi

  # Check 2: Message send (should still work via Postgres)
  SEND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST "${API_BASE}/api/conversations/test/messages" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer test-token" \
    -d '{"body":"chaos test","client_msg_id":"chaos-001","type":"text"}' \
    2>/dev/null || echo "000")
  # 401 is expected (no real token); 000 means connection refused (bad)
  if [ "${SEND_STATUS}" != "000" ]; then
    ok "Message send endpoint reachable (${SEND_STATUS})"
  else
    fail "Message send endpoint unreachable (connection refused)"
    FAIL_COUNT=$((FAIL_COUNT + 1))
  fi

  # Check 3: Look for Redis error in logs (warn level expected, not panic)
  # (In a real setup, you'd tail the log here)
  log "Tip: Check API logs for 'redis: connection refused' warn messages"

  sleep "${INTERVAL}"
done

# ─── Restore ──────────────────────────────────────────────────────────────────

log "=== RESTORING: Starting Redis container ==="
docker start "${REDIS_CONTAINER}"
ok "Redis is UP."

# Wait for Redis to be ready
for i in $(seq 1 10); do
  if docker exec "${REDIS_CONTAINER}" redis-cli ping > /dev/null 2>&1; then
    ok "Redis PING successful after ${i} attempts"
    break
  fi
  sleep 1
done

# ─── Post-restore checks ──────────────────────────────────────────────────────

sleep 5
POST_ERRORS=$(curl -sf "${API_BASE}/metrics" 2>/dev/null | grep 'echoline_http_errors_total' | awk '{print $2}' || echo "0")

log "=== EXPERIMENT RESULTS ==="
log "Duration: ${DURATION}s"
log "Health check passes: ${PASS_COUNT}"
log "Health check failures: ${FAIL_COUNT}"
log "HTTP errors before: ${BASELINE_ERRORS}"
log "HTTP errors after: ${POST_ERRORS}"

if [ "${FAIL_COUNT}" -eq 0 ]; then
  ok "RESULT: PASS — API remained operational during Redis outage"
else
  warn "RESULT: PARTIAL — ${FAIL_COUNT} health check failures during Redis outage"
fi

log ""
log "Next steps:"
log "  1. Check outbox table: SELECT COUNT(*) FROM outbox WHERE status='pending';"
log "  2. Verify outbox drains after Redis restore"
log "  3. Run smoke test: make smoke"
log "  4. Document results in docs/chaos-playbook.md under CHAOS-001"
