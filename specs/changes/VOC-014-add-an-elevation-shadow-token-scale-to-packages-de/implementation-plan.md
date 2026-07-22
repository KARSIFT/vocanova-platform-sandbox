# VOC-014 — Implementation Plan

## Preconditions and protected areas

Do not begin until this package is adopted and implementation is authorized: a
human sets `status`, `approval_status`, `implementation.authorized`, and records
a founder-approved `authority_issue` in `change.yaml` (all at unadopted defaults
in this draft). Depends on the VOC-010→VOC-013 exports already merged to
`develop` (they are present at authoring time). No protected areas are touched.

## File reconciliation and implementation sequence

Existing target: `packages/design-tokens/src/index.ts` currently exports
`spacing` (from `./spacing.js`), `neutral` (from `./colors.js`), `fontSize`
(from `./typography.js`), `radius` (from `./radius.js`), `duration` (from
`./duration.js`), and `easing` (from `./easing.js`). All six must be preserved
unchanged.

New target: `packages/design-tokens/src/elevation.ts` does not yet exist; it is
created fresh.

Ordered steps (single task, `VOC-014-T00`):

1. Create `packages/design-tokens/src/elevation.ts` implementing
   `VOC-014-AC-00` — a readonly `elevation` object with five literal
   `box-shadow` string values, taken verbatim from the table, no computation.
   Match the sibling files' exact shape, e.g.:
   ```ts
   export const elevation: Readonly<Record<string, string>> = {
     none: "none",
     sm: "0 1px 2px 0 rgb(0 0 0 / 0.05)",
     md: "0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)",
     lg: "0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)",
     xl: "0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)",
   };
   ```
2. Update `packages/design-tokens/src/index.ts` to add
   `export { elevation } from "./elevation.js";` (`.js` extension required by
   this package's NodeNext module resolution, as established in VOC-010) without
   touching the six existing export lines.

## Validation and independent verification

Deterministic commands (run from repo root):

```bash
pnpm run lint:packages
pnpm run typecheck:packages
pnpm run build:packages
```

Independent verification: the reviewer re-checks each of the five `elevation`
values in `VOC-014-AC-00`'s table individually against the file contents,
byte-for-byte, and confirms the six pre-existing exports are unchanged, per
`CLAUDE.md`. The verifier binds its verdict to the exact reviewed commit SHA and
confirms the implementer did not self-approve or self-merge.

## Deployment and rollback

`release.deployment: prohibited`. Rollback is a plain `git revert` of the merge
commit; nothing consumes these exports yet. Last-known-good reference is
`develop` at this package's (adoption-time) `base_sha`.
