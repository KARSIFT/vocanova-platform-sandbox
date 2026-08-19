#!/usr/bin/env bash
set -euo pipefail

# VOC-087-T02 — store/scrub decisions for Kuma credential rotation recovery.
#
# No credentials are accepted or printed. Callers pass only GitHub step
# outcomes, proof-fetch booleans, and secret-store completion booleans.
#
# store-decision → STORE | SKIP_UNUSED | RETAIN
# scrub-decision → SCRUB | RETAIN

to_bool() {
  case "${1:-false}" in
    true|TRUE|yes|YES) echo true ;;
    *) echo false ;;
  esac
}

rotate="${ROTATE_HOST_OUTCOME:-}"
preflight="${ROTATION_PREFLIGHT_OUTCOME:-}"
proof_matches="$(to_bool "${PROOF_MATCHES:-false}")"
host_reachable="$(to_bool "${HOST_REACHABLE:-false}")"
proof_fetched="$(to_bool "${PROOF_FETCHED:-false}")"
proof_absent="$(to_bool "${PROOF_ABSENT:-false}")"
password_stored="$(to_bool "${PASSWORD_STORED:-false}")"
username_stored="$(to_bool "${USERNAME_STORED:-false}")"
recover_store_only="$(to_bool "${RECOVER_STORE_ONLY:-false}")"

# Only the remote probe's explicit, reserved "absent" result proves that this
# attempt did not write a reset marker. Reachability or a failed transfer alone
# is inconclusive and must retain the recovery material.
reset_did_not_happen() {
  [ "$rotate" != "success" ] &&
    [ "$host_reachable" = "true" ] &&
    [ "$proof_fetched" = "false" ] &&
    [ "$proof_absent" = "true" ]
}

store_decision() {
  if [ "$proof_matches" = "true" ] || [ "$rotate" = "success" ]; then
    echo STORE
    return
  fi
  if reset_did_not_happen; then
    echo SKIP_UNUSED
    return
  fi
  echo RETAIN
}

scrub_decision() {
  if [ "$password_stored" = "true" ] && [ "$username_stored" = "true" ]; then
    echo SCRUB
    return
  fi
  # recover_store_only never resets. If store did not succeed, keep the host
  # copy so a later recover-store can finish without invoking reset-password.js.
  if [ "$recover_store_only" = "true" ]; then
    echo RETAIN
    return
  fi
  # A failed/skipped preflight may mean it found the retained bundle that this
  # gate exists to protect. Only a successful preflight makes a skipped rotate
  # safe to scrub (for example, password generation/upload then failed).
  if [ "$rotate" = "skipped" ] && [ "$preflight" = "success" ]; then
    echo SCRUB
    return
  fi
  if reset_did_not_happen; then
    echo SCRUB
    return
  fi
  echo RETAIN
}

case "${1:-}" in
  store-decision) store_decision ;;
  scrub-decision) scrub_decision ;;
  *)
    echo "usage: kuma-rotation-recovery-gate.sh store-decision|scrub-decision" >&2
    exit 2
    ;;
esac
