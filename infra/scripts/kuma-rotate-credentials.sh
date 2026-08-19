#!/usr/bin/env bash
set -euo pipefail

# VOC-086-T02 — Reset the Kuma admin password via the official container tool.
#
# Uses a prompt-driven stdin exchange so the credential never appears in argv
# and Kuma's readline handler receives each answer only after it installs the
# corresponding prompt. Alternatively set KUMA_NEW_PASSWORD in the environment
# for workflow handoff (never echo or log that value).
#
# Operator impact: existing Kuma sessions are invalidated once per successful run.
#
# Usage:
#   KUMA_NEW_PASSWORD="$(openssl rand -base64 48)" infra/scripts/kuma-rotate-credentials.sh
#   printf '%s\n' "$pass" | infra/scripts/kuma-rotate-credentials.sh

KUMA_CONTAINER="${KUMA_CONTAINER:-vocanova-uptime-kuma}"
RESET_SCRIPT_PATH="${KUMA_RESET_SCRIPT_PATH:-/app/extra/reset-password.js}"
RESET_APPLIED_FILE="${KUMA_RESET_APPLIED_FILE:-}"
RESET_ATTEMPT_ID="${KUMA_RESET_ATTEMPT_ID:-}"

# A previous successful run must never be accepted as proof for this attempt.
# The marker contains no credential material and is written only after the
# official reset tool returns success.
if [ -n "$RESET_APPLIED_FILE" ]; then
  rm -f "$RESET_APPLIED_FILE"
  if [ -z "$RESET_ATTEMPT_ID" ]; then
    echo "KUMA_RESET_ATTEMPT_ID is required when reset proof is requested" >&2
    exit 1
  fi
fi

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

reset_output=""
prompt_window=""
sent_new_password=false
sent_confirmation=false
coproc KUMA_RESET_PROCESS {
  timeout --signal=TERM "${KUMA_RESET_TIMEOUT_SECONDS:-60}" \
    docker exec -i "$KUMA_CONTAINER" node "$RESET_SCRIPT_PATH" 2>&1
}
reset_pid="$KUMA_RESET_PROCESS_PID"
coproc_output_fd="${KUMA_RESET_PROCESS[0]}"
coproc_input_fd="${KUMA_RESET_PROCESS[1]}"
# Duplicate Bash's transient coprocess descriptors. Bash may close/unset the
# originals as soon as the child exits, including between loop-condition reads.
exec {reset_output_fd}<&"$coproc_output_fd"
exec {reset_input_fd}>&"$coproc_input_fd"
exec {coproc_output_fd}<&-
exec {coproc_input_fd}>&-

while IFS= read -r -N 1 next_character <&"$reset_output_fd"; do
  reset_output+="$next_character"
  prompt_window+="$next_character"
  if [ "$sent_new_password" = false ] && [[ "$prompt_window" == *"New Password: " ]]; then
    printf '%s\n' "$password" >&"$reset_input_fd"
    sent_new_password=true
    prompt_window=""
  elif [ "$sent_new_password" = true ] && [ "$sent_confirmation" = false ] && \
       [[ "$prompt_window" == *"Confirm New Password: " ]]; then
    printf '%s\n' "$password" >&"$reset_input_fd"
    sent_confirmation=true
    prompt_window=""
    exec {reset_input_fd}>&-
  elif [ "${#prompt_window}" -gt 256 ]; then
    prompt_window="${prompt_window: -128}"
  fi
done

set +e
wait "$reset_pid"
reset_status=$?
set -e

if [ "$sent_new_password" != true ] || [ "$sent_confirmation" != true ]; then
  echo "Kuma reset tool did not complete its expected password prompts:" >&2
  redact_sensitive <<<"$reset_output" >&2
  exit 1
fi

if [ "$reset_status" -ne 0 ]; then
  echo "Kuma password reset failed:" >&2
  redact_sensitive <<<"$reset_output" >&2
  exit "$reset_status"
fi

# Uptime Kuma 1.x catches reset errors internally and can still exit zero.
# Require both its durable-reset marker and its immediate authenticated login
# marker before treating the operation as successful. This prevents a false
# reset proof from causing the workflow to store or scrub unusable credentials.
if ! grep -qx 'Password reset successfully\.' <<<"$reset_output" ||
   ! grep -qx 'Logged in\.' <<<"$reset_output"; then
  echo "Kuma reset tool exited zero without verified reset and login markers:" >&2
  redact_sensitive <<<"$reset_output" >&2
  exit 1
fi

if [ -n "$RESET_APPLIED_FILE" ]; then
  umask 077
  printf 'KUMA_RESET_APPLIED=%s\n' "$RESET_ATTEMPT_ID" > "$RESET_APPLIED_FILE"
fi

username=""
if grep -q 'Found user:' <<<"$reset_output"; then
  username="$(sed -n 's/.*Found user: //p' <<<"$reset_output" | head -n 1 | tr -d '\r')"
fi

if [ -z "${username:-}" ]; then
  echo "Kuma password reset succeeded but no username was reported; refusing to continue without a preserved username" >&2
  exit 1
fi

# Username and reset-attempt identity only (never the password). The workflow
# downloads this file host→runner via OpenSSH scp — appleboy/scp-action is
# upload-only in this repository. Binding the metadata to the reset proof keeps
# a later store-only recovery from combining files from different attempts.
echo "KUMA_USERNAME=${username}"
if [ -n "${KUMA_ROTATE_METADATA_FILE:-}" ]; then
  umask 077
  {
    printf 'KUMA_ROTATION_ATTEMPT_ID=%s\n' "$RESET_ATTEMPT_ID"
    printf 'KUMA_USERNAME=%s\n' "$username"
  } > "$KUMA_ROTATE_METADATA_FILE"
fi

echo "Kuma admin password reset completed via ${RESET_SCRIPT_PATH} (existing sessions invalidated)."
