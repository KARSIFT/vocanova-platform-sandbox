#!/usr/bin/env bash
set -euo pipefail

# VOC-037-T03 / VOC-037-TEST-03 - kill-switch and rollback rehearsal
# against the real production target VOC-037-T06 provisioned.
#
# Usage (on the production host, as the production deploy user):
#   rehearse-production-killswitch-rollback.sh <previous_image_sha_tag>
#
# <previous_image_sha_tag> is the immutable tag of an already-published
# earlier build (e.g. sha-1a2b3c4), which is what DOC-11 §3's
# "redeploy the previous artifact" rollback model names. The currently
# running tag is read from the running container and restored at the
# end; it is never assumed to be `prod`.
#
# What this does, and does not, touch:
#   * It edits /opt/vocanova/production/secrets/api.env in place and
#     restores it from a 0600 backup on every exit path, including
#     failure and interrupt. The deploy workflow rewrites the same
#     kill-switch lines on the next deploy regardless.
#   * It creates disposable `t03-rehearsal-*@rehearsal.invalid`
#     identities and their magic-link/session rows, and deletes them
#     again. No other row is read, written, or deleted.
#   * It never calls the AI provider (the generation gate is checked
#     before any provider request), never sends email (no production
#     email provider is configured), and never prints a secret.
#
# Every check either passes or exits non-zero. A probe that could not
# be evaluated is a failure, not a pass.

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/killswitch-probes.sh disable=SC1091
. "$script_dir/lib/killswitch-probes.sh"

if [ "$#" -ne 1 ]; then
  echo "usage: $0 <previous_image_sha_tag>" >&2
  exit 1
fi

rollback_target_tag="$1"

production_root="${VOCANOVA_PRODUCTION_ROOT:-/opt/vocanova/production}"
compose_project="${VOCANOVA_PRODUCTION_PROJECT:-vocanova-production}"
compose_file="$production_root/docker-compose.production.yml"
api_env="$production_root/secrets/api.env"
container_prefix="${VOCANOVA_PRODUCTION_CONTAINER_PREFIX:-vocanova-production}"
api_container="$container_prefix-api"
postgres_container="$container_prefix-postgres"
api_base_url="${VOCANOVA_PRODUCTION_API_BASE_URL:-https://api-production.vocanova.site:8443}"
web_base_url="${VOCANOVA_PRODUCTION_WEB_BASE_URL:-https://production.vocanova.site:8443}"

email_prefix="t03-rehearsal-"
oauth_return_url="$web_base_url/home"
environment="production"

export KS_PSQL="docker exec -i $postgres_container psql -U vocanova -d vocanova -qtAX"

started_at="$(date -u +'%Y-%m-%d %H:%M:%S')"
api_env_backup="$(mktemp)"
consume_headers="$(mktemp)"
chmod 600 "$api_env_backup" "$consume_headers"

for required in "$compose_file" "$api_env"; do
  if [ ! -f "$required" ]; then
    echo "ERROR: $required is missing; this host has no provisioned production target" >&2
    exit 1
  fi
done

cp "$api_env" "$api_env_backup"

running_image_tag="$(docker inspect --format '{{.Config.Image}}' "$api_container" | sed 's/.*://')"
if [ -z "$running_image_tag" ]; then
  echo "ERROR: could not read the running production api image tag" >&2
  exit 1
fi

export VOCANOVA_PRODUCTION_ROOT="$production_root"

compose() {
  local image_tag="$1"
  shift
  PRODUCTION_IMAGE_TAG="$image_tag" \
    docker compose -f "$compose_file" -p "$compose_project" "$@"
}

# Compose bakes env_file content into a container at creation time, so
# `docker restart` would leave the old kill-switch values in place and
# every probe below would report the previous state (found live during
# VOC-037-T06's first real deploy).
recreate_services_at() {
  local image_tag="$1"
  shift
  local service

  compose "$image_tag" up -d --force-recreate --no-deps "$@" > /dev/null
  for service in "$@"; do
    wait_until_healthy "$container_prefix-$service"
  done
}

wait_until_healthy() {
  local container="$1" status=""

  for _ in $(seq 1 60); do
    status="$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}missing{{end}}' "$container" 2>/dev/null || true)"
    if [ "$status" = "healthy" ]; then
      return 0
    fi
    sleep 2
  done
  echo "ERROR: $container did not become healthy (last status: ${status:-unknown})" >&2
  return 1
}

set_switch() {
  local key="$1" value="$2"

  sed -i "/^$key=/d" "$api_env"
  printf '%s=%s\n' "$key" "$value" >> "$api_env"
  chmod 600 "$api_env"
}

apply_switch_state() {
  local ai="$1" magic="$2" oauth="$3" signups="$4"

  set_switch AI_FEATURES_ENABLED "$ai"
  set_switch EMAIL_MAGIC_LINK_ENABLED "$magic"
  set_switch GOOGLE_OAUTH_ENABLED "$oauth"
  set_switch NEW_USER_SIGNUP_ENABLED "$signups"
  recreate_services_at "$running_image_tag" api
}

disposable_email() {
  printf '%s%s@rehearsal.invalid' "$email_prefix" "$(openssl rand -hex 6)"
}

