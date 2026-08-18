#!/usr/bin/env bash
set -euo pipefail

# VOC-086-T03 — dispatch a canonical scheduled synthetic check by registry
# check_ref. Reuses deploy-time harness scripts; does not mutate real users.
#
# Usage:
#   run-scheduled-synthetic.sh <check_ref>
#
# Environment (per check_ref):
#   staging-oauth-expected-state:
#     EXPECT_OAUTH_ENABLED (optional)
#   production-oauth-expected-state:
#     EXPECT_OAUTH_ENABLED (optional)
#   production-journey-content:
#     SMOKE_TEST_SESSION_COOKIE (required)
#   production-authenticated-route-content-sweep:
#     SMOKE_TEST_SESSION_COOKIE (required)
#
# staging-authenticated-core-journey is executed by the workflow via Playwright
# directly (see scheduled-synthetics.yml) because it needs Node/Playwright setup.

repo_root="$(cd "$(dirname "$0")/../.." && pwd)"

if [ "$#" -ne 1 ]; then
  echo "usage: $0 <check_ref>" >&2
  exit 1
fi

check_ref="$1"

case "$check_ref" in
  staging-oauth-expected-state)
    exec bash "$repo_root/infra/scripts/verify-staging-oauth-start.sh"
    ;;
  production-oauth-expected-state)
    exec bash "$repo_root/infra/scripts/verify-production-oauth-start.sh"
    ;;
  production-journey-content)
    if [ -z "${SMOKE_TEST_SESSION_COOKIE:-}" ]; then
      echo "SMOKE_TEST_SESSION_COOKIE is required for production journey content checks" >&2
      exit 1
    fi
    export SMOKE_TEST_PROFILE=journey-content
    exec bash "$repo_root/infra/scripts/smoke-test-production.sh" \
      "${PRODUCTION_API_BASE_URL:-https://api-production.vocanova.site}" \
      "${PRODUCTION_WEB_BASE_URL:-https://production.vocanova.site}"
    ;;
  production-authenticated-route-content-sweep)
    if [ -z "${SMOKE_TEST_SESSION_COOKIE:-}" ]; then
      echo "SMOKE_TEST_SESSION_COOKIE is required for the production route sweep" >&2
      exit 1
    fi
    export SMOKE_TEST_PROFILE=route-sweep
    exec bash "$repo_root/infra/scripts/smoke-test-production.sh" \
      "${PRODUCTION_API_BASE_URL:-https://api-production.vocanova.site}" \
      "${PRODUCTION_WEB_BASE_URL:-https://production.vocanova.site}"
    ;;
  staging-authenticated-core-journey)
    echo "staging-authenticated-core-journey is executed by scheduled-synthetics.yml via Playwright" >&2
    exit 1
    ;;
  *)
    echo "unknown scheduled synthetic check_ref: $check_ref" >&2
    exit 1
    ;;
esac
