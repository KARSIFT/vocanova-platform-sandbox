# VOC-028 — P3 AI Feedback Mock Disposition Inventory

## Scope and authority

This document is the T05 inventory required by `VOC-028-D05`/`VOC-028-D06`:
every mock/placeholder source touched by P3 is listed with its disposition.
This inventory is produced under the **adopted** resolutions `D00`–`D06`
(2026-07-25). The key boundaries are:

- `VOC-028-D00`: P3 carries forward the A1/P1/P2 foundation; no
  A1/P1/P2 mechanic is re-litigated. P3 adds `learner_sentences` +
  `ai_feedback_attempts` only (no P4 tables; the mission tables are confirmed
  absent from `apps/api/ent/schema/` and `apps/api/migrations/`).
- `VOC-028-D01`: the mission-completion step is a **stub/interface point flagged
  for P4**; no `daily_mission_snapshots`/`daily_activity_summaries`/
  `streak_states`/`confidence_point_ledger`/`grace_day_ledger` table is invented.
- `VOC-028-D02`: production provider resolved at adoption to the founder's
  OpenCode Go account, model `opencode-go/deepseek-v4-pro`, integrated via
  `opencode serve`. The production adapter is implemented against the narrow T00
  `FeedbackProvider` interface; backend-only secrets are read from configuration
  and are never hard-coded. Formal privacy verification (training-data use,
  retention, processing regions, subprocessors, deletion) remains a
  pre-production gate per `D04`/`DOC-09 §21`.
- `VOC-028-D03`: AI-disable and cost-ceiling seams implemented with
  DOC-09-documented starting defaults; activation values are founder-controlled
  at runtime.
- `VOC-028-D05`: the reusable feedback component is wired into the Home /
  Word-Detail / Review-Completion entry points per the adopted placement.

**No P4 API route, table, or behavior is invented.** CI never depends on a paid
provider; all deterministic/integration tests use the mock provider (DOC-09
§23).

## Inventory

| File / area                                              | VOC source        | Mock identifier / field                                   | Disposition                                | Follow-up / note                                                                                                                                                       |
| -------------------------------------------------------- | ----------------- | --------------------------------------------------------- | ------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `apps/web/src/app/(app)/home/page.tsx`                   | VOC-019 / VOC-026 | `MOCK_HOME_STATE` (remaining P4 fields, e.g. mission target, reviewed-today, streak) | **Retain as mock-pending-P4** | P3 does not touch the retained P4 mocks; only the feedback entry-point affordance (per `D05`) is added. No mission/streak data is invented.                            |
| `apps/web/src/app/(app)/progress/page.tsx`               | VOC-020 / VOC-026 | `MOCK_PROGRESS_STATE` (Confidence Points, streaks, completion history) | **Retain as mock-pending-P4** | P3 touches none of the Progress P4 mocks. Explicitly labelled; not presented as real P3 data.                                                                          |
| `apps/web/src/app/(app)/discover/[situation]/[word]/*`   | VOC-024 / VOC-026 | n/a (new practice entry)                                  | **New real P3** (per `D05`)               | Word-Detail placement of the reusable feedback component, wired to the real `/api/v1` feedback endpoint under the A1 session. No mock import.                          |
| `apps/web/src/app/(app)/home/*`                          | VOC-028 (new)     | n/a (new practice entry, per `D05`)                       | **New real P3** (per `D05`)               | Home entry point to the reusable feedback component, if adopted per `D05`. No mission/streak affordance.                                                              |
| `apps/web/src/app/(app)/reviews/*`                       | VOC-027 / VOC-028 | n/a (Review-Completion entry, per `D05`)                  | **New real P3** (per `D05`)               | Review-Completion entry point to the reusable feedback component, wired to the real feedback endpoint. No regression to the P2 review flow.                            |
| `apps/api/business/aifeedback` (new module)              | VOC-028 (new)     | n/a                                                       | **New real P3**                            | Narrow `FeedbackProvider`/`ModerationProvider` interfaces, mock provider, orchestration service, prompt package, validation, safety. No generic `Generate(any)`.     |
| `apps/api/ent/schema/learnersentence.go` (new schema)    | VOC-028 (new)     | n/a                                                       | **New real P3**                            | `learner_sentences` table per DOC-05 §11. No P4 table.                                                                                                                |
| `apps/api/ent/schema/aifeedbackattempt.go` (new schema)  | VOC-028 (new)     | n/a                                                       | **New real P3**                            | `ai_feedback_attempts` table per DOC-05 §11; immutable history.                                                                                                       |
| mock feedback provider (`aifeedback` adapter)            | VOC-028 (new)     | `MockProvider`                                            | **Retain as mock-in-CI**                  | Deterministic mock provider used in CI; never a paid provider. Production adapter (`OpenCodeFeedbackProvider`) is implemented against the narrow T00 interface.                                                                     |
| production feedback provider (`aifeedback` adapter)      | VOC-028 (new)     | `OpenCodeFeedbackProvider` / `opencode-go/deepseek-v4-pro` | **New real P3** (configured behind `D02`) | Persistent HTTP client to `opencode serve`; backend-only API key from configuration; no credential in source. Privacy verification remains a pre-production gate.                                                                    |
| evaluation layer (`aifeedback` evaluation)               | VOC-028 (new)     | n/a                                                       | **New real P3**                            | `initial-dataset-v1` (>=200 synthetic cases) + `golden-set-v1` (~50 cases) + `RunMockEvaluation`/`RunGoldenEvaluation`. Protected live-model evaluation runs outside CI only after privacy/staging gates.                             |
| privacy-safe observability (`aifeedback` metrics)        | VOC-028 (new)     | `MetricsRecorder` / `MetricsEvent`                          | **New real P3**                            | Metrics grouped by prompt/schema/provider/model/release; no learner text or user identity in labels.                                                                                                                                 |
| AI-disable / cost-ceiling seam (`aifeedback` gate)     | VOC-028 (new)     | `GenerationGate` / `MemoryGenerationGate`                 | **New real P3**                            | Emergency disable switch + global daily/monthly cost seams; non-AI learning features remain available when generation is disabled.                                                                                                  |
| Sentence-History insight screen                         | DOC-09 §26        | (not built)                                               | **Excluded — post-MVP**                    | Explicitly a post-MVP opportunity (DOC-09 §26); P3 does not build it. No mock retained for it.                                                                                                                                       |

