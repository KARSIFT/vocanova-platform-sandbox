#!/usr/bin/env bash
# VOC-112-T03 — explicit operator setup for the repository-local Graphify pilot.
# Creates scripts/graphify/.venv from the reviewed hash-locked requirements.lock only.
# Ordinary agent sessions and the check/run scripts never invoke this automatically.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GRAPHIFY_HOME="${GRAPHIFY_HOME:-$SCRIPT_DIR}"
IDENTITY_FILE="$GRAPHIFY_HOME/runtime-identity.yaml"
REQUIREMENTS_IN="$GRAPHIFY_HOME/requirements.in"
REQUIREMENTS_LOCK="$GRAPHIFY_HOME/requirements.lock"
VENV_DIR="$GRAPHIFY_HOME/.venv"

die() {
  echo "graphify-pilot setup: $*" >&2
  exit 1
}

require_file() {
  [[ -f "$1" ]] || die "missing required file: $1"
}

sha256_file() {
  sha256sum "$1" | awk '{print $1}'
}

read_identity_field() {
  local key="$1"
  awk -F': ' -v key="$key" '$1 == key {gsub(/^["'\'']|["'\'']$/, "", $2); print $2; exit}' "$IDENTITY_FILE"
}

require_file "$IDENTITY_FILE"
require_file "$REQUIREMENTS_IN"
require_file "$REQUIREMENTS_LOCK"

expected_in_hash="$(read_identity_field requirements_in_sha256)"
expected_lock_hash="$(read_identity_field requirements_lock_sha256)"
actual_in_hash="$(sha256_file "$REQUIREMENTS_IN")"
actual_lock_hash="$(sha256_file "$REQUIREMENTS_LOCK")"

[[ "$actual_in_hash" == "$expected_in_hash" ]] || die "requirements.in digest mismatch (expected $expected_in_hash, got $actual_in_hash)"
[[ "$actual_lock_hash" == "$expected_lock_hash" ]] || die "requirements.lock digest mismatch (expected $expected_lock_hash, got $actual_lock_hash)"

if ! command -v python3 >/dev/null 2>&1; then
  die "python3 is required (see runtime-identity.yaml python_min)"
fi

python3 -m venv "$VENV_DIR"
# Install only from the reviewed hash-locked lockfile — no unpinned bootstrap upgrades.
"$VENV_DIR/bin/python" -m pip install --require-hashes -r "$REQUIREMENTS_LOCK" >/dev/null

installed_version="$("$VENV_DIR/bin/graphify" --version 2>/dev/null | awk '{print $NF}')"
expected_version="$(read_identity_field pypi_version)"
[[ "$installed_version" == "$expected_version" ]] || die "installed graphify version mismatch (expected $expected_version, got $installed_version)"

echo "graphify-pilot setup: repository-local runtime ready at $VENV_DIR (graphify $installed_version)"
