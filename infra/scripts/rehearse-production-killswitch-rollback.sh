#!/usr/bin/env bash
set -euo pipefail

# VOC-037-T03 / VOC-037-TEST-03 (INS-12 through INS-16)
#
# Exercises DOC-11 §3's four required launch kill switches
# (AI_FEATURES_ENABLED, EMAIL_MAGIC_LINK_ENABLED, GOOGLE_OAUTH_ENABLED,
# NEW_USER_SIGNUP_ENABLED) and its rollback model ("redeploy previous
# known-good artifact by digest") against the real production target
# VOC-037-T06 provisioned, then restores the tier to its starting state.
#
# Usage:
#   rehearse-production-killswitch-rollback.sh <api_host> <rollback_image_tag>
#
#   api_host            the production API hostname (the server_name in
#                       infra/nginx-production/conf.d/20-api-production.conf)
#   rollback_image_tag  an already-published, previously-deployed image tag
#                       (sha-<short-sha>) to redeploy as the rollback
#                       rehearsal. Required, not optional: an unexercised
#                       rollback path is an unproven one.
#
# Overridable so the same script can run against a disposable mirror of
# the production shape: VOCANOVA_PRODUCTION_ROOT,
# VOCANOVA_PRODUCTION_HTTPS_PORT, VOCANOVA_HEALTH_TIMEOUT_SECONDS,
# VOCANOVA_UPSTREAM_DNS_TTL_SECONDS, VOCANOVA_PROBE_MODE.
#
# VOCANOVA_PROBE_MODE=origin (the default) resolves the production
# hostname to the loopback address, so every probe reaches this host's
# own nginx directly. That is deliberate: Cloudflare's edge answers with
# its own error pages (T06 hit both a 520 and a 526 that way), and an
# edge error is indistinguishable from the origin behavior these checks
# assert. Origin-direct probing skips certificate-chain verification,
# which T06 already evidenced through the edge, and asserts what only
# this script can - that the switch state on disk changes what the
# application actually does. VOCANOVA_PROBE_MODE=edge probes through
# Cloudflare instead.
#
# Every check either passes or exits non-zero. A check that cannot be
# evaluated is a FAILURE, never a silent pass. The tier's original
# switch values and image tag are restored on exit - including on
# failure - and the restoration is verified rather than assumed.
#
# This script writes disposable rows (one synthetic user, its session,
# and two magic links) under a single reserved e-mail namespace and
# deletes them again. It never reads, modifies, or deletes any other
# row, and never records a token, secret, or personal datum.

if [ "$#" -ne 2 ]; then
  echo "usage: $0 <api_host> <rollback_image_tag>" >&2
  exit 1
fi

api_host="$1"
rollback_image_tag="$2"

production_root="${VOCANOVA_PRODUCTION_ROOT:-/opt/vocanova/production}"
compose_file="$production_root/docker-compose.production.yml"
api_env_file="$production_root/secrets/api.env"
compose_project="vocanova-production"
api_container="vocanova-production-api"
web_container="vocanova-production-web"
postgres_container="vocanova-production-postgres"
api_image_repository="ghcr.io/karsift/vocanova-api"

https_port="${VOCANOVA_PRODUCTION_HTTPS_PORT:-8443}"
probe_mode="${VOCANOVA_PROBE_MODE:-origin}"
health_timeout_seconds="${VOCANOVA_HEALTH_TIMEOUT_SECONDS:-120}"

# nginx resolves its upstreams through Docker's embedded DNS with
# `valid=10s` (infra/nginx-production/conf.d/02-docker-dns.conf), so
# a freshly recreated api container is reachable only after that TTL
# expires. Probing sooner reports nginx's stale-upstream 502, not the
# switch behavior under test.
upstream_dns_ttl_seconds="${VOCANOVA_UPSTREAM_DNS_TTL_SECONDS:-12}"

