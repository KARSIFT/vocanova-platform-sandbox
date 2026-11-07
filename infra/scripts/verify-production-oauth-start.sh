#!/usr/bin/env bash
set -euo pipefail

# VOC-086-T03 — scheduled production OAuth-start initiation check.
# When OAuth is expected enabled, also checks the configured signup policy
# against the public and controlled signup flags in /healthz.
#
# POSTs /api/v1/auth/oauth/google/start without following Google or
# completing OAuth. When EXPECT_OAUTH_ENABLED=true, requires HTTP 200,
# an accounts.google.com authorization URL, redirect_uri exactly
# matching the canonical production callback, and the expected public or
# controlled signup state on /healthz. When false, requires coherent 503 disabled behavior
# (not a fake 200).
#
# Mirrors verify-staging-oauth-start.sh (VOC-084-T02) for production hosts.
#
# Usage:
#   verify-production-oauth-start.sh
#   verify-production-oauth-start.sh <api_base_url> <web_redirect_uri>
#
# Environment:
#   EXPECT_OAUTH_ENABLED        (default: false)
#   EXPECT_NEW_SIGNUPS_ENABLED   (default: false)
#   PRODUCTION_API_HOST          (default: api-production.vocanova.site)
#
# Requires: curl, python3

API_BASE_URL="${1:-https://api-production.vocanova.site}"
WEB_REDIRECT_URI="${2:-https://production.vocanova.site/home}"
EXPECT_OAUTH_ENABLED="${EXPECT_OAUTH_ENABLED:-false}"
EXPECT_NEW_SIGNUPS_ENABLED="${EXPECT_NEW_SIGNUPS_ENABLED:-false}"
API_HOST="${PRODUCTION_API_HOST:-api-production.vocanova.site}"

CANONICAL_CALLBACK="https://api-production.vocanova.site/api/v1/auth/oauth/google/callback"

api_base_url="${API_BASE_URL%/}"

failures=0
record_fail() {
  echo "FAIL: $*" >&2
  failures=$((failures + 1))
}

record_pass() {
  echo "PASS: $*"
}

case "$EXPECT_NEW_SIGNUPS_ENABLED" in
  true|false) ;;
  *)
    echo "FAIL: invalid production signup policy expectation" >&2
    exit 1
    ;;
esac

echo "VOC-086 production OAuth-start verification (EXPECT_OAUTH_ENABLED=$EXPECT_OAUTH_ENABLED, EXPECT_NEW_SIGNUPS_ENABLED=$EXPECT_NEW_SIGNUPS_ENABLED)"

response_file="$(mktemp)"
trap 'rm -f "$response_file"' EXIT

http_status="$(curl -sS --max-time 20 \
  -o "$response_file" \
  -w "%{http_code}" \
  -X POST \
  -H "Host: ${API_HOST}" \
  -H "Content-Type: application/json" \
  -d "{\"redirectUri\":\"${WEB_REDIRECT_URI}\"}" \
  "${api_base_url}/api/v1/auth/oauth/google/start" 2>/dev/null || true)"
if [ -z "$http_status" ]; then
  http_status="000"
fi

if [ "$EXPECT_OAUTH_ENABLED" = "true" ]; then
  if [ "$http_status" != "200" ]; then
    record_fail "OAuth start returned HTTP $http_status (expected 200 while OAuth is enabled)"
  else
    record_pass "OAuth start returned HTTP 200"

    if python3 - "$response_file" "$CANONICAL_CALLBACK" <<'PY'
import json
import sys
from urllib.parse import parse_qs, urlparse

with open(sys.argv[1], encoding="utf-8") as fh:
    body = fh.read()
canonical = sys.argv[2]

payload = json.loads(body)
url = payload.get("url")
if not url:
    raise SystemExit("response JSON missing url field")

parsed = urlparse(url)
if parsed.scheme != "https" or parsed.hostname != "accounts.google.com":
    raise SystemExit(
        f"authorization URL host is not accounts.google.com: {parsed.netloc!r}"
    )

redirect_uris = parse_qs(parsed.query).get("redirect_uri", [])
if redirect_uris != [canonical]:
    raise SystemExit(
        f"redirect_uri mismatch: got {redirect_uris!r}, expected {[canonical]!r}"
    )
PY
    then
      record_pass "authorization URL targets accounts.google.com with canonical redirect_uri"
    else
      record_fail "OAuth start response did not contain a valid Google authorization URL with the canonical callback"
    fi
  fi

  healthz_file="$(mktemp)"
  healthz_status="$(curl -sS --max-time 20 \
    -o "$healthz_file" \
    -w "%{http_code}" \
    -H "Host: ${API_HOST}" \
    "${api_base_url}/healthz" 2>/dev/null || true)"
  if [ -z "$healthz_status" ]; then
    healthz_status="000"
  fi

  if [ "$healthz_status" != "200" ]; then
    record_fail "/healthz returned HTTP $healthz_status (expected 200 while OAuth is enabled)"
  else
    record_pass "/healthz returned HTTP 200"
    if python3 - "$healthz_file" "$EXPECT_NEW_SIGNUPS_ENABLED" <<'PY'
import json
import re
import sys

with open(sys.argv[1], encoding="utf-8") as fh:
    body = fh.read()

payload = json.loads(body)
expect_public = sys.argv[2] == "true"
kill_switches = payload.get("kill_switches")
if not isinstance(kill_switches, dict) or kill_switches.get("new_signups_enabled") is not expect_public:
    raise SystemExit("kill_switches.new_signups_enabled does not match the expected signup policy")
if payload.get("controlled_signup_ready") is not (not expect_public):
    raise SystemExit("controlled_signup_ready does not match the expected signup policy")

if re.search(r'"[^"]*@[^"]+"', body):
    raise SystemExit("healthz response must not expose email-like strings")
PY
    then
      if [ "$EXPECT_NEW_SIGNUPS_ENABLED" = "true" ]; then
        record_pass "public signup is enabled and controlled signup is inactive"
      else
        record_pass "controlled_signup_ready is true without exposing cohort metadata"
      fi
    else
      record_fail "/healthz did not report the expected signup policy without email metadata"
    fi
  fi
  rm -f "$healthz_file"
else
  if [ "$http_status" = "503" ]; then
    record_pass "OAuth start correctly refused with HTTP 503 while disabled"
  else
    record_fail "OAuth start returned HTTP $http_status while disabled (expected 503)"
  fi
fi

if [ "$failures" -gt 0 ]; then
  echo "PRODUCTION OAUTH-START VERIFICATION FAILED: $failures check(s) failed" >&2
  exit 1
fi

echo "PRODUCTION OAUTH-START VERIFICATION PASSED"
