#!/usr/bin/env bash
set -euo pipefail

# VOC-037-T06 / VOC-037-TEST-01 (INS-9 through INS-11)
#
# Proves the logical isolation VOC-037-D01's corrected 4A mechanism
# depends on, now that VOC-037-D00 co-locates both tiers on one host:
#
#   * /opt/vocanova/infra/secrets      (staging)
#   * /opt/vocanova/production/secrets (production)
#
# Usage:
#   rehearse-production-secrets-boundary.sh <staging_user> <production_user>
#
# Both tree roots can be overridden (VOCANOVA_STAGING_ROOT /
# VOCANOVA_PRODUCTION_ROOT) so the same script can be run against a
# disposable rehearsal host that mirrors the production shape without
# touching the real one - which is exactly what VOC-037-TEST-01's
# preconditions call for before a real production host exists.
#
# Every check either passes or exits non-zero. A check that cannot be
# evaluated (missing file, `sudo -u` unavailable) is a FAILURE, never a
# silent pass: an unevaluated negative-access check is indistinguishable
# from an unenforced one. The one narrow carve-out is INS-11's live
# impersonation probe when the invoking identity correctly has no
# `sudo -u <other>` right (granting one cannot be done safely - see
# report_unprobeable below): if the target user is independently
# confirmed to have broad sudo of their own, that is affirmative proof
# of a broken boundary and still FAILS the rehearsal; only the case
# where neither a live probe nor independent-sudo evidence is available
# downgrades to a WARN, because there is genuinely nothing further this
# script can safely check.

if [ "$#" -ne 2 ]; then
  echo "usage: $0 <staging_user> <production_user>" >&2
  exit 1
fi

staging_user="$1"
production_user="$2"

production_root="${VOCANOVA_PRODUCTION_ROOT:-/opt/vocanova/production}"
staging_root="${VOCANOVA_STAGING_ROOT:-/opt/vocanova/infra}"
production_tree="$production_root/secrets"
staging_tree="$staging_root/secrets"

failures=0

fail() {
  echo "  FAIL: $*" >&2
  failures=$((failures + 1))
}

pass() {
  echo "  ok: $*"
}

warnings=0

warn() {
  echo "  WARN: $*" >&2
  warnings=$((warnings + 1))
}

# Modes are compared as an integer bitmask rather than a string so that
# "no wider than 0750" is checked, not "exactly 0750" - a stricter mode
# than the baseline must not be reported as a violation.
require_mode_no_wider_than() {
  target="$1"
  max_mode="$2"

  if [ ! -e "$target" ]; then
    fail "$target does not exist (cannot verify mode $max_mode)"
    return
  fi

  actual_mode="$(stat -c "%a" "$target")"
  if [ "$(( 0"$actual_mode" & ~0"$max_mode" ))" -ne 0 ]; then
    fail "$target mode is $actual_mode, wider than the required $max_mode"
    return
  fi
  pass "$target mode $actual_mode is within $max_mode"
}

require_owner_not() {
  target="$1"
  forbidden_owner="$2"

  if [ ! -e "$target" ]; then
    fail "$target does not exist (cannot verify ownership)"
    return
  fi

  actual_owner="$(stat -c "%U" "$target")"
  if [ "$actual_owner" = "$forbidden_owner" ]; then
    fail "$target is owned by $forbidden_owner, which must not own production material"
    return
  fi
  pass "$target is owned by $actual_owner (not $forbidden_owner)"
}

# A negative-access probe has three outcomes and only one of them is a
# pass. Readable => the boundary is broken. Unreadable => the boundary
# holds. Probe never ran (no sudo rights, missing account) => unknown,
# handed to report_unprobeable(): a FAILURE if independent evidence
# confirms the boundary is broken anyway, a WARN only when genuinely
# nothing further can be safely established. See report_unprobeable's
# own comment for why "unknown" is not automatically a silent pass.
#
# The probe's own exit status cannot be used to tell "unreadable" from
# "sudo refused": both are exit 1. The inner shell therefore echoes a
# marker carrying the real status, and a missing marker means the probe
# did not run.
probe_as_user() {
  user="$1"
  shift

  set +e
  probe_output="$(sudo -n -u "$user" sh -c '"$@" >/dev/null 2>&1; echo "PROBE:$?"' _ "$@" 2>/dev/null)"
  set -e

  case "$probe_output" in
    *PROBE:*) printf '%s' "${probe_output##*PROBE:}" ;;
    *) printf 'unknown' ;;
  esac
}

