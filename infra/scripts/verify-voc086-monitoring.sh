#!/usr/bin/env bash
set -euo pipefail

# VOC-086-T05 — external availability + monitor-host + topology verification.
#
# Confirms:
#   1. Four app hostnames and monitor.vocanova.site (VOC-081 verifier).
#   2. All five canonical availability monitor URLs respond as expected.
#   3. Repository topology still has single shared-edge and no 8081/8443 publish.
#   4. External :8081/:8443 do not serve HTTP 2xx (bridge absence).
#   5. Optional read-only Socket.IO inventory proof when KUMA_* creds exist.
#
# Usage:
#   infra/scripts/verify-voc086-monitoring.sh
#   infra/scripts/verify-voc086-monitoring.sh --skip-socket-proof
#   infra/scripts/verify-voc086-monitoring.sh --skip-apps

repo_root="$(cd "$(dirname "$0")/../.." && pwd)"
voc081_script="$repo_root/infra/scripts/verify-voc081-monitor.sh"
prove_script="$repo_root/infra/scripts/prove-kuma-inventory.sh"

skip_apps=false
skip_socket_proof=false

while [ $# -gt 0 ]; do
  case "$1" in
    --skip-apps)
      skip_apps=true
      ;;
    --skip-socket-proof)
      skip_socket_proof=true
      ;;
    *)
      echo "usage: $0 [--skip-apps] [--skip-socket-proof]" >&2
      exit 1
      ;;
  esac
  shift
done

failures=0
record_fail() {
  echo "FAIL: $*" >&2
  failures=$((failures + 1))
}
record_pass() {
  echo "PASS: $*"
}

check_url() {
  local label="$1"
  local url="$2"
  local status
  status="$(curl -sS -L --max-time 25 -o /dev/null -w "%{http_code}" "$url" || echo "000")"
  if [[ "$status" =~ ^2 ]]; then
    record_pass "$label ($url) -> HTTP $status"
    return 0
  fi
  record_fail "$label ($url) -> HTTP $status (expected 2xx)"
  return 1
}

# Fail only when a retired bridge port still serves a successful app response.
# TCP-open with non-2xx (edge noise) is recorded but is not HTTP 2xx success.
check_bridge_absent() {
  local label="$1"
  local url="$2"
  local status
  status="$(curl -sS -L --max-time 8 -o /dev/null -w "%{http_code}" "$url" || true)"
  status="${status:-000}"
  if [[ "$status" =~ ^2 ]]; then
    record_fail "$label ($url) -> HTTP $status (retired 8081/8443 must not serve 2xx)"
    return 1
  fi
  record_pass "$label ($url) -> HTTP $status (not 2xx)"
  return 0
}

echo "VOC-086 monitoring verification — external + repository topology"

echo ""
echo "VOC-081 monitor hostname and app-tier regression"
if [ "$skip_apps" = true ]; then
  if ! "$voc081_script" --skip-apps; then
    record_fail "verify-voc081-monitor.sh --skip-apps failed"
  fi
else
  if ! "$voc081_script"; then
    record_fail "verify-voc081-monitor.sh failed"
  fi
fi

echo ""
echo "Canonical availability monitor URL probes"
check_url "kuma.availability.staging.web" "https://staging.vocanova.site/" || true
check_url "kuma.availability.staging.api-healthz" "https://api-staging.vocanova.site/healthz" || true
check_url "kuma.availability.production.web" "https://production.vocanova.site/" || true
check_url "kuma.availability.production.api-healthz" "https://api-production.vocanova.site/healthz" || true
check_url "kuma.availability.monitor-host" "https://monitor.vocanova.site/" || true

echo ""
echo "Retired production bridge ports must not serve HTTP 2xx"
check_bridge_absent "production web :8081" "https://production.vocanova.site:8081/" || true
check_bridge_absent "production api :8081" "http://production.vocanova.site:8081/" || true
check_bridge_absent "production web :8443" "https://production.vocanova.site:8443/" || true
check_bridge_absent "production api :8443" "http://production.vocanova.site:8443/" || true

echo ""
echo "Repository topology invariants (single shared-edge; no 8081/8443 publish)"
if ! node --test "$repo_root/scripts/foundation/voc081-monitoring-topology.test.mjs" >/dev/null 2>&1; then
  record_fail "voc081-monitoring-topology.test.mjs did not pass"
else
  record_pass "voc081-monitoring-topology repository assertions"
fi

if [ "$skip_socket_proof" = false ] && [ -n "${KUMA_USERNAME:-}" ] && [ -n "${KUMA_PASSWORD:-}" ]; then
  echo ""
  echo "Read-only Socket.IO inventory proof (host credentials present)"
  if [ -x "$prove_script" ]; then
    if ! "$prove_script"; then
      record_fail "prove-kuma-inventory.sh failed"
    fi
  else
    record_fail "missing executable $prove_script"
  fi
else
  echo ""
  echo "Skipping Socket.IO inventory proof (no KUMA_USERNAME/KUMA_PASSWORD in environment)"
fi

if [ "$failures" -gt 0 ]; then
  echo ""
  echo "$failures check(s) failed." >&2
  exit 1
fi

echo ""
echo "All VOC-086 monitoring verification checks passed."
