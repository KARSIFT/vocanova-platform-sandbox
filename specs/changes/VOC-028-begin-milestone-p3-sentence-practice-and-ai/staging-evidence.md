# VOC-028 — P3 AI Feedback Staging Evidence

## Scope and authority

This document collects the in-repository staging evidence required by
`VOC-028-AC-09` and `VOC-028-T05`. It is produced under the **adopted**
resolutions `VOC-028-D00` through `VOC-028-D06` (2026-07-25). This document does
not declare the DOC-12 P3 milestone gate complete.

The in-repository evaluation fixtures, observability seams, and AI-disable/cost
seams are implemented and ready. Live staging exercises (EV-28, EV-29, EV-30)
and protected live-model provider evaluation are blocked by two remaining
dependencies:
- `VOC-028-DEP-02` (`D02`): the production provider/model is resolved (OpenCode Go
  / `opencode-go/deepseek-v4-pro` via `opencode serve`), but the **formal privacy
  verification** (training-data use, retention, processing regions, subprocessors,
  deletion) remains a pre-production legal-review gate and is not itself resolved by
  adoption (DOC-09 §18, §21, §24; DOC-12 §5 P3 gate).
- `VOC-028-DEP-04`: the F3 staging environment does not exist yet (carried from
  VOC-027-DEP-02).

The procedures below are documented and ready to run once both are available.
Until then, all CI-gated feedback paths use the mock provider only — CI never
depends on a paid provider (DOC-09 §23).

## In-repository evidence (ready now, mock-provider basis)

| Evidence | Requirement | Status | Location / note |
| -------- | ----------- | ------ | --------------- |
| `EV-00`/`EV-01` | `learner_sentences`/`ai_feedback_attempts` migration invariants and recovery | Ready | reviewed Atlas migrations + disposable forward/recovery rehearsal; DOC-05 §18 order after `review_attempts`; no P4 tables. |
| `EV-02`/`EV-03` | Narrow provider/moderation interfaces + mock provider; DTO contracts | Ready | `apps/api/business/aifeedback` interfaces/adapter + mock provider; no generic `Generate(any)`; internal/public DTOs enforce DOC-09 §§9,10. |
| `EV-04`..`EV-06` | Deterministic input validation + target-word matching + normalization | Ready | validation codes, no-synonym rule, original-display-text preservation (DOC-09 §6). |
| `EV-07` | Pending-row workflow; provider call never inside a transaction | Ready | DOC-05 §15 order asserted by a lifecycle test double/hook. |
| `EV-08` | Dedup, rate limiting, learning-vs-operational state separation | Ready | dedup key + `request_hash`; `AI_FEEDBACK_RATE_LIMITED`. |
| `EV-09` | Mission-completion stub (`D01`); no P4 tables; honest `missionCompleted` | Ready | `MissionUpdater` seam; mock-inventory asserts no P4 tables. |
| `EV-10`..`EV-12`, `EV-14` | Prompt architecture, structured-output validation, injection resistance, hidden refusals | Ready | layered + versioned prompts; one repair; injection/exploitation graded as text. |
| `EV-13`..`EV-16` | Safety outcomes, self-harm intervention, legitimate discussion allowed | Ready | five-outcome mapping via mock moderation provider; blocked/self-harm never complete. |
| `EV-17`..`EV-21` | API contract, CSRF/idempotency/ownership, states, component wiring, report action, OpenAPI/client drift | Ready | `/api/v1` endpoint + reusable component at the adopted `D05` entry points; matched `@vocanova/api-client`. |
| `EV-22`..`EV-27` | Evaluation fixtures, golden set, CI-pays-free, privacy-safe observability, AI-disable seam, change metadata | Ready | `apps/api/business/aifeedback/evaluation.go`: `initial-dataset-v1` (>=200 cases) + `golden-set-v1` (~50 cases) + `RunMockEvaluation`/`RunGoldenEvaluation`; `metrics.go`: `MetricsRecorder`/`MetricsEvent` grouped by prompt/schema/provider/model/release, no learner text; `gate.go`: `GenerationGate`/`MemoryGenerationGate` with emergency disable + global daily/monthly cost seams; non-AI features remain available when AI generation is disabled. |
| `EV-31` | Installed deterministic/security suite + extended mock-inventory | Ready | `pnpm validate`/`test`/`build`, Go format/vet/test/build, `scripts/governance/*` as applicable, extended `mock-inventory.mjs` (`evaluation.go`/`metrics.go`/`gate.go` present). |

## Staging exercise plan (blocked by F3 + `D02`)

Once `VOC-028-DEP-04` and `VOC-028-DEP-02` are resolved and a non-production F3
environment with the adopted provider/model is available, the following
exercises must be executed and their results recorded here or in PR evidence.

### EV-28 — Validate → feedback → persist → mission-stub → display

1. With a non-production learner identity, open the feedback component at the
   adopted entry point (per `D05`) with a target word.
2. Submit a valid sentence containing the target word and a CSRF token +
   `Idempotency-Key`.
