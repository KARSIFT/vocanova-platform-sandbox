# VOC-076 — Tasks

## VOC-076-T00 — Confirm the disabled-button root cause with direct evidence

- Requirement source: issue #575; `specification.md` findings and
  `VOC-076-DEP-00`
- Acceptance criteria: `VOC-076-AC-00`
- Tests: `VOC-076-TEST-00`
- Evidence: `VOC-076-EV-00` (`t00-evidence.md`)
- Status: pending

No product fix is written in this task. Confirm which cause explains the
disabled multiple-choice answer button timeout on run 31748423831 and the
general failure mode.

Required investigation order (stop early only if a candidate is *confirmed*
with direct evidence; otherwise continue):

1. **Product disabled-state path.** In
   `apps/web/src/app/(app)/reviews/_components/review-session.tsx`, trace when
   MC option buttons are `disabled` (`isLoading || phase === "feedback"`;
   `isLoading = isSubmitting || isRefetching`). Against staging or a faithful
   reproduction, determine whether `isSubmitting`, `isRefetching`, or `phase`
   leaves buttons disabled while the "Choose the meaning for …" group is
   visible and `aria-pressed` remains false (matching the failure log).
2. **E2E readiness gap.** In `reviewOneCard`
   (`core-loop.staging.spec.ts`), confirm the helper waits only for group
   visibility before clicking, and whether that alone can reproduce a
   disabled-click hang under staging timing.
3. **Staging API latency / errors.** Check available staging logs or network
   timing around run 31748423831 for slow or failing `listDueWords` /
   `submitReview` that could keep `isRefetching` / `isSubmitting` elevated.
   Do not invent log access as a pass if unavailable — mark inconclusive.
4. **Detach / re-render.** Explain how the button could remain disabled for
   ~240s then detach (refetch replacing `dueWords`, advance, completed state,
   error remount).

Record in `t00-evidence.md`: the confirmed cause (or honest remaining
ambiguity), exact evidence, whether product fix and/or E2E fix is required,
and why each other candidate was ruled out or left inconclusive. Do not start
T01 until this evidence names a specific cause — or adoption explicitly scopes
an E2E-only / product-only path despite residual ambiguity.

## VOC-076-T01 — Fix the confirmed disabled-button cause

- Requirement source: `specification.md` scope item 2; `VOC-076-D00`; depends
  on `VOC-076-T00`
- Acceptance criteria: `VOC-076-AC-01`, `VOC-076-AC-02`, `VOC-076-AC-04`
- Tests: `VOC-076-TEST-01`, `VOC-076-TEST-02`, `VOC-076-TEST-03`
- Evidence: `VOC-076-EV-01` (`t01-evidence.md`)
- Status: pending — blocked on `VOC-076-T00` naming a specific
  evidence-backed cause

Implement the narrowest fix for the cause T00 confirmed.

Examples keyed to drafting-time candidates (use only what T00 confirms):

- If `isSubmitting` / `isRefetching` stuck: fix the clear path in
  `review-session.tsx` (e.g. ensure `finally` / error paths clear flags; avoid
  leaving MC options disabled after refetch settles). Preserve accessibility
  attributes and the `max-w-*` workaround (`.karsift/lessons.md`).
- If E2E race only: harden `reviewOneCard` to wait for an **enabled** MC
  button (or equivalent readiness) before click; do not merely raise the
  global timeout without a real readiness signal.
- If both: fix product hang first, then add the enabled wait as defense in
  depth.
- If staging API hang: fix the confirmed API defect in scope, or document why
  a separate package is required — do not expand into unrelated API refactors.

Add regression coverage appropriate to the cause (component/unit test for a
product hang; spec assertion review for E2E readiness). Do not edit
`deploy-staging.yml` unless adoption expanded `VOC-076-DEP-01`. Do not weaken
VOC-074's `reviewedCards >= 1` check.

## VOC-076-T02 — Verify the fix on real staging

- Requirement source: issue #575; `specification.md` scope item 3;
  `VOC-076-DEP-02`
- Acceptance criteria: `VOC-076-AC-03`, `VOC-076-AC-04`
- Tests: `VOC-076-TEST-04`
- Evidence: `VOC-076-EV-02` (`t02-evidence.md`)
- Status: pending — depends on `VOC-076-T01` merging to `develop`

No further source change is expected unless verification surfaces a narrow gap.
After T01 merges, record a real `deploy-staging.yml` run of
`tests/staging-e2e/core-loop.staging.spec.ts`:

- Step 5 completes without the disabled-button 240s timeout at the
  `reviewOneCard` MC click site.
- Prefer a run that exercises at least one multiple-choice card; if only
  self-check appears, satisfy `VOC-076-AC-03`'s MC-coverage rule before
  claiming closure.
- Confirm VOC-050 gate and VOC-074 hardening remain in place.

## Task ordering notes

- T00 blocks T01.
- T02 depends on T01 merge to `develop`.
- No task may be dispatched before this package is adopted.

Tasks preserve scope, separation of duties, and rollback safety.