restore_everything() {
  local exit_code=$?

  echo
  echo "[cleanup] restoring api.env, the deployed tag, and the database"
  cp "$api_env_backup" "$api_env" 2>/dev/null || true
  chmod 600 "$api_env" 2>/dev/null || true
  compose "$running_image_tag" up -d --force-recreate --no-deps api web > /dev/null 2>&1 || true
  ks_cleanup_rehearsal_rows "$email_prefix" "$oauth_return_url" "$started_at" 2>/dev/null || true
  rm -f "$api_env_backup" "$consume_headers"

  local leftover
  leftover="$(ks_count_rehearsal_users "$email_prefix" 2>/dev/null || echo unknown)"
  echo "[cleanup] remaining rehearsal identities: $leftover"

  exit "$exit_code"
}
trap restore_everything EXIT

ks_reset_checks

echo "[S0] baseline - all four switches on"
apply_switch_state true true true true
ks_expect "healthz" 200 "$(ks_probe_healthz "$api_base_url")" || true
ks_expect "magic-link request accepted" 204 \
  "$(ks_probe_magic_link_request "$api_base_url" "$(disposable_email)")" || true
ks_expect_not "oauth start not disabled" 503 \
  "$(ks_probe_oauth_start "$api_base_url" "$oauth_return_url")" || true

rehearsal_email="$(disposable_email)"
baseline_token="$(ks_seed_magic_link "$rehearsal_email" "$environment")"
ks_expect "new sign-up accepted" 200 \
  "$(ks_probe_magic_link_consume "$api_base_url" "$rehearsal_email" "$baseline_token" "$consume_headers")" || true

session_cookie="$(ks_cookie_from_headers "$consume_headers" vocanova_session)"
csrf_cookie="$(ks_cookie_from_headers "$consume_headers" vocanova_csrf)"
ks_expect "ai generation gate open" "404/" \
  "$(ks_probe_sentence_feedback "$api_base_url" "$session_cookie" "$csrf_cookie")" || true

echo "[S1] EMAIL_MAGIC_LINK_ENABLED=false"
apply_switch_state true false true true
ks_expect "magic-link request disabled" 503 \
  "$(ks_probe_magic_link_request "$api_base_url" "$(disposable_email)")" || true
ks_expect "healthz unaffected" 200 "$(ks_probe_healthz "$api_base_url")" || true
ks_expect_not "oauth start unaffected" 503 \
  "$(ks_probe_oauth_start "$api_base_url" "$oauth_return_url")" || true
ks_expect "ai feedback unaffected" "404/" \
  "$(ks_probe_sentence_feedback "$api_base_url" "$session_cookie" "$csrf_cookie")" || true

echo "[S2] GOOGLE_OAUTH_ENABLED=false"
apply_switch_state true true false true
ks_expect "oauth start disabled" 503 \
  "$(ks_probe_oauth_start "$api_base_url" "$oauth_return_url")" || true
ks_expect "magic-link request unaffected" 204 \
  "$(ks_probe_magic_link_request "$api_base_url" "$(disposable_email)")" || true
ks_expect "healthz unaffected" 200 "$(ks_probe_healthz "$api_base_url")" || true

echo "[S3] NEW_USER_SIGNUP_ENABLED=false"
apply_switch_state true true true false
new_email="$(disposable_email)"
new_token="$(ks_seed_magic_link "$new_email" "$environment")"
ks_expect "new sign-up refused" 503 \
  "$(ks_probe_magic_link_consume "$api_base_url" "$new_email" "$new_token" "$consume_headers")" || true

# The switch closes new-user creation only; an already-known identity
# must still be able to sign in (apps/api/business/auth/killswitches.go).
returning_token="$(ks_seed_magic_link "$rehearsal_email" "$environment")"
ks_expect "returning sign-in still allowed" 200 \
  "$(ks_probe_magic_link_consume "$api_base_url" "$rehearsal_email" "$returning_token" "$consume_headers")" || true
ks_expect "healthz unaffected" 200 "$(ks_probe_healthz "$api_base_url")" || true

echo "[S4] AI_FEATURES_ENABLED=false"
apply_switch_state false true true true
ks_expect "ai generation gate closed" "200/AI_FEEDBACK_GENERATION_DISABLED" \
  "$(ks_probe_sentence_feedback "$api_base_url" "$session_cookie" "$csrf_cookie")" || true
ks_expect "magic-link request unaffected" 204 \
  "$(ks_probe_magic_link_request "$api_base_url" "$(disposable_email)")" || true
ks_expect_not "oauth start unaffected" 503 \
  "$(ks_probe_oauth_start "$api_base_url" "$oauth_return_url")" || true
ks_expect "healthz unaffected" 200 "$(ks_probe_healthz "$api_base_url")" || true

echo "[S5] rollback to $rollback_target_tag and forward again"
users_before="$(ks_count_rehearsal_users "$email_prefix")"
ks_expect_not "disposable identity exists before rollback" 0 "$users_before" || true

compose "$rollback_target_tag" pull api web > /dev/null
recreate_services_at "$rollback_target_tag" api web
ks_expect "healthz after rollback" 200 "$(ks_probe_healthz "$api_base_url")" || true
ks_expect "rolled-back tag is running" "$rollback_target_tag" \
  "$(docker inspect --format '{{.Config.Image}}' "$api_container" | sed 's/.*://')" || true
ks_expect "no data loss across rollback" "$users_before" "$(ks_count_rehearsal_users "$email_prefix")" || true

recreate_services_at "$running_image_tag" api web
ks_expect "healthz after roll-forward" 200 "$(ks_probe_healthz "$api_base_url")" || true
ks_expect "original tag restored" "$running_image_tag" \
  "$(docker inspect --format '{{.Config.Image}}' "$api_container" | sed 's/.*://')" || true
ks_expect "no data loss across roll-forward" "$users_before" "$(ks_count_rehearsal_users "$email_prefix")" || true

echo
ks_summary "production kill-switch and rollback"