switch_keys="AI_FEATURES_ENABLED EMAIL_MAGIC_LINK_ENABLED GOOGLE_OAUTH_ENABLED NEW_USER_SIGNUP_ENABLED"

# Rows in these tables must be identical before and after the rollback
# rehearsal: DOC-11 §3 permits rolling application artifacts back, never
# losing learner data while doing it.
invariant_tables="users sessions magic_links user_words review_attempts learner_sentences ai_feedback_attempts"

disposable_email_domain="rehearsal.invalid"
disposable_email_prefix="voc037-t03-"
run_tag="$(date -u +%Y%m%dT%H%M%SZ)-$$"

failures=0
last_check_passed=1
http_body_file="$(mktemp)"

# The tag every compose invocation deploys unless one is named
# explicitly. Preflight sets it to whatever production is running now.
intended_image_tag=""

fail() {
  echo "  FAIL: $*" >&2
  failures=$((failures + 1))
  last_check_passed=0
}

pass() {
  echo "  ok: $*"
  last_check_passed=1
}

# ---------------------------------------------------------------------
# Host primitives
# ---------------------------------------------------------------------

# PRODUCTION_IMAGE_TAG is always passed explicitly, because the compose
# file falls back to the mutable `prod` tag when it is unset - a bare
# recreate would silently move production off the immutable sha tag it
# was deployed with, which is precisely the artifact identity the
# rollback checks below depend on.
compose() {
  VOCANOVA_PRODUCTION_ROOT="$production_root" \
  PRODUCTION_IMAGE_TAG="${PRODUCTION_IMAGE_TAG:-$intended_image_tag}" \
    docker compose -f "$compose_file" -p "$compose_project" "$@"
}

# Reads SQL from stdin. -qtAX gives one bare, pipe-separated row per
# line, which is what the count helpers below expect.
sql() {
  docker exec -i "$postgres_container" \
    psql -qtAX -U vocanova -d vocanova
}

api_env_get() {
  sed -n "s/^$1=//p" "$api_env_file" | tail -1
}

# Re-asserts 0600 because `sed -i` writes a new file under the current
# umask and renames it over the original, which would otherwise widen
# the mode the D01 baseline requires. The rewritten key moves to the end
# of the file; env-file semantics are order-independent, and the next
# deploy rewrites the file wholesale in any case.
api_env_set() {
  sed -i "/^$1=/d" "$api_env_file"
  printf '%s=%s\n' "$1" "$2" >> "$api_env_file"
  chmod 600 "$api_env_file"
}

api_env_unset() {
  sed -i "/^$1=/d" "$api_env_file"
  chmod 600 "$api_env_file"
}

container_image_id() {
  docker inspect --format '{{.Image}}' "$1" 2>/dev/null || true
}

container_image_reference() {
  docker inspect --format '{{.Config.Image}}' "$1" 2>/dev/null || true
}

tagged_image_id() {
  docker image inspect --format '{{.Id}}' "$1" 2>/dev/null || true
}

container_is_healthy() {
  status="$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}missing{{end}}' "$1" 2>/dev/null || true)"
  [ "$status" = "healthy" ]
}

wait_for_healthy() {
  deadline=$((SECONDS + health_timeout_seconds))
  while [ "$SECONDS" -lt "$deadline" ]; do
    if container_is_healthy "$1"; then
      return 0
    fi
    sleep 2
  done
  return 1
}

# The switch state is baked into the container by `env_file` at creation
# time, so every toggle needs a recreate rather than a restart
# (VOC-037-T06 found this live). A container that never becomes healthy
# is a hard stop: no probe result after it would mean anything.
apply_api_configuration() {
  compose up -d --force-recreate --no-deps api >/dev/null
  if ! wait_for_healthy "$api_container"; then
    echo "FATAL: api container did not become healthy after $1" >&2
    docker logs --tail 50 "$api_container" >&2 2>&1 || true
    exit 1
  fi
  sleep "$upstream_dns_ttl_seconds"
}

