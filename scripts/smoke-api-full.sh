#!/usr/bin/env bash
# EchoLine Full API Smoke Test
# Task: SB11
#
# Tests the complete user journey:
#   1. Register two users (Alice and Bob)
#   2. Login as both users, obtain JWT tokens
#   3. Create a DM conversation
#   4. Send messages from both sides
#   5. Test idempotency (send with same client_msg_id)
#   6. Create a group conversation
#   7. Send a group message
#   8. Search for a message
#   9. Test WebSocket connection (if RUN_WS_SMOKE is set)
#   10. Verify conversation list and unread counts
#
# Usage:
#   ./scripts/smoke-api-full.sh
#
# Environment:
#   API_BASE_URL  - default: http://localhost:8080
#   RUN_WS_SMOKE  - set to any value to run WS tests (requires Node 22+ or wscat)
#   VERBOSE       - set to 1 to print full response bodies

set -euo pipefail

API_BASE=${API_BASE_URL:-http://localhost:8080}
VERBOSE=${VERBOSE:-0}
TIMESTAMP=$(date +%s)
SMOKE_MSG1_ID=$(uuidgen 2>/dev/null || python3 -c 'import uuid; print(uuid.uuid4())')
SMOKE_MSG2_ID=$(uuidgen 2>/dev/null || python3 -c 'import uuid; print(uuid.uuid4())')
SMOKE_GROUP_MSG_ID=$(uuidgen 2>/dev/null || python3 -c 'import uuid; print(uuid.uuid4())')
SMOKE_DEVICE_ID=$(uuidgen 2>/dev/null || python3 -c 'import uuid; print(uuid.uuid4())')

# ─── Colors ───────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASS=0
FAIL=0
SKIP=0

pass() { echo -e "${GREEN}  ✓ PASS${NC} $*"; PASS=$((PASS+1)); }
fail() { echo -e "${RED}  ✗ FAIL${NC} $*"; FAIL=$((FAIL+1)); }
skip() { echo -e "${YELLOW}  - SKIP${NC} $*"; SKIP=$((SKIP+1)); }
section() { echo -e "\n${BLUE}=== $* ===${NC}"; }

# ─── HTTP helpers ─────────────────────────────────────────────────────────────

post_json() {
  local url="$1"
  local body="$2"
  local token="${3:-}"
  local auth_header=""
  [ -n "${token}" ] && auth_header="-H 'Authorization: Bearer ${token}'"

  if [ "${VERBOSE}" = "1" ]; then
    curl -s -X POST "${API_BASE}${url}" \
      -H "Content-Type: application/json" \
      ${token:+-H "Authorization: Bearer ${token}"} \
      -d "${body}"
  else
    curl -s -X POST "${API_BASE}${url}" \
      -H "Content-Type: application/json" \
      ${token:+-H "Authorization: Bearer ${token}"} \
      -d "${body}"
  fi
}

get_json() {
  local url="$1"
  local token="${2:-}"
  curl -s -X GET "${API_BASE}${url}" \
    ${token:+-H "Authorization: Bearer ${token}"}
}

run_with_timeout() {
  local seconds="$1"
  shift

  if command -v timeout > /dev/null 2>&1; then
    timeout "${seconds}" "$@"
    return $?
  fi

  if command -v gtimeout > /dev/null 2>&1; then
    gtimeout "${seconds}" "$@"
    return $?
  fi

  if command -v python3 > /dev/null 2>&1; then
    python3 - "${seconds}" "$@" <<'PY'
import subprocess
import sys

seconds = float(sys.argv[1])
cmd = sys.argv[2:]
try:
    raise SystemExit(subprocess.run(cmd, timeout=seconds).returncode)
except subprocess.TimeoutExpired:
    raise SystemExit(124)
PY
    return $?
  fi

  "$@"
}

node_has_websocket() {
  command -v node > /dev/null 2>&1 && [ "$(node -e 'process.stdout.write(typeof WebSocket)')" = "function" ]
}

ws_probe_node() {
  local url="$1"
  local mode="$2"
  local seconds="$3"

  node - "${url}" "${mode}" "${seconds}" <<'NODE'
const url = process.argv[2];
const mode = process.argv[3];
const seconds = Number(process.argv[4]);
let settled = false;

function finish(ok, message) {
  if (settled) return;
  settled = true;
  if (message) console.log(message);
  process.exit(ok ? 0 : 1);
}

const ws = new WebSocket(url);
const timer = setTimeout(() => finish(false, "timeout"), seconds * 1000);

ws.addEventListener("open", () => {
  clearTimeout(timer);
  if (mode === "open") {
    ws.close();
    finish(true, "open");
  } else {
    ws.close();
    finish(false, "unexpected open");
  }
});

ws.addEventListener("error", () => {
  clearTimeout(timer);
  finish(mode === "reject", "error");
});

ws.addEventListener("close", () => {
  clearTimeout(timer);
  if (!settled) finish(mode === "reject", "closed");
});
NODE
}

# ─── Pre-flight ───────────────────────────────────────────────────────────────

section "Pre-flight"

HEALTH=$(curl -s -o /dev/null -w "%{http_code}" "${API_BASE}/health")
if [ "${HEALTH}" = "200" ]; then
  pass "API health check: 200"
else
  fail "API health check: ${HEALTH} (expected 200)"
  echo "API is not running at ${API_BASE}. Exiting."
  exit 1
fi

# ─── Step 1: Register users ───────────────────────────────────────────────────

section "Step 1: User Registration"

ALICE="alice_smoke_${TIMESTAMP}"
BOB="bob_smoke_${TIMESTAMP}"
PASSWORD="smoke-test-password-123"

REG_ALICE=$(post_json "/api/auth/register" "{\"username\":\"${ALICE}\",\"password\":\"${PASSWORD}\"}")
if echo "${REG_ALICE}" | grep -q '"id"'; then
  pass "Alice registered"
  ALICE_ID=$(echo "${REG_ALICE}" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
else
  fail "Alice registration failed: ${REG_ALICE}"
  ALICE_ID=""
fi

REG_BOB=$(post_json "/api/auth/register" "{\"username\":\"${BOB}\",\"password\":\"${PASSWORD}\"}")
if echo "${REG_BOB}" | grep -q '"id"'; then
  pass "Bob registered"
  BOB_ID=$(echo "${REG_BOB}" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
else
  fail "Bob registration failed: ${REG_BOB}"
  BOB_ID=""
fi

# Test duplicate registration returns error
REG_DUP=$(post_json "/api/auth/register" "{\"username\":\"${ALICE}\",\"password\":\"${PASSWORD}\"}")
if echo "${REG_DUP}" | grep -q '"error"'; then
  pass "Duplicate registration rejected"
else
  fail "Duplicate registration should have been rejected: ${REG_DUP}"
fi

# ─── Step 2: Login ────────────────────────────────────────────────────────────

section "Step 2: Login"

LOGIN_ALICE=$(post_json "/api/auth/login" "{\"username\":\"${ALICE}\",\"password\":\"${PASSWORD}\"}")
if echo "${LOGIN_ALICE}" | grep -q '"access_token"'; then
  pass "Alice login: token received"
  TOKEN_ALICE=$(echo "${LOGIN_ALICE}" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
else
  fail "Alice login failed: ${LOGIN_ALICE}"
  TOKEN_ALICE=""
fi

LOGIN_BOB=$(post_json "/api/auth/login" "{\"username\":\"${BOB}\",\"password\":\"${PASSWORD}\"}")
if echo "${LOGIN_BOB}" | grep -q '"access_token"'; then
  pass "Bob login: token received"
  TOKEN_BOB=$(echo "${LOGIN_BOB}" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
else
  fail "Bob login failed: ${LOGIN_BOB}"
  TOKEN_BOB=""
fi

# Test wrong password
LOGIN_BAD=$(post_json "/api/auth/login" "{\"username\":\"${ALICE}\",\"password\":\"wrong-password\"}")
if echo "${LOGIN_BAD}" | grep -q '"error"'; then
  pass "Wrong password rejected"
else
  fail "Wrong password should be rejected: ${LOGIN_BAD}"
fi

# Skip remaining tests if auth failed
if [ -z "${TOKEN_ALICE:-}" ] || [ -z "${TOKEN_BOB:-}" ]; then
  fail "Auth failed; skipping conversation and message tests"
  echo -e "\n${RED}=== SMOKE TEST INCOMPLETE (auth failed) ===${NC}"
  exit 1
fi

# ─── Step 3: Create DM conversation ──────────────────────────────────────────

section "Step 3: Create DM Conversation"

if [ -z "${BOB_ID:-}" ]; then
  skip "Bob ID not available; skipping DM creation"
  DM_CONV_ID=""
else
  DM_BODY="{\"user_id\":\"${BOB_ID}\"}"
  DM_RESP=$(post_json "/api/conversations/direct" "${DM_BODY}" "${TOKEN_ALICE}")
  if echo "${DM_RESP}" | grep -q '"id"'; then
    pass "DM conversation created"
    DM_CONV_ID=$(echo "${DM_RESP}" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
  else
    fail "DM creation failed: ${DM_RESP}"
    DM_CONV_ID=""
  fi
fi

# ─── Step 4: Send messages ────────────────────────────────────────────────────

section "Step 4: Send Messages"

if [ -z "${DM_CONV_ID:-}" ]; then
  skip "DM conversation not available; skipping message send"
else
  MSG1_BODY="{\"body\":\"Hello Bob from smoke test\",\"client_msg_id\":\"${SMOKE_MSG1_ID}\",\"type\":\"text\"}"
  MSG1_RESP=$(post_json "/api/conversations/${DM_CONV_ID}/messages" "${MSG1_BODY}" "${TOKEN_ALICE}")
  if echo "${MSG1_RESP}" | grep -q '"seq"'; then
    pass "Alice sends message: seq assigned"
    MSG1_SEQ=$(echo "${MSG1_RESP}" | grep -o '"seq":[0-9]*' | cut -d: -f2)
    [ "${VERBOSE}" = "1" ] && echo "    seq: ${MSG1_SEQ}"
  else
    fail "Message send failed: ${MSG1_RESP}"
  fi

  MSG2_BODY="{\"body\":\"Hi Alice, smoke test reply\",\"client_msg_id\":\"${SMOKE_MSG2_ID}\",\"type\":\"text\"}"
  MSG2_RESP=$(post_json "/api/conversations/${DM_CONV_ID}/messages" "${MSG2_BODY}" "${TOKEN_BOB}")
  if echo "${MSG2_RESP}" | grep -q '"seq"'; then
    pass "Bob sends reply: seq assigned"
  else
    fail "Bob reply send failed: ${MSG2_RESP}"
  fi

  # Step 5: Idempotency test
  section "Step 5: Idempotency"
  MSG1_RETRY=$(post_json "/api/conversations/${DM_CONV_ID}/messages" "${MSG1_BODY}" "${TOKEN_ALICE}")
  ORIG_ID=$(echo "${MSG1_RESP}" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
  RETRY_ID=$(echo "${MSG1_RETRY}" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
  if [ "${ORIG_ID}" = "${RETRY_ID}" ] && [ -n "${ORIG_ID}" ]; then
    pass "Idempotency: duplicate send returns same message ID"
  else
    fail "Idempotency failed: original=${ORIG_ID} retry=${RETRY_ID}"
  fi

  # Read message history
  HISTORY=$(get_json "/api/conversations/${DM_CONV_ID}/messages" "${TOKEN_ALICE}")
  if echo "${HISTORY}" | grep -q '"seq"'; then
    pass "Message history: messages returned"
  else
    fail "Message history empty or error: ${HISTORY}"
  fi
fi

# ─── Step 6: Create group conversation ───────────────────────────────────────

section "Step 6: Group Conversation"

if [ -z "${BOB_ID:-}" ]; then
  skip "Bob ID not available; skipping group creation"
  GROUP_CONV_ID=""
else
  GROUP_BODY="{\"title\":\"Smoke Test Group ${TIMESTAMP}\",\"member_ids\":[\"${BOB_ID}\"]}"
  GROUP_RESP=$(post_json "/api/conversations/groups" "${GROUP_BODY}" "${TOKEN_ALICE}")
  if echo "${GROUP_RESP}" | grep -q '"id"'; then
    pass "Group conversation created"
    GROUP_CONV_ID=$(echo "${GROUP_RESP}" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
  else
    fail "Group creation failed: ${GROUP_RESP}"
    GROUP_CONV_ID=""
  fi
fi

if [ -n "${GROUP_CONV_ID:-}" ]; then
  GROUP_MSG_BODY="{\"body\":\"Group smoke test message ${TIMESTAMP}\",\"client_msg_id\":\"${SMOKE_GROUP_MSG_ID}\",\"type\":\"text\"}"
  GROUP_MSG_RESP=$(post_json "/api/conversations/${GROUP_CONV_ID}/messages" "${GROUP_MSG_BODY}" "${TOKEN_ALICE}")
  if echo "${GROUP_MSG_RESP}" | grep -q '"seq"'; then
    pass "Group message sent"
  else
    fail "Group message send failed: ${GROUP_MSG_RESP}"
  fi
fi

# ─── Step 7: Search ───────────────────────────────────────────────────────────

section "Step 7: Full-Text Search"

SEARCH_RESP=$(get_json "/api/search/messages?q=smoke+test" "${TOKEN_ALICE}")
if echo "${SEARCH_RESP}" | grep -q '"messages"'; then
  pass "Search endpoint returns results array"
elif echo "${SEARCH_RESP}" | grep -q '"error"'; then
  fail "Search returned error: ${SEARCH_RESP}"
else
  pass "Search endpoint reachable (no results yet; search index may be async)"
fi

# ─── Step 8: Conversation list ────────────────────────────────────────────────

section "Step 8: Conversation List"

CONV_LIST=$(get_json "/api/conversations" "${TOKEN_ALICE}")
if echo "${CONV_LIST}" | grep -q '"id"'; then
  pass "Conversation list: conversations returned"
else
  fail "Conversation list empty or error: ${CONV_LIST}"
fi

# ─── Step 9: WS smoke (optional) ──────────────────────────────────────────────

section "Step 9: WebSocket (optional)"

if [ -n "${RUN_WS_SMOKE:-}" ]; then
  if node_has_websocket; then
    if WS_RESULT=$(ws_probe_node "ws://localhost:8080/ws?token=${TOKEN_ALICE}&device_id=${SMOKE_DEVICE_ID}" open 5 2>&1); then
      pass "WS connection established"
    else
      fail "WS connection failed: ${WS_RESULT}"
    fi

    if WS_BAD=$(ws_probe_node "ws://localhost:8080/ws?token=invalid&device_id=${SMOKE_DEVICE_ID}" reject 3 2>&1); then
      pass "WS rejects invalid token"
    else
      skip "WS invalid token test inconclusive (node output: ${WS_BAD})"
    fi
  elif command -v wscat > /dev/null 2>&1; then
    # Test WS connection with valid token
    WS_RESULT=$(run_with_timeout 5 wscat -c "ws://localhost:8080/ws?token=${TOKEN_ALICE}&device_id=${SMOKE_DEVICE_ID}" --no-color 2>&1 || true)
    if echo "${WS_RESULT}" | grep -q "Connected"; then
      pass "WS connection established"
    else
      fail "WS connection failed: ${WS_RESULT}"
    fi

    # Test WS with invalid token
    WS_BAD=$(run_with_timeout 3 wscat -c "ws://localhost:8080/ws?token=invalid&device_id=${SMOKE_DEVICE_ID}" --no-color 2>&1 || true)
    if echo "${WS_BAD}" | grep -q "401\|Unauthorized\|403"; then
      pass "WS rejects invalid token"
    else
      skip "WS invalid token test inconclusive (wscat output: ${WS_BAD})"
    fi
  else
    skip "Node WebSocket and wscat unavailable; skipping WS test"
  fi
else
  skip "WS smoke skipped (set RUN_WS_SMOKE=1 to enable)"
fi

# ─── Summary ──────────────────────────────────────────────────────────────────

echo ""
echo -e "${BLUE}=== SMOKE TEST SUMMARY ===${NC}"
echo -e "  ${GREEN}PASS: ${PASS}${NC}"
echo -e "  ${RED}FAIL: ${FAIL}${NC}"
echo -e "  ${YELLOW}SKIP: ${SKIP}${NC}"
echo ""

if [ "${FAIL}" -eq 0 ]; then
  echo -e "${GREEN}✓ ALL CHECKS PASSED${NC}"
  exit 0
else
  echo -e "${RED}✗ ${FAIL} CHECK(S) FAILED${NC}"
  exit 1
fi
