# VOC-082 — Test Plan

## VOC-082-TEST-00 — Today completed in snapshots with currentCompletion=true succeeds

- Covers: `VOC-082-AC-00`, `VOC-082-AC-02`
- Preconditions: `VOC-082-T00` tree
- Procedure:
  1. Build a `ReconcileStreak` fixture where today is present in
     `snapshots` with status `completed`, yesterday (or empty history
     for first-completion) is set per the scenario, and
     `currentCompletion=true`.
  2. Call `ReconcileStreak`.
  3. Assert no `ErrInvalidStreakSnapshot` and expected streak
     advancement / first-completion state.
- Expected result: matches issue #675's required regression shape.
- Evidence: `VOC-082-EV-00`

## VOC-082-TEST-01 — Completing-review path no longer 500s from this defect

- Covers: `VOC-082-AC-00`
- Preconditions: `VOC-082-T00` tree
- Procedure:
  1. Prefer a deterministic API/repository test that drives the
     mark-snapshot-completed then reconcile sequence used by
     `applyP4ReviewWiring`, **or** document reliance on TEST-00 plus
     staging TEST-04 if full transactional harness is unavailable.
  2. Assert the completing submission path does not error with
     `ErrInvalidStreakSnapshot` / HTTP 500 from this cause.
- Expected result: completing path succeeds or limitation is honestly
  recorded with T01 as the live proof.
- Evidence: `VOC-082-EV-00`

## VOC-082-TEST-02 — Future lastGood still rejected

- Covers: `VOC-082-AC-01`
- Preconditions: `VOC-082-T00` tree
- Procedure:
  1. Provide a completed snapshot (or force `lastGood`) dated after
     today.
  2. Call `ReconcileStreak` (with and/or without `currentCompletion` as
     needed to hit the guard).
  3. Assert `ErrInvalidStreakSnapshot` (or equivalent fail-closed
     error).
- Expected result: future-date defensive rejection remains.
- Evidence: `VOC-082-EV-00`

## VOC-082-TEST-03 — Idempotency / atomic success invariants retained

- Covers: `VOC-082-AC-02`
- Preconditions: `VOC-082-T00` tree
- Procedure:
  1. Confirm existing unique/idempotency coverage for daily-mission
     completion rewards and streak ledger keys still passes.
  2. Assert the new today+currentCompletion path does not introduce a
     second completion grant on replay if existing guards already cover
     that — add a focused assertion only if a gap is found in-scope.
- Expected result: no weakened idempotency; completing success does not
   require rolling back the transaction for this defect.
- Evidence: `VOC-082-EV-00`

## VOC-082-TEST-04 — Real staging core-loop through daily target

- Covers: `VOC-082-AC-03`
- Preconditions: `VOC-082-T00` merged to `develop`; staging deploy of
  that tip
- Procedure:
  1. Run (or record) `deploy-staging.yml`'s staging core-loop journey
     (`tests/staging-e2e/core-loop.staging.spec.ts`).
  2. Confirm success through the review that reaches the daily target;
     no HTTP 500 on that submission from this defect.
  3. Capture run URL and, if available, post-run mission snapshot /
     progress evidence.
- Expected result: PASS with run URL, or honest FAIL with run URL.
- Evidence: `VOC-082-EV-01`

## VOC-082-TEST-05 — Diff excludes VOC-081 monitor paths

- Covers: `VOC-082-AC-04`
- Preconditions: T00/T01 evidence
- Procedure:
  1. Inspect task PR file lists for
     `infra/docker-compose.monitoring.yml`, shared-edge monitor vhosts,
     Cloudflare/monitor deploy topology, or other VOC-081 paths.
  2. FAIL if those land under this package without an explicit adoption
     scope expansion (which this draft forbids by default).
- Expected result: VOC-082 diffs stay on gamification/reviews/E2E
  evidence paths named in `change.yaml`.
- Evidence: `VOC-082-EV-00`, `VOC-082-EV-01`

Include positive, negative, failure, and rollback-oriented coverage as
above. Tests must not use secrets or production data.
