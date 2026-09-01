#!/usr/bin/env bash
# Shared changed-file loading for governance scripts (VOC-146).
# Source this file; do not execute directly.

GOVERNANCE_CHANGED_FILES=()

resolve_governance_commit() {
  local rev="$1"
  git rev-parse --verify --end-of-options "${rev}^{commit}"
}

# load_changed_files files_from base head
# Populates GOVERNANCE_CHANGED_FILES. Returns nonzero on partial range,
# unresolved commit, or invalid three-dot diff range.
load_changed_files() {
  local files_from="${1:-}"
  local base="${2:-}"
  local head="${3:-}"

  GOVERNANCE_CHANGED_FILES=()

  if [[ ( -n "$base" && -z "$head" ) || ( -z "$base" && -n "$head" ) ]]; then
    echo "A changed-file range requires both --base and --head." >&2
    return 1
  fi

  if [[ -n "$files_from" ]]; then
    mapfile -t GOVERNANCE_CHANGED_FILES < "$files_from"
    return 0
  fi

  if [[ -n "$base" && -n "$head" ]]; then
    local resolved_base resolved_head diff_file
    if ! resolved_base="$(resolve_governance_commit "$base")"; then
      echo "Unable to resolve --base commit: $base" >&2
      return 1
    fi
    if ! resolved_head="$(resolve_governance_commit "$head")"; then
      echo "Unable to resolve --head commit: $head" >&2
      return 1
    fi

    diff_file="$(mktemp)"
    if ! git diff --no-renames --name-only --diff-filter=ACDMRTUXB \
      "$resolved_base...$resolved_head" >"$diff_file"; then
      rm -f "$diff_file"
      echo "Unable to load changed files for range ${base}...${head}" >&2
      return 1
    fi

    if [[ -s "$diff_file" ]]; then
      mapfile -t GOVERNANCE_CHANGED_FILES < "$diff_file"
    fi
    rm -f "$diff_file"
    return 0
  fi

  mapfile -t GOVERNANCE_CHANGED_FILES < <(
    {
      git diff --no-renames --name-only --diff-filter=ACDMRTUXB
      git diff --cached --no-renames --name-only --diff-filter=ACDMRTUXB
      git ls-files --others --exclude-standard
    } | sort -u
  )
}
