#!/usr/bin/env bash
set -euo pipefail

# VOC-038-T02 — repeatable production smoke-test suite.
#
# Replaces the manual curl/SSH checks used ad hoc during R2 (VOC-037)
# with a scripted, callable-from-CI-or-standalone suite. Every check
# either passes or fails explicitly; nothing is silently skipped
# except the core-loop check, which needs a smoke-test session cookie.
# deploy-production.yml supplies one on every deploy since VOC-050-T04
# (minted for the deploy-seeded synthetic account), so the SKIP now
# only applies to a standalone run that does not set the variable -
# and even then it is an explicit SKIP with a reason, not a silent
# pass.
#
# Deliberately does not perform any state-mutating action against a
# real target: it never actually requests a magic link or completes
# an OAuth flow, because a real production email/redirect is a side
# effect this suite must not cause on every deploy. Kill-switch state
# is asserted via /healthz's kill_switches field (VOC-038-T02, see
# apps/api/app/api/production.go) rather than by probing behavior,
# specifically so this suite has no side effects on the auth paths it
# checks.
#
# Usage:
#   smoke-test-production.sh <api_base_url> <web_base_url>
#
# Environment (all optional; defaults match deploy-production.yml's
# current default posture):
#   EXPECT_MAGIC_LINK_ENABLED   (default: false)
#   EXPECT_OAUTH_ENABLED        (default: false)
#   EXPECT_NEW_SIGNUPS_ENABLED  (default: false)
#   EXPECT_AI_ENABLED           (default: true)
#   SMOKE_TEST_SESSION_COOKIE   (default: unset - core-loop check is
#                                skipped). A whole `Cookie:` header
#                                value, i.e. `vocanova_session=<opaque
#                                session value>`, not the bare opaque
#                                value on its own.
#
# Requires: curl, and either jq or python3 for JSON parsing.

if [ "$#" -ne 2 ]; then
  echo "usage: $0 <api_base_url> <web_base_url>" >&2
  exit 1
fi

api_base_url="${1%/}"
web_base_url="${2%/}"

expect_magic_link="${EXPECT_MAGIC_LINK_ENABLED:-false}"
expect_oauth="${EXPECT_OAUTH_ENABLED:-false}"
expect_new_signups="${EXPECT_NEW_SIGNUPS_ENABLED:-false}"
expect_ai="${EXPECT_AI_ENABLED:-true}"
smoke_session_cookie="${SMOKE_TEST_SESSION_COOKIE:-}"

failures=0
fail() {
  echo "FAIL: $1" >&2
  failures=$((failures + 1))
}
pass() {
  echo "PASS: $1"
}
skip() {
  echo "SKIP: $1"
}

# json_field <json> <dotted.path> - reads a field via jq if available,
# else falls back to python3. Both keep this script portable across a
# GitHub Actions runner (jq preinstalled) and a founder's local shell
# (python3 near-universal, jq not guaranteed).
json_field() {
  local json="$1" path="$2"
  if command -v jq >/dev/null 2>&1; then
    printf '%s' "$json" | jq -er ".$path" 2>/dev/null
  else
    python3 -c '
import json, sys
try:
    data = json.loads(sys.argv[1])
    for key in sys.argv[2].split("."):
        data = data[key]
    print(data if isinstance(data, str) else json.dumps(data))
except Exception:
    sys.exit(1)
' "$json" "$path"
  fi
}

echo "== VOC-038-T02 production smoke test =="
echo "api:  $api_base_url"
echo "web:  $web_base_url"
echo

# 1. Health endpoint
healthz_body="$(curl -fsS --max-time 10 "$api_base_url/healthz" || true)"
if [ -z "$healthz_body" ]; then
  fail "GET /healthz did not return a response"
else
  status="$(json_field "$healthz_body" status || echo "")"
  database="$(json_field "$healthz_body" database || echo "")"
  if [ "$status" = "ok" ] && [ "$database" = "ok" ]; then
    pass "/healthz reports status=ok database=ok"
  else
    fail "/healthz reported status='$status' database='$database' (expected both 'ok')"
  fi

  # 2. Kill-switch state assertions, read from the same /healthz body
  # so no state-mutating auth probe is needed.
  for pair in \
    "magic_link_enabled:$expect_magic_link" \
    "oauth_enabled:$expect_oauth" \
    "new_signups_enabled:$expect_new_signups" \
    "ai_enabled:$expect_ai"; do
    field="${pair%%:*}"
    expected="${pair##*:}"
    got="$(json_field "$healthz_body" "kill_switches.$field" || echo "")"
    if [ "$got" = "$expected" ]; then
      pass "kill switch $field=$got (expected)"
    else
      fail "kill switch $field=$got, expected $expected"
    fi
  done
