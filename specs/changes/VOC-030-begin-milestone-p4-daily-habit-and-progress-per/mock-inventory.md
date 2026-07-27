# VOC-030 — P4 Daily Habit and Progress Mock Disposition Inventory

## Scope and authority

This document is the `T06` mock-disposition inventory. It is updated at
`T06` implementation time to record the post-`T00`–`T05` state, the
new real P4 modules/tables/routes, and the boundaries enforced by the
extended `scripts/foundation/mock-inventory.mjs`.

The previous draft text (which was written **before adoption**,
2026-07-26, and predated the `T00`–`T05` work) is preserved in the
spec file's git history; this version supersedes it.

## Current mock state (post-`T00`–`T05`, recorded at `T06` time)

| File / area | VOC source | Mock identifier / field | Disposition |
| --- | --- | --- | --- |
| `apps/web/src/app/(app)/home/page.tsx` | VOC-019 / VOC-026 | `MOCK_HOME_STATE.missionTargetWords` | **Retired** (T05) — sourced from `GET /api/v1/daily-mission` `reviewTarget` |
| `apps/web/src/app/(app)/home/page.tsx` | VOC-019 / VOC-026 | `MOCK_HOME_STATE.reviewedWordsToday` | **Retired** (T05) — sourced from `GET /api/v1/daily-mission` `reviewsCompleted` |
| `apps/web/src/app/(app)/home/page.tsx` | VOC-019 / VOC-026 | `MOCK_HOME_STATE.currentStreakDays` | **Retired** (T05) — sourced from `GET /api/v1/daily-mission` `streak.currentStreakCount` |
| `apps/web/src/app/(app)/home/page.tsx` | VOC-027-D05 | `dueReviewWords` (real P2 due-queue) | **Real** — unchanged, sourced from `GET /api/v1/reviews/due` |
| `apps/web/src/app/(app)/progress/page.tsx` | VOC-020 / VOC-026 | `MOCK_PROGRESS_STATE.confidencePointsTotal` | **Retired** (T05) — sourced from `GET /api/v1/progress` `confidencePointsBalance` |
| `apps/web/src/app/(app)/progress/page.tsx` | VOC-020 / VOC-026 | `MOCK_PROGRESS_STATE.currentStreakDays` | **Retired** (T05) — sourced from `GET /api/v1/progress` `streak.currentStreakCount` |
| `apps/web/src/app/(app)/progress/page.tsx` | VOC-020 / VOC-026 | `MOCK_PROGRESS_STATE.longestStreakDays` | **Retired** (T05) — sourced from `GET /api/v1/progress` `streak.longestStreakCount` |
| `apps/web/src/app/(app)/progress/page.tsx` | VOC-020 / VOC-026 | `MOCK_PROGRESS_STATE.completionHistory` | **Retired** (T05) — sourced from `GET /api/v1/progress` `completionHistory` |
| `apps/api/business/aifeedback/mission.go` | VOC-028-D01 | `StubMissionUpdater` | **Kept as documented rollback fallback** — production wiring in T03 replaced it with `missions.MissionUpdater` |
| `apps/api/business/aifeedback/aifeedback_p4_test.go` | VOC-030-T03 | `missions.NewMissionUpdater(...)` fixture wiring | **Real** — the T06 mock-inventory check (`mock-inventory.mjs`) asserts this fixture is present, so a regression that removes the real-updater wiring would be caught at check time |

## New real P4 modules/tables/routes (added by `T00`–`T04`)

