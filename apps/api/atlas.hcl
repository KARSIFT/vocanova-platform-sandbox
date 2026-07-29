# apps/api/atlas.hcl
#
# Atlas CLI configuration for the versioned SQL migrations in
# apps/api/migrations. This is the configuration file Atlas
# reads when invoked with the `--env dev` flag (or as the
# implicit default when `atlas migrate <subcommand>` is run
# from the apps/api directory).
#
# Package: VOC-032 (Begin Milestone R1 - Staging Readiness)
# Task:    VOC-032-T06
# Risk:    R3 (per docs/governance/change-risk-classification.md;
#          apps/api/migrations is in the `*/migrations/*`
#          path-class, which the classifier floors at R3).
# Authority: VOC-032 change package (adopted 2026-07-28;
#            specs/changes/VOC-032-begin-milestone-r1-staging-readiness-docs-product/).
# Depends on: T00 (apps/api/cmd/api); consulted by T07 (deploy
#             workflow) and T09 (live rehearsal) via
#             apps/api/scripts/migrate.sh.
#
# Cross-references:
#   * apps/api/migrations/*.sql - the versioned forward
#     migrations Atlas applies in timestamp order. The
#     matching *.down.sql.example files are deliberately
#     outside Atlas's file-discovery glob (Atlas's default
#     versioned glob is `*.sql`); this preserves
#     VOC-032-D08's rule that recovery down-files are
#     never executable by the forward-apply path.
#   * apps/api/migrations/atlas.sum - the integrity hash of
#     the migration directory; Atlas's `migrate apply`
#     aborts with an error if the directory state does not
#     match this file. Regenerate it with
#     `atlas migrate hash --dir "file://migrations"`
#     whenever a new migration is added or an existing
#     migration is edited.
#   * apps/api/migrations/README.md - the existing
#     operational note declaring "Versioned Atlas SQL
#     remains the migration authority" and the
#     `.down.sql.example` discipline.
#   * apps/api/scripts/migrate.sh - the wrapper that invokes
#     `atlas migrate apply --url <runtime-url> --dir
#     "file://migrations"`; defined once so T07's deploy
#     workflow and T09's rehearsal do not duplicate the
#     apply command.
#   * apps/api/ent/README.md - the parallel rule that
#     schema creation is never run by API startup;
#     migrations and ent schema are out of scope for the
#     API process either way.
#   * docs/operations/11-devops-and-ci-cd.md §3 (kill
#     switches) and the migration-tooling rules in
#     `apps/api/migrations/README.md` - the operational
#     constraints this config respects.
#
# Why a dev env block AND a runtime `--url` flag:
# Atlas's `env` block is a *named* connection profile. The
# `dev` env below is for Atlas's own diff/lint workflow
# (`atlas migrate diff --env dev`), which compares the
# migration directory against a scratch Postgres the
# developer (or CI) is expected to have running locally.
# The runtime deploy path - both the T07 CI/CD workflow
# and the T09 manual rehearsal - does NOT use `--env dev`.
# It passes the staging `DATABASE_URL` directly via
# `--url`, so the staging target's URL is never hardcoded
# anywhere in this file and the same wrapper works against
# any future environment (production, a second staging,
# a developer's local Postgres) by changing only the
# caller-provided env var.
#
# Secret handling: no credential is written into this file.
# The dev `url` below is a non-routable localhost value
# intended only for a developer-machine disposable Postgres
# (e.g. `docker run --rm -p 5432:5432 postgres:16-alpine`).
# The founder never sets `DATABASE_URL` to a real value
# here; runtime apply uses `--url "$DATABASE_URL"` where
# `DATABASE_URL` is sourced from the untracked, host-only
# secret file infra/secrets/api.env on the staging host
# (matching the convention T01's .env.example documents
# and T04's docker-compose.yml enforces).

env "dev" {
  # Local-development-only Postgres URL. Atlas's `migrate
  # diff` and `migrate validate --env dev` operations
  # connect here; it is intentionally a localhost value
  # with no embedded credential. Override on the command
  # line with `atlas migrate apply --url "$DATABASE_URL"`
  # for any non-local target.
  url = "postgres://vocanova:vocanova@127.0.0.1:5432/vocanova?sslmode=disable"

  # Where the versioned forward-migration files live.
  # Atlas's default file glob inside a versioned directory
  # is `*.sql`; the `*.down.sql.example` files in this
  # directory are deliberately outside that glob and
  # therefore not discoverable by the forward-apply
  # command (VOC-032-D08, VOC-032-TEST-15).
  migration {
    dir = "file://migrations"
  }
}
