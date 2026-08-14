# VOC-078 — Tasks

## VOC-078-T00 — Obtain real staging proof for step 5 / MC click

- Requirement source: issue #608; issue #575; `VOC-078-D00`
- Acceptance criteria: `VOC-078-AC-00`, `VOC-078-AC-01`, `VOC-078-AC-02`,
  `VOC-078-AC-03`
- Tests: `VOC-078-TEST-00`, `VOC-078-TEST-01`, `VOC-078-TEST-03`
- Evidence: `VOC-078-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

Fresh task / new PR — **do not** redispatch VOC-076-T02 under an exhausted
attempt budget. **Do not** invent a green staging run.

### Required work

1. **Confirm tip under test.** Record the `develop` SHA (or the staging
   deploy SHA) that includes VOC-076 T01 and the PR #598 gap fix
   (`shouldShowReviewCardPrompt` / prompt-ready E2E waits). If the tip has
   moved further, confirm those fixes are still present.

2. **Locate or obtain a real `deploy-staging.yml` run** of
   `tests/staging-e2e/core-loop.staging.spec.ts` against that tip.
   Prefer an existing post-#598 Actions run if one exists (`VOC-078-DEP-02`).
   If none exists yet, wait for / rely on the normal post-merge
   `deploy-staging` path — this role does not invent workflow_dispatch
   authority or production secrets. Do not treat a missing run as PASS.

3. **Record honest outcome in `t00-evidence.md`:**
   - Run URL, head SHA, conclusion, step 5 result.
   - Whether MC cards were exercised (VOC-076-AC-03 coverage rule).
   - Explicit comparison to run #227 / issue #575 failure mode.
   - On **PASS**: update VOC-076 `t02-evidence.md` and `VOC-076-AC-03`
     Result to satisfied with the green URL; close issue #575 only then.
   - On **FAIL**: document the failure with file/line / call-log excerpts;
     leave VOC-076-AC-03 unmet; leave #575 open; do **not** claim PASS.
     Name whether T01 product, E2E, or both is required (or mark
     inconclusive with what was tried).

4. **Independent review** must PASS on the exact tip. A PR that claims
   staging PASS without a green run URL must FAIL review (PR #598 class).

### Explicitly out of scope for this task

- Product/E2E code changes (those belong to T01 only after FAIL).
- `deploy-staging.yml` edits.
- Merge-gate / FAIL-merge process hardening (`VOC-078-DEP-01`).
- Weakening VOC-074 or VOC-050 gates.

## VOC-078-T01 — Remediate remaining staging failure and re-prove

- Requirement source: issue #608; depends on `VOC-078-T00`
- Acceptance criteria: `VOC-078-AC-00`, `VOC-078-AC-01`, `VOC-078-AC-02`,
  `VOC-078-AC-03`
- Tests: `VOC-078-TEST-02`, `VOC-078-TEST-03`
- Evidence: `VOC-078-EV-01` (`t01-evidence.md`)
- Status: pending — **required only if T00 recorded FAIL**; if T00 recorded
  PASS with AC-00 met, mark this task N/A in `t01-evidence.md` citing
  `VOC-078-EV-00` and make **no** source change

### Required work (FAIL path only)

1. Using T00's failing run evidence, apply the **narrowest** remaining fix
   in `review-session.tsx` / `review-session-prompt.ts` and/or
   `core-loop.staging.spec.ts` (and readiness unit tests as needed).
   Preserve accessibility attributes and the `max-w-[28rem]` workaround
   (`.karsift/lessons.md`). Do not weaken VOC-074's `reviewedCards >= 1`
   check.

2. Run applicable local web validation per `docs/development.md` (at least
   the review-session readiness tests, typecheck, and staging-e2e
   typecheck/lint as touched).

3. After the fix merges to `develop`, record a **new** green
   `deploy-staging` run URL in `t01-evidence.md` and update VOC-076
   `t02-evidence.md` / AC-03 Result. Close #575 only when that PASS is
   real.

4. Independent review must PASS on the exact tip; missing green URL ⇒ FAIL.

### Explicitly out of scope for this task

- Re-deriving VOC-076-T00 from scratch unless the new failure contradicts
  prior cause evidence.
- `deploy-staging.yml` edits unless adoption expanded scope.
- Process investigation of the #598 FAIL merge (`VOC-078-DEP-01`).

## Task ordering notes

- T00 runs first.
- T01 is blocked on T00 FAIL (or is N/A on T00 PASS).
- No task may be dispatched before this package is adopted.
- Closing #575 is gated on AC-00 PASS evidence, not on task issue closure
  alone.

Tasks preserve scope, separation of duties, and rollback safety.
