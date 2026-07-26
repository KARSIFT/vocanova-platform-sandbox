# VOC-028 — Acceptance Criteria

Acceptance criteria are observable, stable, security-aware, and bidirectionally
traceable to requirements (`D00`–`D06`), tasks (`T00`–`T05`), tests
(`VOC-028-TEST-*`), and evidence. `D01`–`D05` were resolved at adoption
(2026-07-25) and recorded in `D06`; the criteria below are written against the
adopted resolutions. Formal privacy/legal review and F3 staging remain
pre-production gates, not adoption blockers.

## VOC-028-AC-00 — AI domain persistence and migration integrity

- Requirement source: `VOC-028-D00`, DOC-05 §§11,15,16,18–20, DOC-09 §§7,9,10,20
- Tasks: `VOC-028-T00`
- Tests: `VOC-028-TEST-00`, `VOC-028-TEST-01`
- Evidence: `VOC-028-EV-00`, `VOC-028-EV-01`
- Result: pending

Ent/Atlas create `learner_sentences` (nullable `meaning_id`/`user_word_id`,
`sentence_text`, `normalized_sentence_text`, `source`, `status`, `submitted_at`,
`deleted_at`; check `char_length(sentence_text) <= 1000`) and
`ai_feedback_attempts` (`learner_sentence_id`, `status`
`pending`/`succeeded`/`failed`/`cancelled`, `provider`, `model`,
`prompt_version`, `request_hash`, `feedback_json jsonb`, `feedback_text`,
`error_code`, `error_message`, `started_at`, `completed_at` — `completed_at`
required when succeeded, `error_code` required when failed) with FKs to
`users`/`learner_sentences`/`word_meanings`, no `ON DELETE CASCADE`, and
immutable `ai_feedback_attempts` semantics. The DOC-05 §18 order is respected
(added after `review_attempts`; no P4 tables created). Empty-db migration and
disposable recovery rehearsal preserve integrity; production migration never
runs at API startup. No A1/P1/P2 schema is changed.

## VOC-028-AC-01 — Narrow provider/moderation interfaces and a mock provider

- Requirement source: `VOC-028-D00`, DOC-09 §§10,17,23
- Tasks: `VOC-028-T00`
- Tests: `VOC-028-TEST-02`, `VOC-028-TEST-03`
- Evidence: `VOC-028-EV-02`, `VOC-028-EV-03`
- Result: pending

The `aifeedback` module defines a **narrow** `FeedbackProvider` interface and a
separate **narrow** `ModerationProvider` interface — not a generic
`Generate(ctx, input any)` — with provider SDK types confined to an adapter
layer (DOC-09 §17). A mock provider returns deterministic, schema-valid
feedback and deterministic moderation outcomes, so orchestration and CI never
depend on a paid provider (DOC-09 §23). The internal provider schema (DOC-09 §10)
and the public `SentenceFeedbackResult` contract (DOC-09 §9) are defined as
explicit DTOs; the model never controls DB IDs, timestamps, mission completion,
ownership, or persistence state.

## VOC-028-AC-02 — Deterministic input validation and target-word matching

- Requirement source: `VOC-028-D00`, DOC-09 §6
- Tasks: `VOC-028-T01`
- Tests: `VOC-028-TEST-04`..`VOC-028-TEST-06`
- Evidence: `VOC-028-EV-04`..`VOC-028-EV-06`
- Result: pending

Validation enforces DOC-09 §6 exactly: ≥3 words, ≤300 characters, primarily
English, one meaningful sentence, includes the target word / an accepted
inflection (`work`→`works/worked/working`) / a configured phrase variant, and
belongs to an eligible attempt owned by the authenticated learner. Backend
normalizes (trim, collapse whitespace, Unicode-normalize) while preserving the
original display text. Validation codes `too_short`/`too_long`/`missing_target`/
`invalid_input`/`unsupported_language`/`attempt_not_eligible` are stable.
Validation failures never call the model and never complete a mission.
Target-word matching accepts capitalization and configured variants but does not
silently accept unrelated synonyms (`good`≠`better` unless configured).

## VOC-028-AC-03 — Orchestration service (mock only), pending-row workflow, and mission-completion stub

- Requirement source: `VOC-028-D00`, `VOC-028-D01`, DOC-09 §§8,17,19,20, DOC-05 §15
- Tasks: `VOC-028-T01`
- Tests: `VOC-028-TEST-07`..`VOC-028-TEST-09`
- Evidence: `VOC-028-EV-07`..`VOC-028-EV-09`
- Result: pending

