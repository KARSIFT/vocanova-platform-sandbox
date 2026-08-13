# VOC-076 — Release Plan

## Release and deployment authorization

This package requests no new deployment authority beyond what AGENTS.md already
describes for adopted packages. Once adopted and implemented, each task PR
follows the existing governed path: independent review, automatic merge into
`develop` (per `karsift-ai-infra`'s `merge-gate.yml` and this package's
`automatic_merge_allowed: true` draft unless adoption changes it), then — when
this package's task roster closes — automatic promotion to `main` and automatic
production deployment per the 2026-08-08 founder delegation documented in
AGENTS.md.

A merged package does **not** itself authorize production deployment outside
that existing mechanism. This package does not alter release/deploy workflows
by default (`VOC-076-DEP-01`).

If T01 confirms a product-side stuck-disabled state, the production effect is
user-visible and desirable: multiple-choice review options become reliably
interactive when the prompt is ready. If T01 is E2E-only, production UI is
unchanged and the release value is restoring the staging core-loop gate signal.

## Preconditions, monitoring, and outcome

Preconditions:

- Package adopted with explicit stance on `VOC-076-DEP-01` (default: exclude
  `deploy-staging.yml`) and `VOC-076-DEP-02` (VOC-074 coordination).
- `VOC-076-T00` evidence names the cause before T01 proceeds (or adoption
  explicitly scopes the fix path).
- `VOC-076-T01` merges before T02 records staging verification.

Monitoring after T01:

- `deploy-staging.yml` core-loop E2E: step 5 must complete; watch for recurrence
  of disabled-button timeouts at `reviewOneCard`'s MC click.
- If product was fixed: watch staging/production review-submission error rates
  and any reports of double-submit (intentional disabled-during-submit must
  still work).

Outcome owner: implementer of `VOC-076-T02` records the passing staging run in
`VOC-076-EV-02`.

This package's completion closes issue #575's disabled-button timeout symptom.
It does not claim to resolve VOC-074's `reviews_completed` increment work.

## Rollback

Trigger: review submissions fail after T01; MC options incorrectly stay enabled
during submit/feedback causing double-submit; E2E gate fails for unrelated
reasons; or independent review finds the fix incorrect.

Mechanism: revert `VOC-076-T01` commit(s).

Validation: confirm a subsequent staging deploy restores prior behavior. The
disabled-button timeout may return — that is the pre-fix state, not a failed
rollback.

Accountable owner: implementer of the affected task.

Last-known-good reference: tree immediately preceding the reverted task's merge.

## Independent verification, human approvals, and closure

Independent verification must confirm, against the exact implemented revision's
commit SHA:

- Root cause in T00 matches work in T01.
- T01 meets `VOC-076-AC-01` / `VOC-076-AC-02` as applicable.
- T02 staging evidence meets `VOC-076-AC-03` including MC coverage rule — not
  only Playwright green on a self-check-only path if MC was the failure mode.
- Package boundaries respected (`VOC-076-AC-04`).
- Path floor re-measured; proposed R2 still justified or adjusted with evidence.
- Active authority model is A-003; no standing steward approval assumed solely
  for risk class; R4/EHR not triggered unless verifier escalates.
- Implementer did not approve or merge their own work.

Under active A-003, replace routine R3 steward approval with strengthened
applicable technical evidence; preserve R4 founder authority and triggered EHR
evidence.

Repository merge into `develop` and production release/deployment are not the
same event as closure — closure requires this package's acceptance criteria
recorded as passing with linked evidence.
