#!/usr/bin/env bash
set -euo pipefail

# VOC-086-T04: changed-file enforcement for monitoring_impact.
# GitHub Actions does not set GITHUB_BASE_SHA. PR range must come from
# --base/--head (same pattern as classify-change-risk.sh), --files-from,
# or a parseable GITHUB_EVENT_PATH. A pull_request event without a resolved
# range is fail-closed so CI cannot skip package/route checks.

base=""
head=""
files_from=""
declarations_only=false

usage() {
  echo "Usage: $0 [--base SHA --head SHA | --files-from FILE] [--declarations-only]" >&2
}

while (($#)); do
  case "$1" in
    --base) base="${2:-}"; shift 2 ;;
    --head) head="${2:-}"; shift 2 ;;
    --files-from) files_from="${2:-}"; shift 2 ;;
    --declarations-only) declarations_only=true; shift ;;
    -h|--help) usage; exit 0 ;;
    *) usage; echo "Unknown argument: $1" >&2; exit 2 ;;
  esac
done

read_pr_range_from_event() {
  python3 - <<'PY'
import json
import os
import sys

path = os.environ.get("GITHUB_EVENT_PATH", "")
if not path:
    sys.exit(1)
with open(path, encoding="utf-8") as handle:
    event = json.load(handle)
pull_request = event.get("pull_request") or {}
base = ((pull_request.get("base") or {}).get("sha") or "").strip()
head = ((pull_request.get("head") or {}).get("sha") or "").strip()
if not base or not head:
    sys.exit(1)
sys.stdout.write(f"{base}\n{head}\n")
PY
}

if [[ -z "$files_from" && ( -z "$base" || -z "$head" ) && "${GITHUB_EVENT_NAME:-}" == "pull_request" ]]; then
  if event_range="$(read_pr_range_from_event)"; then
    mapfile -t event_shas <<<"$event_range"
    base="${event_shas[0]:-}"
    head="${event_shas[1]:-}"
  fi
fi

if [[ -z "$files_from" && ( -z "$base" || -z "$head" ) && "${GITHUB_EVENT_NAME:-}" == "pull_request" ]]; then
  echo "pull_request monitoring_impact validation requires --base/--head, --files-from, or a parseable GITHUB_EVENT_PATH" >&2
  exit 1
fi

declare -a files=()
if [[ -n "$files_from" ]]; then
  mapfile -t files < "$files_from"
elif [[ -n "$base" && -n "$head" ]]; then
  mapfile -t files < <(git diff --no-renames --name-only --diff-filter=ACDMRTUXB "$base...$head")
else
  mapfile -t files < <(
    {
      git diff --no-renames --name-only --diff-filter=ACDMRTUXB
      git diff --cached --no-renames --name-only --diff-filter=ACDMRTUXB
      git ls-files --others --exclude-standard
    } | sort -u
  )
fi

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
validator="$repository_root/infra/monitoring/validate-monitoring-impact.mjs"
changed_files_file="$(mktemp)"
trap 'rm -f "$changed_files_file"' EXIT

if ((${#files[@]} > 0)); then
  printf '%s\n' "${files[@]}" > "$changed_files_file"
fi

args=(--repository-root "$repository_root")
if [[ -s "$changed_files_file" ]]; then
  args+=(--changed-files-file "$changed_files_file")
fi
if [[ "$declarations_only" == true ]]; then
  args+=(--declarations-only)
fi

node "$validator" "${args[@]}"
