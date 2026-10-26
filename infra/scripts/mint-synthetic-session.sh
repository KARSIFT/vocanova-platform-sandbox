#!/usr/bin/env bash
set -euo pipefail

# VOC-086-T03 — mint a synthetic smoke-test session via VOC-050-T01's
# token-gated POST /ops/synthetic-smoke-test/session endpoint.
#
# Reuses the reserved synthetic account and existing mint secrets from
# deploy workflows. Never prints session or CSRF values except inside
# ::add-mask:: commands, including on errors; when GITHUB_OUTPUT is set,
# writes masked outputs for workflow steps.
#
# Usage:
#   mint-synthetic-session.sh <api_base_url> [api_host_header]
#
# Environment:
#   SMOKE_TEST_SESSION_MINT_TOKEN — required bearer token
#   SYNTHETIC_COOKIE_ENVIRONMENT  — required expected tier, staging|production.
#                                   The response's authoritative cookie names
#                                   must be exactly vocanova_<tier>_session and
#                                   vocanova_<tier>_csrf; anything else fails
#                                   closed with no legacy-name fallback.
#
# Outputs (when GITHUB_OUTPUT is set):
#   session_cookie        — opaque session value
#   csrf_token            — opaque CSRF value; mint fails closed if empty
#                           (deploy-staging parity so staging Playwright
#                           gets a clear ops signal)
#   session_cookie_name   — validated tier session cookie name
#   csrf_cookie_name      — validated tier CSRF cookie name
#   session_cookie_header — `<session_cookie_name>=<session_cookie>`, the
#                           whole Cookie header smoke-test-production.sh sends

if [ "$#" -lt 1 ]; then
  echo "usage: $0 <api_base_url> [api_host_header]" >&2
  exit 1
fi

api_base_url="${1%/}"
api_host_header="${2:-}"

if [ -z "${SMOKE_TEST_SESSION_MINT_TOKEN:-}" ]; then
  echo "SMOKE_TEST_SESSION_MINT_TOKEN is required to mint a synthetic session" >&2
  exit 1
fi

case "${SYNTHETIC_COOKIE_ENVIRONMENT:-}" in
  staging | production) cookie_environment="$SYNTHETIC_COOKIE_ENVIRONMENT" ;;
  *)
    echo "SYNTHETIC_COOKIE_ENVIRONMENT must be staging or production" >&2
    exit 1
    ;;
esac

# The removed prefix mode mixed raw and prefixed values in one output; fail
# closed rather than silently ignore a stale caller.
if [ -n "${SESSION_COOKIE_PREFIX:-}" ]; then
  echo "SESSION_COOKIE_PREFIX is no longer supported; use session_cookie_header" >&2
  exit 1
fi

curl_args=(
  -fsS
  --max-time 20
  -X POST
  -H "Authorization: Bearer ${SMOKE_TEST_SESSION_MINT_TOKEN}"
  -H "Accept: application/json"
)
if [ -n "$api_host_header" ]; then
  curl_args+=(-H "Host: ${api_host_header}")
fi

response_json="$(curl "${curl_args[@]}" "${api_base_url}/ops/synthetic-smoke-test/session")"

# Validate everything before any value is printed or written. Errors name
# fields, never values. Opaque values must be RFC 6265 cookie-octets, which
# excludes CR/LF, whitespace, `;`, `,`, `"` and `\`, so they cannot inject
# headers, extra cookies or GITHUB_OUTPUT lines.
parsed="$(printf '%s' "$response_json" | COOKIE_ENVIRONMENT="$cookie_environment" python3 -c '
import json, os, re, sys

def fail(message):
    print("the session-mint endpoint " + message, file=sys.stderr)
    sys.exit(1)

try:
    data = json.load(sys.stdin)
except Exception:
    fail("returned a response that is not valid JSON")
if not isinstance(data, dict):
    fail("returned a response that is not a JSON object")

tier = os.environ["COOKIE_ENVIRONMENT"]
cookie_octets = re.compile(r"[\x21\x23-\x2b\x2d-\x3a\x3c-\x5b\x5d-\x7e]+")
values = []
for key in ("session_cookie", "csrf_token"):
    value = data.get(key)
    if not isinstance(value, str) or not value:
        fail("returned an empty session or csrf value")
    if not cookie_octets.fullmatch(value):
        fail("returned a " + key + " with characters not allowed in a cookie value")
    values.append(value)
for key, suffix in (("session_cookie_name", "session"), ("csrf_cookie_name", "csrf")):
    value = data.get(key)
    if not isinstance(value, str) or not value:
        fail("did not return " + key)
    if value != "vocanova_" + tier + "_" + suffix:
        fail("returned a " + key + " that is not the expected " + tier + " cookie name")
    values.append(value)
print("\n".join(values))
')"

{
  IFS= read -r session_cookie
  IFS= read -r csrf_token
  IFS= read -r session_cookie_name
  IFS= read -r csrf_cookie_name
} <<< "$parsed"
session_cookie_header="${session_cookie_name}=${session_cookie}"

if [ -n "${GITHUB_OUTPUT:-}" ]; then
  echo "::add-mask::${session_cookie}"
  echo "::add-mask::${csrf_token}"
  {
    printf 'session_cookie=%s\n' "$session_cookie"
    printf 'csrf_token=%s\n' "$csrf_token"
    printf 'session_cookie_name=%s\n' "$session_cookie_name"
    printf 'csrf_cookie_name=%s\n' "$csrf_cookie_name"
    printf 'session_cookie_header=%s\n' "$session_cookie_header"
  } >> "$GITHUB_OUTPUT"
else
  echo "Minted a synthetic smoke-test session (values withheld from stdout)."
fi