The orchestration service runs the full DOC-09 §17 lifecycle against **only the
mock provider** (no real provider call), uses the DOC-05 §15 pending-row
workflow — insert `learner_sentences` → insert `ai_feedback_attempts` (pending) →
commit → call the provider **outside** the transaction → update attempt status →
update sentence status — and **never** holds a DB transaction across the
provider call. Operational attempt states are separated from the three learning
statuses (DOC-09 §7); only a `succeeded`/`completed` result carries
`correct`/`needs_improvement`/`incorrect`. Dedup (learner + attempt + target
word + normalized sentence + prompt version) prevents duplicate provider calls,
duplicate feedback, and double mission completion. Rate limits (one active
generation per learner, 5/min/learner, 30/day/learner — configurable) are
enforced with stable `AI_FEEDBACK_RATE_LIMITED`. Per the adopted `D01`, the
mission-completion step is a **stub/interface point flagged for P4**: a
`MissionUpdater` seam returns a backend-decided result after successful
persistence, writes no `daily_mission_snapshots`/streak/point rows (no P4 tables
invented), and surfaces `missionCompleted` honestly (not-yet-wired / false)
rather than fabricating completion. Blocked/self-harm/moderation-unavailable
outcomes never complete a mission.

## VOC-028-AC-04 — Prompt architecture, versioning, and structured-output validation with one repair

- Requirement source: `VOC-028-D00`, `VOC-028-D02`, DOC-09 §§9,10,14
- Tasks: `VOC-028-T02`
- Tests: `VOC-028-TEST-10`..`VOC-028-TEST-12`, `VOC-028-TEST-14`
- Evidence: `VOC-028-EV-10`..`VOC-028-EV-12`, `VOC-028-EV-14`
- Result: pending

Three backend-built prompt layers exist (system/developer/user-payload; DOC-09
§14); the user payload is serialized as data, never concatenated into
instruction text; the frontend never constructs prompts. Prompt and output-schema
versions are recorded per request (`sentence-feedback-v1` / `feedback-schema-v1`)
and material prompt changes create a new version; prompts live in
version-controlled code. Structured-output validation (DOC-09 §10) rejects
inconsistent combinations, invalid enums, empty required fields, excessive
lengths, off-target feedback, unexpected markup, contradicted explanations,
unsafe output, and leaked instructions; **one** constrained repair attempt is
permitted when budget allows. Injection resistance holds: embedded learner
instructions are graded as text, never followed, and must not reveal/change the
schema or perform unrelated tasks.

## VOC-028-AC-05 — Production provider adapter boundary (`D02` resolved: OpenCode Go / `opencode-go/deepseek-v4-pro`; privacy verification pre-production gate)

- Requirement source: `VOC-028-D00`, `VOC-028-D02`, `VOC-028-D06`, DOC-09 §§17,18,21,24, DOC-12 §5
- Tasks: `VOC-028-T02`
- Tests: `VOC-028-TEST-12`, `VOC-028-TEST-14`
- Evidence: `VOC-028-EV-12`, `VOC-028-EV-14`
- Result: pending — provider/model and adapter implemented; privacy verification remains a pre-production gate

A production adapter is implemented against the narrow T00 `FeedbackProvider`
interface, calling the adopted OpenCode Go provider via `opencode serve` with model
`opencode-go/deepseek-v4-pro` (`D02`). The provider/model is read from
configuration and backend-only secrets are never hard-coded or committed.
Timeouts/retries match DOC-09 §18 (provider request 8s, total backend target 10s;
at most one transport retry for a clearly transient failure and one repair
attempt — never both indefinitely; no retry for invalid input, blocked content,
auth failure, invalid credentials, persistent schema incompatibility, or learner
cancellation). One primary provider/model is operated at a time; no automatic
multi-provider fallback. The formal privacy verification (training-data use,
retention, processing regions, subprocessors, deletion) is a pre-production gate
per `D04` and is recorded as a limitation, not a pass.

## VOC-028-AC-06 — Safety and moderation outcomes with injection resistance

- Requirement source: `VOC-028-D00`, `VOC-028-D02`, DOC-09 §§14,15,20
- Tasks: `VOC-028-T03`
- Tests: `VOC-028-TEST-13`..`VOC-028-TEST-16`
- Evidence: `VOC-028-EV-13`..`VOC-028-EV-16`
- Result: pending

The DOC-09 §15 flow maps to `allowed`/`allowed_sensitive`/`blocked`/
`self_harm_intervention`/`moderation_unavailable` (never shown directly).
Blocked categories (credible threats, serious-harm instructions, weapon/dangerous-
substance instructions, sexual exploitation of minors, suicide/self-harm
encouragement, targeted hateful incitement, malicious off-topic, harassment-
intent personal data) are blocked; legitimate discussion of difficult subjects is
allowed. Clear self-harm content interrupts with a crisis-resource message and
never provides therapy/diagnosis/counselling. Provider refusals return a safe
temporary failure, preserve input, allow retry, and never show raw refusal text.
Injection resistance holds across the moderation + feedback path. Blocked /
self-harm / moderation-unavailable outcomes never complete a mission, and stable
public error codes hide internal failure categories (DOC-09 §20).

