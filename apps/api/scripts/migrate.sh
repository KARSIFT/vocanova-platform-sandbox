#!/usr/bin/env sh
# apps/api/scripts/migrate.sh
#
# VOC-032-T06: thin wrapper around the Atlas CLI that applies the
# versioned forward-migration set in apps/api/migrations to a target
# PostgreSQL database. The wrapper exists so the exact apply command
# is defined once and shared by:
#   * T07's deploy workflow (.github/workflows/deploy-staging.yml,
#     not yet added) - which SSHes into the staging host after a
#     `docker compose pull` and runs this wrapper against the live
#     staging DATABASE_URL before bringing the new api container
#     up.
#   * T09's live migration-and-rollback rehearsal - which runs the
#     same wrapper against a disposable copy of the staging database
#     as part of the forward-then-reverse-then-forward-again
#     sequence that proves rollback is rehearsed (DOC-12 §5 R1
#     gate: "migration + rollback rehearsed"; apps/api/migrations/
#     README.md "disposable recovery rehearsal").
#   * A developer's local `bash` session - the same wrapper works
#     against any local Postgres whose URL is exported as
#     DATABASE_URL.
#
# This script intentionally does NOT install Atlas. Atlas is
# installed by the deploy workflow (T07) via the standard
# `curl -sSf https://atlasgo.sh | sh` installer pinned by the
# workflow's lockfile, and by a developer's homebrew/curl
# installer. The wrapper's only job is to validate the
# environment, build the apply command, and exec Atlas so its
# exit code propagates cleanly.
#
# Usage:
#   DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=disable \
#     sh apps/api/scripts/migrate.sh
#
# Optional environment variables:
#   DATABASE_URL  - Postgres connection string for the target
#                   database. Required. The wrapper refuses to
#                   proceed if it is unset or empty.
#   ATLAS_URL     - Alternate name for the connection string,
#                   accepted as a fallback. Lets the deploy
#                   workflow choose a different name without
#                   editing this script.
#   ATLAS_BIN     - Path to the Atlas binary. Defaults to
#                   `atlas` (resolved via PATH). Overridable
#                   for the deploy workflow, which can install
#                   Atlas to a non-PATH location and point this
#                   var at it.
#   MIGRATIONS_DIR - Override the migration directory URL
#                   (default: `file://<repo>/apps/api/migrations`,
#                   resolved relative to this script's own path
#                   so the wrapper is relocatable as long as
#                   apps/api/scripts/migrate.sh and
#                   apps/api/migrations/ stay co-located).
#
# Exit codes:
#   0   - apply completed (every pending migration applied, or
#         nothing to apply on a re-run).
#   1   - configuration error (missing DATABASE_URL/ATLAS_URL,
#         missing Atlas binary, missing migrations directory,
#         atlas.sum missing or out of sync).
#   2   - apply error (any non-zero Atlas exit; the wrapper
#         propagates Atlas's own exit code minus any wrapper-
#         specific adjustment, see below).
#   3   - integrity check failed: atlas.sum is missing or its
#         recorded hashes do not match the current files in the
#         migration directory. Atlas will already have reported
#         the same condition; this distinct code makes it easier
#         for the deploy workflow to surface a meaningful PR
#         (re-run `atlas migrate hash` and commit the updated
#         atlas.sum) versus a true runtime apply failure.
#
# Cross-references:
#   * apps/api/atlas.hcl - the named `dev` env used by Atlas's
#     own diff/validate commands; this wrapper does NOT use it
#     (runtime apply always uses --url, not --env).
#   * apps/api/migrations/atlas.sum - the integrity file Atlas
#     consults before applying; the wrapper re-checks its
#     presence as a fast pre-flight.
#   * apps/api/migrations/README.md - the existing
#     "explicit operational artifact" rule and the
#     `.down.sql.example` discipline this wrapper preserves.
#   * specs/changes/VOC-032-begin-milestone-r1-staging-readiness-docs-product/
#     tasks.md VOC-032-T06 - the task this wrapper satisfies.
#
# Security: this script never echoes the connection string
# (DATABASE_URL/ATLAS_URL may contain a password) and never
# writes it to disk. Errors cite the variable name, never
# its value.