# The api logs its effective switch state once per start
# (apps/api/cmd/api/main.go). Comparing that line against api.env is
# what proves the file the deploy writes is the state the process
# actually loaded, instead of assuming the two agree.
require_startup_flag() {
  flag_name="$1"
  expected="$2"

  startup_line="$(docker logs "$api_container" 2>&1 | grep -F 'api: listening on' | tail -1)"
  if [ -z "$startup_line" ]; then
    fail "api logged no startup line; cannot confirm $flag_name=$expected"
    return 0
  fi

  actual="$(printf '%s\n' "$startup_line" | sed -n "s/.*[ (]$flag_name=\([a-z]*\).*/\1/p")"
  if [ "$actual" != "$expected" ]; then
    fail "api reports $flag_name=${actual:-<unparseable>} at startup, expected $expected"
    return 0
  fi
  pass "api reports $flag_name=$expected at startup"
}

# ---------------------------------------------------------------------
# HTTP probes
# ---------------------------------------------------------------------

# Prints the HTTP status code and leaves the response body in
# $http_body_file. A transport failure prints 000, which every caller
# treats as a failed check rather than as a disabled response.
http_request() {
  method="$1"
  path="$2"
  shift 2

  set -- -sS --max-time 20 -o "$http_body_file" -w '%{http_code}' \
    -X "$method" -H 'Content-Type: application/json' "$@"
  if [ "$probe_mode" = "origin" ]; then
    set -- "$@" --insecure --resolve "$api_host:$https_port:127.0.0.1"
  fi

  : > "$http_body_file"
  status="$(curl "$@" "https://$api_host:$https_port$path" 2>/dev/null || true)"
  printf '%s' "${status:-000}"
}

require_status() {
  expected="$1"
  actual="$2"
  label="$3"

  if [ "$actual" != "$expected" ]; then
    fail "$label returned HTTP $actual, expected $expected"
    return 0
  fi
  pass "$label returned HTTP $expected"
}

require_status_not() {
  forbidden="$1"
  actual="$2"
  label="$3"

  if [ "$actual" = "$forbidden" ]; then
    fail "$label returned HTTP $forbidden, which is the disabled response"
    return 0
  fi
  if [ "$actual" = "000" ]; then
    fail "$label did not complete; the probe result is unknown"
    return 0
  fi
  pass "$label returned HTTP $actual (not the disabled $forbidden)"
}

require_body_contains() {
  if ! grep -qF "$1" "$http_body_file"; then
    fail "$2 response body does not contain $1"
    return 0
  fi
  pass "$2 response body contains $1"
}

new_uuid() {
  cat /proc/sys/kernel/random/uuid
}

probe_magic_link_request() {
  http_request POST /api/v1/auth/magic-links --data "{\"email\":\"$1\"}"
}

probe_magic_link_consume() {
  http_request POST /api/v1/auth/magic-links/consume \
    --data "{\"token\":\"$1\",\"email\":\"$2\"}"
}

probe_oauth_start() {
  http_request POST /api/v1/auth/oauth/google/start \
    --data '{"redirectUri":"https://example.invalid/onboarding"}'
}

# A random, non-existent attemptId is deliberate. With AI enabled the
# request passes the generation gate and 404s on the target lookup; with
# AI disabled the gate short-circuits it first with a 200 carrying
# AI_FEEDBACK_GENERATION_DISABLED
# (apps/api/business/aifeedback/service.go). The two outcomes are
# therefore distinguishable without creating a real saved word, a real
# review, or any AI-provider call.
probe_sentence_feedback() {
  session_token="$1"
  csrf_token="voc037-t03-csrf-$run_tag"

  http_request POST /api/v1/sentence-feedback \
    -H "Cookie: vocanova_session=$session_token; vocanova_csrf=$csrf_token" \
    -H "X-CSRF-Token: $csrf_token" \
    -H "Idempotency-Key: $disposable_email_prefix$run_tag-$(new_uuid)" \
    --data "{\"sentenceText\":\"This is a disposable rehearsal sentence.\",\"source\":\"word_detail\",\"attemptId\":\"$(new_uuid)\"}"
}

probe_healthz() {
  http_request GET /healthz
}

# ---------------------------------------------------------------------
# Disposable fixtures
# ---------------------------------------------------------------------

disposable_email() {
  printf '%s%s-%s@%s' "$disposable_email_prefix" "$run_tag" "$1" "$disposable_email_domain"
}

# Emits "<token> <sha256-hex>" for one freshly generated 32-byte secret,
# encoded exactly the way auth.generateToken does: padded URL-safe
# base64 of the raw bytes as the token, SHA-256 of the same raw bytes as
# the persisted hash. The raw bytes never leave this function.
new_token_and_hash() {
  raw_file="$(mktemp)"
  head -c 32 /dev/urandom > "$raw_file"
  printf '%s %s' \
    "$(base64 -w0 < "$raw_file" | tr '+/' '-_')" \
    "$(sha256sum "$raw_file" | cut -d' ' -f1)"
  rm -f "$raw_file"
}

create_disposable_session() {
  read -r session_token session_hash <<EOF
$(new_token_and_hash)
EOF

  sql >/dev/null <<SQL
INSERT INTO users (id, email, status, onboarding_status, email_verified_at, created_at, updated_at)
VALUES ('$(new_uuid)', '$1', 'active', 'completed', now(), now(), now());
INSERT INTO sessions (id, user_id, token_hash, created_at, expires_at)
SELECT '$(new_uuid)', id, decode('$session_hash', 'hex'), now(), now() + interval '1 hour'
FROM users WHERE email = '$1';
SQL

  printf '%s' "$session_token"
}

# The consume endpoint is the only path that observes
# NEW_USER_SIGNUP_ENABLED, and it needs a valid link row. Writing the
# row directly, rather than requesting one, keeps the rehearsal from
# depending on real e-mail delivery.
create_disposable_magic_link() {
  read -r link_token link_hash <<EOF
$(new_token_and_hash)
EOF

  sql >/dev/null <<SQL
INSERT INTO magic_links (id, email, token_hash, environment, created_at, expires_at)
VALUES ('$(new_uuid)', '$1', decode('$link_hash', 'hex'), '$2', now(), now() + interval '10 minutes');
SQL

  printf '%s' "$link_token"
}

delete_disposable_rows() {
  sql >/dev/null <<SQL
DELETE FROM sessions WHERE user_id IN (
  SELECT id FROM users WHERE email LIKE '$disposable_email_prefix%@$disposable_email_domain'
);
DELETE FROM magic_links WHERE email LIKE '$disposable_email_prefix%@$disposable_email_domain';
DELETE FROM users WHERE email LIKE '$disposable_email_prefix%@$disposable_email_domain';
SQL
}

count_disposable_users() {
  sql <<SQL
SELECT count(*) FROM users WHERE email LIKE '$disposable_email_prefix%@$disposable_email_domain';
SQL
}

row_counts() {
  for table in $invariant_tables; do
    printf '%s|%s\n' "$table" "$(sql <<SQL
SELECT count(*) FROM $table;
SQL
)"
  done
}

