---
id: DOC-05
title: VocaNova Database Design
version: 1.0
document_type: database-design
status: approved
owner: founder
canonical_path: docs/engineering/05-database-design.md
approved_at: 2026-07-21
last_reviewed_at: 2026-07-21
review_cycle: quarterly
supersedes: null
related_documents:
  - DOC-04
  - DOC-06
  - DOC-07
  - DOC-09
related_decisions: []
adoption_change: VOC-008
source_files:
  - path: 05-database-design.md
    sha256: cc2efd5b6356f41bfc9075bd58297b301e6274a708943c16369600e6f0d5d1c9
---
# 05 — VocaNova Database Design

## 1. Direction

PostgreSQL + Ent ORM + Atlas migrations, one database, one modular-monolith Go backend. UUID
identifiers (UUIDv7 where practical), UTC timestamps in storage, IANA timezones for daily learning
logic. Ent schemas define the persistence model; Atlas handles migrations. Production schema changes
never run automatically at API startup.

## 2. Core principles

- **Normalized core, JSONB only where genuinely flexible** (AI feedback content, audit metadata,
  idempotency response bodies, optional operational metadata).
- **Canonical content vs. user state are separate tables.** Canonical: `canonical_words`,
  `word_meanings`, `word_examples`, `usage_notes`. User state: `user_words`, `review_attempts`,
  `learner_sentences`, `daily_mission_snapshots`.
- **Current state vs. immutable history are separate tables** (e.g. `user_words` current vs.
  `review_attempts` history; `streak_states` current vs. `daily_mission_snapshots` history;
  `confidence_point_ledger` history).
- **The database enforces integrity as the final layer**: PKs, FKs, unique constraints (including
  partial unique indexes), check constraints, not-null.
- **Transactions protect multi-step flows**: review submission, word addition, daily mission
  completion, point award, streak update, grace-day application, account anonymization.
- **No premature event sourcing** — pragmatic history tables and audit logs, not a full event store.

## 3. Naming conventions

Plural snake_case tables, snake_case columns. Every main table: `id uuid primary key` (backend-
generated, prefer UUIDv7). Most tables: `created_at timestamptz not null`, `updated_at timestamptz
not null`, all UTC. Soft deletion (`deleted_at timestamptz null`) only where useful, not everywhere.
Enums are plain `text` columns with `check` constraints, not native PostgreSQL enum types (easier to
evolve during early iteration, still strongly validated).

## 4. Timezone strategy

Current timezone lives in `user_settings.timezone` (IANA string). Historical daily records
(`daily_mission_snapshots.timezone`, `daily_activity_summaries.timezone`) copy the *effective*
timezone at creation time, so a later timezone change doesn't rewrite what timezone was actually used
on a past day. All timestamps stay `timestamptz`/UTC; daily logic uses a separate `local_date date` +
`timezone text` pair.

## 5. Table overview

```text
users, external_identities, user_onboarding_profiles, user_settings
canonical_words, word_meanings, word_examples, usage_notes
journey_situations, journey_words
user_words, review_attempts
daily_mission_snapshots, daily_activity_summaries
learner_sentences, ai_feedback_attempts
confidence_point_ledger, streak_states, grace_day_ledger
idempotency_keys, feature_audit_logs
```

## 6. Identity and account tables

### `users`
`id uuid pk`, `email text null`, `display_name text null`, `avatar_url text null`,
`status text not null` (`active`/`disabled`/`deleted`), `onboarding_status text not null`
(`not_started`/`in_progress`/`completed`), `email_verified_at`, `last_login_at`, `deleted_at`,
`created_at`, `updated_at`. Unique partial index on `lower(email) where email is not null and
deleted_at is null`. `email` is nullable because the external identity provider is the stronger
identity signal.

### `external_identities`
Links a user to `google` or `email` provider. `id`, `user_id → users`, `provider text not null check
in ('google','email')`, `provider_subject text not null`, `provider_email`,
`provider_email_verified boolean default false`, `deleted_at`, timestamps. Unique on
`(provider, provider_subject) where deleted_at is null`. `provider_subject` is anonymized on account
deletion.

### `user_onboarding_profiles`
One row per user. `english_level` (`a1`/`a2`/`b1`/`b2`/`unknown`), `native_language`,
`learning_goal` (`general`/`work`/`travel`/`study`/`conversation`/`exam`), `main_use_case`
(`daily_life`/`work`/`travel`/`study`/`social`), `daily_review_target integer` (check 5–100),
`completed_at`.

