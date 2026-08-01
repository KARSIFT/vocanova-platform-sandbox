#!/usr/bin/env bash
set -euo pipefail

# VOC-037-T03 / VOC-037-EV-03 — deterministic harness for
# rehearse-production-killswitch-rollback.sh.
#
# The rehearsal script itself can only run against a live production
# tier, which means the one thing nobody can check while writing it is
# whether it would actually FAIL when a kill switch or the rollback path
# is broken. VOC-037-T06's review found exactly that defect in the
# secrets-boundary rehearsal: a script that printed observations without
# asserting them and passed unconditionally.
#
# This harness closes that gap without a production host. It builds a
# fake production tier out of `docker` and `curl` stubs that model the
# real contracts the rehearsal depends on:
#
#   * kill-switch values are baked in at container creation, so a
#     recreate - not a restart - is what changes behavior;
#   * apps/api/cmd/api/main.go logs one "api: listening on ... (env=,
#     ai=, magic=, oauth=, signups=)" line per start;
#   * the disabled response is 503 for magic-link, OAuth, and sign-up,
#     and a 200 carrying AI_FEEDBACK_GENERATION_DISABLED for AI;
#   * the enabled responses are, respectively, non-503, 404
#     ("not configured"), 200, and 404 ("target not found");
#   * redeploying a different image tag changes the container's image.
#
# It then runs the real script against that shape once expecting a pass,
# and once per deliberately broken control expecting a failure.
#
# The stubs model behavior, not identity: any non-empty session cookie
# with a matching CSRF header authenticates, because the fake has no
# token verification. Token/hash handling is covered by the API's own
# Go tests; what this harness proves is that the rehearsal script's
# assertions, ordering, and restoration are correct.
#
# Usage: infra/scripts/rehearse-production-killswitch-rollback.selftest.sh
# Requires no root, no docker, and no network.

repo_root="$(cd "$(dirname "$0")/../.." && pwd)"
rehearsal_script="$repo_root/infra/scripts/rehearse-production-killswitch-rollback.sh"

harness_root="$(mktemp -d)"
production_root="$harness_root/production"
stub_bin="$harness_root/bin"
fake_state="$harness_root/state"

api_host="api-production.rehearsal.invalid"
deployed_tag="sha-1111111"
rollback_tag="sha-2222222"

cleanup() {
  rm -rf "$harness_root"
}
trap cleanup EXIT

# ---------------------------------------------------------------------
# Stubs
# ---------------------------------------------------------------------