# ---------------------------------------------------------------------
# Deploy, rollback, and restoration
# ---------------------------------------------------------------------

deploy_image_tag() {
  PRODUCTION_IMAGE_TAG="$1" compose pull api web >/dev/null
  PRODUCTION_IMAGE_TAG="$1" compose up -d --force-recreate --no-deps api web >/dev/null
  intended_image_tag="$1"
  wait_for_healthy "$api_container" || return 1
  wait_for_healthy "$web_container" || return 1
  sleep "$upstream_dns_ttl_seconds"
}

original_switch_state=""
original_image_tag=""
restore_done=0

restore_tier() {
  [ "$restore_done" -eq 0 ] || return 0
  restore_done=1

  echo "[restore] returning the production tier to its starting state"
  delete_disposable_rows || fail "could not delete the disposable rehearsal rows"

  printf '%s\n' "$original_switch_state" | while IFS='=' read -r key value; do
    [ -n "$key" ] || continue
    # An originally-absent key is restored by absence, not by an empty
    # assignment: `KEY=` and no KEY at all are the same to the api's
    # config loader today, but only one of them is what was there.
    if [ -n "$value" ]; then
      api_env_set "$key" "$value"
    else
      api_env_unset "$key"
    fi
  done

  if [ "$(container_image_reference "$api_container")" = "$api_image_repository:$original_image_tag" ]; then
    apply_api_configuration "restore"
  elif ! deploy_image_tag "$original_image_tag"; then
    fail "could not redeploy the original image tag $original_image_tag"
  fi

  rm -f "$http_body_file"
}

