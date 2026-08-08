#!/usr/bin/env sh
# apps/api/scripts/seed-synthetic-smoke-user.sh
#
# VOC-050-T00: ensure the synthetic smoke-test account exists on the
# target environment. The deploy workflows run this immediately after
# apps/api/scripts/migrate.sh, while only the Postgres service is up,
# so the account is present before the rest of the stack starts and
# before any smoke check tries to authenticate as it.
#
# The SQL lives next to this script in seed-synthetic-smoke-user.sql
# and is idempotent (see its header): rerunning it on every deploy
# neither duplicates the account nor errors.
#
# Usage:
#   sh apps/api/scripts/seed-synthetic-smoke-user.sh
#
# Environment variables:
#   VOCANOVA_SYNTHETIC_SMOKE_TEST_EMAIL - reserved, non-deliverable
#                   identity for the account. Must match the API's own
#                   VOCANOVA_SYNTHETIC_SMOKE_TEST_EMAIL, which is what
#                   blocks the address from every real sign-in path.
#                   Defaults to the same reserved .invalid address the
#                   API defaults to.
#   DOCKER_COMPOSE_CMD - Compose invocation that resolves the target
#                   stack (production passes its own -f/-p flags).
#                   Defaults to `docker compose`.
#   POSTGRES_USER / POSTGRES_DB - database role and name inside the
#                   Postgres container. Default to `vocanova`, matching
#                   infra/docker-compose.yml and
#                   infra/docker-compose.production.yml.

set -eu

default_synthetic_email="smoke-test-bot@synthetic.vocanova.invalid"
synthetic_display_name="VOC-050 Synthetic Smoke Test User"

synthetic_email="${VOCANOVA_SYNTHETIC_SMOKE_TEST_EMAIL:-}"
if [ -z "$synthetic_email" ]; then
  synthetic_email="$default_synthetic_email"
fi

postgres_user="${POSTGRES_USER:-vocanova}"
postgres_db="${POSTGRES_DB:-vocanova}"

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
seed_sql="$script_dir/seed-synthetic-smoke-user.sql"
if [ ! -f "$seed_sql" ]; then
  echo "ERROR: seed SQL not found at $seed_sql" >&2
  exit 1
fi

# Deliberately unquoted: DOCKER_COMPOSE_CMD carries multiple words
# (e.g. `docker compose -f docker-compose.production.yml -p vocanova-production`)
# and must word-split into separate arguments.
# shellcheck disable=SC2086
${DOCKER_COMPOSE_CMD:-docker compose} exec -T postgres \
  psql \
    --set=ON_ERROR_STOP=1 \
    --username "$postgres_user" \
    --dbname "$postgres_db" \
    --set=synthetic_email="$synthetic_email" \
    --set=synthetic_display_name="$synthetic_display_name" \
    --file - < "$seed_sql"

echo "seeded synthetic smoke-test user: ${synthetic_email}"