write_docker_stub() {
  cat > "$stub_bin/docker" <<'STUB'
#!/usr/bin/env bash
set -euo pipefail

state="$FAKE_STATE"

read_count() { cat "$state/counts/$1"; }
write_count() { printf '%s\n' "$2" > "$state/counts/$1"; }
bump_count() { write_count "$1" "$(( $(read_count "$1") + $2 ))"; }

image_id_for_tag() {
  printf 'sha256:%s' "$(printf '%s' "$1" | md5sum | cut -d' ' -f1)"
}

# A recreate is the only thing that promotes on-disk switch values into
# the "running" configuration, mirroring compose's env_file behavior.
recreate_api() {
  printf '%s\n' "$(( $(cat "$state/recreates") + 1 ))" > "$state/recreates"
  cp "$state/api.env" "$state/running.env"
  if [ -z "${FAKE_IGNORE_IMAGE_TAG:-}" ] && [ -n "${PRODUCTION_IMAGE_TAG:-}" ]; then
    printf '%s\n' "$PRODUCTION_IMAGE_TAG" > "$state/api_tag"
  fi

  flag() {
    value="$(sed -n "s/^$1=//p" "$state/running.env" | tail -1)"
    if [ "${value:-$2}" = "true" ]; then echo on; else echo off; fi
  }
  printf 'api: listening on :8080 (env=production, ai=%s, magic=%s, oauth=%s, signups=%s)\n' \
    "$(flag AI_FEATURES_ENABLED true)" \
    "$(flag EMAIL_MAGIC_LINK_ENABLED false)" \
    "$(flag GOOGLE_OAUTH_ENABLED false)" \
    "$(flag NEW_USER_SIGNUP_ENABLED false)" \
    >> "$state/logs"

  if [ -n "${FAKE_ROLLBACK_DROPS_ROW:-}" ] && [ "$(cat "$state/api_tag")" != "$FAKE_ORIGINAL_TAG" ]; then
    bump_count users -1
  fi
}

case "${1:-}" in
  compose)
    shift
    services=""
    for arg in "$@"; do
      case "$arg" in
        api|web) services="$services $arg" ;;
      esac
    done
    case " $* " in
      *" up "*)
        case "$services" in *api*) recreate_api ;; esac
        ;;
    esac
    exit 0
    ;;

  exec)
    # docker exec -i <postgres> psql ... , SQL on stdin
    statements="$(cat)"
    while IFS= read -r line; do
      case "$line" in
        "SELECT count(*) FROM users WHERE email LIKE"*)
          cat "$state/disposable/users"
          ;;
        "SELECT count(*) FROM "*)
          table="${line#SELECT count(*) FROM }"
          read_count "${table%;}"
          ;;
        "INSERT INTO users "*)
          bump_count users 1
          printf '%s\n' "$(( $(cat "$state/disposable/users") + 1 ))" > "$state/disposable/users"
          printf '%s\n' "$statements" | grep -oE "'[^']+@[^']+'" | tr -d "'" >> "$state/users"
          ;;
        "INSERT INTO sessions "*)
          bump_count sessions 1
          printf '%s\n' "$(( $(cat "$state/disposable/sessions") + 1 ))" > "$state/disposable/sessions"
          ;;
        "INSERT INTO magic_links "*)
          bump_count magic_links 1
          printf '%s\n' "$(( $(cat "$state/disposable/magic_links") + 1 ))" > "$state/disposable/magic_links"
          ;;
        "DELETE FROM "*)
          [ -z "${FAKE_SKIP_CLEANUP:-}" ] || continue
          table="$(printf '%s' "$line" | cut -d' ' -f3)"
          bump_count "$table" "-$(cat "$state/disposable/$table")"
          printf '0\n' > "$state/disposable/$table"
          if [ "$table" = users ]; then
            : > "$state/users"
          fi
          ;;
      esac
    done <<< "$statements"
    exit 0
    ;;

  logs)
    cat "$state/logs"
    exit 0
    ;;

  inspect)
    format="$3"
    container="$4"
    case "$format" in
      *State.Health*)
        # The fault starts only after the tier's initial start, so the
        # rehearsal's preflight passes and its first toggle is what
        # hits the never-healthy container.
        if [ -n "${FAKE_NEVER_HEALTHY:-}" ] &&
           [ "$container" = vocanova-production-api ] &&
           [ "$(cat "$state/recreates")" -gt 1 ]; then
          echo starting
        else
          echo healthy
        fi
        ;;
      *.Config.Image*)
        case "$container" in
          *-api) printf 'ghcr.io/karsift/vocanova-api:%s\n' "$(cat "$state/api_tag")" ;;
          *-web) printf 'ghcr.io/karsift/vocanova-web:%s\n' "$(cat "$state/api_tag")" ;;
          *) echo postgres:16-alpine ;;
        esac
        ;;
      *.Image*)
        image_id_for_tag "$(cat "$state/api_tag")"
        echo
        ;;
      *)
        echo "fake docker: unsupported inspect format $format" >&2
        exit 1
        ;;
    esac
    exit 0
    ;;

  image)
    # docker image inspect --format '{{.Id}}' <repo>:<tag>
    reference="$5"
    tag="${reference##*:}"
    case " $FAKE_PUBLISHED_TAGS " in
      *" $tag "*) image_id_for_tag "$tag"; echo ;;
      *) echo "fake docker: no such image $reference" >&2; exit 1 ;;
    esac
    exit 0
    ;;
esac

echo "fake docker: unsupported command: $*" >&2
exit 1
STUB
  chmod +x "$stub_bin/docker"
}

write_curl_stub() {
  cat > "$stub_bin/curl" <<'STUB'
#!/usr/bin/env bash
set -euo pipefail

state="$FAKE_STATE"

method=GET
body_file=/dev/null
url=""
request_body=""
cookie_header=""
csrf_header=""

while [ "$#" -gt 0 ]; do
  case "$1" in
    -X) method="$2"; shift 2 ;;
    -o) body_file="$2"; shift 2 ;;
    --data) request_body="$2"; shift 2 ;;
    -H)
      case "$2" in
        Cookie:*) cookie_header="$2" ;;
        X-CSRF-Token:*) csrf_header="$2" ;;
      esac
      shift 2
      ;;
    -w|--resolve|--max-time) shift 2 ;;
    -sS|--insecure) shift ;;
    https://*) url="$1"; shift ;;
    *) shift ;;
  esac
done

