# VOC-085 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01 → T02**.

## VOC-085-T00 — Production P1 content seed build, bundle, ordered run, and tests

- Requirement source: issue #702; `VOC-085-D00`–`D02`; `VOC-085-D06`;
  `VOC-085-DEP-00`; `VOC-085-DEP-01`
- Acceptance criteria: `VOC-085-AC-00`, `VOC-085-AC-01`, `VOC-085-AC-02`
  (deterministic half), `VOC-085-AC-06` (seed posture half),
  `VOC-085-AC-07` (seed tests)
- Tests: `VOC-085-TEST-00`, `VOC-085-TEST-01`, `VOC-085-TEST-02`,
  `VOC-085-TEST-03`, `VOC-085-TEST-08`
- Evidence: `VOC-085-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. In `.github/workflows/deploy-production.yml`, mirror staging's P1 content
   seed pattern:
   - Set up Go with `go-version-file: apps/api/go.mod`.
   - Build static Linux binary to `p1-content-seed` from `apps/api/cmd/seed`.
   - Bundle and copy it with the production deploy artifacts.
   - On the host, after migrations and `seed-synthetic-smoke-user.sh`, run
     the seed with the same private Postgres `DATABASE_URL` bridge resolution
     already used by migrations.
   - Place the run **before** `docker compose up -d`.
2. Ensure seed failure aborts the deploy (no `continue-on-error`) before
   application convergence.
3. Add deterministic tests covering bundling presence, ordering
   (migrations → synthetic-user seed → P1 seed → `up -d`), and fail-closed
   abort behavior.
4. Do not redesign the seed dataset, do not truncate production tables, and
   do not manually edit the live database.
5. Record commands, results, and AC mapping in `t00-evidence.md`.

### Explicitly out of scope for this task

- Content-aware smoke body assertions (T01).
- Route sweep and live Cloudflare proof (T02).
- Staging seed regression or rewrite.
- Cloudflare or topology redesign.

## VOC-085-T01 — Content-aware production smoke and detail API checks

- Requirement source: issue #702; `VOC-085-D03`; `VOC-085-D04`
- Acceptance criteria: `VOC-085-AC-03`, `VOC-085-AC-04`, `VOC-085-AC-07`
  (smoke tests)
- Tests: `VOC-085-TEST-04`, `VOC-085-TEST-05`
- Evidence: `VOC-085-EV-01` (`t01-evidence.md`)
- Status: pending — depends on `VOC-085-T00` merging to `develop` for live
  non-empty production content; the smoke script change may land first with
  selftests that prove empty→fail / non-empty→pass

### Required work

1. Update `infra/scripts/smoke-test-production.sh` so authenticated
   `GET /api/v1/journey-situations` fails when the JSON list is empty even if
   HTTP status is 200.
2. Parse the non-empty response and perform non-mutating checks for at least
   one situation detail and one word detail using identifiers from that
   response (exact API paths as already used by the product).
3. Extend `infra/scripts/smoke-test-production.selftest.sh` (or equivalent)
   for empty-content rejection, response parsing, and positive fixtures.
4. Continue to reuse `SMOKE_TEST_SESSION_COOKIE`; do not invent magic-link or
   OAuth completion paths.
5. Record evidence in `t01-evidence.md` (no secrets).

### Explicitly out of scope for this task

- Deploy seed bundling (T00).
- Full ten-route page sweep (T02).
- Dataset redesign.

## VOC-085-T02 — Authenticated route sweep and live Cloudflare verification

- Requirement source: issue #702; `VOC-085-D05`–`D07`; `VOC-085-DEP-02`
- Acceptance criteria: `VOC-085-AC-02` (live half), `VOC-085-AC-03` (live),
  `VOC-085-AC-05`, `VOC-085-AC-06`, `VOC-085-AC-07`, `VOC-085-AC-08`
- Tests: `VOC-085-TEST-06`, `VOC-085-TEST-07`, `VOC-085-TEST-09`,
  `VOC-085-TEST-10`
- Evidence: `VOC-085-EV-02` (`t02-evidence.md`)
- Status: pending — depends on `VOC-085-T00` and `VOC-085-T01` merging to
  `develop` and promoting through the normal release/deploy path

### Required work

1. Add a non-mutating authenticated production route sweep (or equivalent
   browser-level functional gate) covering:
   - `/`, `/signin`, `/auth/magic`, `/onboarding`, `/home`, `/discover`,
     `/reviews`, `/progress`, `/settings`, `/settings/account`
   - at least one real `/discover/[situation]` and one real
     `/discover/[situation]/[word]` derived from the API response
2. Reuse the workflow-minted synthetic session. Do not request magic links,
   complete OAuth, mutate a real user, or perform state-changing learning
   actions.
3. Wire the gate into `deploy-production.yml` (or the smoke suite it already
   invokes) so deploy fails closed on coverage/auth failures.
4. Add deterministic self-tests for route coverage, auth-cookie handling, and
   failure behavior.
5. After protected deployment, verify through Cloudflare that production
   returns non-empty content and dynamic details; confirm single shared-edge
   nginx, isolation, and absence of 8081/8443.
6. Record run URLs and redacted evidence in `t02-evidence.md`.

### Explicitly out of scope for this task

- Manual host/DB edits.
- Naive unauthenticated page monitors / monitoring-inventory ID creation.
- Cloudflare configuration changes.

## Task ordering notes

- T00 blocks meaningful live non-empty content for T01/T02.
- T01 may land selftests before live content exists, but must not claim
  production empty-content closure until T00 has converged a seeded
  production database.
- T02 is the package's live Cloudflare + route-coverage proof task.
- No task may be dispatched before this package is adopted and
  implementation-authorized.
- Closing issue #702 is gated on AC results with evidence, not on task issue
  closure alone.

Tasks preserve scope, separation of duties, and rollback safety.