3. Verify the pending state preserves input and disables duplicate submission.
4. Verify success returns the DOC-09 §9 result fields plus the honest
   `missionCompleted` (not-yet-wired) state, and persisted
   `learner_sentences` + `ai_feedback_attempts` rows.
5. Once formal privacy verification is complete and F3 staging is available,
   repeat with the real OpenCode Go provider under the protected dev/staging
   evaluation + cost ceiling and record latency/cost.

### EV-29 — Safety, cross-user, CSRF, idempotency, AI-disable

1. **Safety:** submit self-harm content (assert crisis-resource interruption, no
   completion) and blocked content (assert safe failure, no completion); submit
   legitimate difficult-subject discussion (assert it proceeds).
2. **Cross-user:** as learner A submit against learner B's attempt (assert 404;
   no cross-learner sentence/feedback inference).
3. **CSRF:** submit without `X-CSRF-Token` (assert 403, no persistence).
4. **Idempotency:** replay the same key/fingerprint (assert idempotent, no
   duplicate feedback); reuse the key with a changed fingerprint (assert 409).
5. **AI-disable:** toggle the disable seam (assert non-AI review/discover/save
   remain available and stored feedback stays readable).

### EV-30 — learner_sentences / ai_feedback_attempts rollback rehearsal

1. Record current `learner_sentences` and `ai_feedback_attempts` state for a
   test learner.
2. Apply the VOC-028 build and migration in staging.
3. Run several feedback flows, then perform a rollback to the previously
   known-good revision.
4. Verify that:
   - All committed `ai_feedback_attempts` rows created before the rollback
     remain immutable (no accidental cascade or migration deletion).
   - Committed `learner_sentences` content is preserved.
   - No daily-mission, streak, point, or P4 tables were created.
   - Health checks and the feedback endpoint remain functional (or degrade
     cleanly with AI disabled).

## Provider evaluation + privacy verification (blocked by `D02` formal review)

The adopted provider is the founder's OpenCode Go account, model
`opencode-go/deepseek-v4-pro`, integrated via `opencode serve` (DOC-09 §18).
Before production, verify the provider's training-data use, retention,
processing regions, subprocessors, and deletion procedures; prefer
configurations where API content isn't used for training and retention is
disabled/minimized (DOC-09 §21). Record the chosen provider/model/config, the
privacy review result, the secrets-provisioning plan (secrets stay backend-only),
the evaluation scores against the MVP acceptance thresholds (DOC-09 §23), and the
cost/ceiling plan. This section is a placeholder until the formal privacy review
and F3 staging are available.

## Rollback triggers

Per `VOC-028` implementation-plan §Deployment and rollback / release-plan
§Rollback, initiate rollback on (DOC-09 §25):

- Unsafe feedback reaching learners (hateful/sexual/threatening/demeaning, wrong
  self-harm handling, raw provider output shown).
- Suspected cross-user exposure of sentences or feedback.
- Prompt injection revealing protected information.
- Material increase in wrong corrections or a spike in learner reports.
- Schema failures exceeding threshold or unusable latency.
- Cost overrun or exceeding daily/monthly ceilings.
- Inconsistent mission state (should not occur under the `D01` stub, but any
  fabricated completion is a trigger/bug).
- Incorrect provider privacy configuration or a serious provider outage/breaking
  change.

## Rollback procedure

1. Preserve immutable `ai_feedback_attempts` history: never drop committed
   feedback rows.
2. Preserve committed `learner_sentences` content.
3. Keep stored feedback readable if AI generation is disabled (non-AI features
   stay available).
4. Revert the deployment to the last-known-good revision.
5. Validate with non-production identities.
6. Record the last-known-good revision and the rollback reason.

## Limitations / open dependencies

- `VOC-028-DEP-04`: F3 staging does not exist, so EV-28/EV-29/EV-30 cannot be
  run live. Procedures and in-repository evidence are complete; live execution
  is recorded as blocked.
- `VOC-028-DEP-02` (`D02`): the provider/model is resolved (OpenCode Go /
  `opencode-go/deepseek-v4-pro`), but the formal privacy/retention/subprocessor
  verification remains a pre-production legal-review gate. The production
  adapter is implemented, but protected live-model evaluation and privacy
  verification evidence are recorded as placeholders until that review and F3
  staging are available. CI uses the mock provider only.
- Accessibility automation for the feedback component is not yet implemented;
  recorded as a limitation, not a pass.

## Follow-up work

- Complete the formal privacy/retention review for the adopted OpenCode Go
  provider, then run protected offline live-model evaluation + the real-provider
  staging exercises once F3 staging is available.
- P4: replace the `D01` mission-completion stub with real
  `daily_mission_snapshots`/streak/point wiring once those tables exist.
- `D03`: set production AI-disable/cost-ceiling activation values
  (founder-controlled) before general availability.
- `D04`: complete formal legal review of retention defaults before production.
- DOC-09 §22: account-deletion/anonymization of AI content is owned by the
  future account-deletion work.