| Area | New addition | VOC source |
| --- | --- | --- |
| Business module | `apps/api/business/missions/` | DOC-06 §3 (T00) |
| Business module | `apps/api/business/gamification/` | DOC-06 §3 (T00) |
| Ent schema | `dailymissionsnapshot.go` | DOC-05 §10 (T00) |
| Ent schema | `dailyactivitysummary.go` | DOC-05 §10 (T00) |
| Ent schema | `confidencepointledger.go` | DOC-05 §12 (T00) |
| Ent schema | `streakstate.go` | DOC-05 §12 (T00) |
| Ent schema | `gracedayledger.go` | DOC-05 §12 (T00) |
| Ent schema | `usersettings.go` | DOC-05 §6 (T00, D01) |
| Migration | `20260725130000_voc030_p4_user_settings.sql` | DOC-05 §6 (T00) |
| Migration | `20260725130001_voc030_p4_mission_tables.sql` | DOC-05 §10 (T00) |
| Migration | `20260725130002_voc030_p4_gamification_tables.sql` | DOC-05 §12 (T00) |
| API route | `GET /api/v1/daily-mission` | DOC-07 (T04) |
| API route | `GET /api/v1/progress` | DOC-07 (T04) |
| API client method | `@vocanova/api-client` `getDailyMission()` | DOC-07 (T04) |
| API client method | `@vocanova/api-client` `getProgress()` | DOC-07 (T04) |
| Web page wiring | `apps/web/src/app/(app)/home/page.tsx` (real `getDailyMission()`) | DOC-08 (T05) |
| Web page wiring | `apps/web/src/app/(app)/progress/page.tsx` (real `getProgress()`) | DOC-08 (T05) |

## Verified boundaries (enforced by the extended `scripts/foundation/mock-inventory.mjs`)

The `T06` extension of the mock-inventory check enforces the following
boundaries, beyond the T04/T05 baseline:

- **No `MOCK_HOME_STATE` or `MOCK_PROGRESS_STATE` object** may appear
  in any `(app)` route. Both are retired (T05) and replaced with real
  `GET /api/v1/daily-mission` and `GET /api/v1/progress` reads. The
  check inspects the page source directly, so a regression that
  re-introduces either mock object — or a renamed equivalent — is
  caught at check time.
- **The real `missions.NewMissionUpdater(...)` fixture wiring** must
  be present in `apps/api/business/aifeedback/aifeedback_p4_test.go`.
  T03 replaced the `StubMissionUpdater` seam with the real
  `missions.MissionUpdater`; a regression that removes the fixture
  wiring would silently re-introduce the always-`false`
  `missionCompleted` behavior in production.
- **The T06 cross-cutting safety test files** must exist
  (`apps/api/business/missions/cross_cutting_safety_test.go` and
  `apps/api/app/api/missions_cross_cutting_test.go`). Their presence
  is part of the T06 acceptance criteria for the
  duplicate/failed/unauthorized-safety guarantees
  (`VOC-030-AC-08`).
- **No P5 route, table, or behavior** is introduced. The existing
  API path allow list, schema allow list, migration allow list, and
  business-module allow list enforce the route/table side; the T06
  extension adds a P5-forbidden regex sweep over the T06
  deliverable files (a future P5 feature must not be introduced
  through the T06 cross-cutting test files).

## T05/T06 follow-up notes

- The `StubMissionUpdater` is **kept** as a documented rollback
  fallback (per `implementation-plan.md`'s rollback section), so the
  `apps/api/business/aifeedback/service.go` default `if mission == nil {
  mission = NewStubMissionUpdater() }` is intentional and not a
  regression. The check asserts the **fixture** wires the real
  updater, not that the production code path is stub-free.
- `D03` follow-up: if the founder activates the optional new-word /
  sentence-practice mission goals after MVP launch, extend `T01` /
  `T03`'s wiring accordingly under a new `policy_version`. The T06
  boundary check is forward-compatible — the existing
  `dailymissionsnapshot.go` and `dailyactivitysummary.go` schemas
  already include the optional columns.

## Evidence pointers

- `VOC-030-EV-26..EV-29` (in `staging-evidence.md`): T05 mock
  retirement evidence — the home/progress pages render real
  `getDailyMission` / `getProgress` data, with `MOCK_HOME_STATE` /
  `MOCK_PROGRESS_STATE` no longer driving the UI.
- `VOC-030-EV-30..EV-33` (in `staging-evidence.md`): T06
  cross-cutting duplicate/failed/unauthorized-safety test evidence
  — `TEST-30..TEST-33` in
  `apps/api/business/missions/cross_cutting_safety_test.go` and
  `apps/api/app/api/missions_cross_cutting_test.go`.
- `VOC-030-EV-34` (in `staging-evidence.md`): T06 deterministic
  mock-inventory check pass, including the T06 extension.
