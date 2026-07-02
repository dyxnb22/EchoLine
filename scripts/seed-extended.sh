#!/usr/bin/env bash
# scripts/seed-extended.sh
# Extended seed: 5 users, 1 group, 1 channel, sample messages.
# Requires a running EchoLine API at BASE_URL.
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@echoline.dev}"
ADMIN_PASS="${ADMIN_PASS:-changeme123}"

log()  { echo "[seed-extended] $*"; }
die()  { echo "[seed-extended] ERROR: $*" >&2; exit 1; }

# ── helpers ────────────────────────────────────────────────────────────────
register_user() {
  local username="$1" email="$2" password="$3"
  local resp
  resp=$(curl -sf -X POST "$BASE_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$username\",\"email\":\"$email\",\"password\":\"$password\"}" \
    || true)
  echo "$resp"
}

login_user() {
  local email="$1" password="$2"
  local token
  token=$(curl -sf -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$email\",\"password\":\"$password\"}" \
    | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('access_token',''))")
  echo "$token"
}

auth_header() { echo "Authorization: Bearer $1"; }

create_conversation() {
  local token="$1" peer_id="$2"
  curl -sf -X POST "$BASE_URL/conversations" \
    -H "Content-Type: application/json" \
    -H "$(auth_header "$token")" \
    -d "{\"peer_id\":\"$peer_id\",\"type\":\"direct\"}" \
    | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('id',''))"
}

send_message() {
  local token="$1" conv_id="$2" text="$3" client_id="$4"
  curl -sf -X POST "$BASE_URL/conversations/$conv_id/messages" \
    -H "Content-Type: application/json" \
    -H "$(auth_header "$token")" \
    -d "{\"text\":\"$text\",\"client_msg_id\":\"$client_id\"}" > /dev/null
}

create_group() {
  local token="$1" name="$2"
  curl -sf -X POST "$BASE_URL/groups" \
    -H "Content-Type: application/json" \
    -H "$(auth_header "$token")" \
    -d "{\"name\":\"$name\"}" \
    | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('id',''))"
}

invite_to_group() {
  local token="$1" group_id="$2" user_id="$3"
  curl -sf -X POST "$BASE_URL/groups/$group_id/members" \
    -H "Content-Type: application/json" \
    -H "$(auth_header "$token")" \
    -d "{\"user_id\":\"$user_id\"}" > /dev/null || true
}

create_channel() {
  local token="$1" name="$2"
  curl -sf -X POST "$BASE_URL/channels" \
    -H "Content-Type: application/json" \
    -H "$(auth_header "$token")" \
    -d "{\"name\":\"$name\",\"description\":\"Seeded channel\"}" \
    | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('id',''))"
}

# ── check API liveness ──────────────────────────────────────────────────────
log "Checking API health at $BASE_URL/health …"
curl -sf "$BASE_URL/health" > /dev/null \
  || die "API not reachable. Start with: make api-run"

# ── create 5 users ──────────────────────────────────────────────────────────
USERS=("alice" "bob" "carol" "dave" "eve")
EMAILS=("alice@echoline.dev" "bob@echoline.dev" "carol@echoline.dev" "dave@echoline.dev" "eve@echoline.dev")
PASS="Seed1234!"
declare -A TOKENS
declare -A USER_IDS

for i in "${!USERS[@]}"; do
  u="${USERS[$i]}"
  e="${EMAILS[$i]}"
  log "Registering $u ($e) …"
  resp=$(register_user "$u" "$e" "$PASS")
  uid=$(echo "$resp" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('user',{}).get('id',''))" 2>/dev/null || echo "")
  if [[ -z "$uid" ]]; then
    log "  $u may already exist; logging in …"
  fi
  tok=$(login_user "$e" "$PASS")
  if [[ -z "$tok" ]]; then
    die "Failed to get token for $u"
  fi
  TOKENS[$u]="$tok"
  uid=$(curl -sf "$BASE_URL/users/me" \
    -H "$(auth_header "$tok")" \
    | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('id',''))" 2>/dev/null || echo "")
  USER_IDS[$u]="$uid"
  log "  $u → id=${uid:-unknown}"
done

# ── direct conversations + messages ─────────────────────────────────────────
log "Creating alice↔bob direct conversation …"
conv_ab=$(create_conversation "${TOKENS[alice]}" "${USER_IDS[bob]:-}")
if [[ -n "$conv_ab" ]]; then
  send_message "${TOKENS[alice]}" "$conv_ab" "Hey Bob, this is a seeded message!" "seed-ab-01"
  send_message "${TOKENS[bob]}"   "$conv_ab" "Hi Alice! Got your message." "seed-ab-02"
  send_message "${TOKENS[alice]}" "$conv_ab" "Great, seed script is working." "seed-ab-03"
  log "  Sent 3 messages in conv $conv_ab"
fi

log "Creating carol↔dave direct conversation …"
conv_cd=$(create_conversation "${TOKENS[carol]}" "${USER_IDS[dave]:-}")
if [[ -n "$conv_cd" ]]; then
  send_message "${TOKENS[carol]}" "$conv_cd" "Carol to Dave: seeded." "seed-cd-01"
  send_message "${TOKENS[dave]}"  "$conv_cd" "Dave to Carol: received." "seed-cd-02"
  log "  Sent 2 messages in conv $conv_cd"
fi

# ── group ────────────────────────────────────────────────────────────────────
log "Creating group 'seed-group' owned by alice …"
group_id=$(create_group "${TOKENS[alice]}" "seed-group")
if [[ -n "$group_id" ]]; then
  for u in bob carol dave eve; do
    log "  Inviting $u to group $group_id …"
    invite_to_group "${TOKENS[alice]}" "$group_id" "${USER_IDS[$u]:-}"
  done
  log "  Group $group_id created with 5 members"
else
  log "  Warning: group creation returned no ID (API may not be running)"
fi

# ── channel ──────────────────────────────────────────────────────────────────
log "Creating channel 'announcements' owned by alice …"
chan_id=$(create_channel "${TOKENS[alice]}" "announcements")
if [[ -n "$chan_id" ]]; then
  log "  Channel $chan_id created"
else
  log "  Warning: channel creation returned no ID (API may not be running)"
fi

# ── summary ──────────────────────────────────────────────────────────────────
log ""
log "Done. Seeded:"
log "  Users  : ${USERS[*]}"
log "  Group  : seed-group (id=$group_id)"
log "  Channel: announcements (id=$chan_id)"
log "  Convs  : alice↔bob ($conv_ab), carol↔dave ($conv_cd)"
log ""
log "Login with any user at $BASE_URL"
log "  Email   : alice@echoline.dev"
log "  Password: $PASS"
