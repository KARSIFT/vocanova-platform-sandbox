# VOC-031 — P5 Integrated Core Loop Mock Disposition Inventory

## Scope and authority

This document is the `T05` mock-disposition and no-new-backend-scope inventory. It is a **draft**,
written before adoption and before `T00`–`T05` implementation; it records the state confirmed at
draft time (`VOC-031-D00`) and the check this package's `T05` will add. It must be updated at `T05`
implementation time to record the actual post-`T00`–`T04` state, and this draft text preserved in
git history at that point, matching the pattern established by `VOC-027`/`VOC-028`/`VOC-030`'s own
mock-inventory documents.

## Current mock state (confirmed at draft time, 2026-07-27)

`scripts/foundation/mock-inventory.mjs`'s `expectedMocks` array (retained P4-pending placeholders)
is **empty**. Its `decommissionedMocks` array confirms every mock introduced by VOC-010–VOC-024 —
`MOCK_DISCOVER_SITUATIONS` (VOC-021), `MOCK_SITUATION_WORD_LISTS` (VOC-022, two call sites plus a
retired `_lib/mock-word-data.ts` file), `MOCK_HOME_STATE` (VOC-019, retired by VOC-030-T05), and
`MOCK_PROGRESS_STATE` (VOC-020, retired by VOC-030-T05) — has already been decommissioned to a real
backend source. **No mock remains for this package to decommission.** P5 is integration/
reliability/accessibility/performance work over already-real capability, not a mock-retirement
milestone.

## No-new-backend-scope inventory (to be verified at `T05`)

DOC-12 §5's P5 objective is a cross-feature/reliability/accessibility/performance/consistency pass,
not a new domain capability. `T05` extends `scripts/foundation/mock-inventory.mjs` with an explicit
sweep asserting, at the final `T00`–`T04` SHA:

| Area | Adopted-base state (confirmed 2026-07-27) | `T05` assertion |
| --- | --- | --- |
| `apps/api/business/*` modules | `auth`, `content`, `learning`, `reviews`, `aifeedback`, `missions`, `gamification` | No addition, removal, or rename |
| `apps/api/ent/schema/*` | one schema file per existing table through P4 | No addition, removal, or rename |
| `apps/api/migrations/*` | runs through `20260725130002_voc030_p4_gamification_tables.sql` | No new migration file |
| `apps/api/app/api/*` routes | existing A1/P1/P2/P3/P4 routes only | No new operation ID/path in the committed OpenAPI document |
| `apps/web/src/app/onboarding`, `apps/web/src/app/(app)/settings` | do not exist | Do not exist unless `VOC-031-D01` is resolved to require them (recorded explicitly, not silently) |

## New real P5 additions (added by `T00`–`T04`, to be filled in at `T05`)

| Area | New addition | Task |
| --- | --- | --- |
| Web route file | `apps/web/src/app/(app)/discover/loading.tsx` | T00 |
| Web route file | `apps/web/src/app/(app)/discover/error.tsx` | T00 |
| Web route file | `apps/web/src/app/(app)/discover/[situation]/loading.tsx` | T00 |
| Web route file | `apps/web/src/app/(app)/discover/[situation]/error.tsx` | T00 |
| Web route file | `apps/web/src/app/(app)/discover/[situation]/[word]/loading.tsx` | T00 |
| Web route file | `apps/web/src/app/(app)/discover/[situation]/[word]/error.tsx` | T00 |
| Web route file | `apps/web/src/app/(app)/reviews/loading.tsx` | T00 |
| Web route file | `apps/web/src/app/(app)/reviews/error.tsx` | T00 |
| Web route file | `apps/web/src/app/global-error.tsx` | T00 |
| Middleware change | `apps/web/src/middleware.ts` fail-closed auth-check reliability | T01 |
| Test/tooling | accessibility/performance automation, only if `D02`/`D03` activate it | T03/T04 |

_(This table is a draft placeholder reflecting the planned `T00`–`T04` deliverables; `T05` must
replace it with the actual files/changes present at that point, not merely confirm this plan was
followed.)_

## Verified boundaries (to be enforced by the extended `scripts/foundation/mock-inventory.mjs`)

- **No new `apps/api/business` module, `apps/api/ent/schema` file, or `apps/api/migrations` file**
  — a regression that adds one would silently expand P5 into P1–P4-deferred backend scope.
- **No new API route** unless the committed OpenAPI document changes, which this package's scope
  does not call for.
- **No `apps/web/src/app/onboarding` or `apps/web/src/app/(app)/settings` route** unless
  `VOC-031-D01` is recorded as resolved to require them.
- **No re-introduction of a retired mock** (`MOCK_HOME_STATE`, `MOCK_PROGRESS_STATE`,
  `MOCK_DISCOVER_SITUATIONS`, `MOCK_SITUATION_WORD_LISTS`) — the existing `decommissionedMocks`
  check already guards this; `T05` re-confirms it still passes after `T00`–`T04`.

## Evidence pointers

- `VOC-031-EV-33`..`EV-35` (in `staging-evidence.md`): `T05`'s no-new-backend-scope diff evidence
  and the extended mock-inventory check pass.
