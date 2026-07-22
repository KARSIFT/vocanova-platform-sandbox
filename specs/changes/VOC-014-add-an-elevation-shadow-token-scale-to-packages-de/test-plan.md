# VOC-014 — Test Plan

## VOC-014-TEST-00 — Elevation scale matches the exact declared values

- Covers: `VOC-014-AC-00`
- Preconditions: `packages/design-tokens/src/elevation.ts` exists.
- Procedure: run `pnpm run typecheck:packages` and `pnpm run build:packages`;
  compare each of the five `elevation` values against both the acceptance
  criteria table and the actual file contents, value by value.
- Expected result: zero typecheck/build errors; all five values match exactly,
  including the `none` keyword and the exact `rgb(0 0 0 / <alpha>)` and
  multi-layer comma formatting of `sm`/`md`/`lg`/`xl`.
- Evidence: `VOC-014-EV-00`

## VOC-014-TEST-01 — Entry point gains elevation without disturbing existing exports

- Covers: `VOC-014-AC-01`
- Preconditions: `VOC-014-T00` complete.
- Procedure: inspect `packages/design-tokens/src/index.ts`; confirm `elevation`
  is re-exported and that the pre-existing `spacing`/`neutral`/`fontSize`/
  `radius`/`duration`/`easing` export lines are byte-for-byte unchanged from the
  merged VOC-013 state.
- Expected result: seven named exports resolve from the package entry point; no
  regression to the six pre-existing ones.
- Evidence: `VOC-014-EV-01`

## VOC-014-TEST-02 — Full deterministic check suite passes

- Covers: `VOC-014-AC-02`
- Preconditions: all tasks complete.
- Procedure: run, from repo root:
  ```bash
  pnpm run lint:packages
  pnpm run typecheck:packages
  pnpm run build:packages
  ```
- Expected result: all three exit zero, no new findings anywhere else in the
  workspace.
- Evidence: `VOC-014-EV-02`

No security/authorization/migration/rollback test is applicable — same reasoning
as VOC-010/011/012/013 (purely additive static string literals, zero consumers).
No secrets or production data used.
