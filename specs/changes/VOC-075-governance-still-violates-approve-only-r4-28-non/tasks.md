# VOC-075 — Tasks

## VOC-075-T00 — Rewrite AGENTS.md drafting rule to approve-only-R4

- Requirement source: issue #573; `VOC-075-D00`
- Acceptance criteria: `VOC-075-AC-00`, `VOC-075-AC-05`
- Tests: `VOC-075-TEST-00`, `VOC-075-TEST-05`
- Evidence: `VOC-075-EV-00`
- Status: pending

Edit `AGENTS.md` subsection "Drafting `automatic_merge_allowed` in
`change.yaml`" (and any adjacent sentences that restate the R3 carve-out).

Required meaning (wording may vary):

1. R0–R3 → `automatic_merge_allowed: true` always.
2. R4 → `automatic_merge_allowed: false` always (self-describing; redundant
   with merge-gate's R4 hard block).
3. Delete the R3 case-by-case / "auth, secrets, or production infrastructure"
   founder-eyes carve-out.
4. Delete R0–R2 "unless the package records a specific reason to require
   founder eyes" opt-out encouragement.
5. Keep semantics reminders: field is not a substitute for risk
   classification, path floors, CI, independent verification, R4 founder
   authority, or EHR.
6. Update the VOC-068-DEP-00 / DOC-15 note if it still describes the old
   drafting defaults; point implementers/adopters at `VOC-075-DEP-00` for
   whether DOC-15 itself changes in T02.

Do not modify workflows, autonomy switches, or application code.
Do not mark any package adopted.

May land in the same PR as `VOC-075-T01` when both remain docs/template-only.

## VOC-075-T01 — Align change-package template with approve-only-R4

- Requirement source: issue #573; `VOC-075-D00`
- Acceptance criteria: `VOC-075-AC-01`, `VOC-075-AC-05`
- Tests: `VOC-075-TEST-01`, `VOC-075-TEST-05`
- Evidence: `VOC-075-EV-01`
- Status: pending — may land with or immediately after `VOC-075-T00`; must
  not contradict the AGENTS.md rule T00 lands

Edit:

- `specs/templates/change-package/change.yaml` (comment block above
  `automatic_merge_allowed`)
- `specs/templates/change-package/README.md` (`automatic_merge_allowed`
  drafting blurb)

Literal default may remain `true`. Comments must say R0–R3 → `true`, R4 →
`false`, with no justified non-R4 `false` path. README must not say the
literal matches only "routine R0–R2" or that "deliberate opt-outs" are a
normal non-R4 drafting choice.

Do not change existing `specs/changes/VOC-*/change.yaml` here — that is
`VOC-075-T03`.

## VOC-075-T02 — Reconcile governance docs (and DOC-15 per DEP-00)

- Requirement source: issue #573; AGENTS.md doc-reconciliation rule;
  `VOC-075-DEP-00`
- Acceptance criteria: `VOC-075-AC-02`, `VOC-075-AC-05`
- Tests: `VOC-075-TEST-02`, `VOC-075-TEST-05`
- Evidence: `VOC-075-EV-02`
- Status: pending — after or with T00 so AGENTS.md is the reference wording

Edit or evidence-only check:

- `docs/governance/change-risk-classification.md`
- `docs/governance/approval-matrix.md`
- Conditionally
  `docs/operations/15-ai-native-product-and-engineering-operating-model.md`
  **if and only if** adoption chose `VOC-075-DEP-00` option a (raises path
  floor to R4).

If adoption chose option b: do not edit DOC-15; record cited paragraphs and
the residual-risk note in evidence.

Do not weaken R4 founder authority language anywhere.

## VOC-075-T03 — Backfill VOC-072 (and adoption-scoped packages)

- Requirement source: issue #573; `VOC-075-DEP-01`
- Acceptance criteria: `VOC-075-AC-03`, `VOC-075-AC-05`
- Tests: `VOC-075-TEST-03`, `VOC-075-TEST-05`
- Evidence: `VOC-075-EV-03`
- Status: pending — may proceed after T00 lands or in parallel if the field
  flip does not depend on doc text; prefer after T00 so PR description can
  cite the new rule

Required:

- Set `automatic_merge_allowed: true` on
  `specs/changes/VOC-072-same-request-as-github-issue-535-voc-067-t05/change.yaml`.
- Add a short comment citing VOC-075 / issue #573.

If adoption chose `VOC-075-DEP-01` option b or c: flip every additional
in-scope non-R4 package the same way; list them in evidence. Do not change
`risk`, adoption, or authorization fields on those packages.

Do not flip R4 packages to `true`.

## VOC-075-T04 — Add non-R4 `automatic_merge_allowed: false` lint (conditional)

- Requirement source: issue #573; `VOC-075-DEP-02`
- Acceptance criteria: `VOC-075-AC-04`, `VOC-075-AC-05`
- Tests: `VOC-075-TEST-04`, `VOC-075-TEST-05`
- Evidence: `VOC-075-EV-04`
- Status: pending — **skip entirely if adoption chose DEP-02 option b**

If in scope:

1. Implement a deterministic check under `scripts/governance/` and/or
   `tooling/governance/` (exact file settled by implementer; both paths are
   R4 floors).
2. Wire it into the existing governance validation entrypoint used by CI /
   `validate-governance.sh` (or document the exact CI invocation if a
   separate entrypoint is required).
3. Enforce: `automatic_merge_allowed: false` requires `risk: R4` under the
   scan/exemption policy recorded at adoption (full tree vs
   post-VOC-075-only vs PR-diff-only).
4. Add positive/negative test coverage in the existing governance test suite
   if one is extended, or an equivalent deterministic fixture check.
5. Re-declare package risk as R4 if not already raised by DOC-15.

Do not change merge-gate in `karsift-ai-infra`. Do not flip autonomy markers
in `a003-transition-state.yaml` or `protected-paths.yaml`.

## Task ordering notes

- T00 + T01 are tightly coupled; one combined PR is preferred.
- T02 should see the final AGENTS.md wording (same PR as T00/T01 is fine when
  DOC-15 is out of scope; if DOC-15 is in scope, keep the doc set coherent in
  one PR and record R4 founder approval against that revision).
- T03 can be its own PR so VOC-072 unblocks quickly; may combine with T00–T02
  if adoption wants a single land.
- T04 must not land before backfill scope and exemption policy are settled;
  otherwise the lint immediately fails on historical packages.
- No task may be dispatched before this package is adopted.
