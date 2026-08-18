#!/usr/bin/env bash
set -euo pipefail

# VOC-086-T02 — disposable harness for kuma-rotate-credentials.sh.
#
# Usage: infra/scripts/kuma-rotate-credentials.selftest.sh

repo_root="$(cd "$(dirname "$0")/../.." && pwd)"
rotate_script="$repo_root/infra/scripts/kuma-rotate-credentials.sh"
test_root="$(mktemp -d)"

cleanup() {
  rm -rf "$test_root"
}
trap cleanup EXIT

cat > "$test_root/docker" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
if [ "$1" = "inspect" ]; then
  exit 0
fi
if [ "$1" = "exec" ] && [ "$2" = "-i" ]; then
  container="$3"
  shift 3
  if [ "$1" = "node" ] && [ "$2" = "/app/extra/reset-password.js" ]; then
    read -r line1 || true
    read -r line2 || true
    if [ -z "${line1:-}" ] || [ "$line1" != "$line2" ]; then
      echo "password mismatch" >&2
      exit 2
    fi
    echo "== Uptime Kuma Reset Password Tool =="
    echo "Found user: harness-admin"
    echo "Password reset successfully."
    exit 0
  fi
fi
echo "unexpected docker invocation: $*" >&2
exit 99
EOF
chmod +x "$test_root/docker"

export PATH="$test_root:$PATH"
export KUMA_CONTAINER="fake-kuma"
export KUMA_NEW_PASSWORD="harness-strong-password-not-for-logs"
export KUMA_ROTATE_METADATA_FILE="$test_root/kuma-rotate-metadata.env"

output="$( "$rotate_script" )"
if ! grep -q 'KUMA_USERNAME=harness-admin' <<<"$output"; then
  echo "expected KUMA_USERNAME line in output" >&2
  echo "$output" >&2
  exit 1
fi
if grep -q 'harness-strong-password-not-for-logs' <<<"$output"; then
  echo "password must not appear in stdout" >&2
  exit 1
fi
if [ ! -f "$KUMA_ROTATE_METADATA_FILE" ]; then
  echo "expected metadata file to be written" >&2
  exit 1
fi
if ! grep -qx 'KUMA_USERNAME=harness-admin' "$KUMA_ROTATE_METADATA_FILE"; then
  echo "metadata file missing preserved username" >&2
  exit 1
fi

cat > "$test_root/docker" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
if [ "$1" = "inspect" ]; then
  exit 0
fi
if [ "$1" = "exec" ] && [ "$2" = "-i" ]; then
  read -r _line1 || true
  read -r _line2 || true
  echo "== Uptime Kuma Reset Password Tool =="
  echo "Password reset successfully."
  exit 0
fi
echo "unexpected docker invocation: $*" >&2
exit 99
EOF
chmod +x "$test_root/docker"

set +e
missing_user_output="$(KUMA_NEW_PASSWORD="another-harness-password" "$rotate_script" 2>&1)"
missing_user_status=$?
set -e
if [ "$missing_user_status" -eq 0 ]; then
  echo "rotate must fail closed when username is missing from reset output" >&2
  exit 1
fi
if grep -q 'another-harness-password' <<<"$missing_user_output"; then
  echo "password must not appear in failure output" >&2
  exit 1
fi

echo "kuma-rotate-credentials.selftest.sh: PASS"
