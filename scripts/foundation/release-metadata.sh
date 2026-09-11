#!/usr/bin/env bash
# Public identity shared by the API and web images in one deployment build.
set -euo pipefail
release_version="$(tr -d '\r\n' < VERSION)"
if ! [[ "$release_version" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[A-Za-z0-9.-]+)?$ ]]; then
  echo "VERSION must contain a semantic version" >&2
  exit 1
fi
release_commit="$(git rev-parse HEAD)"
if ! [[ "$release_commit" =~ ^[0-9a-f]{40}$ ]]; then
  echo "The checked-out release commit must be a full SHA" >&2
  exit 1
fi
echo "version=$release_version"
echo "commit=$release_commit"
echo "built_at=$(date -u +'%Y-%m-%dT%H:%M:%SZ')"
