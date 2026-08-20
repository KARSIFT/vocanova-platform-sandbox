#!/usr/bin/env bash
set -euo pipefail

# VOC-092-T00 — local operator rehearsal for the controlled-signup OAuth
# callback E2E harness. Runs the Go integration tests against disposable
# loopback Postgres only; never accepts staging/production hosts or real
# credentials.
#
# Usage:
#   run-controlled-signup-oauth-e2e.sh
#
# Requirements: Docker on PATH, Go toolchain, repository root as cwd or
# any descendant path.

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/../.." && pwd)"

for arg in "$@"; do
  case "$arg" in
    *vocanova.site*|*vocanova.com*|*://api-*|*://staging*|*://production*)
      echo "refusing to run with staging/production host argument: $arg" >&2
      exit 1
      ;;
  esac
done

if [ "$#" -ne 0 ]; then
  echo "usage: $0 (no arguments)" >&2
  exit 1
fi

cd "$repo_root/apps/api"
exec go test ./app/api/... -run ControlledSignupOAuth -count=1 -v
