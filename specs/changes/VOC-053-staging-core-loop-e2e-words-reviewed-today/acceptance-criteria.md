# VOC-053 — Acceptance Criteria

## VOC-053-AC-00 — Root cause is confirmed with real evidence, not assumed

- Requirement source: issue #450's "candidate root causes ... not
  prescriptive" framing; `specification.md` open question 1
- Tasks: `VOC-053-T00`
- Tests: `VOC-053-TEST-00`
- Evidence: `VOC-053-EV-00`
- Result: pending

The implementer records which of issue #450's three candidates (or a fourth,
newly identified cause) actually explains the observed same-run decrease,
backed by real evidence — live staging HTTP response headers/cache-status,
and/or a traced backend request-handling code path — not by static reading
alone and not by inference from the failure symptom without direct
confirmation.

## VOC-053-AC-01 — The confirmed root cause is fixed without weakening the test

- Requirement source: `specification.md` scope item 2
- Tasks: `VOC-053-T01`
- Tests: `VOC-053-TEST-01`
- Evidence: `VOC-053-EV-01`
- Result: pending

`tests/staging-e2e/core-loop.staging.spec.ts`'s step 7 assertion
(`reviewedAfter >= reviewedBefore + reviewedCards`) is unchanged in the fix's
diff. The fix addresses the confirmed root cause from `VOC-053-AC-00`
specifically (e.g. an explicit `cache: "no-store"` fetch option if caching is
confirmed; a corrected local-date/timezone resolution if a backend bug is
confirmed), not a retry, poll-until-pass, or assertion-loosening workaround.

## VOC-053-AC-02 — "Words reviewed today" is genuinely monotonically
non-decreasing within the same local day

- Requirement source: issue #450's reported failure; `specification.md`
  objective
- Tasks: `VOC-053-T01`
- Tests: `VOC-053-TEST-02`
- Evidence: `VOC-053-EV-01`
- Result: pending

Two reads of the daily mission's `reviewsCompleted` value for the same user
and the same local calendar day, taken seconds apart with no intervening
review completed, return the same value — not a decrease. This is verified
directly (not only inferred from the E2E spec passing).

## VOC-053-AC-03 — Real staging core-loop E2E step 7 passes reliably, including
under the exact failure condition observed

- Requirement source: issue #450's reported failure (run 31332238452)
- Tasks: `VOC-053-T02`
- Tests: `VOC-053-TEST-03`
- Evidence: `VOC-053-EV-02`
- Result: pending

`tests/staging-e2e/core-loop.staging.spec.ts` passes step 7 on a real staging
run where `reviewedBefore >= 1` from a prior run's residue (the same
persistent-synthetic-account condition the 2026-08-09 failure occurred
under), not only on a run starting from a freshly-reset or zero-`reviewedBefore`
state, which would not actually exercise the failure condition.

## VOC-053-AC-04 — VOC-052-T01's staging E2E evidence requirement is unblocked

- Requirement source: issue #450's "Impact" section
- Tasks: `VOC-053-T02`
- Tests: `VOC-053-TEST-03`
- Evidence: `VOC-053-EV-02`
- Result: pending

After this package's fix lands and is verified on real staging, the staging
core-loop E2E check (`deploy-staging.yml`'s post-deploy gate) passes in full
on a real run, providing the evidence VOC-052-T01 needs to close, which in
turn unblocks VOC-052's completion and the pending PR #435 `develop` → `main`
release. This criterion records the unblocking effect; it does not re-open or
modify VOC-052's or PR #435's own scope.
