#!/usr/bin/env bash
# Stable public identity shared by staging and production for a checked-out commit.
set -euo pipefail
if [[ "$(git rev-parse --is-shallow-repository)" != false ]]; then
  echo "Release version calculation requires full Git history (fetch-depth: 0)" >&2
  exit 1
fi
if ! git diff --quiet HEAD -- VERSION || ! git diff --cached --quiet HEAD -- VERSION; then
  echo "VERSION must match the checked-out release commit" >&2
  exit 1
fi
release_base="$(git show HEAD:VERSION)"
if ! [[ "$release_base" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; then
  echo "VERSION must contain a stable semantic version without leading zeroes" >&2
  exit 1
fi
release_major="${BASH_REMATCH[1]}"
release_minor="${BASH_REMATCH[2]}"
release_patch="${BASH_REMATCH[3]}"
if (( ${#release_major} > 9 || ${#release_minor} > 9 || ${#release_patch} > 9 )); then
  echo "VERSION components must have at most nine digits" >&2
  exit 1
fi
# Only the mainline matters: a multi-commit PR is one release when merged.
# Updating VERSION deliberately starts a new major/minor/patch base.
release_base_commit="$(git log --first-parent -1 --format=%H -- VERSION)"
if [[ -z "$release_base_commit" ]]; then
  echo "VERSION must have a committed first-parent history" >&2
  exit 1
fi
release_distance="$(git rev-list --first-parent --count "$release_base_commit..HEAD")"
release_version="$release_major.$release_minor.$((release_patch + release_distance))"
release_commit="$(git rev-parse HEAD)"
if ! [[ "$release_commit" =~ ^[0-9a-f]{40}$ ]]; then
  echo "The checked-out release commit must be a full SHA" >&2
  exit 1
fi
echo "version=$release_version"
echo "commit=$release_commit"
echo "built_at=$(date -u +'%Y-%m-%dT%H:%M:%SZ')"