# ---------------------------------------------------------------------
# Preflight
# ---------------------------------------------------------------------

echo "[preflight] the production target is present, healthy, and identifiable"
for required in "$compose_file" "$api_env_file"; do
  if [ ! -f "$required" ]; then
    echo "FATAL: $required is missing; this is not a provisioned production host" >&2
    exit 1
  fi
done
for required_container in "$postgres_container" "$api_container" "$web_container"; do
  if ! container_is_healthy "$required_container"; then
    echo "FATAL: $required_container is not healthy; rehearse against a healthy tier only" >&2
    exit 1
  fi
done

environment="$(api_env_get ENVIRONMENT)"
if [ "$environment" != "production" ]; then
  echo "FATAL: ENVIRONMENT is '${environment:-<unset>}', expected production" >&2
  exit 1
fi

original_image_reference="$(container_image_reference "$api_container")"
original_image_tag="${original_image_reference##*:}"
if [ -z "$original_image_tag" ] || [ "$original_image_tag" = "$original_image_reference" ]; then
  echo "FATAL: could not determine the currently deployed image tag" >&2
  exit 1
fi
if [ "$rollback_image_tag" = "$original_image_tag" ]; then
  echo "FATAL: rollback tag $rollback_image_tag is already deployed; rolling back to the running artifact proves nothing" >&2
  exit 1
fi
intended_image_tag="$original_image_tag"

original_switch_state="$(
  for key in $switch_keys; do
    printf '%s=%s\n' "$key" "$(api_env_get "$key")"
  done
)"
trap restore_tier EXIT

pass "tier healthy, ENVIRONMENT=production, deployed tag $original_image_tag"

if [ "$(count_disposable_users)" != "0" ]; then
  fail "the reserved $disposable_email_domain namespace already holds rows; an earlier rehearsal did not clean up"
fi

baseline_counts="$(row_counts)"

# ---------------------------------------------------------------------
# INS-12 — EMAIL_MAGIC_LINK_ENABLED
# ---------------------------------------------------------------------

echo "[INS-12] EMAIL_MAGIC_LINK_ENABLED observably gates magic-link sign-in"
magic_probe_email="$(disposable_email magic)"

api_env_set EMAIL_MAGIC_LINK_ENABLED false
apply_api_configuration "EMAIL_MAGIC_LINK_ENABLED=false"
require_startup_flag magic off
require_status 503 "$(probe_magic_link_request "$magic_probe_email")" \
  "magic-link request with EMAIL_MAGIC_LINK_ENABLED=false"

api_env_set EMAIL_MAGIC_LINK_ENABLED true
apply_api_configuration "EMAIL_MAGIC_LINK_ENABLED=true"
require_startup_flag magic on
# Any status other than 503 proves the switch opened the path. A non-2xx
# here would mean the request reached the sender and the sender failed -
# an e-mail-provider problem to investigate, not a switch problem, since
# the probe address is deliberately undeliverable.
require_status_not 503 "$(probe_magic_link_request "$magic_probe_email")" \
  "magic-link request with EMAIL_MAGIC_LINK_ENABLED=true"

