# VOC-078 — VOC-076-T02: No Real Staging Proof the Review-Queue Button Fix Works

**Status: draft, not adopted.** Nothing in this package is implementation-authorized.
It is a draft response to
[issue #608](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/608),
prepared for founder/steward review at adoption time.

## Identity and lifecycle

- Package ID: VOC-078
- Title: VOC-076-T02: No Real Staging Proof the Review-Queue Button Fix Works
- Canonical path:
  `specs/changes/VOC-078-voc-076-t02-no-real-staging-proof-the-review`
- Lifecycle state: `draft` (not adopted, not authorized for implementation)
- Proposed risk: `R2` (draft proposal only — see `change.yaml`'s
  `planned_implementation_risk_floor`; path floor measured at drafting time
  is `R1` for evidence and for optional review UI / staging-E2E remediation;
  `R3` if `deploy-staging.yml` enters scope)
- Owner: unassigned (see `change.yaml`'s `owners` block)
- Approval evidence: none yet — `approval_status: not-approved`,
  `implementation_authorized: false`
- Target branch: `develop`
- Linked GitHub issues:
  - [#608](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/608)
    (this package's requirement source)
  - [#575](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/575)
    (original staging disabled-button bug — still open until real staging
    PASS)
- Related packages, PRs, and runs:
  - [VOC-076](specs/changes/VOC-076-staging-core-loop-e2e-review-queue-answer-button)
    — T00/T01 passed; T02 merged without green staging proof
  - [PR #598](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/598)
    — VOC-076-T02; merged to `develop` despite `VERDICT: FAIL`
  - Staging run
    [#227](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31791701520)
    — T01 tip still failed at `toBeEnabled` on disabled MC button (pre-gap-fix
    proof target cited by the FAIL review)

## Why this exists

VOC-076's product/E2E fix path (T00 investigation, T01 fix) landed with
passing independent verification. T02's job was to prove step 5 of
`tests/staging-e2e/core-loop.staging.spec.ts` completes past the
multiple-choice click on a **real** `deploy-staging.yml` run. That proof
never arrived: PR #598 merged with an independent-verification
`VERDICT: FAIL` whose High finding was exactly the missing staging PASS.
Existing `t02-evidence.md` correctly records AC-03 as unmet and cites run
#227 as still failing; it also documents a narrow gap fix (hide leftover
feedback MC during `isRefetching`; prompt-ready E2E waits) that is now on
`develop` via the #598 merge — but a green post-fix staging run is still
absent from the package record. Issue #575 therefore correctly stays open.

## What this package does

1. **Obtain real staging proof** (`VOC-078-T00`): against current
   `develop`/staging (tip that includes VOC-076 T01 + the #598 gap fix),
   confirm step 5 completes past the MC interaction without the prior
   disabled-button failure mode, and record honest PASS evidence (or honest
   FAIL with run URL — never invent a green run).
2. **Remediate only if still broken** (`VOC-078-T01`): if T00 records FAIL,
   apply the narrowest remaining product/E2E fix and obtain a subsequent
   green staging run. If T00 already PASSed, T01 is N/A with citation — no
   code change.
3. **Close the loop on #575 / VOC-076-AC-03** only when real PASS evidence
   exists (update VOC-076 `t02-evidence.md` and AC-03 Result; keep this
   package's own evidence files as the VOC-078 audit trail).

## What this package deliberately does NOT do

- Not re-investigating VOC-076-T00's root cause from scratch unless T00's
  new staging failure contradicts the prior evidence.
- Not re-litigating VOC-074's `reviews_completed` / vacuous-pass hardening.
- Not assumed edits to `deploy-staging.yml` (`VOC-078` inherits VOC-076's
  default exclude; expanding raises path floor to R3).
- Not force-closing #575 without a green staging run URL.
- Not diagnosing or fixing why merge-gate allowed a FAIL merge unless
  adoption expands `VOC-078-DEP-01` (default: separate unlabeled issue).
- Does not adopt itself. Every adoption/authorization field stays at the
  unadopted default.

## Open questions for the reviewing human

See `specification.md`. The most important at adoption:

1. Disposition of VOC-076-T02 / PR #598 once VOC-078 supplies PASS evidence
   (`VOC-078-DEP-00`).
2. Whether FAIL-merge process hardening is in-scope follow-up
   (`VOC-078-DEP-01`).
3. Accept proposed `R2` semantic elevation above the measured `R1` path
   floor (or lower to `R1` if adoption scopes evidence-only).

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`. This
package carries no standing approval; adoption, implementation authorization,
independent verification, and any required human approval remain to be
recorded against the exact implemented revision, per AGENTS.md and CLAUDE.md.
