#!/usr/bin/env bash
set -euo pipefail

# VOC-038-T02 — repeatable production smoke-test suite.
# VOC-085-T01 — content-aware journey-situations checks and detail API
# verification (fail closed on HTTP 200 with an empty list).
# VOC-085-T02 — authenticated non-mutating learning-route sweep (ten fixed
# web routes plus API-derived discover situation/word routes).
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
situation_slug=""
word_slug=""

# VOC-085-TEST-06 fixed route inventory (keep in sync with
# scripts/foundation/voc085-production-route-sweep.test.mjs).
PRODUCTION_PUBLIC_WEB_ROUTES=(
  "/"
  "/signin"
  "/auth/magic"
)
PRODUCTION_AUTHENTICATED_WEB_ROUTES=(
  "/onboarding"
  "/home"
  "/discover"
  "/reviews"
  "/progress"
  "/settings"
  "/settings/account"
)

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

# http_get_authed <url> <cookie> - performs one GET with the smoke cookie.
# Sets _http_last_status and _http_last_body.
_http_last_status=""
_http_last_body=""
http_get_authed() {
  local url="$1" cookie="$2"
  local tmp
  tmp="$(mktemp)"
  _http_last_status=""
  _http_last_body=""
  local status
  status="$(curl -sS --max-time 10 -o "$tmp" -w "%{http_code}" \
    -H "Cookie: $cookie" \
    "$url" 2>/dev/null || echo "000")"
  if [ -f "$tmp" ]; then
    _http_last_body="$(cat "$tmp")"
  fi
  rm -f "$tmp"
  _http_last_status="$status"
}

# web_get <path> <cookie> - GET a web route; sets _web_last_status and
# _web_last_location (when the server returns a redirect).
_web_last_status=""
_web_last_location=""
web_get() {
  local route_path="$1" cookie="$2"
  local url="${web_base_url}${route_path}"
  local tmp hdr
  tmp="$(mktemp)"
  hdr="$(mktemp)"
  _web_last_status=""
  _web_last_location=""
  local status
  if [ -n "$cookie" ]; then
    status="$(curl -sS --max-time 15 -o "$tmp" -D "$hdr" -w "%{http_code}" \
      -H "Cookie: $cookie" \
      "$url" 2>/dev/null || echo "000")"
  else
    status="$(curl -sS --max-time 15 -o "$tmp" -D "$hdr" -w "%{http_code}" \
      "$url" 2>/dev/null || echo "000")"
  fi
  if [ -f "$hdr" ]; then
    _web_last_location="$(
      grep -i '^location:' "$hdr" | head -1 | cut -d' ' -f2- | tr -d '\r' || true
    )"
  fi
  rm -f "$hdr" "$tmp"
  _web_last_status="$status"
}

# assert_web_route_reachable <label> <path> <cookie> [require_auth]
# Pass on 2xx. Pass on 3xx unless require_auth is true and Location targets
# /signin (unauthenticated redirect). Fails on 4xx/5xx/network errors.
assert_web_route_reachable() {
  local label="$1" route_path="$2" cookie="$3"
  local require_auth="${4:-false}"
  web_get "$route_path" "$cookie"
  local status="$_web_last_status"
  local location="$_web_last_location"
  case "$status" in
    2*)
      pass "$label returned HTTP $status"
      ;;
    3*)
      if [ "$require_auth" = "true" ] && echo "$location" | grep -q '/signin'; then
        fail "$label redirected to sign-in (HTTP $status Location: $location)"
      else
        pass "$label returned HTTP $status"
      fi
      ;;
    *)
      fail "$label returned HTTP $status (expected 2xx/3xx)"
      ;;
  esac
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

