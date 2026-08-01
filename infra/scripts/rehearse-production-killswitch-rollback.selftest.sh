#!/usr/bin/env bash
set -euo pipefail

# VOC-037-T03 self-test for the kill-switch rehearsal.
#
# rehearse-production-killswitch-rollback.sh can only run on the
# provisioned production host. This self-test runs the *same* probe
# library against a disposable Postgres container and a locally built
# api binary, so three things can be checked before (and independently
# of) any production run:
#
#   1. Each probe reports the documented state for a switch that is on
#      and for the same switch turned off.
#   2. Each assertion actually fails when the observed state is wrong.
#      VOC-037-T06 found a rehearsal script that could pass without
#      checking anything; a probe that cannot fail proves nothing.
#   3. The host-only parts of the rehearsal script (api.env restore,
#      explicit image tag, container recreate rather than restart)
#      are present, since this self-test cannot execute them.
#
# Requires docker, go, curl, openssl, and jq. Everything it creates is
# named after this run and removed on exit.

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/../.." && pwd)"
# shellcheck source=lib/killswitch-probes.sh disable=SC1091
. "$script_dir/lib/killswitch-probes.sh"

host_script="$script_dir/rehearse-production-killswitch-rollback.sh"
run_id="$$"
postgres_container="vocanova-t03-selftest-postgres-$run_id"
postgres_port="$(shuf -i 20000-29999 -n 1)"
postgres_password="selftest-$(openssl rand -hex 8)"
api_port="$(shuf -i 30000-39999 -n 1)"
api_base_url="http://127.0.0.1:$api_port"
web_base_url="http://127.0.0.1:$api_port"
oauth_return_url="$web_base_url/home"
email_prefix="t03-rehearsal-"
environment="production"
workdir="$(mktemp -d)"
api_binary="$workdir/vocanova-api"
api_pid=""

export KS_PSQL="docker exec -i -e PGPASSWORD=$postgres_password $postgres_container psql -U vocanova -d vocanova -qtAX"

cleanup() {
  local exit_code=$?
  stop_api
  docker rm -f "$postgres_container" > /dev/null 2>&1 || true
  rm -rf "$workdir"
  exit "$exit_code"
}
trap cleanup EXIT

start_postgres() {
  docker run -d --name "$postgres_container" \
    -e POSTGRES_USER=vocanova \
    -e POSTGRES_DB=vocanova \
    -e POSTGRES_PASSWORD="$postgres_password" \
    -p "127.0.0.1:$postgres_port:5432" \
    postgres:16-alpine > /dev/null

  # pg_isready alone is not enough: the official image answers it
  # during initdb, while the bootstrap server is still about to shut
  # down. Readiness means a real query succeeds.
  for _ in $(seq 1 60); do
    if ks_sql "SELECT 1;" > /dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  echo "ERROR: disposable Postgres did not become ready" >&2
  return 1
}

