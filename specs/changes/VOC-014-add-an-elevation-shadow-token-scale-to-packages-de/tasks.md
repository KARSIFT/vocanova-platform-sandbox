# VOC-014 — Tasks

## VOC-014-T00 — Add typed elevation scale and wire into the package entry point

- Requirement source: `VOC-014-D00`
- Acceptance criteria: `VOC-014-AC-00`, `VOC-014-AC-01`, `VOC-014-AC-02`
- Tests: `VOC-014-TEST-00`, `VOC-014-TEST-01`, `VOC-014-TEST-02`
- Evidence: `VOC-014-EV-00`, `VOC-014-EV-01`, `VOC-014-EV-02`
- Status: pending

Single task (per VOC-010's, VOC-011's, and VOC-012's precedent: splitting one
tightly-coupled change across multiple per-task PRs fails review, since each PR
is reviewed against the whole package's acceptance criteria independently):

1. Create `packages/design-tokens/src/elevation.ts` exporting a readonly
   `elevation` object with the exact five keys and values in `VOC-014-AC-00`'s
   table.
2. Update `packages/design-tokens/src/index.ts` to add
   `export { elevation } from "./elevation.js";`, preserving the existing
   `spacing`, `neutral`, `fontSize`, `radius`, `duration`, and `easing` export
   lines unchanged.

Scoped to introduce no new dependency, script, or protected-path change, and to
preserve the VOC-010→VOC-013 existing exports exactly as they are. This task is
independently implementable and reviewable in one pull request. It carries no
migration or rollback complication beyond a plain revert.
