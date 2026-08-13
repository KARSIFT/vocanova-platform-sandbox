# VOC-075 — Implementation Plan

## Preconditions and protected areas

Do not begin any task until this package is adopted and implementation is
authorized. Adoption must record decisions for `VOC-075-DEP-00`,
`VOC-075-DEP-01`, and `VOC-075-DEP-02` so implementers are not guessing.
If DEP-00 option a and/or DEP-02 option a are chosen, re-declare the
package risk as **R4** at adoption and obtain founder approval against the
exact implemented revision(s) that touch those paths.

Protected areas in default / conditional scope:

- `AGENTS.md` — R3 (`VOC-075-T00`)
- `specs/templates/change-package/` — R3 (`VOC-075-T01`)
- `docs/governance/change-risk-classification.md`,
  `docs/governance/approval-matrix.md` — R3 (`VOC-075-T02`)
- `specs/changes/VOC-072-…/change.yaml` (+ optional others) — ordinary
  change-package paths (`VOC-075-T03`)
- Conditionally DOC-15 — R4 (`VOC-075-T02` if DEP-00 option a)
- Conditionally `scripts/governance/` / `tooling/governance/` — R4
  (`VOC-075-T04` if DEP-02 option a)

Explicitly out of scope: `karsift-ai-infra` workflow semantics,
`apps/`, `packages/`, migrations, production secrets/configuration,
autonomy/authorization marker flips in `a003-transition-state.yaml` /
`protected-paths.yaml`.

## File reconciliation and implementation sequence

1. **Confirm adoption decisions** recorded in
   `requirement_approval_status` / `blocking_reasons` (or the adoption PR
   description): DEP-00, DEP-01, DEP-02, and the resulting declared risk
   class.
2. **`VOC-075-T00` + `VOC-075-T01`** — Rewrite AGENTS.md drafting rule;
   align template comment + README. Prefer one PR.
3. **`VOC-075-T02`** — Reconcile governance docs; apply DOC-15 only if
   DEP-00 option a. May share the T00/T01 PR when coherent.
4. **`VOC-075-T03`** — Flip VOC-072 (required) and any additional packages
   from DEP-01. Prefer a fast follow PR if VOC-072 is still blocking at
   implementation time.
5. **`VOC-075-T04`** — Only if DEP-02 option a; land after backfill /
   exemption policy will not immediately fail CI.
6. Run governance validation commands below before claiming complete.

Preserve AGENTS.md's "Safety" / release-authority historical records; do not
rewrite the 2026-08-08 deployment-delegation section while editing the
drafting subsection.

## Validation and independent verification

Deterministic commands before claiming any task complete:

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

If T04 adds a new check, invoke that entrypoint as well (exact command
recorded in `VOC-075-EV-04`).

Confirm `classify-change-risk.sh`'s detected floor is not higher than the
package's declared class (or raise the declaration when DOC-15 / lint paths
are included).

Independent verification (per `CLAUDE.md`) must confirm against the exact
implemented revision:

- AGENTS.md and template match approve-only-R4; VOC-068 carve-out text is
  gone.
- Doc reconciliation (`VOC-075-AC-02`) is satisfied with evidence.
- VOC-072 (and adoption-scoped backfills) are `true` while non-R4.
- If T04 present: lint enforces the rule under the adopted scan policy;
  tests pass; autonomy markers untouched.
- No workflow / autonomy-switch drift beyond allowed lint implementation
  files.
- Codex (or current implementer-role occupant) did not approve or merge its
  own implementation.

## Deployment and rollback

No application deployment effect is intended. Rollback is a documentation /
metadata / lint revert.

Rollback trigger: guidance instructs planners to auto-merge R4, to skip
verification, or leaves contradictory docs; or lint falsely blocks
legitimate R4 packages / fails closed incorrectly.

Rollback mechanism: revert the AGENTS.md, template, doc, backfill, and/or
lint commits for the affected task(s). Last-known-good: revisions
immediately preceding this package's implementation merge(s).

Owner: implementer of the affected task.
