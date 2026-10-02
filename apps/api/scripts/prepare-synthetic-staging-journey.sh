#!/usr/bin/env sh
# Explicit fresh fixture for the staging browser journey. Source the staging
# API environment first; never call this from production or the default seed.
# The workflow must hold its shared staging account lock through browser exit.
set -eu

fail() { printf '%s\n' "$1" >&2; exit 1; }
[ "$#" -eq 0 ] || fail 'synthetic journey preparation takes no arguments'
# The application environment authorizes preparation; monitoring is optional.
[ "${ENVIRONMENT:-}" = staging ] || fail 'synthetic journey preparation requires staging'
for override in "${DOCKER_COMPOSE_CMD:-}" "${COMPOSE_FILE:-}" "${COMPOSE_PROJECT_NAME:-}" "${DOCKER_HOST:-}" "${DOCKER_CONTEXT:-}"; do
  [ -z "$override" ] || fail 'synthetic journey preparation refuses Docker scope overrides'
done

synthetic_email="${VOCANOVA_SYNTHETIC_SMOKE_TEST_EMAIL:-smoke-test-bot@synthetic.vocanova.invalid}"
# Keep email limits and pattern in sync with prepare-synthetic-staging-journey.sql.
case "$synthetic_email" in *[!a-z0-9._%+@-]*) fail 'synthetic journey requires a canonical lowercase .invalid email' ;; esac
[ "${#synthetic_email}" -le 254 ] || fail 'synthetic journey requires a canonical lowercase .invalid email'
printf '%s\n' "$synthetic_email" | LC_ALL=C grep -Eq '^[a-z0-9][a-z0-9._%+-]{0,63}@[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(\.[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)*\.invalid$' || fail 'synthetic journey requires a canonical lowercase .invalid email'

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/../../.." && pwd)
prepare_sql="$script_dir/prepare-synthetic-staging-journey.sql"
[ -f "$prepare_sql" ] || fail 'synthetic journey preparation SQL is missing'
[ -f "$repo_root/infra/docker-compose.yml" ] || fail 'staging Compose file is missing'

docker --context default compose \
  --project-name vocanova-staging \
  --file "$repo_root/infra/docker-compose.yml" \
  exec -T postgres psql -X --set=ON_ERROR_STOP=1 \
  --username "${POSTGRES_USER:-vocanova}" \
  --dbname "${POSTGRES_DB:-vocanova}" \
  --set=journey_environment="$ENVIRONMENT" \
  --set=synthetic_email="$synthetic_email" \
  --file - < "$prepare_sql"
