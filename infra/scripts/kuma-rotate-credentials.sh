#!/usr/bin/env bash
set -euo pipefail

# VOC-086-T02 — Reset the Kuma admin password via the official container tool.
#
# Uses stdin (two lines: new password + confirm) so the credential never appears
# in argv. Alternatively set KUMA_NEW_PASSWORD in the environment for workflow
# handoff (never echo or log that value).
#
# Operator impact: existing Kuma sessions are invalidated once per successful run.
#
# Usage:
#   KUMA_NEW_PASSWORD="$(openssl rand -base64 48)" infra/scripts/kuma-rotate-credentials.sh
#   printf '%s\n%s\n' "$pass" "$pass" | infra/scripts/kuma-rotate-credentials.sh

KUMA_CONTAINER="${KUMA_CONTAINER:-vocanova-uptime-kuma}"
RESET_SCRIPT_PATH="${KUMA_RESET_SCRIPT_PATH:-/app/extra/reset-password.js}"

redact_sensitive() {
  sed -E \
    -e 's/(password[=: ]+)[^[:space:]]+/\1[REDACTED]/gi' \
    -e 's/(KUMA_PASSWORD[=: ]+)[^[:space:]]+/\1[REDACTED]/gi' \
    -e 's/(KUMA_NEW_PASSWORD[=: ]+)[^[:space:]]+/\1[REDACTED]/gi' \
    -e 's/(Bearer +)[A-Za-z0-9._-]+/\1[REDACTED]/gi'
}

if [ -n "${KUMA_NEW_PASSWORD_FILE:-}" ] && [ -f "$KUMA_NEW_PASSWORD_FILE" ]; then
  password="$(cat "$KUMA_NEW_PASSWORD_FILE")"
elif [ -n "${KUMA_NEW_PASSWORD:-}" ]; then
  password="$KUMA_NEW_PASSWORD"
else
  IFS= read -r password || true
fi

if [ -z "${password:-}" ]; then
  echo "KUMA_NEW_PASSWORD or stdin password is required" >&2
  exit 1
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required to run Kuma credential rotation" >&2
  exit 1
fi

if ! docker inspect "$KUMA_CONTAINER" >/dev/null 2>&1; then
  echo "Kuma container ${KUMA_CONTAINER} is not present on this host" >&2
  exit 1
fi

set +e
reset_output="$(printf '%s\n%s\n' "$password" "$password" | \
  docker exec -i "$KUMA_CONTAINER" node "$RESET_SCRIPT_PATH" 2>&1)"
reset_status=$?
set -e

if [ "$reset_status" -ne 0 ]; then
  echo "Kuma password reset failed:" >&2
  redact_sensitive <<<"$reset_output" >&2
  exit "$reset_status"
fi

username=""
if grep -q 'Found user:' <<<"$reset_output"; then
  username="$(sed -n 's/.*Found user: //p' <<<"$reset_output" | head -n 1 | tr -d '\r')"
fi

if [ -z "${username:-}" ]; then
  echo "Kuma password reset succeeded but no username was reported; refusing to continue without a preserved username" >&2
  exit 1
fi

# Username only (never the password). Workflow downloads this file host→runner via
# OpenSSH scp — appleboy/scp-action is upload-only in this repository.
echo "KUMA_USERNAME=${username}"
if [ -n "${KUMA_ROTATE_METADATA_FILE:-}" ]; then
  umask 077
  printf 'KUMA_USERNAME=%s\n' "$username" > "$KUMA_ROTATE_METADATA_FILE"
fi

echo "Kuma admin password reset completed via ${RESET_SCRIPT_PATH} (existing sessions invalidated)."