### `user_settings`
One row per user. `timezone text default 'UTC'`, `daily_review_target integer default 20` (check
5–100), `review_interval_preset text default 'vocanova_default'` (`vocanova_default`/`wordup_like`/
`custom`), `notifications_enabled`, `marketing_emails_enabled`, `app_language default 'en'`. No
detailed custom-interval tables in MVP v1 — presets only, add detail storage later if the UI needs
it.

## 7. Canonical vocabulary content

Platform-owned, not user-specific.

### `canonical_words`
`text`, `normalized_text`, `word_type` (`word`/`phrase`/`phrasal_verb`/`idiom`/`collocation`),
`language_code default 'en'`, `status` (`draft`/`active`/`archived`), `difficulty_level`
(`a1`..`c1`/`unknown`, nullable), `frequency_rank`. Unique on `(language_code, normalized_text)`.

### `word_meanings`
The core learning unit is a **meaning**, not just a word spelling (e.g. "book" noun vs. "book" verb
are different meanings). `word_id → canonical_words`, `part_of_speech` (14 values: noun, verb,
adjective, adverb, preposition, conjunction, interjection, pronoun, determiner, phrase, idiom,
phrasal_verb, collocation, other), `short_definition`, `learner_definition`, `meaning_order integer`,
`status`, `difficulty_level`. Unique on `(word_id, meaning_order)`.

### `word_examples`
Per-meaning example sentences. `meaning_id → word_meanings`, `example_text`, `example_order`,
`difficulty_level`, `situation_label`, `status`. Unique on `(meaning_id, example_order)`. No
translation tables in MVP.

### `usage_notes`
Per-meaning usage/grammar/collocation/mistake/register/pronunciation notes. `note_type` (6 values
above), `note_text`, `note_order`, `status`. Unique on `(meaning_id, note_order)`.

## 8. Journey situations and discovery

Discovery is organized around real-life situations (Airport, Restaurant, Job Interview, etc.), each
containing many word meanings; a meaning can appear in many situations (many-to-many).

### `journey_situations`
`slug` (unique), `title`, `short_description`, `level_band` (`a1_a2`/`a2_b1`/`b1_b2`/`mixed`,
nullable), `category` (`daily_life`/`travel`/`work`/`study`/`social`), `status`, `display_order`.

### `journey_words`
Join table. `journey_situation_id`, `meaning_id`, `relevance_score integer default 50` (1–100),
`display_order` (nullable), `is_core boolean default false`. Unique on
`(journey_situation_id, meaning_id)`.

**Discovery query pattern:** active word meanings for a situation, excluding meanings already in the
learner's `user_words`, ordered by `is_core desc, display_order, relevance_score desc`.

## 9. User learning state

### `user_words`
One row per user × saved meaning. `status` (`new`/`learning`/`reviewing`/`mastered`/`ignored`/
`archived`), `source` (`journey`/`search`/`ai_suggestion`/`manual`/`seed`), `review_step integer
default 0` **(check between 0 and 7)**, `next_review_at`, `last_reviewed_at`, `last_result`
(`correct`/`incorrect`/`skipped`, nullable), `last_rating`
(`again`/`hard`/`good`/`easy`, nullable), `consecutive_correct_count`,
`consecutive_incorrect_count`, `total_review_count`, `correct_review_count` (check `<=`
`total_review_count`), `added_at`, `mastered_at`, `ignored_at`, `deleted_at`. Unique on
`(user_id, meaning_id) where deleted_at is null`.

