#!/bin/sh

# Replace a deployed Atlas migration directory with the exact directory from a
# release bundle. Extracting a tarball over an existing directory is unsafe for
# Atlas because migrations removed or renamed in the bundle survive on disk and
# invalidate atlas.sum.

set -eu

usage() {
  echo "usage: $0 <bundle.tgz> <deployed-migrations-directory>" >&2
  exit 64
}

[ "$#" -eq 2 ] || usage

bundle=$1
target=$2
archive_path=./apps/api/migrations

[ -f "$bundle" ] || {
  echo "ERROR: migration bundle does not exist" >&2
  exit 1
}

parent=$(dirname "$target")
[ -d "$parent" ] || {
  echo "ERROR: migration parent directory does not exist" >&2
  exit 1
}

stage_root=$(mktemp -d "$parent/.migrations-stage.XXXXXX")
backup=""

cleanup() {
  rm -rf "$stage_root"
  # A backup is disposable only while an installed directory exists. If both
  # installation and rollback fail, retain the previous artifact for recovery.
  if [ -n "$backup" ] && [ -d "$target" ]; then
    rm -rf "$backup"
  fi
}
trap cleanup EXIT HUP INT TERM

tar -xzf "$bundle" -C "$stage_root" "$archive_path"
stage="$stage_root/apps/api/migrations"

if [ ! -f "$stage/atlas.sum" ] ||
  ! find "$stage" -maxdepth 1 -type f -name '*.sql' -print -quit | grep -q .; then
  echo "ERROR: staged migration artifact is incomplete; deployed migrations were not changed" >&2
  exit 1
fi

if [ -e "$target" ]; then
  backup=$(mktemp -d "$parent/.migrations-backup.XXXXXX")
  rmdir "$backup"
  if ! mv -T "$target" "$backup"; then
    echo "ERROR: could not create the migration rollback backup" >&2
    exit 1
  fi
fi

if ! mv -T "$stage" "$target"; then
  if [ -n "$backup" ] && mv -T "$backup" "$target"; then
    echo "ERROR: could not install the migration artifact; restored the previous directory" >&2
  elif [ -n "$backup" ]; then
    echo "ERROR: migration installation and rollback failed; previous artifact retained at $backup" >&2
  else
    echo "ERROR: could not install the migration artifact" >&2
  fi
  exit 1
fi

cleanup
trap - EXIT HUP INT TERM

