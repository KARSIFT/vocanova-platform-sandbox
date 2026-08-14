# VOC-075-EV-02 — Doc reconciliation (T02)

## Decision

`VOC-075-DEP-00` option **(a)**: edit DOC-15 in the same change set so it no
longer describes a general non-R4 founder-approval opt-out via
`automatic_merge_allowed`. Package risk re-declared **R4**;
`automatic_merge_allowed: false` (self-describing under approve-only-R4).

## Files reconciled

| File | Action |
|---|---|
| `docs/governance/change-risk-classification.md` | Added `automatic_merge_allowed` drafting section: R0–R3 → `true`; R4 → `false`; no sensitive-R3 carve-out |
| `docs/governance/approval-matrix.md` | Added `develop` / `automatic_merge_allowed` section matching the same rule |
| `docs/operations/15-ai-native-product-and-engineering-operating-model.md` | §17.1, §17.2, §17.3, and DG5-08 correction updated to approve-only-R4 |
| `AGENTS.md` | DEP-00 residual “until T02” note updated to record option (a) resolution (doc-reconciliation) |
| `specs/changes/VOC-075-…/change.yaml` | `risk: R4`; `automatic_merge_allowed: false`; DEP-00 status resolved as option (a) |
| `specs/changes/VOC-075-…/impact-analysis.md` | DEP-00 marked resolved option (a) |
| `specs/changes/VOC-075-…/release-plan.md` | Package field note updated for R4 |

## AC-02 / AC-05

- No VOC-068-style R3 carve-out remains in docs touched by this task.
- No edits to `karsift-ai-infra` merge-gate, `auto_merge_enabled`, release/deploy
  autonomy switches, `a003-transition-state.yaml`, or `protected-paths.yaml`.

## Approvals still required

Founder `approved` on the exact revision that includes the DOC-15 edit (R4 path
floor). Independent verification of that remediated revision (separate from the
implementer).
