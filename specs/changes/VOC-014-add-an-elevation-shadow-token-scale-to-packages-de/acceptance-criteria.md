# VOC-014 — Acceptance Criteria

## VOC-014-AC-00 — Elevation scale is exported and typed with exact values

- Requirement source: `VOC-014-D00`
- Tasks: `VOC-014-T00`
- Tests: `VOC-014-TEST-00`
- Evidence: `VOC-014-EV-00`
- Result: pending

`packages/design-tokens/src/elevation.ts` exports a `const elevation` object,
typed as `Readonly<Record<string, string>>`, with exactly these five keys and
values:

| key    | value                                                              |
|--------|-------------------------------------------------------------------|
| `none` | `"none"`                                                          |
| `sm`   | `"0 1px 2px 0 rgb(0 0 0 / 0.05)"`                                 |
| `md`   | `"0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)"` |
| `lg`   | `"0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)"` |
| `xl`   | `"0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)"` |

Every value must match the table exactly, byte-for-byte, including internal
spacing, commas, and the modern `rgb(0 0 0 / <alpha>)` color syntax.

## VOC-014-AC-01 — Re-exported from the package entry point without disturbing existing exports

- Requirement source: `VOC-014-D00`
- Tasks: `VOC-014-T00`
- Tests: `VOC-014-TEST-01`
- Evidence: `VOC-014-EV-01`
- Result: pending

`packages/design-tokens/src/index.ts` exports `elevation` (named export) in
addition to the existing `spacing`, `neutral`, `fontSize`, `radius`, `duration`,
and `easing` exports — none of those six may be removed, renamed, or have their
values altered.

## VOC-014-AC-02 — Deterministic checks pass

- Requirement source: `VOC-014-D00`
- Tasks: `VOC-014-T00`
- Tests: `VOC-014-TEST-02`
- Evidence: `VOC-014-EV-02`
- Result: pending

`pnpm run lint:packages`, `pnpm run typecheck:packages`, and
`pnpm run build:packages` all exit zero, with no new lint or type errors
introduced anywhere else in the workspace.

Acceptance criteria are observable, stable, and bidirectionally traceable to
`VOC-014-D00`, task `VOC-014-T00`, tests `VOC-014-TEST-00..02`, and evidence
`VOC-014-EV-00..02`.