# ---------------------------------------------------------------------
# INS-13 — GOOGLE_OAUTH_ENABLED
# ---------------------------------------------------------------------

echo "[INS-13] GOOGLE_OAUTH_ENABLED observably gates Google sign-in"
api_env_set GOOGLE_OAUTH_ENABLED false
apply_api_configuration "GOOGLE_OAUTH_ENABLED=false"
require_startup_flag oauth off
require_status 503 "$(probe_oauth_start)" \
  "oauth start with GOOGLE_OAUTH_ENABLED=false"

api_env_set GOOGLE_OAUTH_ENABLED true
apply_api_configuration "GOOGLE_OAUTH_ENABLED=true"
require_startup_flag oauth on
# With the switch on but no production Google client provisioned yet
# (VOC-032-DEP-07's production-tier equivalent), the service answers
# "not configured" instead of "disabled" - a different code path, and
# exactly the distinction this check exists to observe.
require_status_not 503 "$(probe_oauth_start)" \
  "oauth start with GOOGLE_OAUTH_ENABLED=true"

# ---------------------------------------------------------------------
# INS-14 — NEW_USER_SIGNUP_ENABLED
# ---------------------------------------------------------------------

echo "[INS-14] NEW_USER_SIGNUP_ENABLED observably gates first-time sign-up"
signup_probe_email="$(disposable_email signup)"

api_env_set EMAIL_MAGIC_LINK_ENABLED true
api_env_set NEW_USER_SIGNUP_ENABLED false
apply_api_configuration "NEW_USER_SIGNUP_ENABLED=false"
require_startup_flag signups off
require_status 503 \
  "$(probe_magic_link_consume "$(create_disposable_magic_link "$signup_probe_email" "$environment")" "$signup_probe_email")" \
  "first-time sign-up with NEW_USER_SIGNUP_ENABLED=false"
if [ "$(count_disposable_users)" != "0" ]; then
  fail "a user row was created while NEW_USER_SIGNUP_ENABLED=false"
else
  pass "no user row was created while NEW_USER_SIGNUP_ENABLED=false"
fi

api_env_set NEW_USER_SIGNUP_ENABLED true
apply_api_configuration "NEW_USER_SIGNUP_ENABLED=true"
require_startup_flag signups on
require_status 200 \
  "$(probe_magic_link_consume "$(create_disposable_magic_link "$signup_probe_email" "$environment")" "$signup_probe_email")" \
  "first-time sign-up with NEW_USER_SIGNUP_ENABLED=true"
if [ "$last_check_passed" -eq 1 ]; then
  if [ "$(count_disposable_users)" = "1" ]; then
    pass "exactly one disposable user row was created while NEW_USER_SIGNUP_ENABLED=true"
  else
    fail "expected exactly one disposable user row while NEW_USER_SIGNUP_ENABLED=true"
  fi
fi

# ---------------------------------------------------------------------
# INS-15 — AI_FEATURES_ENABLED
# ---------------------------------------------------------------------

echo "[INS-15] AI_FEATURES_ENABLED observably gates AI feedback generation"
ai_session_token="$(create_disposable_session "$(disposable_email ai)")"

api_env_set AI_FEATURES_ENABLED false
apply_api_configuration "AI_FEATURES_ENABLED=false"
require_startup_flag ai off
require_status 200 "$(probe_sentence_feedback "$ai_session_token")" \
  "sentence feedback with AI_FEATURES_ENABLED=false"
if [ "$last_check_passed" -eq 1 ]; then
  require_body_contains "AI_FEEDBACK_GENERATION_DISABLED" \
    "sentence feedback with AI_FEATURES_ENABLED=false"
fi

api_env_set AI_FEATURES_ENABLED true
apply_api_configuration "AI_FEATURES_ENABLED=true"
require_startup_flag ai on
# 404 means the request passed the generation gate and reached the
# deliberately non-existent target: the gate is open.
require_status 404 "$(probe_sentence_feedback "$ai_session_token")" \
  "sentence feedback with AI_FEATURES_ENABLED=true"