apply_migrations() {
  local migration
  for migration in "$repo_root"/apps/api/migrations/*.sql; do
    $KS_PSQL -v ON_ERROR_STOP=1 -f - < "$migration" > /dev/null
  done
}

stop_api() {
  if [ -n "$api_pid" ] && kill -0 "$api_pid" 2> /dev/null; then
    kill "$api_pid" 2> /dev/null || true
    wait "$api_pid" 2> /dev/null || true
  fi
  api_pid=""
}

# Mirrors the host script's container recreate: the process is
# replaced, not signalled, so the new kill-switch values are actually
# read.
start_api_with_switches() {
  local binary="$1" ai="$2" magic="$3" oauth="$4" signups="$5"

  stop_api
  PORT="$api_port" \
  DATABASE_URL="postgres://vocanova:$postgres_password@127.0.0.1:$postgres_port/vocanova?sslmode=disable" \
  ENVIRONMENT="$environment" \
  BASE_URL="$api_base_url" \
  OAUTH_REDIRECT_URI="$api_base_url/auth/oauth/google/callback" \
  OAUTH_REDIRECT_ALLOWLIST="$oauth_return_url" \
  SESSION_COOKIE_DOMAIN="127.0.0.1" \
  SESSION_COOKIE_SECURE=false \
  AI_FEATURES_ENABLED="$ai" \
  EMAIL_MAGIC_LINK_ENABLED="$magic" \
  GOOGLE_OAUTH_ENABLED="$oauth" \
  NEW_USER_SIGNUP_ENABLED="$signups" \
    "$binary" >> "$workdir/api.log" 2>&1 &
  api_pid=$!

  for _ in $(seq 1 60); do
    if [ "$(ks_probe_healthz "$api_base_url")" = "200" ]; then
      return 0
    fi
    sleep 1
  done
  echo "ERROR: disposable api did not become healthy; last log lines:" >&2
  tail -n 20 "$workdir/api.log" >&2
  return 1
}

disposable_email() {
  printf '%s%s@rehearsal.invalid' "$email_prefix" "$(openssl rand -hex 6)"
}

# The negative half of the self-test: the same assertion, given a
# deliberately wrong expectation, must report a failure.
require_assertion_fails() {
  local label="$1" wrong_expectation="$2" observed="$3"

  if (ks_expect "$label" "$wrong_expectation" "$observed") > /dev/null 2>&1; then
    ks_fail "negative proof: '$label' still passed against the wrong expectation $wrong_expectation"
    return
  fi
  ks_pass "negative proof: '$label' fails when the observed state is wrong"
}

require_host_script_contains() {
  local description="$1" pattern="$2"

  if grep -qE -- "$pattern" "$host_script"; then
    ks_pass "host script: $description"
    return
  fi
  ks_fail "host script: $description (no line matching /$pattern/)"
}

require_host_script_lacks() {
  local description="$1" pattern="$2"

  if grep -qE -- "$pattern" "$host_script"; then
    ks_fail "host script: $description (unexpected line matching /$pattern/)"
    return
  fi
  ks_pass "host script: $description"
}

ks_reset_checks

echo "[L] assertion-library guards"
if (ks_expect "inconclusive probe" 200 inconclusive) > /dev/null 2>&1; then
  ks_fail "an inconclusive observation was accepted as a pass"
else
  ks_pass "an inconclusive observation is a failure, not a pass"
fi
if (ks_expect_not "inconclusive probe" 503 inconclusive) > /dev/null 2>&1; then
  ks_fail "ks_expect_not accepted an inconclusive observation"
else
  ks_pass "ks_expect_not rejects an inconclusive observation"
fi
if (ks_reset_checks && ks_summary "empty") > /dev/null 2>&1; then
  ks_fail "a run with zero checks reported PASS"
else
  ks_pass "a run with zero checks reports FAIL"
fi

echo "[H] host-only rehearsal steps are present in the host script"
# Each entry is "<description>|<pattern>". The patterns are literal
# search strings for the host script's source text; the dollar signs in
# them are part of what must appear there, not expansions here.
# shellcheck disable=SC2016
host_script_required=(
  'recreates containers instead of restarting them|--force-recreate'
  'restores api.env on every exit path|trap restore_everything EXIT'
  'pins the deployed image tag on every compose call|PRODUCTION_IMAGE_TAG="\$image_tag"'
  'rehearses rollback to the operator-named tag|compose "\$rollback_target_tag" pull'
  'restores the originally running tag|recreate_services_at "\$running_image_tag" api web'
  'deletes the disposable identities it created|ks_cleanup_rehearsal_rows'
)
for required in "${host_script_required[@]}"; do
  require_host_script_contains "${required%%|*}" "${required#*|}"
done
require_host_script_lacks "never touches staging's tree" '/opt/vocanova/infra'

echo "[E] disposable environment"
start_postgres
apply_migrations
(cd "$repo_root/apps/api" && go build -o "$api_binary" ./cmd/api)
cp "$api_binary" "$api_binary.rollback-target"
ks_pass "disposable Postgres, migrations, and api binary are ready"

echo "[S0] baseline - all four switches on"
start_api_with_switches "$api_binary" true true true true
ks_expect "magic-link request accepted" 204 \
  "$(ks_probe_magic_link_request "$api_base_url" "$(disposable_email)")" || true
ks_expect_not "oauth start not disabled" 503 \
  "$(ks_probe_oauth_start "$api_base_url" "$oauth_return_url")" || true

rehearsal_email="$(disposable_email)"
consume_headers="$workdir/consume-headers.txt"
baseline_token="$(ks_seed_magic_link "$rehearsal_email" "$environment")"
ks_expect "new sign-up accepted" 200 \
  "$(ks_probe_magic_link_consume "$api_base_url" "$rehearsal_email" "$baseline_token" "$consume_headers")" || true

session_cookie="$(ks_cookie_from_headers "$consume_headers" vocanova_session)"
csrf_cookie="$(ks_cookie_from_headers "$consume_headers" vocanova_csrf)"
ai_open_observation="$(ks_probe_sentence_feedback "$api_base_url" "$session_cookie" "$csrf_cookie")"
ks_expect "ai generation gate open" "404/" "$ai_open_observation" || true

echo "[S1] EMAIL_MAGIC_LINK_ENABLED=false"
start_api_with_switches "$api_binary" true false true true
magic_off_observation="$(ks_probe_magic_link_request "$api_base_url" "$(disposable_email)")"
ks_expect "magic-link request disabled" 503 "$magic_off_observation" || true
require_assertion_fails "magic-link request disabled" 204 "$magic_off_observation"
ks_expect_not "oauth start unaffected" 503 \
  "$(ks_probe_oauth_start "$api_base_url" "$oauth_return_url")" || true
ks_expect "ai feedback unaffected" "404/" \
  "$(ks_probe_sentence_feedback "$api_base_url" "$session_cookie" "$csrf_cookie")" || true

echo "[S2] GOOGLE_OAUTH_ENABLED=false"
start_api_with_switches "$api_binary" true true false true
oauth_off_observation="$(ks_probe_oauth_start "$api_base_url" "$oauth_return_url")"
ks_expect "oauth start disabled" 503 "$oauth_off_observation" || true
require_assertion_fails "oauth start disabled" 404 "$oauth_off_observation"
ks_expect "magic-link request unaffected" 204 \
  "$(ks_probe_magic_link_request "$api_base_url" "$(disposable_email)")" || true

echo "[S3] NEW_USER_SIGNUP_ENABLED=false"
start_api_with_switches "$api_binary" true true true false
new_email="$(disposable_email)"
new_token="$(ks_seed_magic_link "$new_email" "$environment")"
signup_off_observation="$(ks_probe_magic_link_consume "$api_base_url" "$new_email" "$new_token" "$consume_headers")"
ks_expect "new sign-up refused" 503 "$signup_off_observation" || true
require_assertion_fails "new sign-up refused" 200 "$signup_off_observation"

returning_token="$(ks_seed_magic_link "$rehearsal_email" "$environment")"
ks_expect "returning sign-in still allowed" 200 \
  "$(ks_probe_magic_link_consume "$api_base_url" "$rehearsal_email" "$returning_token" "$workdir/returning-headers.txt")" || true

echo "[S4] AI_FEATURES_ENABLED=false"
start_api_with_switches "$api_binary" false true true true
ai_off_observation="$(ks_probe_sentence_feedback "$api_base_url" "$session_cookie" "$csrf_cookie")"
ks_expect "ai generation gate closed" "200/AI_FEEDBACK_GENERATION_DISABLED" "$ai_off_observation" || true
require_assertion_fails "ai generation gate closed" "$ai_open_observation" "$ai_off_observation"
ks_expect "magic-link request unaffected" 204 \
  "$(ks_probe_magic_link_request "$api_base_url" "$(disposable_email)")" || true

echo "[S5] artifact swap preserves data (local stand-in for the host tag rollback)"
users_before="$(ks_count_rehearsal_users "$email_prefix")"
ks_expect_not "disposable identity exists before the swap" 0 "$users_before" || true
start_api_with_switches "$api_binary.rollback-target" true true true true
ks_expect "healthz after swapping the running artifact" 200 "$(ks_probe_healthz "$api_base_url")" || true
ks_expect "no data loss across the swap" "$users_before" "$(ks_count_rehearsal_users "$email_prefix")" || true
start_api_with_switches "$api_binary" true true true true
ks_expect "healthz after swapping back" 200 "$(ks_probe_healthz "$api_base_url")" || true
ks_expect "no data loss across the swap back" "$users_before" "$(ks_count_rehearsal_users "$email_prefix")" || true

echo "[C] cleanup removes every row the rehearsal created"
started_at="1970-01-01 00:00:00"
ks_cleanup_rehearsal_rows "$email_prefix" "$oauth_return_url" "$started_at"
ks_expect "no rehearsal identity survives cleanup" 0 "$(ks_count_rehearsal_users "$email_prefix")" || true

echo
ks_summary "kill-switch rehearsal self-test"
