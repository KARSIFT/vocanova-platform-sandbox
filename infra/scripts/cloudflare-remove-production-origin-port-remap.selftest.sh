#!/usr/bin/env bash
set -euo pipefail

# VOC-067-T05 — offline harness for cloudflare-remove-production-origin-port-remap.sh
# (no live Cloudflare calls). Exercises the production script path, not a
# duplicated mutation copy.
#
# Usage: infra/scripts/cloudflare-remove-production-origin-port-remap.selftest.sh

repo_root="$(cd "$(dirname "$0")/../.." && pwd)"
cutover_script="$repo_root/infra/scripts/cloudflare-remove-production-origin-port-remap.sh"
verify_script="$repo_root/infra/scripts/verify-voc067-cutover.sh"

sample_ruleset="$(mktemp)"
removed_ruleset="$(mktemp)"
restored_ruleset="$(mktemp)"
trap 'rm -f "$sample_ruleset" "$removed_ruleset" "$restored_ruleset"' EXIT
cat >"$sample_ruleset" <<'JSON'
{
  "result": {
    "id": "ruleset-test",
    "rules": [
      {
        "id": "rule-prod-web",
        "expression": "(http.host eq \"production.vocanova.site\")",
        "action": "route",
        "action_parameters": {
          "origin": { "port": 8443 }
        }
      },
      {
        "id": "rule-staging",
        "expression": "(http.host eq \"staging.vocanova.site\")",
        "action": "route",
        "action_parameters": {}
      }
    ]
  }
}
JSON

run_cutover() {
  VOC067_OFFLINE_RULESET_FILE="$1" VOC067_OFFLINE_OUTPUT="${2:-}" \
    "$cutover_script" "$3"
}

echo "selftest: live --verify-only without token fails closed (TEST-06)"
set +e
token_err="$(
  env -u CLOUDFLARE_API_TOKEN -u PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN \
    -u PRODUCTION_CLOUDFLARE_API_TOKEN -u VOC067_OFFLINE_RULESET_FILE \
    "$cutover_script" --verify-only 2>&1
)"
token_rc=$?
set -e
if [ "$token_rc" -eq 1 ] && echo "$token_err" | grep -q "is required"; then
  echo "PASS"
else
  echo "FAIL: expected token-required exit 1, got rc=$token_rc err=$token_err" >&2
  exit 1
fi

echo "selftest: verify detects port 8443 remap (production script, offline)"
set +e
verify_out="$(run_cutover "$sample_ruleset" "" --verify-only 2>&1)"
verify_rc=$?
set -e
if [ "$verify_rc" -eq 2 ] && echo "$verify_out" | grep -q "FOUND:"; then
  echo "PASS"
else
  echo "FAIL: expected FOUND exit 2, got rc=$verify_rc out=$verify_out" >&2
  exit 1
fi

echo "selftest: apply strips port from production rule only (production script)"
run_cutover "$sample_ruleset" "$removed_ruleset" --apply >/dev/null
python3 - "$removed_ruleset" <<'PY'
import json, sys
ruleset = json.load(open(sys.argv[1], encoding="utf-8"))
rules = ruleset["result"]["rules"]
prod = [r for r in rules if "production.vocanova.site" in r.get("expression", "")]
staging = [r for r in rules if "staging.vocanova.site" in r.get("expression", "")]
assert prod and "port" not in (prod[0].get("action_parameters") or {}).get("origin", {})
assert staging, "staging rule must be preserved"
print("ok")
PY
echo "PASS"

echo "selftest: verify on stripped ruleset reports remap absent"
set +e
absent_out="$(run_cutover "$removed_ruleset" "" --verify-only 2>&1)"
absent_rc=$?
set -e
if [ "$absent_rc" -eq 0 ] && echo "$absent_out" | grep -q "OK: no origin rules remap"; then
  echo "PASS"
else
  echo "FAIL: expected OK exit 0, got rc=$absent_rc out=$absent_out" >&2
  exit 1
fi

echo "selftest: restore re-adds port 8443 (production script)"
run_cutover "$removed_ruleset" "$restored_ruleset" --restore >/dev/null
python3 - "$restored_ruleset" <<'PY'
import json, sys
ruleset = json.load(open(sys.argv[1], encoding="utf-8"))
rules = ruleset["result"]["rules"]
prod = [r for r in rules if "production.vocanova.site" in r.get("expression", "")]
assert prod[0]["action_parameters"]["origin"]["port"] == 8443
print("ok")
PY
echo "PASS"

echo "selftest: verify-voc067-cutover.sh syntax"
bash -n "$verify_script"
bash -n "$cutover_script"
echo "PASS"

echo "All cloudflare cutover selftests passed."