path="${url#https://*/}"
path="/${path}"

# Only a recreate promotes on-disk values into the running
# configuration, so the running snapshot - never api.env - is what the
# fake application reads.
switch() {
  value="$(sed -n "s/^$1=//p" "$state/running.env" | tail -1)"
  case " ${FAKE_IGNORE_SWITCH:-} " in
    *" $2 "*) echo true; return ;;
  esac
  printf '%s' "${value:-$3}"
}

respond() {
  printf '%s' "${2:-}" > "$body_file"
  printf '%s' "$1"
  exit 0
}

body_email() {
  printf '%s' "$request_body" | sed -n 's/.*"email":"\([^"]*\)".*/\1/p'
}

authenticated() {
  [ -n "$cookie_header" ] || return 1
  [ -n "$csrf_header" ] || return 1
  csrf_value="${csrf_header#X-CSRF-Token: }"
  case "$cookie_header" in
    *"vocanova_csrf=$csrf_value"*) return 0 ;;
  esac
  return 1
}

case "$path" in
  /healthz)
    respond 200 '{"status":"ok","database":"ok"}'
    ;;

  /api/v1/auth/magic-links)
    [ "$(switch EMAIL_MAGIC_LINK_ENABLED magic false)" = true ] || respond 503 '{"detail":"magic link sign-in is disabled"}'
    printf '%s\n' "$(( $(cat "$state/counts/magic_links") + 1 ))" > "$state/counts/magic_links"
    printf '%s\n' "$(( $(cat "$state/disposable/magic_links") + 1 ))" > "$state/disposable/magic_links"
    respond 200 '{}'
    ;;

  /api/v1/auth/magic-links/consume)
    [ "$(switch EMAIL_MAGIC_LINK_ENABLED magic false)" = true ] || respond 503 '{"detail":"magic link sign-in is disabled"}'
    email="$(body_email)"
    if grep -qxF "$email" "$state/users" 2>/dev/null; then
      respond 200 '{"email":"existing"}'
    fi
    [ "$(switch NEW_USER_SIGNUP_ENABLED signups false)" = true ] || respond 503 '{"detail":"new sign-ups are disabled"}'
    printf '%s\n' "$email" >> "$state/users"
    for table in users sessions; do
      printf '%s\n' "$(( $(cat "$state/counts/$table") + 1 ))" > "$state/counts/$table"
      printf '%s\n' "$(( $(cat "$state/disposable/$table") + 1 ))" > "$state/disposable/$table"
    done
    respond 200 '{"email":"created"}'
    ;;

  /api/v1/auth/oauth/google/start)
    [ "$(switch GOOGLE_OAUTH_ENABLED oauth false)" = true ] || respond 503 '{"detail":"google oauth sign-in is disabled"}'
    respond 404 '{"detail":"oauth provider not configured"}'
    ;;

  /api/v1/sentence-feedback)
    authenticated || respond 401 '{"detail":"authentication required"}'
    if [ "$(switch AI_FEATURES_ENABLED ai true)" = true ]; then
      respond 404 '{"detail":"target not found"}'
    fi
    respond 200 '{"originalSentence":"x","errorCode":"AI_FEEDBACK_GENERATION_DISABLED","canRetry":true}'
    ;;
esac

echo "fake curl: unsupported path $path" >&2
respond 000 ''
STUB
  chmod +x "$stub_bin/curl"
}

# ---------------------------------------------------------------------
# Fake production tier
# ---------------------------------------------------------------------

provision_fake_tier() {
  rm -rf "$production_root" "$stub_bin" "$fake_state"
  mkdir -p "$production_root/secrets" "$stub_bin" \
    "$fake_state/counts" "$fake_state/disposable"

  cp "$repo_root/infra/docker-compose.production.yml" "$production_root/"
  cat > "$production_root/secrets/api.env" <<'ENVFILE'
ENVIRONMENT=production
DATABASE_URL=postgres://vocanova:placeholder@postgres:5432/vocanova
AI_PROVIDER=cloudflare
AI_FEATURES_ENABLED=true
EMAIL_MAGIC_LINK_ENABLED=false
GOOGLE_OAUTH_ENABLED=false
NEW_USER_SIGNUP_ENABLED=false
ENVFILE
  chmod 600 "$production_root/secrets/api.env"
  cp "$production_root/secrets/api.env" "$harness_root/api.env.pristine"

  ln -sf "$production_root/secrets/api.env" "$fake_state/api.env"
  cp "$production_root/secrets/api.env" "$fake_state/running.env"
  printf '%s\n' "$deployed_tag" > "$fake_state/api_tag"
  : > "$fake_state/logs"
  : > "$fake_state/users"
  printf '0\n' > "$fake_state/recreates"

  for table in users sessions magic_links user_words review_attempts learner_sentences ai_feedback_attempts; do
    printf '7\n' > "$fake_state/counts/$table"
    printf '0\n' > "$fake_state/disposable/$table"
  done

  write_docker_stub
  write_curl_stub

  # One start has already happened before the rehearsal begins, exactly
  # as on a real host where the api is already running.
  PRODUCTION_IMAGE_TAG="$deployed_tag" FAKE_STATE="$fake_state" \
    FAKE_ORIGINAL_TAG="$deployed_tag" "$stub_bin/docker" compose up -d api
}

