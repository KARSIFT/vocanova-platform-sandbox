#!/usr/bin/env bash
# VOC-088-T00 — validate the secret-backed staging controlled-signup cohort.
#
# The allowlist is personal data. This script deliberately reports only the
# ready/not-ready fact and fixed diagnostics; it must never print the value or
# cohort cardinality. The API performs its own lowercase/trim normalization.

set -euo pipefail

oauth_enabled="${STAGING_GOOGLE_OAUTH_ENABLED:-false}"
allowlist="${STAGING_NEW_USER_SIGNUP_ALLOWLIST:-}"

if [[ "$oauth_enabled" != "true" && "$oauth_enabled" != "false" ]]; then
  echo "invalid staging OAuth readiness flag" >&2
  exit 1
fi

if [[ "$oauth_enabled" == "false" ]]; then
  echo "controlled signup ready: false"
  exit 0
fi

if [[ -z "$allowlist" ]]; then
  echo "controlled signup cohort is empty; refusing staging deployment with OAuth enabled" >&2
  exit 1
fi

if [[ "$allowlist" == *$'\n'* || "$allowlist" == *$'\r'* ]]; then
  echo "controlled signup cohort is malformed; refusing staging deployment" >&2
  exit 1
fi

IFS=',' read -r -a entries <<< "$allowlist"
if (( ${#entries[@]} == 0 )); then
  echo "controlled signup cohort is empty; refusing staging deployment with OAuth enabled" >&2
  exit 1
fi

for entry in "${entries[@]}"; do
  trimmed="${entry#"${entry%%[![:space:]]*}"}"
  trimmed="${trimmed%"${trimmed##*[![:space:]]}"}"
  if [[ ! "$trimmed" =~ ^[^@,[:space:]]+@[^@,[:space:]]+\.[^@,[:space:]]+$ ]]; then
    echo "controlled signup cohort is malformed; refusing staging deployment" >&2
    exit 1
  fi
done

echo "controlled signup ready: true"
