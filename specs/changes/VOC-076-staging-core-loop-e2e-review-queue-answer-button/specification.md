# VOC-076 — Staging Core-Loop E2E: Review-Queue Answer Button Stays Disabled: Specification

## Objective and requirement source

Close the gap reported in
[GitHub issue #575](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/575):
the staging core-loop E2E journey fails in step 5 because the review-queue
multiple-choice answer button stays `disabled` until the 240s test timeout,
then detaches from the DOM.

Primary evidence (issue #575):

| Item | Value |
|------|--------|
| Run | [31748423831](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31748423831) |
| Job / step | `deploy to staging` → `Run the staging core-loop journey` |
| Spec | `tests/staging-e2e/core-loop.staging.spec.ts:272:3` › authenticated journey › step "5. work the review queue" |
| Failure | `locator.click: Test timeout of 240000ms exceeded` waiting for `getByRole('group', { name: /^Choose the meaning for / }).getByRole('button').first()` |
| Observed DOM | `<button disabled … aria-pressed="false">` — not enabled; later detached |
| Call site | `reviewOneCard` → `core-loop.staging.spec.ts:258` |

**Objective:** after this package's implementation, (a) the staging core-loop
journey can complete step 5's multiple-choice answer interaction without the
disabled-button 240s timeout on a real staging deploy, and (b) if the cause is
product-side, real learners can interact with multiple-choice answer options
when a card is ready for input (buttons not stuck disabled while the prompt is
visible).

## Confirmed findings (from issue #575 and drafting-time code read)

- The failure is **not** the vacuous `reviewedCards = 0` case VOC-074 addressed:
  the MC group and disabled button *did* render, so a card was present; the
  click never succeeded.
- Playwright treated the control as visible but **not enabled** for the entire
  timeout, then saw detachment (consistent with a re-render, advance, or refetch
  replacing the card).
- In `review-session.tsx` (drafting-time read):
  - `isLoading = isSubmitting || isRefetching`
  - MC option buttons: `disabled={isLoading || phase === "feedback"}`
  - `aria-pressed="false"` on the failing button suggests no selection had been
    recorded yet — more consistent with `isLoading === true` (or an equivalent
    stuck disabled path) than with a completed feedback-phase selection.
  - `advance()` sets `isRefetching` while calling `listDueWords`; `submitAttempt`
    sets `isSubmitting` around `submitReview`. Either flag stuck true would keep
    MC buttons disabled.
- `reviewOneCard` currently waits for the MC group (or Show answer / caught-up)
  to be **visible**, then immediately clicks the first MC button **without** an
  explicit wait for `enabled`. That is a plausible E2E race if the group appears
  while buttons are still disabled.

Drafting-time investigation starting points (from issue #575, non-prescriptive):

1. Reproduce against staging (or a faithful local path) and observe whether
   `isSubmitting` / `isRefetching` / `phase` leave MC buttons disabled while the
   legend/group is visible.
2. Inspect staging API latency / errors for `listDueWords` and `submitReview`
   around run 31748423831's timeframe (server logs if available; no production
   secrets).
3. Determine whether hardening the E2E wait (`toBeEnabled` / wait for
   non-loading) alone is sufficient, or whether a product hang must be fixed.
4. Confirm the card-detach behavior (refetch replacing `dueWords`, index reset,
   completed state) against the timeout window.

## Scope and non-goals

In scope:

- `VOC-076-T00`: Confirm, with direct evidence, which cause explains the
  disabled-button timeout on run 31748423831 (and the general failure mode).
  Record evidence in `t00-evidence.md`.
- `VOC-076-T01`: Fix the confirmed cause narrowly in
  `review-session.tsx` and/or `core-loop.staging.spec.ts` (and only
  `deploy-staging.yml` if `VOC-076-DEP-01` expands). Add regression coverage
  appropriate to the cause.
- `VOC-076-T02`: Verify on a real `deploy-staging.yml` run that step 5
  completes (including MC answer click when MC cards appear) without the 240s
  disabled-button timeout.

Non-goals / explicitly excluded:

- Not fixing or reopening VOC-074's `reviews_completed` increment path unless
  T00 proves this timeout is the same defect (not expected).
- Not weakening VOC-050's staging gate or VOC-074's `reviewedCards >= 1` fail.
- Not assumed API/schema migrations or production credential changes.
- Not assumed `deploy-staging.yml` edits (`VOC-076-DEP-01`).
- Not broad review-queue UX redesign beyond restoring interactive MC answers /
  reliable staging waits.

## Risk and protected areas

Builder assessment: expected code touch is
`apps/web/src/app/(app)/reviews/_components/review-session.tsx` and/or
`apps/web/tests/staging-e2e/core-loop.staging.spec.ts`. Path classifier floor
for that set measured at drafting time: **R1**.

Including `.github/workflows/deploy-staging.yml` raises the path floor to
**R3**.

This package proposes **R2** because the staging core-loop gate is failing on
`develop` deploys and a stuck-disabled answer button may affect real learners
if product-side. The independent verifier must re-run
`classify-change-risk.sh` against the real task file list and may raise or lower
per T00's confirmed cause and whether the workflow file is touched.

No governance, secret-handling, or migration area is in default scope. EHR is
not triggered. Under active A-003, routine R3 (if the path floor rises) does not
require standing technical-steward or founder approval merely for being R3;
strengthened verification still applies.

## Decisions, contradictions, security, and privacy

`VOC-076-D00` (recorded here for traceability; formal decision numbering
applies after adoption): The staging core-loop journey must be able to answer a
visible multiple-choice review card without waiting the full test timeout on a
disabled control. A green local/mock suite that never exercises real staging
timing does not close issue #575 — T02 requires a real `deploy-staging` run.

No contradiction with VOC-074: that package addresses vacuous step-7 passes and
mission counter increments; this package addresses a distinct step-5
disabled-button timeout that can block the journey before those assertions run.

Open questions for the reviewing human:

1. **`VOC-076-DEP-00` — Root cause priority.** Adoption may proceed with T00
   still required; note whether product loading-state tracing or E2E
   enabled-wait hardening is the starting priority.
2. **`VOC-076-DEP-01` — `deploy-staging.yml` scope.** Default exclude; expand
   only if T00 needs workflow diagnostics / seeding (R3 path floor).
3. **`VOC-076-DEP-02` — VOC-074 coordination.** Confirm merge-order /
   evidence attribution if VOC-074 tasks remain open on `develop`.
4. **Risk class.** Accept proposed R2, or adjust once T00 clarifies
   product-vs-test severity and workflow scope.

No new secret, credential, or personal-data handling is introduced. Staging
verification continues to use only the existing synthetic smoke-test account.

## Data, migrations, analytics, and accessibility

- **Data / migrations:** None expected.
- **Analytics:** None expected.
- **Accessibility:** Preserve existing review-session patterns (fieldset /
  legend "Choose the meaning for …", `aria-pressed`, focus-visible). If T01
  changes disabled-state timing, ensure buttons are not left permanently
  non-interactive for assistive tech when the prompt is ready. Keep the
  Tailwind `max-w-*` workaround in `review-session.tsx` (see
  `.karsift/lessons.md`).
