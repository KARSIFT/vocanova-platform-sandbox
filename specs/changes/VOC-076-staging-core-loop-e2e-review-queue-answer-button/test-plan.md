# VOC-076 — Test Plan

## VOC-076-TEST-00 — Root-cause evidence is complete and actionable

- Covers: `VOC-076-AC-00`
- Preconditions: Issue #575 thread; run 31748423831 logs; repository sources
  readable; staging access as existing workflows use (if available).
- Procedure: Review `t00-evidence.md`. Confirm it:
  - Names one confirmed cause (or honestly states remaining ambiguity — which
    fails this criterion unless adoption explicitly scoped the fix path).
  - Includes exact file/line and/or live staging evidence.
  - Addresses stuck `isLoading`/`isRefetching`/`isSubmitting`, E2E missing
    enabled wait, staging API latency, and detach/re-render (confirmed, ruled
    out, or inconclusive).
  - States whether T01 must change product, E2E, or both.
- Expected result: A single, evidence-backed cause (or explicit adoption-scoped
  path) suitable to drive T01.
- Evidence: `VOC-076-EV-00`

## VOC-076-TEST-01 — Product interactivity when the prompt is ready (if product in scope)

- Covers: `VOC-076-AC-01`
- Preconditions: T01 branch builds; applicable web test runner available per
  `docs/development.md`.
- Procedure: If T00 confirmed a product hang, run the new or updated unit /
  component test that fails under the pre-fix stuck-disabled behavior and
  passes on the fix. Manually or via test, confirm MC options are enabled in
  prompt phase when not submitting/refetching.
- Expected result: Regression coverage green; options enabled when ready.
  If T00 scoped E2E-only, record N/A with citation to T00 and skip to TEST-03.
- Evidence: `VOC-076-EV-01`

## VOC-076-TEST-02 — Intentional disabled states still work (negative / characterization)

- Covers: `VOC-076-AC-01` (negative)
- Preconditions: Same as TEST-01 when product is in scope.
- Procedure: Confirm buttons remain disabled during intentional
  `isSubmitting` / `isRefetching` and during `phase === "feedback"` (after a
  selection), so the fix does not allow double-submit or re-selection during
  feedback unless product intentionally changed that — default is preserve.
- Expected result: Intentional disabled states preserved.
- Evidence: `VOC-076-EV-01`

## VOC-076-TEST-03 — Staging E2E waits for enabled MC answer before click

- Covers: `VOC-076-AC-02`
- Preconditions: T01 branch; ability to review/lint the spec file.
- Procedure: Review the T01 diff to `core-loop.staging.spec.ts` (if E2E was in
  scope). Confirm `reviewOneCard` (or equivalent) waits for an enabled MC
  button / readiness signal before click, and does not rely solely on raising
  `test.setTimeout`. Confirm VOC-074 `reviewedCards >= 1` and VOC-050 journey
  structure remain.
- Expected result: Disabled-click hang cannot consume the full timeout without
  a clear readiness failure; VOC-074/VOC-050 boundaries intact.
- Evidence: `VOC-076-EV-01`

## VOC-076-TEST-04 — Real staging core-loop E2E completes step 5 past MC click

- Covers: `VOC-076-AC-03`, `VOC-076-AC-04`
- Preconditions: T01 merged to `develop`; a real `deploy-staging.yml` run
  executes against staging with that revision.
- Procedure: Observe the workflow log for
  `tests/staging-e2e/core-loop.staging.spec.ts`. Record run URL, step 5
  outcome, whether MC cards were exercised, and absence of the prior
  disabled-button 240s timeout. Confirm package boundaries (no unauthorized
  workflow expansion; VOC-074 not cited as closure for #575).
- Expected result: Step 5 completes; MC coverage rule in AC-03 satisfied;
  boundaries respected.
- Evidence: `VOC-076-EV-02`

## Rollback coverage

Rolling back means reverting T01 commit(s). Validation: applicable `apps/web`
tests still pass on the reverted tree; a subsequent staging deploy runs. The
disabled-button timeout may return — that is the known pre-fix state.

## Constraints

No test in this plan uses secrets or production user data.
`VOC-076-TEST-04` uses only the existing synthetic smoke-test account already
provisioned for staging E2E by VOC-050.
