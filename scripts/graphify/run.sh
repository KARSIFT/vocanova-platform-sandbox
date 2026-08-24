#!/usr/bin/env bash
# VOC-112-T03 — code-only Graphify extract wrapper (opt-in pilot).
# Requires a valid locked runtime from scripts/graphify/setup.sh.
# Never downloads packages, registers hooks, or enables provider auto-detection.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GRAPHIFY_HOME="$SCRIPT_DIR"
REPO_ROOT="$(cd "$GRAPHIFY_HOME/../.." && pwd)"

bash "$GRAPHIFY_HOME/check" >/dev/null

cli_relative="$(awk -F': ' '$1 == "graphify_cli_relative" {gsub(/^["'\'']|["'\'']$/, "", $2); print $2; exit}' "$GRAPHIFY_HOME/runtime-identity.yaml")"
GRAPHIFY_CLI="$GRAPHIFY_HOME/$cli_relative"

TARGET_INPUT="${1:-$REPO_ROOT}"
OUTPUT_DIR="${GRAPHIFY_OUTPUT_DIR:-$REPO_ROOT/graphify-out}"

[[ -d "$TARGET_INPUT" ]] || {
  echo "graphify-pilot run: target must be an existing directory" >&2
  exit 1
}

TARGET="$(realpath "$TARGET_INPUT")"
case "$TARGET" in
  "$REPO_ROOT"|"$REPO_ROOT"/*) ;;
  *)
    echo "graphify-pilot run: target must remain inside the repository" >&2
    exit 1
    ;;
esac

RUNTIME_HOME="$GRAPHIFY_HOME/.runtime-home"
mkdir -p "$RUNTIME_HOME/cache"

# An empty environment is the fail-closed boundary: provider variables,
# cloud credential chains, user profiles, and agent-session secrets are absent.
exec env -i \
  HOME="$RUNTIME_HOME" \
  XDG_CACHE_HOME="$RUNTIME_HOME/cache" \
  PATH="$GRAPHIFY_HOME/.venv/bin:/usr/bin:/bin" \
  LANG=C.UTF-8 \
  LC_ALL=C.UTF-8 \
  PYTHONNOUSERSITE=1 \
  PYTHONDONTWRITEBYTECODE=1 \
  GRAPHIFY_QUERY_LOG_DISABLE=1 \
  GRAPHIFY_QUERY_LOG_ENABLE=0 \
  GRAPHIFY_QUERY_LOG=/dev/null \
  GRAPHIFY_OUT="$OUTPUT_DIR" \
  "$GRAPHIFY_CLI" extract "$TARGET" --code-only
