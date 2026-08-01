#!/usr/bin/env bash
# shellcheck shell=bash
#
# VOC-037-T03 / VOC-037-TEST-03 shared probe and assertion library.
#
# Sourced by two callers that must agree on exactly what "this kill
# switch is observably off" means:
#
#   * rehearse-production-killswitch-rollback.sh - runs against the
#     real production target on the shared host.
#   * rehearse-production-killswitch-rollback.selftest.sh - runs the
#     same probes against a disposable Postgres + api pair, and also
#     proves each probe reports FAIL when the switch is in the wrong
#     state (a probe that cannot fail proves nothing).
#
# Every probe is an observation of externally visible API behavior, not
# a read of the configuration that produced it: reading
# EMAIL_MAGIC_LINK_ENABLED back out of api.env would only prove the file
# was edited, which is not what VOC-037-AC-03 asks for.
#
# Callers must export:
#   KS_PSQL - a command that reads SQL on stdin and writes unaligned,
#             tuples-only output on stdout (e.g. `docker exec -i
#             <container> psql -U vocanova -d vocanova -qtAX`).
#
# No probe here ever prints a secret, a session token, or a magic-link
# token: tokens are passed between functions by value and only their
# resulting HTTP status is reported.

# The magic-link and OAuth request paths are rate limited to 10
# requests per minute per client IP. Every scenario stays well under
# that, but the limiter is per-IP and shared with real traffic, so a
# 429 is reported as an inconclusive probe rather than silently
# counted as "not disabled".
readonly KS_HTTP_RATE_LIMITED=429

# Postgres stores magic-link tokens as sha256(raw 32 random bytes)
# while the API accepts the base64url encoding of those same raw
# bytes (apps/api/business/auth/tokens.go). Seeding a link directly
# is what makes the NEW_USER_SIGNUP_ENABLED switch observable without
# a working email provider - production has none (kill switch off by
# deploy default), so no probe can wait for a delivered email.
readonly KS_MAGIC_LINK_TTL='15 minutes'

ks_checks_run=0
ks_failures=0

ks_reset_checks() {
  ks_checks_run=0
  ks_failures=0
}

ks_pass() {
  ks_checks_run=$((ks_checks_run + 1))
  echo "  ok: $*"
}

ks_fail() {
  ks_checks_run=$((ks_checks_run + 1))
  ks_failures=$((ks_failures + 1))
  echo "  FAIL: $*" >&2
}

# An assertion whose observed value is the string "inconclusive" is a
# failure, never a pass: an unevaluated switch check is
# indistinguishable from an unenforced one (the same rule
# rehearse-production-secrets-boundary.sh applies to its negative
# access probes).
ks_expect() {
  local label="$1" expected="$2" observed="$3"

  if [ "$observed" = "$expected" ]; then
    ks_pass "$label: $observed"
    return 0
  fi
  ks_fail "$label: expected $expected, observed $observed"
  return 1
}

ks_expect_not() {
  local label="$1" forbidden="$2" observed="$3"

  if [ "$observed" = "inconclusive" ]; then
    ks_fail "$label: probe was inconclusive"
    return 1
  fi
  if [ "$observed" != "$forbidden" ]; then
    ks_pass "$label: $observed (not $forbidden)"
    return 0
  fi
  ks_fail "$label: observed the forbidden value $forbidden"
  return 1
}

ks_summary() {
  local context="$1"

  if [ "$ks_checks_run" -eq 0 ]; then
    echo "FAIL: $context ran zero checks" >&2
    return 1
  fi
  if [ "$ks_failures" -ne 0 ]; then
    echo "FAIL: $ks_failures of $ks_checks_run $context check(s) failed" >&2
    return 1
  fi
  echo "PASS: all $ks_checks_run $context check(s) succeeded"
  return 0
}

ks_sql() {
  printf '%s\n' "$1" | $KS_PSQL
}

# curl's own failures (DNS, TLS, connection refused) must not be
# reported as an HTTP status: `000` would compare unequal to every
# expectation and read as "the switch is off" in a scenario expecting
# 503 to be absent.
ks_http_status() {
  local status
  status="$(curl -sS --max-time 15 -o /dev/null -w '%{http_code}' "$@" 2>/dev/null || true)"
  case "$status" in
    [1-5][0-9][0-9]) printf '%s' "$status" ;;
    *) printf 'inconclusive' ;;
  esac
}

ks_probe_magic_link_request() {
  local base_url="$1" email="$2" status

  status="$(ks_http_status -X POST "$base_url/api/v1/auth/magic-links" \
    -H 'Content-Type: application/json' \
    --data "{\"email\":\"$email\"}")"
  if [ "$status" = "$KS_HTTP_RATE_LIMITED" ]; then
    printf 'inconclusive'
    return
  fi
  printf '%s' "$status"
}

ks_probe_oauth_start() {
  local base_url="$1" redirect_uri="$2" status

  status="$(ks_http_status -X POST "$base_url/api/v1/auth/oauth/google/start" \
    -H 'Content-Type: application/json' \
    --data "{\"redirectUri\":\"$redirect_uri\"}")"
  if [ "$status" = "$KS_HTTP_RATE_LIMITED" ]; then
    printf 'inconclusive'
    return
  fi
  printf '%s' "$status"
}

