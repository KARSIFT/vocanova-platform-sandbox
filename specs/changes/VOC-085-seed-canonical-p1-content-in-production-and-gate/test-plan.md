# VOC-085 — Test Plan

## VOC-085-TEST-00 — Production deploy bundles the P1 content seed binary

- Covers: `VOC-085-AC-00`
- Preconditions: `VOC-085-T00` tree; deterministic workflow/config harness
- Procedure:
  1. Inspect or fixture-drive the production deploy bundle step.
  2. Assert the Go setup uses `go-version-file: apps/api/go.mod`.
  3. Assert a `p1-content-seed` (or equivalent documented path) artifact is
     built from `apps/api/cmd/seed` and included in the production bundle.
- Expected result: binary present in bundle; toolchain pin matches staging
  convention.
- Evidence: `VOC-085-EV-00`

## VOC-085-TEST-01 — Seed runs after migrations and synthetic-user seed, before up -d

- Covers: `VOC-085-AC-00`
- Preconditions: `VOC-085-T00` tree
- Procedure:
  1. Deterministically assert step/script order in
     `deploy-production.yml` (or extracted helper): migrations →
     synthetic-user seed → P1 content seed → `docker compose up -d`.
  2. Assert P1 seed uses the private Postgres bridge `DATABASE_URL`
     resolution pattern already used by migrations.
- Expected result: documented order enforced; no seed after convergence.
- Evidence: `VOC-085-EV-00`

## VOC-085-TEST-02 — Seed failure aborts before application convergence

- Covers: `VOC-085-AC-01`
- Preconditions: `VOC-085-T00` tree
- Procedure:
  1. Drive a disposable fixture where the P1 seed command exits non-zero.
  2. Assert the deploy path does not proceed to `up -d` / does not claim
     success.
  3. Assert no `continue-on-error` on the seed step.
- Expected result: fail-closed abort before convergence.
- Evidence: `VOC-085-EV-00`

## VOC-085-TEST-03 — Idempotent upsert / no duplicate canonical rows

- Covers: `VOC-085-AC-02`
- Preconditions: existing `apps/api/cmd/seed` tests and/or disposable DB
  fixture; do not use production credentials in CI fixtures
- Procedure:
  1. Apply seed twice against a disposable database (or rely on existing
     seed unit/integration coverage if already sufficient and referenced).
  2. Assert canonical situation/word counts remain stable (no duplicates).
  3. Assert no DELETE of user-owned learning tables is introduced.
- Expected result: idempotent upsert semantics preserved.
- Evidence: `VOC-085-EV-00`

## VOC-085-TEST-04 — Empty journey-situations body fails smoke even on HTTP 200

- Covers: `VOC-085-AC-03`
- Preconditions: `VOC-085-T01` tree; smoke selftest harness
- Procedure:
  1. Fixture an HTTP 200 response with an empty situation list.
  2. Run the smoke suite / extracted checker with a synthetic session cookie
     present.
  3. Assert deterministic failure (not SKIP, not PASS).
- Expected result: empty content is a hard fail.
- Evidence: `VOC-085-EV-01`

## VOC-085-TEST-05 — Situation and word detail checks from returned identifiers

- Covers: `VOC-085-AC-04`
- Preconditions: `VOC-085-T01` tree
- Procedure:
  1. Fixture a non-empty journey-situations payload with at least one
     situation identifier and enough linkage to resolve one word detail.
  2. Assert the smoke/detail checker issues non-mutating detail requests and
     requires successful responses.
  3. Assert mutating methods are not used.
- Expected result: detail verification from returned IDs; read-only.
- Evidence: `VOC-085-EV-01`

## VOC-085-TEST-06 — Fixed route coverage list is complete

- Covers: `VOC-085-AC-05`
- Preconditions: `VOC-085-T02` tree
- Procedure:
  1. Assert the route-sweep configuration/script enumerates exactly the ten
     fixed routes required by AC-05 (missing any fails the test).
  2. Assert dynamic situation and word routes are derived from API data, not
     hard-coded to a single staging-only slug unless also validated live.
- Expected result: coverage list matches AC; dynamic routes are data-derived.
- Evidence: `VOC-085-EV-02`

## VOC-085-TEST-07 — Auth-cookie handling and fail-closed route sweep

- Covers: `VOC-085-AC-05`
- Preconditions: `VOC-085-T02` tree
- Procedure:
  1. With a valid `vocanova_session=...` cookie header fixture, assert routes
     are requested with that cookie where authentication is required.
  2. With cookie absent/malformed, assert authenticated protected routes fail
     closed (suite fails; does not silently skip required coverage).
  3. Assert the harness never triggers magic-link send or OAuth completion.
- Expected result: cookie required for protected coverage; no auth side
  effects.
- Evidence: `VOC-085-EV-02`

## VOC-085-TEST-08 — Synthetic account onboarding remains seed-owned

- Covers: `VOC-085-AC-06`
- Preconditions: existing synthetic-user SQL/script; T00 does not replace it
- Procedure:
  1. Confirm repository `seed-synthetic-smoke-user.sql` still sets
     `onboarding_status='completed'` and `is_synthetic_test_account=true`
     idempotently.
  2. Assert no task introduces a manual live-DB edit procedure as the
     acceptance path.
- Expected result: synthetic posture remains repository-seed deterministic.
- Evidence: `VOC-085-EV-00`, `VOC-085-EV-02`

## VOC-085-TEST-09 — Live production Cloudflare verification

- Covers: `VOC-085-AC-03`, `VOC-085-AC-07`
- Preconditions: T00–T02 merged and promoted; protected
  `deploy-production.yml` run succeeded
- Procedure:
  1. From the deploy run (or equivalent controlled post-deploy check through
     Cloudflare hostnames), confirm non-empty
     `/api/v1/journey-situations` and successful situation/word detail
     checks.
  2. Record run URL and redacted outputs in `t02-evidence.md`.
- Expected result: live production content gate green through Cloudflare.
- Evidence: `VOC-085-EV-02`

## VOC-085-TEST-10 — Topology and isolation remain intact

- Covers: `VOC-085-AC-08`
- Preconditions: production deploy of this package's revision
- Procedure:
  1. Confirm single shared-edge nginx remains the public edge.
  2. Confirm staging/production secret, directory, deploy-user, database, and
     network isolation still hold.
  3. Confirm ports 8081/8443 remain absent.
- Expected result: VOC-067 invariants preserved.
- Evidence: `VOC-085-EV-02`

Include positive, negative, authorization, failure, and rollback-relevant
coverage as above. Tests must not use secrets or production data in
repository fixtures.