## VOC-028-AC-07 — API endpoint and frontend integration (Home / Word-Detail / Review-Completion)

- Requirement source: `VOC-028-D00`, `VOC-028-D05`, `VOC-028-D06`, DOC-09 §§3,5,9,16, DOC-07
- Tasks: `VOC-028-T04`
- Tests: `VOC-028-TEST-17`..`VOC-028-TEST-21`
- Evidence: `VOC-028-EV-17`..`VOC-028-EV-21`
- Result: pending — `D05` resolved at adoption; endpoint and component wiring implemented

The `/api/v1` sentence-feedback write endpoint uses `credentials: "include"`,
explicit DTOs (never Ent models), a stable operation ID, committed OpenAPI + a
matched `@vocanova/api-client` method, and requires `X-CSRF-Token` plus a
user+operation-scoped `Idempotency-Key`. The frontend never sends provider
prompts/model settings/authoritative vocabulary metadata. Replay is idempotent;
reused key + changed fingerprint → 409; cross-user key isolated; owner mismatch
→ 404; unauthenticated/disabled → 401. Pending preserves input and disables
duplicate submission; success returns the DOC-09 §9 result fields plus backend-
confirmed mission state; failure preserves input with the safe retryable
message and exposes no provider details. A reusable feedback component is wired
into the Home, Word-Detail, and Review-Completion entry points per the adopted
`D05`; saved/reviewed state stays consistent with `user_words` across
navigation; no client DB access or duplicated authorization. The DOC-09 §16
report action creates one quality-review record with the stated
states/classifications and does not change mission completion or replace the
result. No P4 behavior.

## VOC-028-AC-08 — Evaluation fixtures, golden regression set, and privacy-safe observability (CI never depends on a paid provider)

- Requirement source: `VOC-028-D00`, `VOC-028-D02`, `VOC-028-D03`, `VOC-028-D06`, DOC-09 §§19,20,23,25
- Tasks: `VOC-028-T05`
- Tests: `VOC-028-TEST-22`..`VOC-028-TEST-27`
- Evidence: `VOC-028-EV-22`..`VOC-028-EV-27`
- Result: pending — evaluation fixtures, observability, and AI-disable seam implemented; protected offline live-model evaluation blocked until formal privacy review + F3 staging

An initial dataset of ≥200 synthetic cases and a stable golden regression set
(~50 cases, `golden-set-v1`) exist and are versioned (`apps/api/business/aifeedback/evaluation.go`); no case is removed just
because the current model performs poorly. Every material AI change records
dataset/golden-set/prompt/schema versions, provider, model, config, commit,
scores, critical failures, latency, cost, and reviewer approval. **Normal CI
never depends on a paid provider**; protected offline live-model evaluation runs
outside CI under explicit cost limits only after the formal privacy review and
F3 staging are available. Observability metrics (`apps/api/business/aifeedback/metrics.go`)
are grouped by prompt version/schema version/provider/model/release and never
include learner text or user identity in metric labels; raw provider request/response is not
stored by default. The DOC-09 §23 MVP acceptance thresholds are the release-
blocking targets and the release-blocking critical failures are tracked. The
AI-disable seam + cost-ceiling knobs (`apps/api/business/aifeedback/gate.go`)
exist (`D03`) so non-AI learning features remain available if AI generation is
disabled; activation values are founder-controlled, not guessed.

## VOC-028-AC-09 — P3 evidence, staging, rollback, mock-inventory, and gate readiness

- Requirement source: `VOC-028-D00`, `VOC-028-D06`, DOC-12 §5 P3
- Tasks: `VOC-028-T00`..`VOC-028-T05`
- Tests: `VOC-028-TEST-28`..`VOC-028-TEST-30`
- Evidence: `VOC-028-EV-28`..`VOC-028-EV-30`
- Result: pending — in-repository evidence complete; live staging and protected provider evaluation blocked until F3 staging + formal privacy review

Applicable checks, validation/orchestration/provider-mock/safety/contract/
privacy/evaluation tests, exact-SHA reviews, and the extended deterministic
mock-inventory test pass; mock-inventory verifies no P4 route/table/behavior was
invented and no P3-only mock is presented as real. Staging tests for validate →
mock/real feedback → persist → mission-stub → display, cross-user denial, CSRF,
idempotency, safety outcomes, AI-disable, and the `learner_sentences`/
`ai_feedback_attempts` rollback rehearsal are documented and ready to run once
F3 staging and the formal privacy review are complete (`DEP-02`/`DEP-04`).
Protected provider evaluation and privacy verification evidence is a placeholder
pending that review. This enables — but does not itself declare — the DOC-12 P3
gate evaluation; the milestone gate is not satisfied by package merge or staging
deploy alone.