# ---------------------------------------------------------------------
# INS-16 — rollback by redeploying a previous artifact
# ---------------------------------------------------------------------

echo "[INS-16] rollback redeploys a previous artifact without losing data"
pre_rollback_counts="$(row_counts)"
pre_rollback_image_id="$(container_image_id "$api_container")"

if ! deploy_image_tag "$rollback_image_tag"; then
  fail "rollback to $rollback_image_tag did not reach a healthy state"
else
  rolled_back_image_id="$(container_image_id "$api_container")"
  expected_image_id="$(tagged_image_id "$api_image_repository:$rollback_image_tag")"

  if [ -z "$expected_image_id" ]; then
    fail "could not resolve the image id of $api_image_repository:$rollback_image_tag"
  elif [ "$rolled_back_image_id" = "$pre_rollback_image_id" ]; then
    fail "the api container still runs the pre-rollback image after redeploying $rollback_image_tag"
  elif [ "$rolled_back_image_id" != "$expected_image_id" ]; then
    fail "the api container runs $rolled_back_image_id, not $rollback_image_tag's $expected_image_id"
  else
    pass "api runs $rollback_image_tag's image ($expected_image_id) after rollback"
  fi

  require_status 200 "$(probe_healthz)" "health check after rollback to $rollback_image_tag"
  require_body_contains '"status":"ok"' "health check after rollback to $rollback_image_tag"

  if [ "$(row_counts)" != "$pre_rollback_counts" ]; then
    fail "row counts changed across the rollback; the artifact rollback lost or wrote data"
  else
    pass "row counts are unchanged across the rollback"
  fi
fi

echo "[INS-16] roll forward again to the originally deployed artifact"
if ! deploy_image_tag "$original_image_tag"; then
  fail "roll-forward to $original_image_tag did not reach a healthy state"
else
  if [ "$(container_image_id "$api_container")" != "$pre_rollback_image_id" ]; then
    fail "the api container does not run the original image after rolling forward"
  else
    pass "api runs the original image ($pre_rollback_image_id) again"
  fi
  require_status 200 "$(probe_healthz)" "health check after roll-forward to $original_image_tag"
fi

# ---------------------------------------------------------------------
# Restoration and final invariants
# ---------------------------------------------------------------------

restore_tier
trap - EXIT

echo "[restore] the tier's starting state is verified, not assumed"
for key in $switch_keys; do
  expected="$(printf '%s\n' "$original_switch_state" | sed -n "s/^$key=//p")"
  actual="$(api_env_get "$key")"
  if [ "$actual" != "$expected" ]; then
    fail "$key is '${actual:-<unset>}' after the rehearsal, expected the original '${expected:-<unset>}'"
  else
    pass "$key restored to '${expected:-<unset>}'"
  fi
done

# AI_FEATURES_ENABLED is the one switch whose config default is "on"
# (apps/api/app/api/production.go's getenvBool fallback), so an absent
# value restores to on, not off.
restored_ai_flag=on
if [ "$(api_env_get AI_FEATURES_ENABLED)" = "false" ]; then
  restored_ai_flag=off
fi
require_startup_flag ai "$restored_ai_flag"

if [ "$(count_disposable_users)" != "0" ]; then
  fail "disposable rehearsal rows survived cleanup"
else
  pass "no disposable rehearsal row survived cleanup"
fi

if [ "$(row_counts)" != "$baseline_counts" ]; then
  fail "row counts differ from the pre-rehearsal baseline"
else
  pass "row counts match the pre-rehearsal baseline"
fi

if [ "$failures" -ne 0 ]; then
  echo "FAIL: $failures kill-switch/rollback check(s) failed" >&2
  exit 1
fi

echo "PASS: all four kill switches and the rollback path behaved as documented"