# Any trailing KEY=VALUE arguments are the fault this case injects into
# the stubs. They are passed through `env` rather than set in this
# shell, so no case can leak a fault into the next one.
run_rehearsal() {
  env \
    PATH="$stub_bin:$PATH" \
    FAKE_STATE="$fake_state" \
    FAKE_ORIGINAL_TAG="$deployed_tag" \
    FAKE_PUBLISHED_TAGS="$deployed_tag $rollback_tag" \
    VOCANOVA_PRODUCTION_ROOT="$production_root" \
    VOCANOVA_UPSTREAM_DNS_TTL_SECONDS=0 \
    VOCANOVA_HEALTH_TIMEOUT_SECONDS=4 \
    "$@" \
    "$rehearsal_script" "$api_host" "$rollback_tag_for_run"
}

expect_pass() {
  label="$1"
  shift
  echo "=== case: $label (expect PASS) ==="
  if run_rehearsal "$@"; then
    echo "--- as expected: the rehearsal passed"
  else
    echo "--- UNEXPECTED: the rehearsal failed against a correctly behaving tier" >&2
    exit 1
  fi
  echo
}

expect_fail() {
  label="$1"
  shift
  echo "=== case: $label (expect FAIL) ==="
  if run_rehearsal "$@"; then
    echo "--- UNEXPECTED: the rehearsal passed despite a broken control" >&2
    exit 1
  else
    echo "--- as expected: the rehearsal failed"
  fi
  echo
}

# ---------------------------------------------------------------------
# Cases
# ---------------------------------------------------------------------

rollback_tag_for_run="$rollback_tag"

provision_fake_tier
expect_pass "every kill switch honored and rollback changes the artifact"

echo "=== case: the passing run left no trace behind ==="
if ! diff -u "$harness_root/api.env.pristine" "$production_root/secrets/api.env"; then
  echo "--- UNEXPECTED: api.env was not restored byte-for-byte" >&2
  exit 1
fi
for table in users sessions magic_links; do
  if [ "$(cat "$fake_state/disposable/$table")" != "0" ]; then
    echo "--- UNEXPECTED: disposable $table rows survived the passing run" >&2
    exit 1
  fi
  if [ "$(cat "$fake_state/counts/$table")" != "7" ]; then
    echo "--- UNEXPECTED: $table row count drifted during the passing run" >&2
    exit 1
  fi
done
if [ "$(cat "$fake_state/api_tag")" != "$deployed_tag" ]; then
  echo "--- UNEXPECTED: the originally deployed image tag was not restored" >&2
  exit 1
fi
echo "--- as expected: switch values, image tag, and row counts all restored"
echo

for broken_switch in magic oauth signups ai; do
  provision_fake_tier
  expect_fail "$broken_switch kill switch has no effect on behavior" \
    FAKE_IGNORE_SWITCH="$broken_switch"
done

provision_fake_tier
expect_fail "rollback leaves the pre-rollback artifact running" FAKE_IGNORE_IMAGE_TAG=1

provision_fake_tier
expect_fail "rollback loses a row" FAKE_ROLLBACK_DROPS_ROW=1

provision_fake_tier
expect_fail "disposable rehearsal rows are never cleaned up" FAKE_SKIP_CLEANUP=1

provision_fake_tier
rollback_tag_for_run="sha-9999999"
expect_fail "the requested rollback tag was never published"
rollback_tag_for_run="$rollback_tag"

provision_fake_tier
expect_fail "the api container never becomes healthy after a toggle" FAKE_NEVER_HEALTHY=1

echo "SELFTEST PASS: the rehearsal script accepts a correctly behaving tier and rejects every broken control above"
