#!/usr/bin/env bash
set -euo pipefail

# VOC-081-T04 — external HTTPS, WebSocket, and access-policy verification for
# monitor.vocanova.site after repository deploy convergence (VOC-081-T03).
#
# Confirms:
#   1. Four staging/production app hostnames remain healthy on ordinary :443.
#   2. https://monitor.vocanova.site/ serves the Uptime Kuma application (not
#      Cloudflare/origin 520/502). Kuma redirects `/` → `/dashboard` (SPA).
#   3. The socket.io path works: Engine.IO polling handshake returns 2xx with
#      a session id (WebSocket upgrade is advertised). Bare websocket without
#      a sid may return 400 — that is not treated as origin failure.
#   4. Unauthenticated access does not expose an authenticated admin API
#      payload; the SPA shell / login boundary matches VOC-081-DEP-00.
#
# Usage:
#   verify-voc081-monitor.sh
#   verify-voc081-monitor.sh --skip-apps
#   MONITOR_BASE_URL=http://127.0.0.1:PORT verify-voc081-monitor.sh
#
# Requires: curl, mktemp. Optional: infra/scripts/verify-voc067-cutover.sh

repo_root="$(cd "$(dirname "$0")/../.." && pwd)"
cutover_script="$repo_root/infra/scripts/verify-voc067-cutover.sh"

skip_apps=false
if [ "${1:-}" = "--skip-apps" ]; then
  skip_apps=true
elif [ -n "${1:-}" ]; then
  echo "usage: $0 [--skip-apps]" >&2
  exit 1
fi

monitor_base="${MONITOR_BASE_URL:-https://monitor.vocanova.site}"
monitor_origin="${monitor_base%/}"
monitor_socket_poll="${monitor_origin}/socket.io/?EIO=4&transport=polling"
monitor_socket_ws="${monitor_origin}/socket.io/?EIO=4&transport=websocket"

failures=0
record_fail() {
  echo "FAIL: $*" >&2
  failures=$((failures + 1))
}

record_pass() {
  echo "PASS: $*"
}

echo "VOC-081 monitor verification — external checks (via Cloudflare unless MONITOR_BASE_URL is set)"

if [ "$skip_apps" = false ]; then
  if [ -x "$cutover_script" ]; then
    echo ""
    echo "App tier :443 regression guard (verify-voc067-cutover.sh)"
    if ! "$cutover_script"; then
      record_fail "one or more app-tier :443 checks failed"
    fi
  else
    record_fail "missing executable $cutover_script"
  fi
fi

echo ""
echo "Monitor hostname checks ($monitor_origin)"

monitor_body="$(mktemp)"
entry_body="$(mktemp)"
poll_body="$(mktemp)"
trap 'rm -f "$monitor_body" "$entry_body" "$poll_body"' EXIT

# Follow redirects: Kuma serves `/` as 302 → `/dashboard` SPA shell.
monitor_status="$(
  curl -sS -L --max-time 25 -o "$monitor_body" -w "%{http_code}" "$monitor_origin/" || echo "000"
)"

if [[ "$monitor_status" =~ ^2 ]]; then
  record_pass "monitor web ($monitor_origin/ → final) -> HTTP $monitor_status"
else
  record_fail "monitor web ($monitor_origin/) -> HTTP $monitor_status (expected 2xx after redirects; 520/502 indicate edge/upstream failure)"
fi

if grep -qi 'uptime kuma' "$monitor_body"; then
  record_pass "monitor body includes Uptime Kuma marker"
else
  record_fail "monitor body missing Uptime Kuma marker (body may be an error page)"
fi

# DEP-00: proxied DNS is not authorization. Prefer an explicit login form when
# present; otherwise accept the Kuma SPA shell plus a non-authenticated
# entry-page payload (entryPage null / setup complete without session data).
entry_status="$(
  curl -sS --max-time 25 -o "$entry_body" -w "%{http_code}" "$monitor_origin/api/entry-page" || echo "000"
)"
if grep -Eqi 'type=["'\'']password["'\'']|name=["'\'']password["'\'']|>Login<|sign in' "$monitor_body"; then
  record_pass "unauthenticated response exposes a login surface (VOC-081-DEP-00)"
elif [[ "$entry_status" =~ ^2 ]] && grep -Eqi '"entryPage"|needSetup|setup' "$entry_body"; then
  if grep -Eqi '"token"|"jwt"|authenticated.: *true|"username"' "$entry_body"; then
    record_fail "unauthenticated /api/entry-page looks like an authenticated session"
  else
    record_pass "unauthenticated SPA/entry-page boundary present (VOC-081-DEP-00; Kuma auth retained)"
  fi
elif grep -Eqi 'Add New Monitor' "$monitor_body" && ! grep -qi 'uptime kuma' "$monitor_body"; then
  record_fail "unauthenticated response looks like an administrative dashboard (anonymous access)"
else
  # SPA-only shell with Uptime Kuma title is the normal unauthenticated HTML
  if grep -qi 'uptime kuma' "$monitor_body"; then
    record_pass "unauthenticated Kuma SPA shell served without session HTML (VOC-081-DEP-00)"
  else
    record_fail "could not confirm Kuma login/SPA boundary from unauthenticated response"
  fi
fi

poll_status="$(
  curl -sS --max-time 25 -o "$poll_body" -w "%{http_code}" "$monitor_socket_poll" || echo "000"
)"
if [[ "$poll_status" =~ ^2 ]] && grep -Eq '"sid"|upgrades' "$poll_body"; then
  record_pass "monitor socket.io polling handshake ($monitor_socket_poll) -> HTTP $poll_status (sid present; websocket upgrade advertised)"
elif [ "$poll_status" = "101" ]; then
  record_pass "monitor socket.io upgrade -> HTTP 101"
else
  # Fallback: direct websocket probe (may be 101, or 400 without sid — 400
  # after a failed poll is still not proof of health).
  ws_status="$(
    curl -sS --max-time 25 -o /dev/null -w "%{http_code}" \
      -H 'Connection: Upgrade' \
      -H 'Upgrade: websocket' \
      -H 'Sec-WebSocket-Version: 13' \
      -H 'Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==' \
      "$monitor_socket_ws" 2>/dev/null || true
  )"
  ws_status="${ws_status:-000}"
  if [ "$ws_status" = "101" ] || [[ "$ws_status" =~ ^2 ]]; then
    record_pass "monitor socket.io probe ($monitor_socket_ws) -> HTTP $ws_status"
  else
    record_fail "monitor socket.io polling ($monitor_socket_poll) -> HTTP $poll_status (expected 2xx with sid; got ws=$ws_status)"
  fi
fi

echo ""
if [ "$failures" -eq 0 ]; then
  echo "All VOC-081 monitor verification checks passed."
  exit 0
fi

echo "ERROR: $failures monitor verification check(s) failed." >&2
exit 1