ks_probe_healthz() {
  ks_http_status "$1/healthz"
}

# Creates one magic_links row and echoes the token that consumes it.
# The row is scoped by email, so ks_cleanup_rehearsal_rows removes it
# whether or not the consume probe reached it.
ks_seed_magic_link() {
  local email="$1" environment="$2"
  local raw_file token hash_hex

  raw_file="$(mktemp)"
  openssl rand 32 > "$raw_file"
  token="$(base64 -w0 < "$raw_file" | tr '+/' '-_')"
  hash_hex="$(openssl dgst -sha256 -binary < "$raw_file" | od -An -tx1 | tr -d ' \n')"
  rm -f "$raw_file"

  ks_sql "INSERT INTO magic_links (id, email, token_hash, environment, created_at, expires_at)
          VALUES (gen_random_uuid(), lower('$email'), decode('$hash_hex', 'hex'), '$environment',
                  now(), now() + interval '$KS_MAGIC_LINK_TTL');" > /dev/null

  printf '%s' "$token"
}

# Consumes a seeded link, writing the response headers to
# header_file so the caller can lift the session and CSRF cookies out
# of them.
#
# The raw Set-Cookie headers are used rather than a curl cookie jar
# because the session cookie's Domain is the *web* host while these
# probes talk to the *API* host, so curl's own domain matching would
# refuse to send it back (see the cookie-domain follow-up note in this
# task's evidence document).
ks_probe_magic_link_consume() {
  local base_url="$1" email="$2" token="$3" header_file="$4"

  ks_http_status -X POST "$base_url/api/v1/auth/magic-links/consume" \
    -H 'Content-Type: application/json' \
    -D "$header_file" \
    --data "{\"token\":\"$token\",\"email\":\"$email\"}"
}

ks_cookie_from_headers() {
  local header_file="$1" name="$2"

  grep -i '^set-cookie:' "$header_file" \
    | sed -n "s/.*[[:space:]]$name=\([^;[:space:]]*\).*/\1/p" \
    | tail -n1
}

# Observes AI_FEATURES_ENABLED without calling the AI provider and
# without needing a saved word: the generation gate is checked before
# the attempt is resolved (apps/api/business/aifeedback/service.go),
# so a syntactically valid but unowned attempt id separates the two
# states cleanly.
#
#   gate closed -> 200 with errorCode AI_FEEDBACK_GENERATION_DISABLED
#   gate open   -> 404, the attempt itself is not eligible
#
# Echoes "<status>/<errorCode>" so both halves can be asserted.
ks_probe_sentence_feedback() {
  local base_url="$1" session_cookie="$2" csrf="$3"
  local response status body error_code

  if [ -z "$session_cookie" ] || [ -z "$csrf" ]; then
    printf 'inconclusive'
    return
  fi

  response="$(curl -sS --max-time 20 -w '\n%{http_code}' \
    -X POST "$base_url/api/v1/sentence-feedback" \
    -H 'Content-Type: application/json' \
    -H "Idempotency-Key: $(ks_random_uuid)" \
    -H "X-CSRF-Token: $csrf" \
    -H "Cookie: vocanova_session=$session_cookie; vocanova_csrf=$csrf" \
    --data "{\"sentenceText\":\"VOC-037-T03 rehearsal sentence.\",\"source\":\"free_practice\",\"attemptId\":\"$(ks_random_uuid)\"}" \
    2>/dev/null || true)"

  status="$(printf '%s' "$response" | tail -n1)"
  body="$(printf '%s' "$response" | sed '$d')"
  case "$status" in
    [1-5][0-9][0-9]) ;;
    *) printf 'inconclusive'; return ;;
  esac

  error_code="$(printf '%s' "$body" | jq -r '.errorCode // ""' 2>/dev/null || printf '')"
  printf '%s/%s' "$status" "$error_code"
}

ks_random_uuid() {
  if command -v uuidgen > /dev/null 2>&1; then
    uuidgen | tr '[:upper:]' '[:lower:]'
    return
  fi
  ks_sql "SELECT gen_random_uuid();" | tr -d '[:space:]'
}

# Deletes every row this rehearsal can have created. Emails are
# prefixed, and oauth_states is scoped by the probe's own return URL
# and the rehearsal start time, so no pre-existing row is touched.
ks_cleanup_rehearsal_rows() {
  local email_prefix="$1" oauth_return_url="$2" started_at="$3"

  ks_sql "DELETE FROM sessions WHERE user_id IN
            (SELECT id FROM users WHERE email LIKE '$email_prefix%');
          DELETE FROM magic_links WHERE email LIKE '$email_prefix%';
          DELETE FROM users WHERE email LIKE '$email_prefix%';
          DELETE FROM oauth_states
            WHERE app_return_url = '$oauth_return_url' AND created_at >= '$started_at';" > /dev/null
}

ks_count_rehearsal_users() {
  ks_sql "SELECT count(*) FROM users WHERE email LIKE '$1%';" | tr -d '[:space:]'
}
