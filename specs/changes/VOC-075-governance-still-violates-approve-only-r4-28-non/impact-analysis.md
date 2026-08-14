# VOC-075 — Impact Analysis

## Security and privacy

No new secret, credential, authentication, authorization, or personal-data
handling is introduced. The change is governance drafting text, template
comments, optional deterministic lint, and metadata field flips on existing
change packages.

Residual security consideration: removing the non-R4 opt-out path means
packages that touch auth, secrets, or production infrastructure at **R3**
will auto-merge into `develop` when CI and independent review pass (project
switch on), without a founder `approved` comment — which is exactly the
founder's stated rule. Mitigation: R3 still requires strengthened controls
and independent verification; path floors and merge-gate's R4 hard block
remain; EHR remains available for exceptional triggers; founder override
remains valid when the gate requires approval.

Incorrect guidance that told planners to set `automatic_merge_allowed: true`
for R4, or to skip independent verification, would weaken controls.
Mitigation: acceptance criteria and tests explicitly forbid weakening R4,
EHR, CI, or independent verification.

## Data and migrations

None. No schema, seed, or production data change. Backfill touches only
`automatic_merge_allowed` (and an adjacent comment) on named package
`change.yaml` files.

## Analytics and accessibility

None. No user-facing behavior, analytics, or UI change. Explicit
non-applicability: governance / template / package-metadata package.

## Risks, dependencies, and evidence

- `VOC-075-R00`: **Sensitive R3 work auto-merges without founder comment.**
  This is the intended consequence of the founder rule, not an accident.
  Residual risk is operational (founder sees fewer merge gates), not a
  bypass of R3 technical controls. Mitigation: keep independent verification
  and CI mandatory; do not weaken path floors; do not change R4 hard block.
- `VOC-075-R01`: **DOC-15 / AGENTS.md contradiction if DEP-00 chooses (b).**
  Leaving DOC-15's general opt-out language while forbidding non-R4 `false`
  in AGENTS.md recreates doc drift. Mitigation: draft recommends DEP-00 (a);
  AC-02 forces an explicit recorded choice.
- `VOC-075-R02`: **Lint vs historical packages.** A scan-all lint without
  matching backfill fails CI immediately on the 28 existing packages.
  Mitigation: T04 ordering note; DEP-01/DEP-02 must be co-settled at
  adoption.
- `VOC-075-R03`: **Scope creep into workflow or autonomy switches.**
  Mitigation: AC-05 / TEST-05 forbid those edits.
- `VOC-075-R04`: **VOC-072 field flip alone does not finish VOC-072's
  product work.** Backfill only removes an incorrect founder-approval gate;
  it does not implement VOC-072 tasks. No mitigation needed beyond clarity
  in T03 evidence.
- `VOC-075-DEP-00`: Resolved — option (a). DOC-15 edited under T02 to match
  approve-only-R4; package risk raised to R4; `automatic_merge_allowed: false`.
- `VOC-075-DEP-01`: Unresolved — backfill scope (VOC-072 only vs active vs
  all 28).
- `VOC-075-DEP-02`: Unresolved — lint inclusion / possible R4 raise.
- `VOC-075-EV-00`: AGENTS.md diff showing approve-only-R4 drafting rule.
- `VOC-075-EV-01`: Template `change.yaml` / `README.md` diff.
- `VOC-075-EV-02`: Doc reconciliation evidence (edits or cites).
- `VOC-075-EV-03`: VOC-072 (and any additional) `change.yaml` field flips.
- `VOC-075-EV-04`: Lint implementation + test results (or N/A if deferred).