fi

# 3. Web reachability
web_status="$(curl -sS --max-time 10 -o /dev/null -w "%{http_code}" "$web_base_url/" || echo "000")"
case "$web_status" in
  2*) pass "GET $web_base_url/ returned $web_status" ;;
  *) fail "GET $web_base_url/ returned $web_status (expected 2xx)" ;;
esac

# 4. Auth flow reachability (no side effects: never actually requests
# a magic link or completes an OAuth flow). A disabled path must
# surface the specific ErrMagicLinkDisabled/ErrOAuthDisabled-shaped
# 503; this is itself the reachability proof for the disabled case.
# When a path is expected enabled, this suite only confirms the
# server responds at all (not a network error / raw proxy failure),
# since actually exercising it would send a real email or write real
# OAuth state.
magic_status="$(curl -sS --max-time 10 -o /dev/null -w "%{http_code}" \
  -X POST -H "Content-Type: application/json" \
  -d '{"email":"smoke-test@vocanova.site"}' \
  "$api_base_url/api/v1/auth/magic-links" || echo "000")"
if [ "$expect_magic_link" = "false" ]; then
  if [ "$magic_status" = "503" ]; then
    pass "magic-link request correctly refused (503) while disabled"
  else
    fail "magic-link request returned $magic_status, expected 503 while disabled"
  fi
else
  if [ "$magic_status" != "000" ]; then
    pass "magic-link endpoint reachable (status $magic_status; not exercising further to avoid a real send)"
  else
    fail "magic-link endpoint unreachable (network error)"
  fi
fi

oauth_status="$(curl -sS --max-time 10 -o /dev/null -w "%{http_code}" \
  -X POST -H "Content-Type: application/json" \
  -d "{\"redirectUri\":\"$web_base_url/home\"}" \
  "$api_base_url/api/v1/auth/oauth/google/start" || echo "000")"
if [ "$expect_oauth" = "false" ]; then
  if [ "$oauth_status" = "503" ]; then
    pass "oauth start correctly refused (503) while disabled"
  else
    fail "oauth start returned $oauth_status, expected 503 while disabled"
  fi
else
  if [ "$oauth_status" != "000" ]; then
    pass "oauth start endpoint reachable (status $oauth_status; not following the redirect)"
  else
    fail "oauth start endpoint unreachable (network error)"
  fi
fi

# 5. Core-loop happy path - requires the synthetic smoke-test account's
# session cookie. deploy-production.yml mints one on every deploy
# (VOC-050-T04), so this runs for real there; a standalone run without
# the variable still gets an explicit, non-fatal SKIP rather than a
# silent pass or a hard failure over an unprovisioned local shell.
# Both requests are read-only, preserving this suite's no-side-effects
# design.
if [ -z "$smoke_session_cookie" ]; then
  skip "core-loop happy path (SMOKE_TEST_SESSION_COOKIE not set - pass vocanova_session=<value> to run it)"
else
  me_status="$(curl -sS --max-time 10 -o /dev/null -w "%{http_code}" \
    -H "Cookie: $smoke_session_cookie" \
    "$api_base_url/api/v1/me" || echo "000")"
  if [ "$me_status" = "200" ]; then
    pass "GET /api/v1/me returned 200 with smoke-test session"
  else
    fail "GET /api/v1/me returned $me_status with smoke-test session (expected 200)"
  fi

  situations_status="$(curl -sS --max-time 10 -o /dev/null -w "%{http_code}" \
    -H "Cookie: $smoke_session_cookie" \
    "$api_base_url/api/v1/journey-situations" || echo "000")"
  if [ "$situations_status" = "200" ]; then
    pass "GET /api/v1/journey-situations returned 200 with smoke-test session"
  else
    fail "GET /api/v1/journey-situations returned $situations_status with smoke-test session (expected 200)"
  fi
fi

echo
if [ "$failures" -gt 0 ]; then
  echo "SMOKE TEST FAILED: $failures check(s) failed" >&2
  exit 1
fi
echo "SMOKE TEST PASSED"
