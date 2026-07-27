# VOC-031 — P5 Integrated Core Loop Mock Disposition Inventory

## Scope and authority

This document is the `T09` mock-disposition inventory. It is a **draft
blueprint** written before adoption and before `T00`–`T08` implementation;
it records the mock state confirmed at draft time (`VOC-031-D00`) and the
boundaries the extended `scripts/foundation/mock-inventory.mjs` must enforce
once `T09` actually runs. It is updated at `T09` implementation time to
record the real post-`T00`–`T08` state; this draft text is superseded at
that point, the same way VOC-030's own draft mock-inventory text was
superseded at its `T06`.

## Confirmed mock state at draft time (`VOC-031-D00`, 2026-07-27)

| File / area | VOC source | Mock identifier / field | State at draft time |
| --- | --- | --- | --- |
| `apps/web/src/app/(app)/home/page.tsx` | VOC-030 | `MOCK_HOME_STATE` | **Already retired** — reads real `getDailyMission()`; no P5 work needed here |
| `apps/web/src/app/(app)/progress/page.tsx` | VOC-030 | `MOCK_PROGRESS_STATE` | **Already retired** — reads real `getProgress()`; no P5 work needed here |
| `apps/api/business/aifeedback/mission.go` | VOC-028-D01, VOC-030-T03 | `StubMissionUpdater` | **Kept as documented rollback fallback** (VOC-030 decision, unchanged by this package) |
| `apps/web` (repo-wide) | — | accessibility/performance test coverage | **Does not exist** — no axe-core, `@axe-core/playwright`, `playwright`, Lighthouse, or `@lhci/cli` dependency anywhere; `T06`/`T07` install real tooling, not a mock/manual substitute |
| `apps/api/business/gamification` | VOC-030 | `user_settings` public API | **Does not exist** — only an internal lazy-upsert used for timezone/target resolution; `T02` adds the first real public surface |
| — | — | `user_onboarding_profiles` | **Does not exist** — `T00` creates it for real, not a stub |

No P4-or-earlier mock remains presented as real data in the shipped
application; this package introduces no interim mock of its own — every new
surface (`onboarding`, `settings`, `account`) is built against its real
persistence from the start (`T00`–`T03`), matching how VOC-030 itself built
`missions`/`gamification` without an interim mock layer.

## Planned real P5 modules/tables/routes (to be added by `T00`–`T07`)

| Area | Planned addition | VOC source |
| --- | --- | --- |
| Business module | `apps/api/business/onboarding/` | DOC-05 §6 (T00) |
| Ent schema | `useronboardingprofile.go` | DOC-05 §6 (T00) |
| Migration | one new versioned migration for `user_onboarding_profiles`, ordered after `grace_day_ledger` | DOC-05 §18 (T00) |
| API route | `GET`/`PUT /api/v1/onboarding` | DOC-07 (T01) |
| API route | `GET`/`PATCH /api/v1/settings` | DOC-07 (T02) |
| API route | `GET`/`PATCH /api/v1/account` | DOC-07 (T02) |
| API client methods | `@vocanova/api-client` onboarding/settings/account methods | DOC-07 (T01, T02) |
| Web route group | `apps/web/src/app/(onboarding)/onboarding/page.tsx` | DOC-08 (T01) |
| Web route | `apps/web/src/app/(app)/settings/page.tsx` | DOC-08 (T03) |
| Web route | `apps/web/src/app/(app)/settings/account/page.tsx` | DOC-08 (T03) |
| Shared components | loading/empty/error component set adopted across `(app)`/`(onboarding)` | DOC-08 (T04) |
| Dev dependency | `playwright`, `@axe-core/playwright` | DOC-08 (T06) |
| Dev dependency | `@lhci/cli` | DOC-08 (T07) |
| CI job | accessibility sweep | `VOC-031-D03` (T06) |
| CI job | Lighthouse CI thresholds | `VOC-031-D04` (T07) |

## Boundaries the extended `scripts/foundation/mock-inventory.mjs` must enforce at `T09`

- **No regression of `MOCK_HOME_STATE`/`MOCK_PROGRESS_STATE`'s already-retired
  status** — the check must continue to assert neither reappears in any
  `(app)` route.
- **No P5 route/table/behavior invented beyond this document's planned
  list** — the existing API path allow list, schema allow list, migration
  allow list, and business-module allow list are extended to admit exactly
  the `onboarding` module/routes/schema/migration and the `settings`/
  `account` routes, and no more.
- **The two new accessibility/performance CI jobs must be present and
  referenced from the pipeline workflow** — a regression that silently
  removes either job from CI (rather than removing it from this inventory
  with its own reviewed decision) is caught at check time.
- **No already-shipped route renamed** (`VOC-031-D08`) — the route allow
  list is a fixed list at `T08` time; `/reviews` and the absence of a
  dedicated `/words` route are asserted unchanged.

## Evidence pointers

- `VOC-031-EV-06..EV-11` (in `staging-evidence.md`, once produced): `T01`
  onboarding API/frontend evidence.
- `VOC-031-EV-20..EV-23`: `T03` settings/account frontend evidence.
- `VOC-031-EV-35..EV-41`: `T06`/`T07` accessibility/performance CI evidence.
- `VOC-031-EV-45`: `T09` extended mock-inventory check pass.

This document does not itself constitute evidence of anything produced; it
is the pre-implementation disposition plan `T09` executes against and then
updates with the real, post-implementation record.
