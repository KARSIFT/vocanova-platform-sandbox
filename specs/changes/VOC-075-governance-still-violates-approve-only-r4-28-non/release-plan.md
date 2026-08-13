# VOC-075 — Release Plan

## Release and deployment authorization

This package requests no new deployment authority. Once adopted and
implemented, each task's pull request follows the existing governed path:
independent review, then merge into `develop` per `merge-gate.yml`, then —
per AGENTS.md's "Release and deployment authority" section — automatic
promotion to `main` and automatic production deployment once this package's
task roster closes. This package does not alter any of that mechanism.

No production user-facing effect is expected: changes are confined to
governance drafting text, the change-package template, optional governance
lint, and `automatic_merge_allowed` metadata on named packages.

**Note on this package's own `automatic_merge_allowed`:** drafted as `true`
because the package proposes R3 under the founder rule it encodes. If
adoption raises the package to R4 (DOC-15 and/or lint paths), set this
package's field to `false` at adoption so the record stays self-describing
and consistent with the new rule.

## Preconditions, monitoring, and outcome

Preconditions:

- Package adopted with `VOC-075-DEP-00`, `VOC-075-DEP-01`, and
  `VOC-075-DEP-02` settled in writing.
- Declared risk raised to R4 when DOC-15 and/or `scripts|tooling/governance`
  lint is included.
- Implementation authorized.
- Independent verification PASS (or PASS WITH NON-BLOCKING FINDINGS) on each
  exact implemented revision.
- Founder approval recorded when the effective class is R4.

Monitoring after merge:

- VOC-072 (and any other backfilled) task PRs no longer produce merge-gate
  text requiring founder approval solely because
  `automatic_merge_allowed: false`.
- New plan PRs draft R0–R3 with `true` and R4 with `false`.
- If T04 landed: the lint fails closed on attempted non-R4 `false`.
- R4 hard-block behavior unchanged.

Outcome owner: implementer records task evidence; adopting human owns the
DEP decisions recorded at adoption.

## Rollback

Trigger: guidance weakens R4/EHR/verification; docs contradict each other;
lint mis-fires; or independent review fails closed on the exact revision.

Mechanism: revert the implementation commit(s) for the affected task(s)
(AGENTS.md / template / docs / backfill / lint).

Validation: governance scripts pass on the reverted tree; targeted files
match pre-change text where intended.

Accountable owner: implementer of the reverted task(s).

Last-known-good reference: `AGENTS.md`,
`specs/templates/change-package/`, affected `docs/**`, and any flipped
`specs/changes/VOC-*/change.yaml` revisions immediately preceding this
package's implementation merge(s).

## Independent verification, human approvals, and closure

Independent verification must confirm, against each exact implemented
revision's commit SHA:

- Adoption decisions for open questions were followed, not silently
  re-litigated by the implementer.
- Applicable `VOC-075-AC-00` through `VOC-075-AC-05` hold with linked
  evidence.
- Implementer-role occupant did not approve or merge its own
  implementation.
- Active authority model remains `a003-active`.
- If DOC-15 or `scripts|tooling/governance` lint was edited: R4 founder
  approval is recorded for that revision.
- If DOC-15 was not edited: option-b reconciliation evidence is present.
- No still-required EHR trigger applies.

Under active A-003, no standing technical-steward approval is assumed solely
because baseline work is proposed R3. If the path floor rises to R4, founder
approval is mandatory for that consequence — consistent with the same
approve-only-R4 rule this package enforces for everyone else.

Repository merge into `develop` and production release/deployment are not
the same event as closure — closure requires this package's acceptance
criteria recorded as passing with linked evidence.
