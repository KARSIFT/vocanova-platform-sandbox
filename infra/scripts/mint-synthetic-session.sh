#!/usr/bin/env bash
set -euo pipefail

# VOC-086-T03 — mint a synthetic smoke-test session via VOC-050-T01's
# token-gated POST /ops/synthetic-smoke-test/session endpoint.
#
# Reuses the reserved synthetic account and existing mint secrets from
# deploy workflows. Never prints session or CSRF values to stdout; when
# GITHUB_OUTPUT is set, writes masked outputs for workflow steps.
#
# Usage:
#   mint-synthetic-session.sh <api_base_url> [api_host_header]
#
# Environment:
#   SMOKE_TEST_SESSION_MINT_TOKEN — required bearer token
#   SESSION_COOKIE_PREFIX         — optional; when "vocanova_session", the
#                                   session_cookie output is prefixed for
#                                   smoke-test-production.sh's Cookie header
#
# Outputs (when GITHUB_OUTPUT is set):
#   session_cookie — opaque value or full Cookie header pair per prefix mode
#   csrf_token     — required; mint fails closed if empty (deploy-staging
#                    parity so staging Playwright gets a clear ops signal)

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

session_cookie="$(printf '%s' "$response_json" | python3 -c 'import json,sys; print(json.load(sys.stdin).get("session_cookie") or "")')"
csrf_token="$(printf '%s' "$response_json" | python3 -c 'import json,sys; print(json.load(sys.stdin).get("csrf_token") or "")')"

if [ -z "$session_cookie" ] || [ -z "$csrf_token" ]; then
  echo "the session-mint endpoint returned an empty session or csrf value" >&2
  exit 1
fi

if [ "${SESSION_COOKIE_PREFIX:-}" = "vocanova_session" ]; then
  session_output="vocanova_session=${session_cookie}"
else
  session_output="$session_cookie"
fi

if [ -n "${GITHUB_OUTPUT:-}" ]; then
  echo "::add-mask::${session_cookie}"
  echo "::add-mask::${csrf_token}"
  {
    echo "session_cookie=${session_output}"
    echo "csrf_token=${csrf_token}"
  } >> "$GITHUB_OUTPUT"
else
  echo "Minted a synthetic smoke-test session (values withheld from stdout)."
fi
