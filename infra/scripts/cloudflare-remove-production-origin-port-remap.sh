#!/usr/bin/env bash
set -euo pipefail

# VOC-067-T05 — repository-driven Cloudflare API cutover (VOC-067-DEP-03).
#
# Removes Origin Rules that remap production hostnames to destination port
# 8443, restoring ordinary edge :443 → origin :443. Supports verify-only
# and rollback (restore port 8443) per T00.
#
# Live API (all modes, including --verify-only) requires a token. Missing
# credentials is a TEST-06 failure, not a pass. Offline selftest uses
# VOC067_OFFLINE_RULESET_FILE and does not call Cloudflare.
#
# Environment (accepts production-prefixed names from deploy-production.yml):
#   Token precedence (first non-empty wins):
#     CLOUDFLARE_API_TOKEN
#     PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN — VOC-072 cutover secret
#       (Zone Read + Origin Rules Edit on vocanova.site; wired in
#       deploy-production.yml voc067-cloudflare-cutover / --apply steps)
#     PRODUCTION_CLOUDFLARE_API_TOKEN — legacy fallback during rotation only;
#       Workers-AI-scoped token cannot resolve the zone (issue #535)
#   Required for live --verify-only / --apply / --restore (unless offline fixture)
#   CLOUDFLARE_ZONE_NAME (default: vocanova.site)
#   VOC067_PRODUCTION_WEB_HOST (default: production.vocanova.site)
#   VOC067_PRODUCTION_API_HOST (default: api-production.vocanova.site)
#   VOC067_ORIGIN_PORT (default: 8443 — the cutover bridge port to remove)
#   VOC067_OFFLINE_RULESET_FILE — fixture JSON (Cloudflare GET shape); skips API
#   VOC067_OFFLINE_OUTPUT — where --apply/--restore writes the updated ruleset
#
# Usage:
#   cloudflare-remove-production-origin-port-remap.sh --verify-only
#   cloudflare-remove-production-origin-port-remap.sh --apply
#   cloudflare-remove-production-origin-port-remap.sh --restore
#
# After --apply or when --verify-only reports no remap, run:
#   infra/scripts/verify-voc067-cutover.sh
# Do not retire vocanova-production-nginx until --verify-only exits 0.

API_BASE="https://api.cloudflare.com/client/v4"
ZONE_NAME="${CLOUDFLARE_ZONE_NAME:-vocanova.site}"
WEB_HOST="${VOC067_PRODUCTION_WEB_HOST:-production.vocanova.site}"
API_HOST="${VOC067_PRODUCTION_API_HOST:-api-production.vocanova.site}"
ORIGIN_PORT="${VOC067_ORIGIN_PORT:-8443}"
MUTATE_PY="$(cd "$(dirname "$0")" && pwd)/cloudflare_origin_port_remap.py"

TOKEN="${CLOUDFLARE_API_TOKEN:-${PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN:-${PRODUCTION_CLOUDFLARE_API_TOKEN:-}}}"
OFFLINE_FILE="${VOC067_OFFLINE_RULESET_FILE:-}"

mode=""
case "${1:-}" in
  --verify-only) mode=verify ;;
  --apply) mode=apply ;;
  --restore) mode=restore ;;
  *)
    echo "usage: $0 --verify-only | --apply | --restore" >&2
    exit 1
    ;;
esac

if [ -z "$OFFLINE_FILE" ] && [ -z "$TOKEN" ]; then
  echo "ERROR: CLOUDFLARE_API_TOKEN (or PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN / PRODUCTION_CLOUDFLARE_API_TOKEN) is required for --verify-only/--apply/--restore" >&2
  echo "Missing Cloudflare credentials is a VOC-067-TEST-06 failure, not a pass." >&2
  exit 1
fi

if [ -n "$OFFLINE_FILE" ] && [ ! -f "$OFFLINE_FILE" ]; then
  echo "ERROR: VOC067_OFFLINE_RULESET_FILE not found: $OFFLINE_FILE" >&2
  exit 1
fi

cf_api() {
  local method="$1"
  local path="$2"
  local data="${3:-}"
  local tmp http_code
  tmp="$(mktemp)"
  if [ -n "$data" ]; then
    http_code="$(curl -sS -o "$tmp" -w "%{http_code}" -X "$method" \
      -H "Authorization: Bearer ${TOKEN}" \
      -H "Content-Type: application/json" \
      --data "$data" \
      "${API_BASE}${path}")"
  else
    http_code="$(curl -sS -o "$tmp" -w "%{http_code}" -X "$method" \
      -H "Authorization: Bearer ${TOKEN}" \
      "${API_BASE}${path}")"
  fi
  if [[ ! "$http_code" =~ ^2 ]]; then
    echo "ERROR: Cloudflare API ${method} ${path} returned HTTP ${http_code}:" >&2
    cat "$tmp" >&2
    rm -f "$tmp"
    exit 1
  fi
  cat "$tmp"
  rm -f "$tmp"
}

