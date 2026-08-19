#!/usr/bin/env bash
set -euo pipefail

# VOC-095-T00 — bounded Playwright Chromium install contract for CI workflows.
#
# Separates browser download (Playwright CDN; may no-op when ~/.cache/ms-playwright
# is warm) from system dependency installation (apt via playwright install-deps).
# Deps install uses an explicit per-attempt timeout, process-tree termination,
# fixed retry count, a runner-only HTTPS archive fallback, and binary
# verification. Workflows remain fail-closed on exhaustion or cleanup failure.
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
readonly PLAYWRIGHT_DEPS_KILL_GRACE_SECONDS=10
readonly PLAYWRIGHT_APT_LOCK_WAIT_SECONDS=15
readonly UBUNTU_HTTPS_ARCHIVE="https://archive.ubuntu.com/ubuntu/"

readonly REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
readonly PLAYWRIGHT_CACHE_ROOT="${HOME}/.cache/ms-playwright"
readonly APT_MIRROR_LIST_PATH="${PLAYWRIGHT_APT_MIRROR_LIST_PATH:-/etc/apt/apt-mirrors.txt}"

mirror_backup=""
mirror_fallback_active=false

restore_runner_mirror_list() {
  if [[ -n "$mirror_backup" && -f "$mirror_backup" ]]; then
    sudo -n install -m 0644 "$mirror_backup" "$APT_MIRROR_LIST_PATH" >/dev/null 2>&1 || {
      echo "failed to restore hosted-runner apt mirror configuration" >&2
      return 1
    }
    rm -f "$mirror_backup"
    mirror_backup=""
  fi
}

cleanup() {
  restore_runner_mirror_list
}

trap cleanup EXIT

cd "$REPO_ROOT"

for required_command in pnpm timeout sudo fuser; do
  if ! command -v "$required_command" >/dev/null 2>&1; then
    echo "required Playwright install command is unavailable: ${required_command}" >&2
    exit 1
  fi
done

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

wait_for_package_manager_quiescence() {
  local waited=0
  local lock=""
  local locks=(
    /var/lib/dpkg/lock-frontend
    /var/lib/dpkg/lock
    /var/lib/apt/lists/lock
    /var/cache/apt/archives/lock
  )

  while (( waited < PLAYWRIGHT_APT_LOCK_WAIT_SECONDS )); do
    local busy=false
    for lock in "${locks[@]}"; do
      if sudo -n fuser "$lock" >/dev/null 2>&1; then
        busy=true
        break
      fi
    done
    if [[ "$busy" == false ]]; then
      return 0
    fi
    sleep 1
    waited=$((waited + 1))
  done

  echo "package-manager locks remained busy after timed install cleanup; failing closed" >&2
  return 1
}

activate_https_archive_fallback() {
  local fallback_file=""

  if [[ "${GITHUB_ACTIONS:-false}" != "true" ]]; then
    echo "hosted-runner apt fallback is unavailable outside GitHub Actions" >&2
    return 1
  fi
  if [[ "${RUNNER_OS:-}" != "Linux" || "${RUNNER_ARCH:-}" != "X64" ]]; then
    echo "hosted-runner apt fallback requires the GitHub Linux X64 runner" >&2
    return 1
  fi
  if [[ ! -f "$APT_MIRROR_LIST_PATH" ]]; then
    echo "hosted-runner apt mirror list is unavailable; failing closed" >&2
    return 1
  fi

  mirror_backup="$(mktemp)"
  fallback_file="$(mktemp)"
  cp "$APT_MIRROR_LIST_PATH" "$mirror_backup"
  printf '%s\n' "$UBUNTU_HTTPS_ARCHIVE" >"$fallback_file"

  if ! sudo -n install -m 0644 "$fallback_file" "$APT_MIRROR_LIST_PATH" >/dev/null 2>&1; then
    rm -f "$fallback_file"
    echo "could not activate verified HTTPS Ubuntu archive fallback" >&2
    return 1
  fi
  rm -f "$fallback_file"

  if ! grep -Fxq "$UBUNTU_HTTPS_ARCHIVE" "$APT_MIRROR_LIST_PATH"; then
    echo "HTTPS Ubuntu archive fallback verification failed" >&2
    return 1
  fi

  mirror_fallback_active=true
  echo "activated verified HTTPS Ubuntu archive fallback for hosted runner"
}

install_chromium_deps_with_retries() {
  local attempt=1
  local exit_code=0

  while (( attempt <= PLAYWRIGHT_DEPS_MAX_ATTEMPTS )); do
    set +e
    timeout --signal=TERM \
      --kill-after="${PLAYWRIGHT_DEPS_KILL_GRACE_SECONDS}" \
      "${PLAYWRIGHT_DEPS_ATTEMPT_TIMEOUT_SECONDS}" \
      pnpm --filter @vocanova/web exec playwright install-deps chromium
    exit_code=$?
    set -e

    if (( exit_code == 0 )); then
      echo "playwright install-deps succeeded on attempt ${attempt}/${PLAYWRIGHT_DEPS_MAX_ATTEMPTS}"
      return 0
    fi

    echo "playwright install-deps attempt ${attempt}/${PLAYWRIGHT_DEPS_MAX_ATTEMPTS} failed (exit ${exit_code})" >&2

    if ! wait_for_package_manager_quiescence; then
      return 1
    fi

    if (( attempt >= PLAYWRIGHT_DEPS_MAX_ATTEMPTS )); then
      echo "playwright install-deps exhausted ${PLAYWRIGHT_DEPS_MAX_ATTEMPTS} attempts; failing closed" >&2
      return 1
    fi

    if [[ "$mirror_fallback_active" == false ]]; then
      if ! activate_https_archive_fallback; then
        echo "playwright install-deps cannot continue without a verified fallback" >&2
        return 1
      fi
    fi

    sleep "${PLAYWRIGHT_DEPS_RETRY_SLEEP_SECONDS}"
    attempt=$((attempt + 1))
  done

  return 1
}

install_chromium_browser
install_chromium_deps_with_retries
verify_chromium_binary
