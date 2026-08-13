# VOC-075 — Test Plan

## VOC-075-TEST-00 — AGENTS.md drafting rule is approve-only-R4 and non-weakening

- Covers: `VOC-075-AC-00`
- Preconditions: `VOC-075-T00` diff available.
- Procedure: Read the revised AGENTS.md drafting subsection. Confirm R0–R3 →
  `true`, R4 → `false`; grep the subsection (and nearby Change workflow text)
  for leftover carve-out phrases such as "case-by-case", "founder eyes",
  "auth, secrets, or production infrastructure" as justifications for
  non-R4 `automatic_merge_allowed: false`. Confirm R4/EHR/CI/independent
  verification are preserved and that `true` is not claimed to bypass them.
- Expected result: Rule matches founder instruction; carve-out language
  gone; controls not weakened.
- Evidence: `VOC-075-EV-00`

## VOC-075-TEST-01 — Template matches approve-only-R4

- Covers: `VOC-075-AC-01`
- Preconditions: `VOC-075-T01` diff available.
- Procedure: Read `specs/templates/change-package/change.yaml` and
  `README.md`. Confirm comments/blurb match AC-00. Confirm absence of "R3:
  true or false with stated justification" and of "deliberate opt-outs" as
  a normal non-R4 path.
- Expected result: Template cannot be read as authorizing non-R4 `false`.
- Evidence: `VOC-075-EV-01`

## VOC-075-TEST-02 — Doc reconciliation satisfied

- Covers: `VOC-075-AC-02`
- Preconditions: Adoption chose `VOC-075-DEP-00` option a or b; T02 complete.
- Procedure:
  - Grep `docs/governance/change-risk-classification.md`,
    `docs/governance/approval-matrix.md`, and (if touched) DOC-15 for
    language that would reintroduce a non-R4 founder-approval opt-out via
    `automatic_merge_allowed`.
  - If DEP-00 option a: confirm DOC-15 wording matches AGENTS.md and package
    risk declaration is at least R4.
  - If DEP-00 option b: confirm evidence cites exact DOC-15 paragraphs and
    records residual risk.
- Expected result: AC-02 satisfied; no silent contradictory carve-out in
  touched docs.
- Evidence: `VOC-075-EV-02`

## VOC-075-TEST-03 — Backfill field values

- Covers: `VOC-075-AC-03`
- Preconditions: T03 complete; DEP-01 scope recorded.
- Procedure: Parse each in-scope package `change.yaml`. Assert
  `automatic_merge_allowed: true` and `risk` is not R4. Assert VOC-072 is
  included. List every flipped path in evidence. Spot-check that
  `status` / `implementation_authorized` / other adoption fields were not
  altered by T03.
- Expected result: All in-scope non-R4 packages are `true`; VOC-072 fixed;
  no accidental adoption-field edits.
- Evidence: `VOC-075-EV-03`

## VOC-075-TEST-04 — Lint behavior (conditional)

- Covers: `VOC-075-AC-04`
- Preconditions: T04 in scope; fixtures or unit tests available.
- Procedure:
  - Negative: a fixture/package with `risk: R3` (or R1/R2) and
    `automatic_merge_allowed: false` fails the check.
  - Positive: `risk: R4` with `false` passes; R0–R3 with `true` passes.
  - Confirm exemption/scan policy from adoption is honored (no surprise
    failures on exempted historical packages, and no exemption that would
    allow new non-R4 `false` to land).
  - Run the wired validation entrypoint.
- Expected result: Deterministic pass/fail matches the rule; evidence
  includes command output.
- Evidence: `VOC-075-EV-04`
- N/A: if adoption chose DEP-02 option b, record N/A in evidence and skip.

## VOC-075-TEST-05 — No merge-gate or autonomy-switch drift

- Covers: `VOC-075-AC-05`
- Preconditions: Full PR diff(s) for the task set available.
- Procedure: Inspect file lists. Assert no edits to merge-gate semantics in
  `karsift-ai-infra`, and no autonomy marker changes in
  `docs/governance/a003-transition-state.yaml` or
  `.github/approved-policy/protected-paths.yaml`. Allowed exception: new/edited
  check files under `scripts/governance/` or `tooling/governance/` for T04
  only. Run:

  ```bash
  bash scripts/governance/validate-governance.sh
  bash scripts/governance/classify-change-risk.sh
  git diff --check
  ```

- Expected result: Validation passes; declared risk meets or exceeds
  detected floor; forbidden files unmodified.
- Evidence: `VOC-075-EV-00` through `VOC-075-EV-04` as applicable

## Rollback coverage

Rolling back means reverting the documentation / template / backfill / lint
commits. Validation: re-run `validate-governance.sh` and `git diff --check`
on the reverted tree; confirm AGENTS.md, template, and any flipped package
fields return to their pre-VOC-075 text where intended.

No data rollback is required.

## Constraints

No test in this plan uses secrets or production data. No live merge-gate
experiment against production credentials is required for closure; observing
that VOC-072 task PRs no longer demand founder `approved` solely for
`automatic_merge_allowed: false` is an expected post-merge consequence, not
a gate that requires production access inside this package.