require_unreadable_by() {
  user="$1"
  target="$2"

  if [ ! -e "$target" ]; then
    fail "$target does not exist; the negative-access probe for $user is meaningless"
    return
  fi

  probe_status="$(probe_as_user "$user" test -r "$target")"
  case "$probe_status" in
    0) fail "$user CAN read $target" ;;
    unknown) report_unprobeable "$user" "read" "$target" ;;
    *) pass "$user cannot read $target" ;;
  esac
}

require_untraversable_by() {
  user="$1"
  target="$2"

  probe_status="$(probe_as_user "$user" ls "$target")"
  case "$probe_status" in
    0) fail "$user can list $target" ;;
    unknown) report_unprobeable "$user" "traversal" "$target" ;;
    *) pass "$user cannot list $target" ;;
  esac
}

# When the live impersonation probe can't run (the invoking identity has
# no sudo rights to become $user - correctly not granted, since granting
# it would itself be a privilege-escalation path: sudoers cannot safely
# restrict a `sh -c '"$@"'`-wrapped command to read-only test/ls, so the
# only workable grant is the shell itself, i.e. arbitrary command
# execution as $user), fall back to what CAN be safely established: does
# $user have independent, pre-existing broad sudo rights of their own? If
# so, directory/file permissions cannot prove isolation against that
# user regardless of what INS-9 found - $user can always read anything
# as root through their own grant. This is reported as an explicit WARN
# naming the real residual risk, not a silent pass and not a
# script-breaking FAIL for a condition this rehearsal cannot safely test
# further without creating the very escalation path it exists to catch.
#
# Detection deliberately does NOT use `sudo -n -l -U $user`: that query
# itself requires the INVOKING identity to have sudo list rights over
# $user, which a correctly least-privileged invoker (like
# vocanova-production) does not have - confirmed live, `sudo -n -l -U
# deploy` run as vocanova-production fails with "not allowed to execute
# 'list'", silently returning empty and hiding a real, confirmed grant.
# Group membership is checked instead: Ubuntu's standard blanket-sudo
# mechanism is membership in the `sudo` (or `wheel`/`admin` on other
# distros) group via the shipped `%sudo ALL=(ALL:ALL) ALL` default rule,
# and group membership is queryable by any unprivileged user (`id -nG`
# needs no special rights). This does not catch a custom per-user
# sudoers.d grant that bypasses group membership entirely, but it does
# catch the actual, confirmed mechanism in use on this host.
report_unprobeable() {
  user="$1"
  kind="$2"
  target="$3"

  own_grant=""
  for privileged_group in sudo wheel admin; do
    if id -nG "$user" 2>/dev/null | tr ' ' '\n' | grep -qx "$privileged_group"; then
      own_grant="member of the '$privileged_group' group (blanket sudo via the standard %$privileged_group sudoers default)"
      break
    fi
  done
  if [ -n "$own_grant" ]; then
    # This is not "cannot verify" - it is affirmative, confirmed evidence
    # that the negative-access property does NOT hold. $user already has
    # independent broad sudo ($own_grant) and can therefore read $target
    # as root regardless of what INS-9's file permissions say. A known
    # breach must fail the rehearsal, not be downgraded to a warning -
    # the only legitimate ways past this are narrowing $user's sudoers to
    # remove the blanket grant, or an explicit founder-accepted waiver
    # recorded against this specific finding.
    fail "$user is NOT $kind-blocked from $target: $user already has independent broad sudo ($own_grant) and can read anything as root regardless of file permissions - directory-based isolation does NOT hold against $user on this shared host"
  else
    warn "cannot verify $user is $kind-blocked from $target by live probe (no safe impersonation right granted); $user has no independent broad sudo found, so INS-9's file-permission/ownership checks are the available evidence, not a live-probed guarantee."
  fi
}

echo "[INS-9] production secret tree exists and matches the D01 permission baseline"
require_mode_no_wider_than "$production_root" 750
require_mode_no_wider_than "$production_tree" 700
require_mode_no_wider_than "$production_tree/nginx" 700
require_owner_not "$production_root" "$staging_user"
require_owner_not "$production_tree" "$staging_user"

