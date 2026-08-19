#!/usr/bin/env bash
set -euo pipefail

# VOC-086-T05 — read-only Socket.IO proof that live Kuma matches monitors.yaml.
#
# Runs prove-kuma-inventory.mjs in a disposable container on
# vocanova-monitoring-net. Requires KUMA_USERNAME/KUMA_PASSWORD (GitHub
# monitoring environment secrets). Read-only; never mutates inventory.
#
# Usage (on shared host after credentials exist):
#   KUMA_USERNAME=… KUMA_PASSWORD=… infra/scripts/prove-kuma-inventory.sh

KUMA_CONTAINER="${KUMA_CONTAINER:-vocanova-uptime-kuma}"
KUMA_URL="${KUMA_URL:-http://vocanova-uptime-kuma:3001}"
MONITORING_SYNC_NODE_IMAGE="${MONITORING_SYNC_NODE_IMAGE:-node:24-bookworm-slim}"
MONITORING_SYNC_SOCKET_IO_VERSION="${MONITORING_SYNC_SOCKET_IO_VERSION:-4.8.1}"

script_dir="$(cd "$(dirname "$0")" && pwd)"
MONITORING_SYNC_ROOT="${MONITORING_SYNC_ROOT:-$script_dir/../monitoring}"

if [ -z "${KUMA_USERNAME:-}" ] || [ -z "${KUMA_PASSWORD:-}" ]; then
  echo "KUMA_USERNAME and KUMA_PASSWORD environment variables are required" >&2
  exit 1
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required to prove Kuma inventory" >&2
  exit 1
fi

if ! docker inspect "$KUMA_CONTAINER" >/dev/null 2>&1; then
  echo "Kuma container ${KUMA_CONTAINER} is not present on this host" >&2
  exit 1
fi

if ! docker network inspect vocanova-monitoring-net >/dev/null 2>&1; then
  echo "vocanova-monitoring-net is missing; monitoring topology is not converged" >&2
  exit 1
fi

if [ ! -f "$MONITORING_SYNC_ROOT/prove-kuma-inventory.mjs" ]; then
  echo "prove-kuma-inventory.mjs not found under MONITORING_SYNC_ROOT=${MONITORING_SYNC_ROOT}" >&2
  exit 1
fi

work_dir="$(mktemp -d)"
cleanup() {
  rm -rf "$work_dir"
}
trap cleanup EXIT

cp -a "$MONITORING_SYNC_ROOT/." "$work_dir/"

docker run --rm \
  --network vocanova-monitoring-net \
  -e KUMA_URL="$KUMA_URL" \
  -e KUMA_USERNAME \
  -e KUMA_PASSWORD \
  -v "$work_dir:/work" \
  -w /work \
  "$MONITORING_SYNC_NODE_IMAGE" \
  bash -c "set -euo pipefail; npm install --no-save socket.io-client@${MONITORING_SYNC_SOCKET_IO_VERSION}; node prove-kuma-inventory.mjs"
