# VOC-026 — P1 Learning-Content Mock Decommission Inventory

## Scope and authority

This document is the T05 inventory required by `VOC-026-D05`/`VOC-026-D06`:
every VOC-010–VOC-024 learning-content mock/placeholder source touched by P1 is
listed with its disposition. The adopted resolutions are:

- `VOC-026-D01`: the VOC-022 situation/word set becomes the real canonical seed.
- `VOC-026-D03`: discovery shows all meanings with a requester-scoped `saved`
  overlay.
- `VOC-026-D04`: save writes only the P1-owned `user_words` fields with no P2/P4
  side effects.
- `VOC-026-D05`: Home and Progress use the real saved-words API for saved-word
  content and retain every other P4/P2 mock field explicitly labelled as
  mock-pending.

No P2–P4 API route, table, or behavior is invented.

## Inventory

| File                                                                 | VOC source        | Mock identifier                         | Disposition                   | Follow-up / note                                                                                                                                                                                          |
| -------------------------------------------------------------------- | ----------------- | --------------------------------------- | ----------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `apps/web/src/app/(app)/home/page.tsx`                               | VOC-019           | `MOCK_HOME_STATE`                       | **Retain as mock-pending-P4** | Saved-word list is wired to the real `GET /api/v1/user-words` API. Mission target, reviewed-today, streak, and due-review fields have no P1 equivalent and are explicitly labelled as P4-pending mocks.   |
| `apps/web/src/app/(app)/progress/page.tsx`                           | VOC-020           | `MOCK_PROGRESS_STATE`                   | **Retain as mock-pending-P4** | Saved-vocabulary list is wired to the real `GET /api/v1/user-words` API. Confidence Points, streaks, and weekly completion history have no P1 equivalent and are explicitly labelled as P4-pending mocks. |
| `apps/web/src/app/(app)/discover/page.tsx`                           | VOC-021           | `MOCK_DISCOVER_SITUATIONS`              | **Decommissioned → real P1**  | Replaced by `GET /api/v1/journey-situations`. The mock constant is removed.                                                                                                                               |
| `apps/web/src/app/(app)/discover/[situation]/page.tsx`               | VOC-022           | `MOCK_SITUATION_WORD_LISTS` (imported)  | **Decommissioned → real P1**  | Replaced by `GET /api/v1/journey-situations/{slug}`. The mock import is removed.                                                                                                                          |
| `apps/web/src/app/(app)/discover/[situation]/[word]/page.tsx`        | VOC-022 / VOC-024 | `MOCK_SITUATION_WORD_LISTS` (imported)  | **Decommissioned → real P1**  | Replaced by `GET /api/v1/canonical-words/{wordSlug}` with real save/unsave control. No mock import remains.                                                                                               |
| `apps/web/src/app/(app)/discover/[situation]/_lib/mock-word-data.ts` | VOC-022 / VOC-024 | `MOCK_SITUATION_WORD_LISTS`, `mockWord` | **Decommissioned → real P1**  | Mock source file removed; canonical content now lives in `canonical_words`, `word_meanings`, `word_examples`, and `usage_notes` loaded by the real content API.                                           |
| `apps/web/src/app/(app)/_components/app-header.tsx`                  | VOC-017           | n/a (real component)                    | **Wire to A1**                | Already uses the real `@vocanova/api-client` logout operation with CSRF double-submit. No mock data.                                                                                                      |
| `apps/web/src/app/(app)/_components/bottom-nav.tsx`                  | VOC-017           | n/a (real component)                    | **Wire to A1**                | Pure navigation shell, no data. Already protected by the A1 middleware.                                                                                                                                   |

## Disposition definitions

- **Decommissioned → real P1**: the mock constant/source was removed and the
  screen or file is now served by the real VOC-026 P1 API.
- **Retain as mock-pending-P4**: the local mock data is kept exactly as-is for
  fields that have no P1 equivalent, with an explicit comment identifying them as
  P4-pending. The real P1 saved-word content is wired in alongside them.
- **Wire to A1**: the component/route is already real code and depends on the A1
  auth layer. It is not a mock and is not covered by the decommission/retain
  rule.

## Verified boundaries

The deterministic check in `scripts/foundation/mock-inventory.mjs` enforces the
following boundaries:

- No `MOCK_*` constant remains in any `(app)` source other than the two
  explicitly retained `MOCK_HOME_STATE` and `MOCK_PROGRESS_STATE` constants.
- No decommissioned mock constant (`MOCK_DISCOVER_SITUATIONS`,
  `MOCK_SITUATION_WORD_LISTS`) appears anywhere in the `(app)` source tree.
- The only API routes under `apps/api/app/api/` are the A1 auth/current-user
  routes (`/api/v1/me`, `/api/v1/auth/*`) and the P1 content/learning routes
  (`/api/v1/journey-situations`, `/api/v1/journey-situations/{slug}`,
  `/api/v1/canonical-words/{wordSlug}`, `/api/v1/user-words`,
  `/api/v1/user-words/{meaningId}`).
- The only business modules under `apps/api/business/` are `auth` (A1),
  `content` (P1 reads), and `learning` (P1 owner-data writes).
- The only Ent schemas in `apps/api/ent/schema/` are the A1 identity/session
  schemas (`user`, `externalidentity`, `session`, `magiclink`) and the P1
  content/owner-data schemas (`canonicalword`, `wordmeaning`, `wordexample`,
  `usagenote`, `journeysituation`, `journeyword`, `userword`) plus shared
  mixins.
- The only committed migrations are the A1 identity/OAuth migrations and the
  VOC-026 P1 content/idempotency migrations.

No P2–P4 route, table, or behavior was invented.

## Follow-up work

- VOC-019 follow-up: replace the retained Home P4 mock fields (mission target,
  reviewed-today, streak, due-review count) with real P4 gamification/scheduler
  data when those milestones are implemented.
- VOC-020 follow-up: replace the retained Progress P4 mock fields (Confidence
  Points, streaks, completion history) with real P4 gamification/scheduler data
  when those milestones are implemented.
