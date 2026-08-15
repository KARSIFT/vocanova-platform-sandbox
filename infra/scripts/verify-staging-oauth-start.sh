#!/usr/bin/env bash
set -euo pipefail

# VOC-084-T02 — post-deploy staging OAuth-start initiation check.
#
# POSTs /api/v1/auth/oauth/google/start without following Google or
# completing OAuth. When EXPECT_OAUTH_ENABLED=true, requires HTTP 200,
# an accounts.google.com authorization URL, and redirect_uri exactly
# matching the canonical staging callback. When false, requires coherent
# 503 disabled behavior (not a fake 200).
#
# Usage:
#   verify-staging-oauth-start.sh
#   verify-staging-oauth-start.sh <api_base_url> <web_redirect_uri>
#
# Environment:
#   EXPECT_OAUTH_ENABLED   (default: false)
#   STAGING_API_HOST       (default: api-staging.vocanova.site)
#
# Requires: curl, python3

API_BASE_URL="${1:-https://api-staging.vocanova.site}"
WEB_REDIRECT_URI="${2:-https://staging.vocanova.site/home}"
EXPECT_OAUTH_ENABLED="${EXPECT_OAUTH_ENABLED:-false}"
API_HOST="${STAGING_API_HOST:-api-staging.vocanova.site}"

CANONICAL_CALLBACK="https://api-staging.vocanova.site/api/v1/auth/oauth/google/callback"

api_base_url="${API_BASE_URL%/}"

failures=0
record_fail() {
  echo "FAIL: $*" >&2
  failures=$((failures + 1))
}

record_pass() {
  echo "PASS: $*"
}

echo "VOC-084 staging OAuth-start verification (EXPECT_OAUTH_ENABLED=$EXPECT_OAUTH_ENABLED)"

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
if parsed.scheme != "https" or "accounts.google.com" not in parsed.netloc:
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
else
  if [ "$http_status" = "503" ]; then
    record_pass "OAuth start correctly refused with HTTP 503 while disabled"
  else
    record_fail "OAuth start returned HTTP $http_status while disabled (expected 503)"
  fi
fi

if [ "$failures" -gt 0 ]; then
  echo "STAGING OAUTH-START VERIFICATION FAILED: $failures check(s) failed" >&2
  exit 1
fi

echo "STAGING OAUTH-START VERIFICATION PASSED"
