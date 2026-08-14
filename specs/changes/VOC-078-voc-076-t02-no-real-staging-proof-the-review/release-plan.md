# VOC-078 — Release Plan

## Release and deployment authorization

This package requests no special deployment authority beyond the repository's
existing governed path. Task PRs follow independent review, then merge into
`develop` per `merge-gate.yml`. Per AGENTS.md's "Release and deployment
authority" section, automatic promotion to `main` and automatic production
deployment may follow once this package's task roster closes — that is the
standing 2026-08-08 delegation, not a new grant from this draft.

Evidence-only completion has no user-facing runtime effect.
T01 product remediation, if required, changes review-queue UI behavior for
real users once promoted — still only through that governed path.

This package does **not** authorize inventing staging credentials, editing
`deploy-staging.yml` by default, or closing #575 without a green run.

## Preconditions, monitoring, and outcome

Preconditions:

- Package adopted with `VOC-078-DEP-00` / `VOC-078-DEP-01` settled in writing
  (or explicitly deferred as follow-ups).
- Implementation authorized.
- Independent verification PASS (or PASS WITH NON-BLOCKING FINDINGS) on the
  exact implemented revision — founder `approved` alone does not override
  FAIL (PR #598 / #558 class lessons).
- Real `deploy-staging` PASS recorded before claiming VOC-076-AC-03 /
  closing #575.

Monitoring after merge:

- Issue #575 closed iff AC-00 PASS.
- VOC-076 `t02-evidence.md` / AC-03 Result match the green run.
- Subsequent `develop` deploys continue to pass staging core-loop step 5
  (watch for regression to run #227 signature).
- No secret values in merged commits.

Outcome owner: implementer records `VOC-078-EV-00` / `VOC-078-EV-01`;
adopting human owns DEP decisions.

## Rollback

Trigger: false PASS evidence; secret leakage; unauthorized workflow edits;
or harmful T01 remediation.

Mechanism: revert the implementation commit(s) for the affected task.

Validation: governance scripts (and applicable web tests if product was
touched) pass on the reverted tree.

Accountable owner: implementer of the reverted task.

Last-known-good reference: repository tip immediately preceding this
package's implementation merge (AC-03 unmet / #575 open).

## Independent verification, human approvals, and closure

Independent verification must confirm, against the exact implemented
revision's commit SHA:

- Adoption decisions for open questions were followed.
- `VOC-078-AC-00` through `VOC-078-AC-03` hold with linked evidence.
- A staging PASS claim without a green run URL is absent (or is a blocking
  finding).
- Implementer-role occupant did not approve or merge its own
  implementation.
- Active authority model remains `a003-active`.
- No still-required R4 / EHR gate applies to the default scope; if
  `deploy-staging.yml` entered scope, report the raised floor and any
  remaining approvals.
- Codex (or current implementer-role occupant) did not self-approve.

Under active A-003, no standing technical-steward approval is assumed for
proposed R2. Repository merge into `develop` is not the same event as
closing issue #575 — #575 closes only on genuine staging PASS evidence.

Closure of **this** package requires VOC-078 acceptance criteria recorded
as passing with linked evidence (`VOC-078-EV-00` and, if used,
`VOC-078-EV-01`).
