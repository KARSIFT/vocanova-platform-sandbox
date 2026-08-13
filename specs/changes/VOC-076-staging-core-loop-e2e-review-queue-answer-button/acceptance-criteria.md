# VOC-076 — Acceptance Criteria

## VOC-076-AC-00 — Disabled-button root cause is confirmed with direct evidence

- Requirement source: issue #575; `specification.md` / `VOC-076-DEP-00`
- Tasks: `VOC-076-T00`
- Tests: `VOC-076-TEST-00`
- Evidence: `VOC-076-EV-00`
- Result: pending

`t00-evidence.md` names a specific root cause for the disabled multiple-choice
answer button timeout on run 31748423831 (and the general failure mode), with
exact file/line and/or live staging evidence. Each drafting-time candidate
(stuck `isLoading` / `isRefetching` / `isSubmitting`, E2E missing enabled wait,
staging API latency, phase/detach behavior) is either confirmed, ruled out, or
honestly marked inconclusive with what was tried. No product fix is required in
T00 itself.

## VOC-076-AC-01 — Multiple-choice answer controls become interactable when the prompt is ready

- Requirement source: `VOC-076-D00`; issue #575
- Tasks: `VOC-076-T01`
- Tests: `VOC-076-TEST-01`, `VOC-076-TEST-02`
- Evidence: `VOC-076-EV-01`
- Result: pending

After T01, when a multiple-choice review card is shown as ready for learner
input (prompt phase, not mid-submit / mid-refetch), the meaning-option buttons
are enabled and clickable. A product-side hang that leaves options permanently
disabled while the "Choose the meaning for …" group remains visible fails this
criterion. If T00 scoped the defect as E2E-only, this criterion is satisfied by
documenting that product behavior was already correct and T01's E2E change is
the sole fix — recorded explicitly in `t01-evidence.md`.

## VOC-076-AC-02 — Staging E2E does not click a disabled MC answer button

- Requirement source: issue #575; `VOC-076-D00`
- Tasks: `VOC-076-T01`
- Tests: `VOC-076-TEST-03`
- Evidence: `VOC-076-EV-01`
- Result: pending

`reviewOneCard` in `core-loop.staging.spec.ts` (or equivalent helper) waits for
the target multiple-choice answer control to be **enabled** (or an equivalent
explicit readiness signal that cannot pass while `disabled`) before clicking.
A click attempt against a still-disabled button that burns the full test timeout
fails this criterion. VOC-074's `reviewedCards >= 1` hardening and VOC-050's
gate remain intact.

## VOC-076-AC-03 — Real staging verification proves step 5 completes past the MC click

- Requirement source: issue #575; `specification.md` scope item 3
- Tasks: `VOC-076-T02`
- Tests: `VOC-076-TEST-04`
- Evidence: `VOC-076-EV-02`
- Result: pending

A real `deploy-staging.yml` run after T01 merges executes
`tests/staging-e2e/core-loop.staging.spec.ts` and completes step 5 without the
disabled-button 240s timeout at the `reviewOneCard` MC click site. Evidence
records the run URL, whether MC cards were exercised, and that the prior
failure mode did not recur. If a given run only hits self-check cards, evidence
must either (a) show a subsequent run that exercised MC, or (b) include an
explicit reproduction that forces an MC card under the fixed revision — T00/T01
must not claim closure on MC-disable without MC coverage.

## VOC-076-AC-04 — Package boundaries respected

- Requirement source: issue #575; `VOC-076-DEP-01`; `VOC-076-DEP-02`
- Tasks: `VOC-076-T00`, `VOC-076-T01`, `VOC-076-T02`
- Tests: `VOC-076-TEST-04`
- Evidence: `VOC-076-EV-00`, `VOC-076-EV-01`, `VOC-076-EV-02`
- Result: pending

This package does not weaken VOC-050's staging gate or VOC-074's vacuous-pass
hardening, does not silently expand into `deploy-staging.yml` without
`VOC-076-DEP-01` / adoption, and does not treat VOC-074's increment fix as
closure for issue #575.
