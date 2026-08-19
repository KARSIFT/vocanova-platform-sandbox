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
    echo "Logged in."
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

gate_script="$repo_root/infra/scripts/kuma-rotation-recovery-gate.sh"
if [ ! -x "$gate_script" ]; then
  echo "kuma-rotation-recovery-gate.sh must exist and be executable" >&2
  exit 1
fi

assert_gate() {
  local command="$1"
  local expected="$2"
  local actual
  actual="$("$gate_script" "$command")"
  if [ "$actual" != "$expected" ]; then
    echo "recovery gate $command expected $expected, got $actual (rotate=$ROTATE_HOST_OUTCOME proof_matches=$PROOF_MATCHES host_reachable=$HOST_REACHABLE proof_fetched=$PROOF_FETCHED password_stored=$PASSWORD_STORED recover=$RECOVER_STORE_ONLY)" >&2
    exit 1
  fi
}

# VOC-087-TEST-10 matrix: post-reset rotate failure + failed proof fetch must
# retain the last password copy (High finding on VOC-087-T02 attempt 1).
export ROTATE_HOST_OUTCOME=failure
export PROOF_MATCHES=false
export HOST_REACHABLE=false
export PROOF_FETCHED=false
export PROOF_ABSENT=false
export PASSWORD_STORED=false
export USERNAME_STORED=false
export RECOVER_STORE_ONLY=false
assert_gate store-decision RETAIN
assert_gate scrub-decision RETAIN

export ROTATE_HOST_OUTCOME=success
export PROOF_MATCHES=false
export HOST_REACHABLE=false
export PROOF_FETCHED=false
export PROOF_ABSENT=false
export PASSWORD_STORED=false
assert_gate store-decision STORE
assert_gate scrub-decision RETAIN
export PASSWORD_STORED=true
export USERNAME_STORED=false
assert_gate scrub-decision RETAIN
export USERNAME_STORED=true
assert_gate scrub-decision SCRUB

export ROTATE_HOST_OUTCOME=failure
export PROOF_MATCHES=true
export HOST_REACHABLE=true
export PROOF_FETCHED=true
export PROOF_ABSENT=false
export PASSWORD_STORED=false
assert_gate store-decision STORE
export PASSWORD_STORED=true
export USERNAME_STORED=true
assert_gate scrub-decision SCRUB

export ROTATE_HOST_OUTCOME=failure
export PROOF_MATCHES=false
export HOST_REACHABLE=true
export PROOF_FETCHED=false
export PROOF_ABSENT=true
export PASSWORD_STORED=false
export USERNAME_STORED=false
assert_gate store-decision SKIP_UNUSED
assert_gate scrub-decision SCRUB

export ROTATE_HOST_OUTCOME=skipped
export ROTATION_PREFLIGHT_OUTCOME=success
export PROOF_MATCHES=false
export HOST_REACHABLE=false
export PROOF_FETCHED=false
export PROOF_ABSENT=false
export PASSWORD_STORED=false
export RECOVER_STORE_ONLY=false
assert_gate scrub-decision SCRUB
export ROTATION_PREFLIGHT_OUTCOME=failure
assert_gate scrub-decision RETAIN
export RECOVER_STORE_ONLY=true
assert_gate scrub-decision RETAIN
export PASSWORD_STORED=true
export USERNAME_STORED=false
assert_gate scrub-decision RETAIN
export USERNAME_STORED=true
assert_gate scrub-decision SCRUB

unset ROTATE_HOST_OUTCOME ROTATION_PREFLIGHT_OUTCOME PROOF_MATCHES HOST_REACHABLE PROOF_FETCHED PROOF_ABSENT \
  PASSWORD_STORED USERNAME_STORED RECOVER_STORE_ONLY

export KUMA_RESET_APPLIED_FILE="$test_root/kuma-reset-applied.env"
export KUMA_RESET_ATTEMPT_ID="harness-attempt-1"
rm -f "$KUMA_RESET_APPLIED_FILE"
output_with_proof="$( "$rotate_script" )"
if [ ! -f "$KUMA_RESET_APPLIED_FILE" ]; then
  echo "expected reset-applied proof file to be written" >&2
  exit 1
fi
if ! grep -qx 'KUMA_RESET_APPLIED=harness-attempt-1' "$KUMA_RESET_APPLIED_FILE"; then
  echo "reset-applied proof file missing attempt id" >&2
  cat "$KUMA_RESET_APPLIED_FILE" >&2
  exit 1
fi
if ! grep -qx 'KUMA_ROTATION_ATTEMPT_ID=harness-attempt-1' "$KUMA_ROTATE_METADATA_FILE"; then
  echo "rotation metadata must be bound to the reset attempt" >&2
  cat "$KUMA_ROTATE_METADATA_FILE" >&2
  exit 1
fi
if grep -q 'harness-strong-password-not-for-logs' <<<"$output_with_proof"; then
  echo "password must not appear in stdout when writing reset proof" >&2
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
  echo "Found user: harness-admin"
  echo "Error: simulated internally caught reset failure"
  echo "Finished."
  exit 0
fi
echo "unexpected docker invocation: $*" >&2
exit 99
EOF
chmod +x "$test_root/docker"

rm -f "$KUMA_RESET_APPLIED_FILE" "$KUMA_ROTATE_METADATA_FILE"
set +e
false_success_output="$(KUMA_NEW_PASSWORD="false-success-harness-password" "$rotate_script" 2>&1)"
false_success_status=$?
set -e
if [ "$false_success_status" -eq 0 ]; then
  echo "rotate must reject Kuma's zero exit when reset/login markers are absent" >&2
  exit 1
fi
if [ -e "$KUMA_RESET_APPLIED_FILE" ] || [ -e "$KUMA_ROTATE_METADATA_FILE" ]; then
  echo "false reset success must not create proof or metadata" >&2
  exit 1
fi
if grep -q 'false-success-harness-password' <<<"$false_success_output"; then
  echo "password must not appear in false-success failure output" >&2
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
  echo "Logged in."
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
if [ ! -f "$KUMA_RESET_APPLIED_FILE" ]; then
  echo "reset-applied proof must remain after verified reset with post-reset username failure" >&2
  exit 1
fi
if ! grep -qx 'KUMA_RESET_APPLIED=harness-attempt-1' "$KUMA_RESET_APPLIED_FILE"; then
  echo "post-reset username failure must not rewrite or drop this attempt's proof" >&2
  cat "$KUMA_RESET_APPLIED_FILE" >&2
  exit 1
fi

echo "kuma-rotate-credentials.selftest.sh: PASS"
