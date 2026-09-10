#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
INSTALLER="$SCRIPT_DIR/install-migration-artifact.sh"
REAL_MV=$(command -v mv)

TEST_ROOT=$(mktemp -d)
trap 'rm -rf "$TEST_ROOT"' EXIT

make_bundle() {
  local root=$1
  local bundle=$2
  mkdir -p "$root/apps/api/migrations"
  printf 'h1:test\n' >"$root/apps/api/migrations/atlas.sum"
  printf '%s\n' 'SELECT 1;' >"$root/apps/api/migrations/20260101000000_current.sql"
  tar -czf "$bundle" -C "$root" ./apps/api/migrations
}

test_exact_replacement_removes_stale_files() {
  local case_root="$TEST_ROOT/exact"
  local target="$case_root/deploy/apps/api/migrations"
  local bundle="$case_root/bundle.tgz"
  mkdir -p "$target" "$case_root/source"
  printf '%s\n' 'SELECT stale;' >"$target/20250101000000_stale.sql"
  make_bundle "$case_root/source" "$bundle"

  "$INSTALLER" "$bundle" "$target"

  test -f "$target/20260101000000_current.sql"
  test ! -e "$target/20250101000000_stale.sql"
  test -f "$target/atlas.sum"
}

test_invalid_bundle_preserves_deployed_directory() {
  local case_root="$TEST_ROOT/invalid"
  local target="$case_root/deploy/apps/api/migrations"
  local bundle="$case_root/bundle.tgz"
  mkdir -p "$target" "$case_root/source/apps/api/migrations"
  printf '%s\n' 'SELECT old;' >"$target/20250101000000_old.sql"
  tar -czf "$bundle" -C "$case_root/source" ./apps/api/migrations

  if "$INSTALLER" "$bundle" "$target" 2>/dev/null; then
    echo "invalid bundle unexpectedly succeeded" >&2
    return 1
  fi

  test -f "$target/20250101000000_old.sql"
}

test_install_failure_restores_backup() {
  local case_root="$TEST_ROOT/rollback"
  local target="$case_root/deploy/apps/api/migrations"
  local bundle="$case_root/bundle.tgz"
  local fake_bin="$case_root/bin"
  mkdir -p "$target" "$case_root/source" "$fake_bin"
  printf '%s\n' 'SELECT old;' >"$target/20250101000000_old.sql"
  make_bundle "$case_root/source" "$bundle"

  cat >"$fake_bin/mv" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
count_file=${MV_COUNT_FILE:?}
count=0
if [[ -f "$count_file" ]]; then
  count=$(<"$count_file")
fi
count=$((count + 1))
printf '%s' "$count" >"$count_file"
if [[ "$count" -eq 2 ]]; then
  exit 1
fi
exec "${REAL_MV:?}" "$@"
EOF
  chmod +x "$fake_bin/mv"

  if PATH="$fake_bin:$PATH" MV_COUNT_FILE="$case_root/mv-count" REAL_MV="$REAL_MV" \
    "$INSTALLER" "$bundle" "$target" 2>/dev/null; then
    echo "injected installation failure unexpectedly succeeded" >&2
    return 1
  fi

  test -f "$target/20250101000000_old.sql"
  test ! -e "$target/20260101000000_current.sql"
  test -z "$(find "$(dirname "$target")" -maxdepth 1 -name '.migrations-*' -print -quit)"
}

test_first_install() {
  local case_root="$TEST_ROOT/first"
  local target="$case_root/deploy/apps/api/migrations"
  local bundle="$case_root/bundle.tgz"
  mkdir -p "$(dirname "$target")" "$case_root/source"
  make_bundle "$case_root/source" "$bundle"

  "$INSTALLER" "$bundle" "$target"
  test -f "$target/20260101000000_current.sql"
}

test_exact_replacement_removes_stale_files
test_invalid_bundle_preserves_deployed_directory
test_install_failure_restores_backup
test_first_install

echo "install-migration-artifact self-test passed"
