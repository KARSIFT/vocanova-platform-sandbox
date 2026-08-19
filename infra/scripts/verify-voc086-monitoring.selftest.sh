#!/usr/bin/env bash
set -euo pipefail

# VOC-086-T05 — disposable harness for monitoring verification tooling.
#
# Validates script wiring without live Kuma credentials. Full external closure
# uses verify-voc086-monitoring.sh against production hostnames.
#
# Usage: infra/scripts/verify-voc086-monitoring.selftest.sh

repo_root="$(cd "$(dirname "$0")/../.." && pwd)"

for script in \
  "$repo_root/infra/scripts/verify-voc086-monitoring.sh" \
  "$repo_root/infra/scripts/prove-kuma-inventory.sh"; do
  if [ ! -f "$script" ]; then
    echo "missing: $script" >&2
    exit 1
  fi
  bash -n "$script"
done

node --test "$repo_root/scripts/foundation/voc086-live-verification.test.mjs"

echo "verify-voc086-monitoring.selftest.sh: PASS"
