# VOC-027 — P2 Review Mock Disposition Inventory

## Scope and authority

This document is the T04 inventory required by `VOC-027-D05`/`VOC-027-D06`:
every mock/placeholder source touched by P2 is listed with its disposition. The
adopted resolutions (recorded at adoption) are:

- `VOC-027-D01`: P1's `next_review_at = null` save behavior already satisfies
  the DOC-05 §9 due-word rule; P1 is not revisited and P2 does not set
  `next_review_at = now` on word-add.
- `VOC-027-D02`: the prompt-type contract contradiction (DOC-05 §9
  `multiple_choice`/`self_check` vs. DOC-07 `meaning_choice`/`word_choice`/
  `self_check`) is reconciled at adoption; the adopted `prompt_type` enum
  governs the migration/DTO/tests, and DOC-07 is updated under the DOC-12 §11
  change-control rule.
- `VOC-027-D03`: which prompt type(s) P2 builds first.
- `VOC-027-D04`: the review-session UX/flow and completion/empty states.
- `VOC-027-D05`: Home's `dueReviewWords` is either wired to the real due-count
  or retained mock-pending-P2; every other Home/Progress mock field stays
  mock-pending-P4.

**No P3/P4 API route, table, or behavior is invented.**

## Inventory

| File                                                 | VOC source        | Mock identifier / field                            | Disposition                       | Follow-up / note                                                                                                                                                       |
| ---------------------------------------------------- | ----------------- | -------------------------------------------------- | --------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `apps/web/src/app/(app)/home/page.tsx`               | VOC-019 / VOC-026 | `MOCK_HOME_STATE.dueReviewWords`                   | **Decommissioned → real P2** (option (a) adopted) | Per `VOC-027-D05`. Wired to `GET /api/v1/reviews/due` `totalCount`. Removed from `MOCK_HOME_STATE`; the mission target, reviewed-today count, and streak stay mock-pending-P4. |
| `apps/web/src/app/(app)/home/page.tsx`               | VOC-019 / VOC-026 | `MOCK_HOME_STATE` (mission target, reviewed-today, streak) | **Retain as mock-pending-P4** | No P2 equivalent; mission/reviewed-today/streak are P4 gamification/scheduler aggregates. Explicitly labelled; not presented as real P2 data.                          |
| `apps/web/src/app/(app)/progress/page.tsx`           | VOC-020 / VOC-026 | `MOCK_PROGRESS_STATE` (Confidence Points, streaks, completion history) | **Retain as mock-pending-P4** | No P2 equivalent; Confidence Points and streaks are P4. Explicitly labelled; not presented as real P2 data.                                                            |
| `apps/web/src/app/(app)/reviews/*`                   | VOC-027 (new)     | n/a (new route)                                    | **New real P2**                   | New review route wired to `GET /api/v1/reviews/due` + `POST /api/v1/reviews/submissions` under the A1 session. No mock import.                                          |

## Disposition definitions

- **Decommission → real P2**: the mock field was replaced by the real P2
  due-queue/submission API (per `D05` option (a)).
- **Retain as mock-pending-P2**: the field stays mock at P2's boundary with an
  explicit "P2-pending" label, to be replaced at a follow-up (per `D05` option
  (b)).
- **Retain as mock-pending-P4**: the field has no P2 equivalent and stays
  exactly as VOC-026 left it, explicitly labelled P4-pending.
- **New real P2**: the route is newly created by this package and served by the
  real P2 API.

## Verified boundaries (enforced by `scripts/foundation/mock-inventory.mjs`)

- The only API routes under `apps/api/app/api/` are the A1 auth/current-user
  routes, the VOC-026 P1 content/learning routes, and the **new VOC-027 P2
  review routes** (`/api/v1/reviews/due`, `/api/v1/reviews/submissions`).
- The only business modules under `apps/api/business/` are `auth` (A1),
  `content` (P1 reads), `learning` (P1 owner-data writes), and **`reviews`
  (P2 scheduling + submission)**.
- The only Ent schemas in `apps/api/ent/schema/` are the A1 identity/session
  schemas, the P1 content/owner-data schemas, and the **new P2
  `reviewattempt` schema** plus shared mixins. The `userword` schema is
  unchanged from VOC-026.
- The only committed migrations are the A1 identity/OAuth migrations, the
  VOC-026 P1 migrations, and the **VOC-027 P2 `review_attempts` migration**.
- No `daily_mission_snapshots` / `confidence_point_ledger` / `streak_states` /
  `daily_activity_summaries` table, route, or behavior is created (P4 out of
  scope).

No P3/P4 route, table, or behavior was invented; no P1 save behavior was
revisited.

## Follow-up work

- VOC-019 follow-up (P4): replace the retained Home P4 mock fields (mission
  target, reviewed-today count, streak) with real P4 gamification/scheduler
  data when those milestones are implemented. (carried from VOC-026)
- VOC-020 follow-up (P4): replace the retained Progress P4 mock fields with real
  P4 data when P4 is implemented. (carried from VOC-026)
- VOC-027 follow-up (P3/P4): the `typing`/`sentence_usage` prompt types are a
  later superset (DOC-05 §9) and remain out of scope; the submission
  transaction's daily-mission / point-ledger / streak steps are added by P4 when
  those tables exist.