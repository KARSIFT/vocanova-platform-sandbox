# VOC-075 — Acceptance Criteria

## VOC-075-AC-00 — AGENTS.md states approve-only-R4 `automatic_merge_allowed` drafting

- Requirement source: issue #573; `VOC-075-D00`
- Tasks: `VOC-075-T00`
- Tests: `VOC-075-TEST-00`
- Evidence: `VOC-075-EV-00`
- Result: pending

`AGENTS.md` contains an explicit drafting rule that:

1. Requires **R0–R3** packages to set `automatic_merge_allowed: true`.
2. Requires **R4** packages to set `automatic_merge_allowed: false`.
3. Contains **no** R3 case-by-case carve-out and **no** language that
   auth/secrets/production infrastructure (or similar) justify a non-R4
   founder-approval opt-out via this field.
4. Contains **no** encouragement for R0–R2 deliberate `false` opt-outs.
5. Still states that the field does not bypass risk classification, path
   floors, CI, independent verification, R4 founder authority, or EHR.

## VOC-075-AC-01 — Change-package template matches the approve-only-R4 rule

- Requirement source: issue #573; `VOC-075-D00`
- Tasks: `VOC-075-T01`
- Tests: `VOC-075-TEST-01`
- Evidence: `VOC-075-EV-01`
- Result: pending

`specs/templates/change-package/change.yaml` comment block and
`specs/templates/change-package/README.md` match `VOC-075-AC-00`: R0–R3 →
`true`; R4 → `false`; no R3 justified-false path; no encouraged non-R4
opt-out. A new planner reading only the template cannot infer that a
"sensitive" R3 package should set `false`.

## VOC-075-AC-02 — Doc reconciliation leaves no contradictory carve-out

- Requirement source: AGENTS.md doc-reconciliation rule; `VOC-075-DEP-00`;
  issue #573 suggested doc updates
- Tasks: `VOC-075-T02`
- Tests: `VOC-075-TEST-02`
- Evidence: `VOC-075-EV-02`
- Result: pending

Either:

- DOC-15 is updated in the same change set so it no longer describes a
  general non-R4 founder-approval opt-out via `automatic_merge_allowed`
  (adoption chose `VOC-075-DEP-00` option a; package risk at least R4), **or**
- T02 evidence records why DOC-15 was left unchanged (adoption chose option
  b) and cites the exact paragraphs, with an explicit residual-risk note.

`docs/governance/change-risk-classification.md` and
`docs/governance/approval-matrix.md` either already match approve-only-R4
(with cite) or are edited so they do. No doc touched by this package may
retain the VOC-068-style R3 carve-out for this field.

## VOC-075-AC-03 — VOC-072 (and adoption-scoped backfills) no longer opt out

- Requirement source: issue #573; `VOC-075-DEP-01`
- Tasks: `VOC-075-T03`
- Tests: `VOC-075-TEST-03`
- Evidence: `VOC-075-EV-03`
- Result: pending

`specs/changes/VOC-072-same-request-as-github-issue-535-voc-067-t05/change.yaml`
has `automatic_merge_allowed: true` (VOC-072 remains non-R4). Every
additional package adoption included under `VOC-075-DEP-01` is likewise
flipped with an auditable comment citing VOC-075 / issue #573. No in-scope
backfill package remains at `false` while declaring risk R0–R3.

## VOC-075-AC-04 — Lint prevents non-R4 `false` (if adoption includes T04)

- Requirement source: issue #573; `VOC-075-DEP-02`
- Tasks: `VOC-075-T04`
- Tests: `VOC-075-TEST-04`
- Evidence: `VOC-075-EV-04`
- Result: pending — **N/A if adoption chose `VOC-075-DEP-02` option b**

If T04 is in scope: a deterministic check (wired into the existing
governance validation path or an equivalent CI-invoked script) fails when a
scanned `change.yaml` has `automatic_merge_allowed: false` without
`risk: R4`, using the scan/exemption policy recorded at adoption. Positive
and negative fixtures or inline test cases exist. Path risk declaration for
the package is at least R4.

## VOC-075-AC-05 — Guidance and backfill do not alter merge-gate or autonomy switches

- Requirement source: `specification.md` non-goals
- Tasks: `VOC-075-T00`, `VOC-075-T01`, `VOC-075-T02`, `VOC-075-T03`,
  `VOC-075-T04`
- Tests: `VOC-075-TEST-05`
- Evidence: `VOC-075-EV-00` through `VOC-075-EV-04` as applicable
- Result: pending

Implementation PR(s) do not modify `karsift-ai-infra` merge-gate behavior,
this repo's `auto_merge_enabled` / release/deploy autonomy switches,
`docs/governance/a003-transition-state.yaml` merge/release fields, or
`.github/approved-policy/protected-paths.yaml` merge/release fields except
where a lint task unavoidably touches only the check implementation under
`scripts/governance/` or `tooling/governance/` (not those autonomy markers).
R4 hard-block behavior remains intact.