# json_jq <json> <filter> - small jq-shaped queries used by the core-loop
# content checks. The python3 fallback implements only the filters this
# script needs so local shells without jq still work.
json_jq() {
  local json="$1" filter="$2"
  if command -v jq >/dev/null 2>&1; then
    printf '%s' "$json" | jq -er "$filter" 2>/dev/null
  else
    python3 -c '
import json, sys
try:
    data = json.loads(sys.argv[1])
    f = sys.argv[2]
    if f == ".items | length":
        print(len(data.get("items") or []))
    elif f == ".items[0].slug // empty":
        items = data.get("items") or []
        print(items[0].get("slug") or "" if items else "")
    elif f == ".situation.slug // empty":
        print((data.get("situation") or {}).get("slug") or "")
    elif f == ".meanings | length":
        print(len(data.get("meanings") or []))
    elif f == ".meanings[0].wordSlug // empty":
        meanings = data.get("meanings") or []
        print(meanings[0].get("wordSlug") or "" if meanings else "")
    elif f == ".word.slug // empty":
        print((data.get("word") or {}).get("slug") or "")
    else:
        sys.exit(1)
except Exception:
    sys.exit(1)
' "$json" "$filter"
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

  http_get_authed "$api_base_url/api/v1/journey-situations" "$smoke_session_cookie"
  situations_status="$_http_last_status"
  situations_body="$_http_last_body"
  word_slug=""
  if [ "$situations_status" != "200" ]; then
    fail "GET /api/v1/journey-situations returned $situations_status with smoke-test session (expected 200)"
  elif [ -z "$situations_body" ]; then
    fail "GET /api/v1/journey-situations returned 200 with an empty body"
  else
    situations_count="$(json_jq "$situations_body" '.items | length' || echo "")"
    if [ -z "$situations_count" ] || [ "$situations_count" -lt 1 ] 2>/dev/null; then
      fail "GET /api/v1/journey-situations returned 200 but items list is empty (expected canonical P1 content)"
    else
      pass "GET /api/v1/journey-situations returned 200 with $situations_count situation(s)"
    fi

    situation_slug="$(json_jq "$situations_body" '.items[0].slug // empty' || echo "")"
    if [ -z "$situation_slug" ]; then
      fail "journey-situations response missing first situation slug"
    else
      http_get_authed \
        "$api_base_url/api/v1/journey-situations/$situation_slug" \
        "$smoke_session_cookie"
      situation_detail_status="$_http_last_status"
      situation_detail_body="$_http_last_body"
      if [ "$situation_detail_status" != "200" ]; then
        fail "GET /api/v1/journey-situations/$situation_slug returned $situation_detail_status (expected 200)"
      elif [ -z "$situation_detail_body" ]; then
        fail "GET /api/v1/journey-situations/$situation_slug returned 200 with an empty body"
      else
        detail_situation_slug="$(json_jq "$situation_detail_body" '.situation.slug // empty' || echo "")"
        meanings_count="$(json_jq "$situation_detail_body" '.meanings | length' || echo "")"
        word_slug="$(json_jq "$situation_detail_body" '.meanings[0].wordSlug // empty' || echo "")"
        if [ "$detail_situation_slug" != "$situation_slug" ]; then
          fail "situation detail slug '$detail_situation_slug' does not match list slug '$situation_slug'"
        elif [ -z "$meanings_count" ] || [ "$meanings_count" -lt 1 ] 2>/dev/null; then
          fail "situation detail for $situation_slug has no meanings"
        elif [ -z "$word_slug" ]; then
          fail "situation detail for $situation_slug missing first wordSlug"
        else
          pass "GET /api/v1/journey-situations/$situation_slug returned situation with meanings"
        fi
      fi
    fi

    if [ -n "$word_slug" ]; then
      http_get_authed \
        "$api_base_url/api/v1/canonical-words/$word_slug" \
        "$smoke_session_cookie"
      word_detail_status="$_http_last_status"
      word_detail_body="$_http_last_body"
      if [ "$word_detail_status" != "200" ]; then
        fail "GET /api/v1/canonical-words/$word_slug returned $word_detail_status (expected 200)"
      elif [ -z "$word_detail_body" ]; then
        fail "GET /api/v1/canonical-words/$word_slug returned 200 with an empty body"
      else
        detail_word_slug="$(json_jq "$word_detail_body" '.word.slug // empty' || echo "")"
        if [ "$detail_word_slug" != "$word_slug" ]; then
          fail "canonical word detail slug '$detail_word_slug' does not match requested slug '$word_slug'"
        else
          pass "GET /api/v1/canonical-words/$word_slug returned word detail"
        fi
      fi
    fi
  fi
fi

# 6. Authenticated learning-route sweep (VOC-085-T02). deploy-production.yml
# always supplies SMOKE_TEST_SESSION_COOKIE after VOC-050-T04 session mint, so
# this section fails closed when the cookie is missing rather than silently
# skipping required coverage. Public routes are still requested without a
# cookie; protected routes and API-derived discover paths require it. Every
# request is a non-mutating GET — no magic links, OAuth completion, or
# state-changing learning actions.
if [ -z "$smoke_session_cookie" ]; then
  fail "authenticated route sweep requires SMOKE_TEST_SESSION_COOKIE (fail closed; deploy-production.yml must mint one)"
else
  for route_path in "${PRODUCTION_PUBLIC_WEB_ROUTES[@]}"; do
    assert_web_route_reachable \
      "GET $web_base_url$route_path (public)" \
      "$route_path" \
      ""
  done

  for route_path in "${PRODUCTION_AUTHENTICATED_WEB_ROUTES[@]}"; do
    assert_web_route_reachable \
      "GET $web_base_url$route_path (authenticated)" \
      "$route_path" \
      "$smoke_session_cookie" \
      "true"
  done

  if [ -z "$situation_slug" ] || [ -z "$word_slug" ]; then
    fail "cannot derive dynamic discover routes without journey-situations content (situation_slug/word_slug missing)"
  else
    assert_web_route_reachable \
      "GET $web_base_url/discover/$situation_slug (authenticated)" \
      "/discover/$situation_slug" \
      "$smoke_session_cookie" \
      "true"
    assert_web_route_reachable \
      "GET $web_base_url/discover/$situation_slug/$word_slug (authenticated)" \
      "/discover/$situation_slug/$word_slug" \
      "$smoke_session_cookie" \
      "true"
  fi
fi

echo
if [ "$failures" -gt 0 ]; then
  echo "SMOKE TEST FAILED: $failures check(s) failed" >&2
  exit 1
fi
echo "SMOKE TEST PASSED"
