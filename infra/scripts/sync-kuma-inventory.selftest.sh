#!/usr/bin/env bash
set -euo pipefail

# VOC-086-T02 — static checks for sync-kuma-inventory.sh (no live Kuma required).
#
# Usage: infra/scripts/sync-kuma-inventory.selftest.sh

repo_root="$(cd "$(dirname "$0")/../.." && pwd)"
sync_script="$repo_root/infra/scripts/sync-kuma-inventory.sh"
source="$(cat "$sync_script")"

for pattern in 'kuma\.db' 'sqlite' '/app/data'; do
  if grep -Eiq "$pattern" "$sync_script"; then
    echo "sync-kuma-inventory.sh must not reference SQLite deployment paths ($pattern)" >&2
    exit 1
  fi
done

if ! grep -q 'vocanova-monitoring-net' "$sync_script"; then
  echo "sync-kuma-inventory.sh must attach to vocanova-monitoring-net" >&2
  exit 1
fi

if ! grep -q 'sync-kuma.mjs' "$sync_script"; then
  echo "sync-kuma-inventory.sh must invoke sync-kuma.mjs" >&2
  exit 1
fi

if ! grep -q 'socket.io-client' "$sync_script"; then
  echo "sync-kuma-inventory.sh must install socket.io-client in the disposable Node container" >&2
  exit 1
fi

echo "sync-kuma-inventory.selftest.sh: PASS"
