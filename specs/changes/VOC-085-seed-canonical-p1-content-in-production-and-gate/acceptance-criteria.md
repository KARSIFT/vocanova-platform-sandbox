# VOC-085 — Acceptance Criteria

## VOC-085-AC-00 — Production deploy builds, bundles, and runs the P1 seed in order

- Requirement source: issue #702; `VOC-085-D00`
- Tasks: `VOC-085-T00`
- Tests: `VOC-085-TEST-00`, `VOC-085-TEST-01`
- Evidence: `VOC-085-EV-00`
- Result: pending

Observable outcome:

1. `deploy-production.yml` builds the existing `apps/api/cmd/seed` as a
   static Linux binary using the repository Go toolchain pin
   (`go-version-file: apps/api/go.mod`), bundles it into the production
   deploy artifacts, and runs it against the production-private database.
2. Execution order is: migrations → synthetic-user seed → P1 content seed →
   `docker compose up -d` (application convergence).
3. Staging's existing seed behavior is not regressively removed.

## VOC-085-AC-01 — Seed/migration failure aborts before application convergence

- Requirement source: issue #702; `VOC-085-D01`
- Tasks: `VOC-085-T00`
- Tests: `VOC-085-TEST-02`
- Evidence: `VOC-085-EV-00`
- Result: pending

Observable outcome:

1. A migration, synthetic-user seed, or P1 content seed failure exits
   non-zero before `docker compose up -d`.
2. The workflow does not report partial success for that deploy attempt.

## VOC-085-AC-02 — Idempotent upsert preserves user data and avoids duplicates

- Requirement source: issue #702; `VOC-085-D02`
- Tasks: `VOC-085-T00`
- Tests: `VOC-085-TEST-03`
- Evidence: `VOC-085-EV-00`, `VOC-085-EV-02`
- Result: pending

Observable outcome:

1. Repeated production deploys do not duplicate canonical situations/words.
2. User-owned learning state is not overwritten beyond the seed's fixed
   primary-key upsert semantics for repository-owned canonical rows.
3. Evidence may combine deterministic seed-tool tests with live post-deploy
   non-empty content counts that remain stable across a redeploy.

## VOC-085-AC-03 — Production journey-situations is non-empty on success

- Requirement source: issue #702; `VOC-085-D03`
- Tasks: `VOC-085-T01`, `VOC-085-T02`
- Tests: `VOC-085-TEST-04`, `VOC-085-TEST-09`
- Evidence: `VOC-085-EV-01`, `VOC-085-EV-02`
- Result: pending

Observable outcome:

1. Production `GET /api/v1/journey-situations` returns HTTP 200 and a
   non-empty canonical list when the deploy succeeds.
2. Production smoke fails deterministically on an empty content response even
   when HTTP status is 200.

## VOC-085-AC-04 — Real situation and word details are verified

- Requirement source: issue #702; `VOC-085-D04`
- Tasks: `VOC-085-T01`
- Tests: `VOC-085-TEST-05`
- Evidence: `VOC-085-EV-01`, `VOC-085-EV-02`
- Result: pending

Observable outcome:

1. Using the reserved synthetic session, at least one situation detail and
   one word detail are verified from identifiers returned by production.
2. Checks are non-mutating (GET / read-only equivalent).

## VOC-085-AC-05 — Authenticated non-mutating route sweep covers fixed and dynamic routes

- Requirement source: issue #702; `VOC-085-D05`
- Tasks: `VOC-085-T02`
- Tests: `VOC-085-TEST-06`, `VOC-085-TEST-07`
- Evidence: `VOC-085-EV-02`
- Result: pending

Observable outcome:

1. A non-mutating authenticated gate covers all ten fixed routes:
   `/`, `/signin`, `/auth/magic`, `/onboarding`, `/home`, `/discover`,
   `/reviews`, `/progress`, `/settings`, `/settings/account`.
2. The same gate covers at least one real `/discover/[situation]` and one
   real `/discover/[situation]/[word]` route derived from the API response.
3. The workflow-minted synthetic session is reused; no magic links, OAuth
   completion, real-user mutation, or state-changing learning actions occur.

## VOC-085-AC-06 — Synthetic account posture remains deterministic via repository seed

- Requirement source: issue #702; `VOC-085-D06`
- Tasks: `VOC-085-T00`, `VOC-085-T02`
- Tests: `VOC-085-TEST-08`
- Evidence: `VOC-085-EV-00`, `VOC-085-EV-02`
- Result: pending

Observable outcome:

1. The reserved synthetic account remains active, verified, marked
   synthetic, and has onboarding completed by its idempotent repository
   seed.
2. No manual live-database edit is used to achieve that posture.
3. No real user's state is read through privileged shortcuts or changed by
   verification.

## VOC-085-AC-07 — Deterministic tests and live Cloudflare verification pass

- Requirement source: issue #702; `VOC-085-D00`–`D07`
- Tasks: `VOC-085-T00`, `VOC-085-T01`, `VOC-085-T02`
- Tests: `VOC-085-TEST-00` through `VOC-085-TEST-10`
- Evidence: `VOC-085-EV-00`, `VOC-085-EV-01`, `VOC-085-EV-02`
- Result: pending

Observable outcome:

1. Tests cover positive and fail-closed cases for seed ordering/bundling,
   empty-content rejection, response parsing, route coverage, auth-cookie
   handling, and failure behavior.
2. Deployment and live verification pass through Cloudflare for non-empty
   content and dynamic details.

## VOC-085-AC-08 — Topology and isolation remain intact

- Requirement source: issue #702; `VOC-085-D07`
- Tasks: `VOC-085-T00`, `VOC-085-T02`
- Tests: `VOC-085-TEST-10`
- Evidence: `VOC-085-EV-02`
- Result: pending

Observable outcome:

1. Single shared-edge nginx remains the public edge.
2. Database/secret/network/directory/deploy-user isolation between staging
   and production remains intact.
3. Ports 8081/8443 remain absent.