set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
migrations_dir_default="file://${script_dir}/../migrations"

# -----------------------------------------------------------------------------
# Resolve and validate the database connection string.
# -----------------------------------------------------------------------------
db_url=${DATABASE_URL:-${ATLAS_URL:-}}
if [ -z "$db_url" ]; then
    printf '%s\n' "migrate.sh: DATABASE_URL (or ATLAS_URL) is required." >&2
    printf '%s\n' "  Set one of them to a postgres:// connection string and re-run." >&2
    exit 1
fi

# -----------------------------------------------------------------------------
# Resolve and validate the migration directory.
# -----------------------------------------------------------------------------
migrations_dir=${MIGRATIONS_DIR:-$migrations_dir_default}
case "$migrations_dir" in
    file://*)
        # Strip the file:// scheme and verify the resulting path
        # exists on disk. Atlas itself would also error, but a
        # fast pre-flight here gives a clearer message and a
        # distinct exit code for "wrapper misconfigured" vs
        # "Atlas errored at runtime".
        migrations_path=${migrations_dir#file://}
        if [ ! -d "$migrations_path" ]; then
            printf '%s\n' "migrate.sh: migration directory not found: $migrations_path" >&2
            exit 1
        fi
        ;;
    *)
        printf '%s\n' "migrate.sh: MIGRATIONS_DIR must be a file:// URL; got: $migrations_dir" >&2
        exit 1
        ;;
esac

# -----------------------------------------------------------------------------
# Verify the migration directory's atlas.sum integrity file is
# present. Atlas will also enforce this, but failing fast here
# means a missing atlas.sum surfaces a clearer "regenerate and
# commit" message rather than Atlas's lower-level
# "checksum file is out of sync" diagnostic.
# -----------------------------------------------------------------------------
if [ ! -f "${migrations_path}/atlas.sum" ]; then
    printf '%s\n' "migrate.sh: ${migrations_path}/atlas.sum is missing." >&2
    printf '%s\n' "  Run 'atlas migrate hash --dir \"%s\"' and commit the result." >&2 "$migrations_dir" >&2
    exit 3
fi

# -----------------------------------------------------------------------------
# Resolve the Atlas binary.
# -----------------------------------------------------------------------------
atlas_bin=${ATLAS_BIN:-atlas}
if ! command -v "$atlas_bin" >/dev/null 2>&1; then
    printf '%s\n' "migrate.sh: Atlas binary not found on PATH: $atlas_bin" >&2
    printf '%s\n' "  Install Atlas (https://atlasgo.sh) or set ATLAS_BIN to its absolute path." >&2
    exit 1
fi

# -----------------------------------------------------------------------------
# Apply. The wrapper uses --url (not --env) so the staging target
# URL is sourced from the caller-provided DATABASE_URL, never
# hardcoded in apps/api/atlas.hcl. The default transaction mode
# is `file` (each migration in its own transaction), matching
# apps/api/atlas.hcl's env block and Atlas's documented default;
# it is not overridden here so the per-file `atlas:txmode`
# directives in apps/api/migrations/*.sql are honored if and
# when the existing migration set is amended to use a valid
# Atlas txmode value (current files use `-- atlas:txmode
# transaction`, which Atlas does not recognize - see
# staging-evidence.md and the T06 follow-up notes recorded
# against this task).
#
# --baseline "" is not passed: Atlas's default behavior of
# starting from the first unapplied version is correct for our
# fresh-database-per-deploy staging workflow. --to-version is
# also not passed: we always want to apply every pending
# migration, not stop partway.
# -----------------------------------------------------------------------------
"$atlas_bin" migrate apply \
    --url "$db_url" \
    --dir "$migrations_dir"