## Disposition definitions

- **New real P3**: created by this package and served by the real P3 endpoint
  / persistence / orchestration (the mock provider is used in CI, not shown as
  a feature).
- **Retain as mock-in-CI**: the deterministic mock provider is the CI provider
  and is never a paid provider call (DOC-09 §23); it is not a learner-facing
  mock.
- **Retain as mock-pending-P4**: carried from earlier milestones, untouched by
  P3, explicitly labelled P4-pending.
- **Excluded — post-MVP**: explicitly out of scope per DOC-09 §26; not built,
  not mocked.

## Verified boundaries (enforced by an extended `scripts/foundation/mock-inventory.mjs`)

- The only business modules under `apps/api/business/` are `auth` (A1),
  `content` (P1 reads), `learning` (P1 owner-data writes), `reviews` (P2), and
  the **new `aifeedback` (P3)** module.
- The only Ent schemas in `apps/api/ent/schema/` are the A1 identity/session
  schemas, the P1 content/owner-data schemas, the P2 `reviewattempt` schema, and
  the **new P3 `learnersentence` + `aifeedbackattempt` schemas** plus shared
  mixins. No `userword` change occurs.
- The only committed migrations are the A1 identity/OAuth migrations, the
  VOC-026 P1 migrations, the VOC-027 P2 `review_attempts` migration, and the
  **new VOC-028 P3 `learner_sentences` + `ai_feedback_attempts` migrations**.
- No `daily_mission_snapshots` / `daily_activity_summaries` / `streak_states` /
  `confidence_point_ledger` / `grace_day_ledger` table, route, or behavior is
  created (P4 out of scope); the mission-completion step is the `D01` stub.
- The P3 `aifeedback` module contains the T05 deliverables: `evaluation.go`,
  `metrics.go`, and `gate.go`.
- No concrete commercial provider/model other than the adopted OpenCode Go
  provider is selected or hard-configured; no real provider credential is
  committed to source (`D02`); CI never depends on a paid provider.

No P4 route, table, or behavior was invented; no A1/P1/P2 mechanic was
revisited; no Sentence-History screen was built.

## Follow-up work

- P4 follow-up: implement the real mission-completion wiring against
  `daily_mission_snapshots`/streak/point tables when P4 creates them, replacing
  the `D01` stub.
- `D02`/`D04` follow-up: complete the formal privacy verification (training-data
  use, retention, processing regions, subprocessors, deletion) and record the
  chosen provider config before production launch.
- `D03` follow-up: set production AI-disable/cost-ceiling activation values
  (founder-controlled).
- Post-MVP (DOC-09 §26): the Sentence-History insight screen and other
  opportunities, each requiring separate approval.