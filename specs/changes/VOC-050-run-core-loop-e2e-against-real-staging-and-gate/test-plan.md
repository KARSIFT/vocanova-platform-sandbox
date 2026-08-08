# VOC-050 — Test Plan

## VOC-050-TEST-00 — Synthetic account seed is idempotent

- Covers: `VOC-050-AC-00`
- Preconditions: `VOC-050-T00` implemented; a disposable Postgres instance
  with the migration applied.
- Procedure: Run the seed mechanism twice in a row against the same
  database.
- Expected result: The second run does not create a duplicate account, does
  not error, and the account's identity (email/flag) is unchanged.
- Evidence: `VOC-050-EV-00`

## VOC-050-TEST-01 — Synthetic account is unreachable via real signup paths

- Covers: `VOC-050-AC-01`
- Preconditions: `VOC-050-T00` implemented.
- Procedure: Attempt to request a magic link, and attempt to complete OAuth
  sign-in, using the synthetic account's reserved email/identity, against a
  test or disposable-staging-shaped environment.
- Expected result: Both attempts are rejected or otherwise fail to produce a
  usable session for that identity through the real signup/sign-in paths;
  only `T01`'s dedicated minting mechanism can authenticate as this account.
- Evidence: `VOC-050-EV-00`

## VOC-050-TEST-02 — Session-minting mechanism fails closed when its gating token is unset

- Covers: `VOC-050-AC-02`
- Preconditions: `VOC-050-T01` implemented.
- Procedure: Invoke the mechanism (a) with no gating token configured, and
  (b) with an incorrect token.
- Expected result: In both cases, the mechanism performs no action and
  produces no valid session - mirroring `RegisterMonitoringSentryTest`'s
  existing unit-test pattern in `apps/api/app/api/production_test.go`.
- Evidence: `VOC-050-EV-01`

## VOC-050-TEST-03 — Real core-loop journey passes against a real staging-shaped target

- Covers: `VOC-050-AC-03`
- Preconditions: `VOC-050-T02` implemented; a real or disposable
  staging-shaped deployment reachable over HTTPS with the synthetic account
  seeded and a minted session available.
- Procedure: Run the staging-targeted core-loop check exactly as
  `deploy-staging.yml`'s new step will invoke it.
- Expected result: The full journey (onboarding-state-appropriate start,
  discover, save, review, sentence feedback, progress, settings, logout,
  auth-gate rejection) passes against the real backend and real Postgres,
  with no reliance on `mock-api-server.mjs`.
- Evidence: `VOC-050-EV-02`

## VOC-050-TEST-04 — A deliberately-failing core-loop run fails deploy-staging.yml's job

- Covers: `VOC-050-AC-04`
- Preconditions: `VOC-050-T02` and `VOC-050-T03` implemented; a disposable
  test branch or environment where the core-loop check can be made to fail
  deliberately (e.g. pointed at a target missing the expected content).
- Procedure: Trigger `deploy-staging.yml` (or its `workflow_dispatch` path)
  against the deliberately-broken target.
- Expected result: The workflow run's overall conclusion is failure, with no
  `continue-on-error` masking the core-loop step's failure. The task's pull
  request additionally documents, in writing, the still-open cross-repo
  dependency on `karsift-ai-infra`'s `release.yml` per
  `specification.md`'s open question 1 - this test does not (and cannot)
  verify that repository's behavior.
- Evidence: `VOC-050-EV-03`

## VOC-050-TEST-05 — Production smoke-test suite's core-loop section runs for real, without side effects

- Covers: `VOC-050-AC-05`
- Preconditions: `VOC-050-T04` implemented; a real or disposable
  production-shaped target with the synthetic account seeded and
  `SMOKE_TEST_SESSION_COOKIE` set to a value minted by `T01`'s mechanism.
- Procedure: Run `infra/scripts/smoke-test-production.sh` against the
  target, exactly as `deploy-production.yml`'s step will invoke it.
  Separately, confirm via code inspection that no request added or changed
  by this package performs a `POST`/`PUT`/`PATCH`/`DELETE` against a
  state-mutating endpoint.
- Expected result: The script's `# 5.` section prints `PASS:` lines for both
  the `/api/v1/me` and `/api/v1/journey-situations` checks instead of the
  previous `SKIP:` line, and the script's overall exit code reflects genuine
  pass/fail rather than a masked skip. No state-mutating request is added.
- Evidence: `VOC-050-EV-04`

Positive coverage (`TEST-00`, `TEST-03`, `TEST-05`), negative/security
coverage (`TEST-01`, `TEST-02`), and failure-path coverage (`TEST-04`) are
included. Migration coverage is folded into `TEST-00` (idempotent rerun).
Accessibility coverage is not applicable per `impact-analysis.md`'s
determination - the underlying journey's accessibility coverage already
exists separately via T07a/T07b. No test in this plan uses real production
data or a real user's credentials; all tests operate on the dedicated
synthetic account or a disposable/test environment.