env_files_found=0
for env_file in "$production_tree"/*.env; do
  [ -e "$env_file" ] || continue
  env_files_found=$((env_files_found + 1))
  require_mode_no_wider_than "$env_file" 600
  require_owner_not "$env_file" "$staging_user"
done
if [ "$env_files_found" -eq 0 ]; then
  fail "no *.env file found under $production_tree; production secrets are not provisioned"
else
  pass "$env_files_found production env file(s) checked"
fi

for key_file in "$production_tree"/nginx/key.pem "$production_tree"/nginx/cert.pem; do
  require_mode_no_wider_than "$key_file" 600
done

# Any host path belonging to the other tier - staging's secrets tree,
# its app tree, or a relative build context that resolved upward out of
# the production root - is a cross-tier reference.
report_stray_paths() {
  source_label="$1"
  text="$2"

  stray_paths="$(printf '%s\n' "$text" \
    | grep -oE "(/opt/vocanova|$staging_root|$(dirname "$production_root"))[^\"' :,]*" \
    | grep -v "^$production_root" \
    | sort -u || true)"

  if [ -n "$stray_paths" ]; then
    fail "$source_label references paths outside $production_root: $(printf '%s' "$stray_paths" | tr '\n' ' ')"
    return
  fi
  pass "no $source_label path outside $production_root"
}

echo "[INS-10] production compose reads the production tree only"
compose_file="$production_root/docker-compose.production.yml"
if [ ! -f "$compose_file" ]; then
  fail "$compose_file is missing"
else
  # Two passes are needed, because neither alone sees every host path.
  # `docker compose config` resolves relative volume and build paths to
  # absolute ones (the raw file cannot show those), but it folds
  # `env_file` entries into `environment:` and drops the file paths
  # themselves - so a compose file whose only cross-tier reference is an
  # env_file pointing at staging renders completely clean. The raw file,
  # with the root variable expanded the same way compose expands it,
  # is what catches that case.
  set +e
  compose_rendered="$(VOCANOVA_PRODUCTION_ROOT="$production_root" \
    docker compose -f "$compose_file" -p vocanova-production config 2>&1)"
  compose_status=$?
  set -e

  if [ "$compose_status" -ne 0 ]; then
    fail "docker compose config failed: $compose_rendered"
  else
    report_stray_paths "rendered compose" "$compose_rendered"

    if printf '%s\n' "$compose_rendered" | grep -qE '^[[:space:]]+build:'; then
      fail "production compose declares a build context; images must be pulled by tag"
    else
      pass "production compose declares no build context"
    fi
  fi

  # Comments are stripped first: this file's own header discusses the
  # staging paths it must never use, and a documentation mention is not
  # a reference.
  compose_source_expanded="$(sed \
    -e 's/[[:space:]]*#.*$//' \
    -e "s#\${VOCANOVA_PRODUCTION_ROOT:-/opt/vocanova/production}#$production_root#g" \
    -e "s#\${VOCANOVA_PRODUCTION_ROOT}#$production_root#g" \
    "$compose_file")"
  report_stray_paths "compose source" "$compose_source_expanded"

  if printf '%s\n' "$compose_source_expanded" | grep -qF "$production_tree"; then
    pass "compose references the production secrets tree"
  else
    fail "compose does not reference $production_tree"
  fi
fi

echo "[INS-11] neither tier's deploy identity can read the other's secrets"
require_unreadable_by "$staging_user" "$production_tree/api.env"
require_untraversable_by "$staging_user" "$production_root"
if [ -e "$staging_tree/api.env" ]; then
  require_unreadable_by "$production_user" "$staging_tree/api.env"
else
  pass "staging secrets tree absent on this host; production-to-staging probe not applicable"
fi

if [ "$failures" -ne 0 ]; then
  echo "FAIL: $failures production/staging secret boundary check(s) failed" >&2
  exit 1
fi

if [ "$warnings" -ne 0 ]; then
  echo "PASS WITH $warnings WARNING(S): every check that could run safely passed; see WARN lines above for what could not be live-probed and why - this does not block deploy, but is a real residual-risk disclosure, not a clean pass." >&2
else
  echo "PASS: production/staging secret boundary rehearsal checks succeeded"
fi