**Review-step rule** (see
[the migration notes](../archive/README-migration-notes.md#2-review-rating-and-scheduling-conflict)
for why this table was selected over other drafts):
`result` records objective correctness while `rating` records the scheduling choice. For objective
prompts, an incorrect answer records `Again`; a correct answer permits Hard/Good/Easy. For
self-check prompts, the selected rating derives the result (`Again` is incorrect; Hard/Good/Easy
are correct). `Again` moves back one step with a floor of 0, and two consecutive
incorrect/`Again` attempts reset to 0; `Hard` stays on the current step; `Good` and `Easy` move
forward one step with a cap of 7. The backend owns the interval-to-step mapping.

A word is "due" when `status in ('new','learning','reviewing') and deleted_at is null and
(next_review_at is null or next_review_at <= now())`.

### `review_attempts`
Immutable review history, one row per submitted answer. `user_id`, `user_word_id`, `meaning_id`,
`attempt_type` (`review`/`practice`/`placement`/`mission`), `prompt_type`
(`multiple_choice`/`self_check`/`typing`/`sentence_usage` — MVP implements `multiple_choice` and
`self_check` first), `result` (`correct`/`incorrect`/`skipped`), `rating`
(`again`/`hard`/`good`/`easy`, nullable only when no rating applies),
`review_step_before/_after` (0–7),
`answered_at`, `response_time_ms`, `selected_option_meaning_id`, `typed_answer`, `was_hint_used`,
`source` (`daily_review`/`word_detail`/`journey_practice`/`manual_practice`), `client_attempt_id`,
`metadata jsonb`. Unique on `(user_id, client_attempt_id) where client_attempt_id is not null` — this
is the idempotency guard for double-submits.

**Review submission is one transaction:** lock `user_words` row → validate ownership → check
idempotency → insert `review_attempts` → update `user_words` (step, counters, last result,
next_review_at) → update daily mission progress → insert point ledger entry if earned → update
streak state if needed → commit.

## 10. Daily mission and daily activity

Deliberately two tables: `daily_mission_snapshots` (goal + completion state) vs.
`daily_activity_summaries` (actual counters) — these serve different query patterns and shouldn't be
merged.

### `daily_mission_snapshots`
One row per user per local date. `local_date date`, `timezone text` (copied from settings at
creation), `review_target integer` (5–100), `reviews_completed integer` (`<=` `review_target`),
optional `new_word_target`/`new_words_completed`, optional
`sentence_practice_target`/`sentence_practices_completed`, `policy_version`,
`status` (`open`/`completed`/`missed`/`protected`), `completed_at` (required when
`status='completed'`), `grace_applied boolean`, `grace_day_id`. Unique on `(user_id, local_date)`.
Optional goals are bonus goals and do not block core mission or streak completion unless a later,
versioned policy explicitly changes that rule.

### `daily_activity_summaries`
One row per user per local date. Counters: `reviews_attempted/_correct/_skipped`,
`words_discovered/_added`, `sentences_submitted`, `ai_feedback_received`,
`confidence_points_earned/_spent`. Unique on `(user_id, local_date)`. Detailed source of truth
remains `review_attempts`, `learner_sentences`, `ai_feedback_attempts`, `confidence_point_ledger` —
this table is a fast aggregate for Home/streak/stats UI, not the record of truth.

## 11. Learner sentences and AI feedback

Two tables, deliberately separate: `learner_sentences` (user-written content) vs.
`ai_feedback_attempts` (attempts to generate feedback on that content).

### `learner_sentences`
`meaning_id` (nullable — future free-writing may not target a specific meaning), `user_word_id`
(nullable), `sentence_text`, `normalized_sentence_text`, `source`
(`word_detail`/`review`/`daily_mission`/`free_practice`), `status`
(`submitted`/`feedback_ready`/`feedback_failed`/`archived`), `submitted_at`, `deleted_at`. Check:
`char_length(sentence_text) <= 1000` at the DB layer (the API-level limit is stricter — 300
characters — see [07](07-api-contract-and-dto-design.md) and [09](09-ai-features.md) §6).

### `ai_feedback_attempts`
`learner_sentence_id`, `status` (`pending`/`succeeded`/`failed`/`cancelled`), `provider`, `model`,
`prompt_version`, `request_hash`, `feedback_json jsonb`, `feedback_text`, `error_code`,
`error_message`, `started_at`, `completed_at` (required when `status='succeeded'`; `error_code`
required when `status='failed'`).

**Feedback status model** (see
[the migration notes](../archive/README-migration-notes.md#1-ai-feedback-label-conflict)): the
attempt status is operational (`pending`/`succeeded`/`failed`/`cancelled`). The public processing
status maps to `pending`/`completed`/`failed`/`skipped`; only a completed response carries the
learning result `correct`/`needs_improvement`/`incorrect`. These layers are defined precisely in
[09](09-ai-features.md) §7–10 and must not be restated with different
wording anywhere else (earlier drafts used "Good/Almost/Needs work" or "Great/Almost/Try again";
those are retired).

**AI feedback workflow (never hold a DB transaction during the external AI call):** insert
`learner_sentences` → insert `ai_feedback_attempts` (pending) → commit → call AI provider outside
the transaction → update attempt status → update sentence status → update daily activity summary if
succeeded.

## 12. Confidence Points, streaks, and grace days

### `confidence_point_ledger`
**Append-only.** Never just a balance field. `amount integer` (nonzero, may be negative),
`balance_after`, `reason` (`review_correct`/`daily_mission_completed`/`sentence_submitted`/
`ai_feedback_received`/`streak_bonus`/`admin_adjustment`), `source_type`
(`review_attempt`/`daily_mission`/`learner_sentence`/`ai_feedback_attempt`/`streak`/`admin`),
`source_id`, `idempotency_key`, `metadata jsonb`, `occurred_at`. Unique on
`(user_id, idempotency_key) where idempotency_key is not null`. Exact reward amounts belong in
backend product configuration, not database constraints (see [06](06-backend-design.md) §11 for the
actual point values).

**Confidence Points read trace (audited VOC-1178, 2026-09-05):** the Progress screen's
displayed total is the ledger's own running `balance_after`, never a separately maintained
mutable balance -
`confidence_point_ledger.balance_after` (written once per row, at insert time, as
`currentBalance + outcome.Amount` inside the caller's existing transaction — see
`gamification.Service.GrantPoint`) →
`gamification.Repository.GetLatestPointBalance` (`SELECT balance_after ... ORDER BY
occurred_at DESC, id DESC LIMIT 1`) → `gamification.Service.CurrentBalance` →
`missions.Service.GetProgressView` (assigns `ConfidencePointsBalance` directly, no
recomputation) → `api.progressViewToDTO` (`GET /api/v1/progress`, passthrough) →
`apps/web/src/app/(app)/progress/page.tsx` (renders `confidencePointsBalance` as-is). No
step in this chain sums, caches, or independently derives the total.
`gamification/service_test.go`'s `TestCurrentBalanceMatchesSumOfLedgerEntries` grants a
sequence of ledger entries and asserts the value this trace ends in equals their exact sum.
One unrelated, pre-existing mutable counter was found and left alone because it sits outside
this read path entirely: `daily_activity_summaries.confidence_points_earned`/`_spent`
(`missions.Service.IncrementConfidencePointsEarned`) is written by some reward call sites but
is never read back into any API response — it's dead weight, not a balance the Progress
screen (or anything else) can drift from.

### `streak_states`
One row per user (unique). `current_streak_count`, `longest_streak_count` (check `>=` current),
`last_completed_local_date`, `last_activity_local_date`, `timezone`, `status`
(`active`/`at_risk`/`broken`). Backend calculates transitions from `daily_mission_snapshots`.

### `grace_day_ledger`
Append-only. `amount` (nonzero), `balance_after`, `reason`
(`earned_by_streak`/`manual_grant`/`used_for_missed_day`/`expired`/`admin_adjustment`),
`source_type` (`daily_mission`/`streak`/`admin`), `applied_to_local_date`, `timezone`,
`idempotency_key`. Same idempotency pattern as the point ledger.

## 13. Idempotency and audit records

### `idempotency_keys`
`user_id`, `key text`, `scope text` (`review_submission`/`daily_mission_completion`/`word_addition`/
`sentence_submission`/`ai_feedback_request`/`point_award`/`grace_day_application`),
`request_hash`, `response_status`, `response_body jsonb`, `status`
(`processing`/`completed`/`failed`/`expired`), `locked_until`, `expires_at`. Unique on
`(user_id, scope, key)`. Idempotency is deliberately scoped to the authenticated user: two users
may safely send the same key without sharing a response or blocking one another. Expired records
are safe to hard-delete.

### `feature_audit_logs`
Lightweight operational audit trail (not event sourcing). `action`, `entity_type`, `entity_id`,
`request_id`, `actor_type` (`user`/`system`/`admin`/`ai`), `actor_id`, `metadata jsonb`.

## 14. Relationships (summary)

```text
users 1→many external_identities, 1→1 user_onboarding_profiles, 1→1 user_settings
canonical_words 1→many word_meanings 1→many word_examples/usage_notes
journey_situations many↔many word_meanings (through journey_words)
users 1→many user_words 1→many review_attempts
users 1→many daily_mission_snapshots/daily_activity_summaries
users 1→many learner_sentences 1→many ai_feedback_attempts
users 1→many confidence_point_ledger, 1→1 streak_states, 1→many grace_day_ledger
users 1→many idempotency_keys/feature_audit_logs
```

## 15. Transaction rules (summary)

Every multi-step flow below is one transaction, in this order: **add word** (validate user/meaning
active → insert/restore `user_words` → update daily counters → audit log); **submit review** (lock
row → validate → idempotency check → insert attempt → update schedule → update mission/activity →
point ledger → streak); **complete daily mission** (mission + activity + point ledger + streak +
grace ledger together); **AI feedback** (pending-row pattern, never holding a transaction across the
external call); **account deletion** (must be transactional or safely staged, always with an audit
log entry).

## 16. Deletion, anonymization, retention

- **Soft-delete pending purge**: `users`, `external_identities`, `user_words`, `learner_sentences`.
- **Status lifecycle instead of deletion**: `canonical_words`, `word_meanings`, `word_examples`,
  `usage_notes`, `journey_situations` (draft/active/archived).
- **Immutable during the active-account lifecycle**: `review_attempts`, `ai_feedback_attempts`,
  `confidence_point_ledger`, `grace_day_ledger`, `feature_audit_logs`; deletion processing must
  delete or irreversibly de-identify learner-linked content as described below.
- **Deletion-dependent**: `user_onboarding_profiles`, `user_settings`, `daily_mission_snapshots`,
  `daily_activity_summaries`, `streak_states`; retain only if irreversibly de-identified and
  unlinkable, otherwise delete.
- **Hard-delete eligible**: expired idempotency keys, non-production test data.

**Account deletion workflow:** immediately deactivate the account and revoke all sessions, then run
a staged, retryable, verified purge/anonymization job. Delete or irreversibly anonymize identifiers,
external identities, learner sentences, AI feedback, and user reports. Learning history and
aggregates may be retained only when de-identified and no longer linkable to the learner; otherwise
delete them. Record lifecycle audit events without retaining deleted learner content. The default
completion target is 30 days, subject to legal review before production.

No automatic `ON DELETE CASCADE` for core business tables — accidental cascades could destroy
learning history. Cascade only where explicitly justified.

## 17. Ent schema direction

One Ent schema file per table under `internal/ent/schema/`. Shared mixins: `IDMixin`, `TimeMixin`,
`SoftDeleteMixin` — no broad generic `StatusMixin` (each table's valid status values differ).
Application-generated UUIDs (prefer UUIDv7), generated before insert. Explicit `Table` annotation per
schema. Atlas migration SQL must add what Ent can't fully express: partial unique indexes,
expression indexes (`lower(email)`), check constraints, JSONB detail, conditional uniqueness.

## 18. Migration strategy

Ent + Atlas. Auto-migration allowed for local/test/temporary databases only. **Production always
uses Atlas versioned migrations with an explicit apply step — never automatic schema mutation at API
startup.** Safety gate is automated checks (Atlas lint, fresh-DB apply, existing-DB apply, Ent
compatibility check, constraint tests, transaction tests, seed rerun), not AI review — AI may help
generate/explain/fix, but GitHub Actions and automated tests are the actual gate. Use
expand-and-contract for potentially-breaking changes.

Migration order: extensions → users → external_identities → user_onboarding_profiles →
user_settings → canonical_words → word_meanings → word_examples → usage_notes →
journey_situations → journey_words → user_words → review_attempts → daily_mission_snapshots →
daily_activity_summaries → learner_sentences → ai_feedback_attempts → confidence_point_ledger →
streak_states → grace_day_ledger → idempotency_keys → feature_audit_logs → indexes/constraints →
seed data.

## 19. Seed data

Versioned, deterministic, reviewable, safe-to-rerun JSON seed files loaded by a Go seed command
inside one transaction, using fixed UUIDs for stable relationships. Covers initial journey
situations, canonical words/meanings/examples/usage-notes, and journey-word relationships.

## 20. Required database tests

Schema constraints, unique indexes, foreign keys, review-transition transactions, idempotency, daily
mission timezone handling, point ledger, streak/grace-day rules, AI feedback lifecycle, account
deletion/anonymization, seed rerun-safety. Critical cases carried from the source document (duplicate
active email rejected, duplicate provider identity rejected, duplicate normalized word rejected,
discovery excludes already-saved meanings, duplicate `client_attempt_id` doesn't duplicate an
attempt, two consecutive incorrect reviews reset `review_step` to 0, only one mission snapshot per
user per local date and it doesn't rewrite historical timezone on settings change, point/grace ledger
amounts can't be zero, account deletion immediately revokes access and completes a verified
purge/anonymization without retaining linkable learner content).

## 21. Final summary

Canonical content is separate from user learning state; word meanings (not bare words) are the core
learning unit; review and AI-feedback history are preserved; daily missions and streaks are
timezone-aware; Confidence Points and grace days are ledgers, not mutable balances; idempotency
guards duplicate operations; account anonymization is a first-class workflow; Ent + Atlas with
automated-test migration gating. Deliberately avoided: event sourcing, microservice databases,
premature custom scheduling tables, unnecessary JSONB, production auto-migration at startup.