resolve_zone_id() {
  ZONE_NAME="$ZONE_NAME" python3 -c '
import json, os, sys
zone_name = os.environ["ZONE_NAME"]
payload = json.load(sys.stdin)
zones = payload.get("result", [])
if zones:
    print(zones[0]["id"])
    sys.exit(0)
errors = payload.get("errors", [])
if errors:
    msg = errors[0].get("message", "Cloudflare API error")
    raise SystemExit(
        f"ERROR: Cloudflare API rejected zone lookup for {zone_name!r}: {msg}"
    )
if not payload.get("success", True):
    raise SystemExit(f"ERROR: Cloudflare zone lookup for {zone_name!r} failed")
raise SystemExit(
    f"ERROR: zone {zone_name!r} not found or token lacks Zone Read for that zone "
    "(empty GET /zones result — check CLOUDFLARE_ZONE_NAME and use "
    "PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN, not the Workers-AI-scoped "
    "PRODUCTION_CLOUDFLARE_API_TOKEN)"
)
' <<< "$(cf_api GET "/zones?name=${ZONE_NAME}")"
}

mutate_ruleset_json() {
  local action="$1"
  shift
  python3 "$MUTATE_PY" "$action" "$WEB_HOST" "$API_HOST" "$ORIGIN_PORT" "$@"
}

load_ruleset_json() {
  if [ -n "$OFFLINE_FILE" ]; then
    cat "$OFFLINE_FILE"
    return
  fi
  ZONE_ID="$(resolve_zone_id)"
  cf_api GET "/zones/${ZONE_ID}/rulesets/phases/http_request_origin/entrypoint"
}

write_offline_output() {
  local body="$1"
  if [ -n "${VOC067_OFFLINE_OUTPUT:-}" ]; then
    printf '%s' "$body" > "$VOC067_OFFLINE_OUTPUT"
  fi
}

ruleset_json="$(load_ruleset_json)"
if [ -z "$OFFLINE_FILE" ]; then
  RULESET_ID="$(python3 -c 'import json,sys; print(json.load(sys.stdin)["result"]["id"])' <<<"$ruleset_json")"
else
  RULESET_ID="offline"
fi

case "$mode" in
  verify)
    printf '%s' "$ruleset_json" | mutate_ruleset_json verify
    ;;
  apply)
    if [ -n "$OFFLINE_FILE" ]; then
      update_payload="$(printf '%s' "$ruleset_json" | mutate_ruleset_json remove --full-ruleset || true)"
      if [ -z "${update_payload:-}" ] || [ "$update_payload" = "NOOP" ]; then
        echo "Remap already absent; nothing to apply."
        printf '%s' "$ruleset_json" | mutate_ruleset_json verify
      else
        echo "Applying origin-port remap removal to ruleset ${RULESET_ID} (offline)..."
        write_offline_output "$update_payload"
        printf '%s' "$update_payload" | mutate_ruleset_json verify
      fi
    else
      update_payload="$(printf '%s' "$ruleset_json" | mutate_ruleset_json remove || true)"
      if [ -z "${update_payload:-}" ] || [ "$update_payload" = "NOOP" ]; then
        echo "Remap already absent; nothing to apply."
      else
        echo "Applying origin-port remap removal to ruleset ${RULESET_ID}..."
        cf_api PUT "/zones/${ZONE_ID}/rulesets/${RULESET_ID}" "$update_payload" >/dev/null
        echo "Cloudflare API update succeeded."
      fi
      printf '%s' "$(cf_api GET "/zones/${ZONE_ID}/rulesets/phases/http_request_origin/entrypoint")" | mutate_ruleset_json verify
    fi
    ;;
  restore)
    if [ -n "$OFFLINE_FILE" ]; then
      update_payload="$(printf '%s' "$ruleset_json" | mutate_ruleset_json restore --full-ruleset || true)"
      if [ -z "${update_payload:-}" ] || [ "$update_payload" = "NOOP" ]; then
        echo "Remap already present; nothing to restore."
      else
        echo "Restoring origin-port remap to ${ORIGIN_PORT} on ruleset ${RULESET_ID} (offline)..."
        write_offline_output "$update_payload"
        echo "Rollback mutation succeeded (offline)."
      fi
    else
      update_payload="$(printf '%s' "$ruleset_json" | mutate_ruleset_json restore || true)"
      if [ -z "${update_payload:-}" ] || [ "$update_payload" = "NOOP" ]; then
        echo "Remap already present; nothing to restore."
      else
        echo "Restoring origin-port remap to ${ORIGIN_PORT} on ruleset ${RULESET_ID}..."
        cf_api PUT "/zones/${ZONE_ID}/rulesets/${RULESET_ID}" "$update_payload" >/dev/null
        echo "Rollback API update succeeded. Re-verify :8443 path before closing incident."
      fi
    fi
    ;;
esac
