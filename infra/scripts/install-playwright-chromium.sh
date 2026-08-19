#!/usr/bin/env bash
set -euo pipefail

# VOC-095-T00 — bounded Playwright Chromium install contract for CI workflows.
#
# Separates browser download (Playwright CDN; may no-op when ~/.cache/ms-playwright
# is warm) from system dependency installation (apt via playwright install-deps).
# Deps install uses explicit per-attempt timeout, fixed retry count, and binary
# verification. Workflows remain fail-closed on exhaustion.
#
# Constants (VOC-095-D03):
#   PLAYWRIGHT_DEPS_MAX_ATTEMPTS=3
#   PLAYWRIGHT_DEPS_ATTEMPT_TIMEOUT_SECONDS=120
#   PLAYWRIGHT_DEPS_RETRY_SLEEP_SECONDS=30
#
# Usage (from repository root):
#   bash infra/scripts/install-playwright-chromium.sh
#
# Logs record attempt counts and exit codes only (VOC-095-D07).

readonly PLAYWRIGHT_DEPS_MAX_ATTEMPTS=3
readonly PLAYWRIGHT_DEPS_ATTEMPT_TIMEOUT_SECONDS=120
readonly PLAYWRIGHT_DEPS_RETRY_SLEEP_SECONDS=30

readonly REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
readonly PLAYWRIGHT_CACHE_ROOT="${HOME}/.cache/ms-playwright"

cd "$REPO_ROOT"

playwright_invoke() {
  pnpm --filter @vocanova/web exec playwright "$@"
}

verify_chromium_binary() {
  local candidate=""
  shopt -s nullglob
  local matches=("$PLAYWRIGHT_CACHE_ROOT"/chromium-*/chrome-linux/chrome)
  shopt -u nullglob

  for candidate in "${matches[@]}"; do
    if [[ -x "$candidate" ]]; then
      return 0
    fi
  done

  echo "Chromium binary not found under ${PLAYWRIGHT_CACHE_ROOT}/chromium-*/chrome-linux/chrome" >&2
  return 1
}

install_chromium_browser() {
  playwright_invoke install chromium
}

install_chromium_deps_with_retries() {
  local attempt=1
  local exit_code=0

  while (( attempt <= PLAYWRIGHT_DEPS_MAX_ATTEMPTS )); do
    set +e
    timeout "${PLAYWRIGHT_DEPS_ATTEMPT_TIMEOUT_SECONDS}" \
      pnpm --filter @vocanova/web exec playwright install-deps chromium
    exit_code=$?
    set -e

    if (( exit_code == 0 )); then
      echo "playwright install-deps succeeded on attempt ${attempt}/${PLAYWRIGHT_DEPS_MAX_ATTEMPTS}"
      return 0
    fi

    echo "playwright install-deps attempt ${attempt}/${PLAYWRIGHT_DEPS_MAX_ATTEMPTS} failed (exit ${exit_code})" >&2

    if (( attempt >= PLAYWRIGHT_DEPS_MAX_ATTEMPTS )); then
      echo "playwright install-deps exhausted ${PLAYWRIGHT_DEPS_MAX_ATTEMPTS} attempts; failing closed" >&2
      return 1
    fi

    sleep "${PLAYWRIGHT_DEPS_RETRY_SLEEP_SECONDS}"
    attempt=$((attempt + 1))
  done

  return 1
}

install_chromium_browser
install_chromium_deps_with_retries
verify_chromium_binary
