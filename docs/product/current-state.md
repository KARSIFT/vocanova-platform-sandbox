# Vocanova current product state

## Connected learner experience — merged, 5 October 2026

The owner requested a candid usability critique and authorized substantial changes.
The integrated `codex/vocanova-learner-experience` branch makes the daily action
clearer, exposes writing on Home, presents each Journey situation once and keeps
practice, writing and stories discoverable. Lesson, practice and story sessions
use a focused shell and reachable action rail. Choice questions now require
explicit Check after selection. Writing saves its selected meaning in place and
compares exact original and suggested sentences beside rewriting. Progress leads
with learning evidence, including two recent server records. The public preview
demonstrates real teaching imagery and explicitly authored feedback.

See [learner experience improvement](../design/learner-experience.md) for the
critique, visual direction, behavior boundaries and exact local acceptance.
The consolidated PR check record establishes hosted acceptance. Production
promotion remains separate. The accepted artwork phase is preserved.
The integrated change is in [PR #1487](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1487).
It merged as `36361f196ce26cc7a96e5e97501fdb61c30b1488` and passed staging
core-loop and maturity journeys. A late review identified a separate resume
presentation defect: correctly graded choice questions can reopen without their
selected radio. The narrow correction restores selection from matching server
feedback; listening practice receives an additive optional confirmed-choice ID
derived from its saved snapshot. No new learning mutation or scoring rule is introduced.

## Visual learning phase — merged, 5 October 2026

The merged visual learning phase simplified Home, Journey, Progress and word teaching,
with optional explanations behind keyboard-accessible disclosures. Meaning-specific
original photos, illustrations and diagrams are mapped to the 92 active canonical meanings; they appear
in teaching and word detail, never in graded questions. See
[visual learning delivery](../design/visual-learning-delivery.md) for artwork coverage,
reference-video decisions and current verification. This phase is not yet a
verified production release; the previously accepted release below remains the
production record.

## Consolidated maturity delivery — 5 October 2026

[PR #1484](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1484), merged
on 4 October as `af3eacb44edacedd4ee0fec0fa9d1f3026e9f742`, adds private named lists and list-selected
practice, direct per-meaning lesson saving, typed/listening guided exercises,
shared authored meaning teaching, six original mini-stories, seventeen situation
guides and topic writing. These additions were delivered together in one PR, version **0.4.0**.
See [maturity delivery](maturity-delivery.md) for navigation, data contracts,
verified outcomes and limitations. Forward migrations are now 40 and account
export is schema 1.5; original lesson definitions and historical snapshots remain
compatible.

Final whole-workspace `pnpm run validate` passes. Independent root PostgreSQL checks pass for accounts,
lists, lessons, practice and stories on the isolated database after all 40 real
migrations and canonical seed. Consolidated regression checks pass (258 foundation,
52 client and 135 web helper tests, plus Go suites). The affected browser matrix passes
177/177 across 360px, 430px and desktop, including keyboard and both themes.
[Staging run 37232617616](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/37232617616)
deployed the merged revision and passed exact release identity verification. Its
core journey failed before feedback: an old test selector chose a hidden word link
inside the new closed guide accordion. Independent review also found that its
generic pressed-button selector can choose the audio speed control instead of
saving a meaning. The original maturity journey was skipped.

The focused correction in [PR #1485](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1485)
passed 18 browser regressions across all three layouts.
[Staging retry 37234463076](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/37234463076)
passed exact release identity and both real core and maturity journeys at
`2d222621ac721e1d81dd2d329c08a97c971ffc76`. Compared with merged `af3eacb44`,
only three test files and two docs differ; application and deployment inputs are
identical. This accepts the exercised staging flows, not physical speech playback,
calendar alarms, broad feedback quality or learner effectiveness.
Production promotion is verified below.

## Latest deployed release — verified 5 October 2026

[Production deployment 37235094589](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/37235094589)
completed successfully from merged PR #1484. Fresh production web and API
`/version` reads agree on **0.4.0**, commit
`af3eacb44edacedd4ee0fec0fa9d1f3026e9f742` and environment **production**.
The public API health probe reports service and database **ok**. Normal deployment
identity, readiness and synthetic smoke gates passed. Promotion used the merged
application source proven equivalent to the accepted staging build above; image
digest equivalence is not asserted.

The consolidated lists, authored teaching, mixed guided lessons, explicit lesson
saving, mini-stories, guides and topic writing are included. This release remains
a controlled-signup product. Live-flow acceptance does not establish broad AI
feedback quality, physical audio, reminder alarms or learner effectiveness.

## Previous production release — verified 4 October 2026

The expanded learning product was released through
[PR #1483](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1483), merged
on 2 October as `73846e2eac6596733bcca1b4a0a7c6a0e4cc2df0`.
[Production deployment 37055589721](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/37055589721)
completed successfully. Fresh web and API version reads on 4 October returned
**0.3.1**, that exact commit and environment **production**. The integrated
[accessibility run 37053713291](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/37053713291)
also reports success. These are release/CI facts, not evidence of live AI quality,
calendar alerts, physical-device pronunciation or learner effectiveness.

The local reference-review checkout is `0b2cee78`, the PR head. Its implemented
30-lesson course, vocabulary/knowledge features, varied practice, preferences,
calendar exporter, achievements and help page are included in the release.
Direct per-meaning saving at lesson completion remains a follow-up. Existing
completion already shows practised words, first-attempt accuracy and practice links.

The owner's recordings are important design sources. Use the
[timestamped video review and maturity priorities](reference-video-review.md)
before future feature or design work. This release is progress toward the full
product goal, not a claim that it is complete.

## Historical inventory and pre-release evidence

The dated local checkpoints below precede the final release and must not be read
as the current deployment status. Their specific test boundaries remain relevant.

The deployed baseline was reviewed on 2026-10-02 against implementation revision
`5b074584cb1799d6b8da7ee02909c703382a3ee5` in the consolidated
[PR #1482](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1482).
The active delivery-branch section also records later local implementation and
verification on that date. This is a dated inventory and readiness guide.
Update it when behavior, deployment, or verified evidence changes; it does
not establish that the final product is complete.

Use the [release-readiness plan](release-readiness.md) to turn the gaps below
into bounded delivery work with explicit acceptance evidence.

The owner's later direction on 2 October is **feature completeness before a
dedicated polish pass**. The [feature-complete delivery scope](feature-complete-delivery.md)
defines the connected product being built. The MVP inventory below is a baseline,
not the intended stopping point.

## Active feature expansion — delivery branch, not deployed

The `codex/vocanova-product-delivery` working tree now contains a versioned
30-lesson path (90 target meanings), durable teaching/recall/context sessions,
canonical vocabulary search with situation/level/personal-knowledge filters,
device pronunciation, a connected practice hub, a visual vocabulary map and
private per-meaning notes and self-assessments. Typed recall, listening choices,
mistake practice, an optional starting-word check, a personal learning plan and
eight history-derived milestones are implemented locally. Home, Journey and Progress connect
these features while retaining the three primary destinations. Self-reported
knowledge, saved words, SRS stages and lesson completion remain distinct.

New domain data is included in account export and deletion. The lesson and
knowledge/practice migrations were first checked with 37 forward migrations in disposable PostgreSQL 16 databases, with
ownership, concurrency, replay, rollback and export/purge checks. Earlier local web
production build, type checks, 46 API-client tests and 127 web helper tests passed.
The integrated baseline then passed `pnpm run validate` (exit 0): formatting,
lint, Go vet, type checks, 256 foundation tests, 50 API-client tests, 133 web helper
tests, the Go package suites and application builds. This baseline predates the
pending self-check/saved-empty interface changes and AI input case-preservation
fix; those changes need their own focused verification and final integration.
Focused browser checks now pass for guided lessons 12/12, vocabulary search 6/6,
pronunciation 18/18 and private knowledge/notes 12/12 across 360px, 430px and desktop.
The practice hub passed 21/21 after a mobile navigation contrast fix; repeatable
practice passed 27/27 including corrected navigation assertions. Initial fixture-origin/CORS and locator
failures were reproduced and corrected without weakening assertions.
This is in-progress local evidence, not staging or final release acceptance.
The self-check's atomic known-status update passed PostgreSQL concurrency,
note-preservation, replay and account-deactivation checks. Achievement threshold
dates and qualifying history passed read-only PostgreSQL checks and independent
review. Self-check, plan and achievement browser checks passed 27/27 across 360px,
430px and desktop, including light/dark, keyboard, accessibility, empty and
unavailable states, note-preserving retry and signed-out redirects.
The [expanded curriculum](starter-curriculum.md) preserves all 400 original seed
rows and the original seven lesson definitions. The canonical inventory is now
17 situations, 89 words/phrases, 92 meanings, 148 examples and 200 notes, with 90
distinct meaning targets across 30 lessons. Disposable PostgreSQL checks passed
repeated seeding, preserved saved references, all guided lessons, current practice
and persisted version-1 practice read/list/replay/answer/continue behavior. These
are content and persistence checks, not a learner-effectiveness study.

Current goal and focus have a separate authenticated, CSRF-protected
[learning-preferences API](../../apps/api/app/api/learning_preferences.go).
It preserves original onboarding answers and existing daily pace/settings;
revision checks reject a conflicting stale change while a retry of already-saved
intent is a no-op. Missing overrides fall back to onboarding, with nulls retained
for an older completed account that has no answers. Migration
[20261002233000](../../apps/api/migrations/20261002233000_learning_preferences.sql)
brings the forward inventory to 38. A fresh disposable PostgreSQL 16 run passed
preference concurrency/preservation, account export/anonymization and current/
historical practice checks. Account export schema is now 1.4. Goal/focus editing
is connected to the plan and starting-word check. A later 42-check mock-browser
run passed learning-direction editing/retry/conflict recovery, lesson-specific
practice selection and canonical search across 360px, 430px and desktop.

Home and the plan use a requester-owned
[lesson recommendation](../../apps/api/app/api/lesson_recommendation.go).
An unfinished session stays resumable. Otherwise the recommendation prefers
the learner's focus and actual useful target coverage, excluding meanings marked
known or currently mastered; saved alone does not mean known. Missing content or
read failures are not represented as completed learning. The read-only API
awards no progress or points.
A later 15/15 mock-browser recommendation run passed across the three layouts,
including Home's redirect when its recommendation read encounters an expired
session, unfinished-session resume and honest unavailable states.

[Calendar reminders](calendar-reminders.md) are implemented as an optional,
explicit download and calendar import. The daily event uses the importing
calendar's local-time behavior and contains a display alarm. Exporter tests pass
6/6 and static review found no actionable issue. All nine calendar browser checks
and 30 affected Settings checks passed across the three layouts, including
downloaded bytes, unchanged preferences, both themes and failure recovery.
A real calendar import and observed alert remain unverified. No email or push reminder is
sent, and the retained legacy reminder preference is not treated as consent.

The branch contains an optional OpenAI moderation/feedback provider and the
`sentence-feedback-v4` prompt with unchanged `feedback-schema-v3` semantics.
It has not activated OpenAI on staging or production. A preregistered synthetic
service gate for GPT-5 nano **failed and stopped** after four provider POSTs:
two cases reached moderation and feedback, while two local validation/safety
controls used no provider calls. The second provider case disagreed with the
frozen original-sentence naturalness reference; six cases remain unrun. This
used isolated memory storage, not PostgreSQL, and exercised no live repair.
It establishes neither model acceptance nor resolution of the observed live
feedback failure. See the [current acceptance record](release-readiness.md#active-feature-expansion-checkpoint--2026-10-02).

A later GPT-4o mini gate completed ten reused development cases with 14 POSTs
and passed the frozen automated references and lifecycle checks. Independent
coordinating AI review accepted only that bounded development gate; technical
headings, overly cheerful grief feedback and a lowercase correction remain
quality notes. It used isolated memory, exercised no live repair and had no human
learner review. Its source/binary hash manifest was recorded after the first two
POSTs; internal frozen case/rubric guards ran before network access, and the same
process continued after the gap was disclosed. No model activation or deployment
followed. This does not supersede the failed nano gate or establish broad quality.

The saved collection now has its own full-collection word/definition search,
learning-stage and due filters. Its page and total share one database snapshot;
cursors bind the requester and normalized filters. Displayed review state keeps
reviewed legacy `new` records in Learning while preserving raw status. Due means
eligible by the existing review schedule, without a daily-target cap. Both
PostgreSQL regressions passed after all 38 migrations and the real canonical seed,
covering more than 50 saved meanings, literal search, isolation, removed cursor
boundaries and exhausted-page counts. A 48-check mock-browser set passed self-check,
plan recommendation and collection-filter flows; the three existing library
save/detail/practice/remove checks then passed separately. The integrated local
baseline above is accepted within its scope; the later changes, combined browser
acceptance, deployment, quality and launch requirements stay open.

## Product and source of truth

Vocanova serves A2–B1 learners through a practical vocabulary loop: discover a
word in a real situation, save it, recall it, write an original sentence, receive
focused feedback, and return for the next daily mission. The
[product bible](00-product-bible.md) and [MVP PRD](01-mvp-prd.md) define the
baseline; later owner-authorized delivery documents describe additions.
[AGENTS.md](../../AGENTS.md) governs implementation and release work.

Product direction recorded on 2026-10-02: Vocanova serves individual learners;
B2B is outside the product scope. The MVP should be free at launch. Any future
monetization requires a separate owner decision informed by learner value and
retention.

A separate ChatGPT Site prototype describes a three-word lesson, recall,
sentence practice, discovery, collection, progress, and mobile navigation with
sample data. It is design input; it does not establish backend functionality,
real learner data, or deployment of this repository. Recovery of the full
ChatGPT project was incomplete during this review; this inventory does not
claim to preserve every conversation or attachment.

## Implemented surfaces

These links locate existing implementation. They do not imply each flow was
freshly tested live on the review date.

| Surface                 | Current implementation                                                                       | Source                                                                                                                                                                                               |
| ----------------------- | -------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Entry                   | Landing, capability-driven sign-in/signup, onboarding                                        | [Web routes](../../apps/web/src/app/), [auth capabilities](../../apps/web/src/lib/auth-capabilities.ts)                                                                                              |
| Home                    | Daily mission, next action, due reviews, saved words, sentence practice                      | [Home](<../../apps/web/src/app/(app)/home/page.tsx>), [missions](../../apps/api/business/missions/)                                                                                                  |
| Journey                 | Situation browsing, situation/word detail, saved vocabulary                                  | [Discovery](<../../apps/web/src/app/(app)/discover/>), [content](../../apps/api/business/content/), [learning](../../apps/api/business/learning/)                                                    |
| Recall                  | Focused review sessions and deterministic spaced scheduling                                  | [Review routes](../../apps/web/src/app/review/), [reviews](../../apps/api/business/reviews/)                                                                                                         |
| Guided learning         | Durable teaching, recall and context lessons; optional starting-word check and personal plan | [Lessons](<../../apps/web/src/app/(app)/learn/>), [self-check](<../../apps/web/src/app/(app)/vocabulary/check/>), [plan](<../../apps/web/src/app/(app)/plan/>)                                       |
| Independent practice    | Repeatable typed recall, device listening choices, mistake practice and resumable sessions   | [Practice hub](<../../apps/web/src/app/(app)/practice/>), [practice domain](../../apps/api/business/practice/)                                                                                       |
| Personal knowledge      | Canonical search, whole-catalog knowledge summary, explicit known status and private notes   | [Vocabulary](<../../apps/web/src/app/(app)/vocabulary/>), [word knowledge](../../apps/api/business/wordknowledge/)                                                                                   |
| Sentence practice       | Feedback, recoverable drafts, retries, sentence history                                      | [Practice](<../../apps/web/src/app/(app)/_components/sentence-feedback.tsx>), [history](<../../apps/web/src/app/(app)/progress/sentences/>), [feedback service](../../apps/api/business/aifeedback/) |
| Progress                | API-backed activity, missions/streaks, Confidence Points                                     | [Progress](<../../apps/web/src/app/(app)/progress/>), [gamification](../../apps/api/business/gamification/)                                                                                          |
| Account                 | Learning preferences, profile, email change, password security, export, deletion             | [Settings](<../../apps/web/src/app/(app)/settings/>), [accounts](../../apps/api/business/accounts/), [passwords](../../apps/api/business/password/)                                                  |
| Appearance and identity | Light/Dark/System; web/API release identity                                                  | [Theme](../../apps/web/src/lib/theme-preference.ts), [web version](../../apps/web/src/app/version/), [release operations](../development/account-and-release-operations.md)                          |
| Calendar reminder       | Optional daily event download; learner imports and manages it in a calendar                  | [Reminder component](<../../apps/web/src/app/(app)/settings/_components/calendar-reminder.tsx>), [scope and limits](calendar-reminders.md)                                                           |

Home, Journey, and Progress remain the three primary destinations. Additions are
documented in [product maturity](product-maturity-delivery.md),
[learning continuity](mature-learning-and-account-experience.md), and
[password/profile/theme scope](password-profile-and-theme.md). These contain
dated evidence or acceptance plans, rather than a fresh test report.

The deployed baseline had seven situations, 51 words, 54 meanings and 72 examples.
The delivery branch's canonical [seed](../../apps/api/cmd/seed/voc026-p1.json)
now contains the [30-lesson starter curriculum](starter-curriculum.md) described
above; deployment and learner usefulness remain separate gates.
[Daily Conversation](daily-conversation-curriculum.md)
has 18 ordered meanings and two examples each, with editorial level rationales
and preserved existing identities. Word pages suppress repeated short/full
definitions while retaining distinct fuller guidance. Generated feedback is
distinct from canonical vocabulary.

Daily Conversation includes optional context practice: three
independently reviewed message/dialogue examples contrast invite/join,
sounds good/keep in touch and reschedule/cancel. Explanations follow the choice;
learners can retry, restart or exit locally. Completion uses existing word-page
links and explicit saving before sentence practice. All six canonical IDs, text
and slugs must match the situation response; missing or drifted references omit
the activity. It adds no points, mission/progress changes, API/schema or provider
calls. Full workspace validation passed (246 foundation, 32 API-client and 98
web helper tests, Go checks and builds). All 21 focused browser checks passed
across 360px, 430px and desktop in both themes. The full browser matrix passed
314 tests with 37 existing skips. The change merged in
[PR #1476](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1476), with
staging release checks passing as recorded below. That release's deployed
synthetic journey did not exercise this specific activity; its interaction
evidence came from the mock-backed browser suite.

[PR #1478](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1478) added a focused context phase to the existing staging
journey, before its save/review/submission actions. It checks the Daily
Conversation inventory, the three examples and explanations, keyboard
retry/continuation, completion/restart/exit, reload reset and canonical word
destinations. It does not save words or submit sentences; the rest of the
existing staging journey still changes the reserved synthetic account and may
call the evaluator. The exact helper is also exercised locally. Its first
[staging run](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36973540873)
passed identity/health/OAuth checks but failed an ambiguous example-text lookup
on the sounds-good page after completing the three examples. That is not a full
journey pass. The follow-up checks one visible example within the main landmark,
with local controls for hidden and visibly duplicated markup.
[PR #1480](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1480) shipped
that correction, and [staging36977504685](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36977504685)
passed the complete context phase. The existing review phase then failed because
the persistent synthetic account had completed its daily target. The full journey
had not passed at that point. The fixture follow-up prepares a fresh guarded
staging account while preserving retired history and normal learner behavior.
No real-account, physical-device, screen-reader speech or learning-effectiveness
claim follows from synthetic checks. See
[curriculum guidance](daily-conversation-curriculum.md#optional-context-practice).

Sentence practice protects pending drafts and retry identities from older
feedback actions. Current session errors remain visible alongside earlier
feedback, and late report responses cannot mark a newer result as reported.
Settings preserve edits made during an earlier save and distinguish those
unsaved edits from confirmed changes. A retry explicitly resends fields whose
previous save response was lost. The deployed baseline's reminder switch stored
a preference only. The delivery branch replaces it with the explicit calendar
download described above, without changing that stored value or sending messages.

The consolidated delivery also protects the separate profile editor: an earlier
save cannot replace newer typing, overlapping saves are ignored while pending,
and the status distinguishes confirmed changes from a newer unsaved name.
Review choices exclude other meanings of the current canonical word because the
word-only prompt cannot distinguish those valid answers. When fewer than three
safe alternatives remain, the existing self-check mode is used. Sign-in recovery
uses known capabilities when suggesting another method; unknown capabilities
fall back to retry guidance. The current-user response preserves an authoritative
completed onboarding status for older accounts without questionnaire answers;
missing answers alone no longer send those learners back through setup.
These working changes require the combined release
verification recorded in [release readiness](release-readiness.md).

## Live deployment observation

The consolidated implementation `5b074584` passed the complete
[staging deployment and learning journey](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36989699136)
on 2026-10-02 at 09:29 UTC. Release identity, health and OAuth initiation also
passed. A second [staging-only scheduled journey](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36990179266)
passed at 09:31 UTC, exercising fresh preparation and the full learning loop
again on the same day. These were pre-merge checks of the frozen PR revision.
They establish synthetic acceptance on staging; production promotion, real
provider-account acceptance and learner usefulness remain separate requirements.

### Earlier deployment observations

Public checks on 2026-10-02 at approximately 03:24 UTC observed staging web and API
serving version `0.3.1`, commit `56bbff54a7ac5872a996af9d6b5c29f94e0ffe5f`, after
[PR #1473](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1473).
Production remained at `5b16186c6471cbbc1c935a703bb5a8a33bf28797`, version
`0.3.1`. Both APIs reported database health `ok` and correct environment labels.
The [staging deployment](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36959705987)
passed release identity, OAuth initiation and the reserved synthetic learner
journey. These establish availability and the tested synthetic flow; they do
not establish real account acceptance, live feedback quality or recoverable
backups. No production release was performed for this milestone.

Later, [PR #1474](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1474)
merged as `ea6ec479b9cb2923430944d51224dfd00f339786` and its
[staging deployment](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36962911361)
passed release identity, health, OAuth initiation and the reserved learner
journey. A fresh local public-endpoint request returned HTTP 403; this later
release observation therefore comes from the deployment checks.

[PR #1475](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1475) then
delivered pending-edit recovery as `bb0241cb2ace5f848a1e28bbd4b3cb9b3d3118db`.
Required CI and Codex/Claude reviews cleared; local browser verification passed
299 tests with 37 existing skips. Its
[staging deployment](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36965361774)
succeeded, verified at 04:41:55 UTC, including strict release identity, web/API
health, OAuth initiation and the reserved synthetic learner core loop. This is
workflow verification, not a new direct public-endpoint observation after the
earlier HTTP 403. Production was not promoted.

[PR #1476](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1476) delivered
context practice as `e4cc4236f1f1c18df0059b005a4ade3b07addc31` at 05:16:59 UTC.
Applicable CI, inspected Codex code/security and Claude reviews, and all seven
merge-queue checks passed. [Staging deployment 36968236895](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36968236895)
succeeded, verified at 05:21 UTC, including exact release identity, web/API health,
OAuth initiation and the reserved synthetic learner core loop. The merged source
tree matches the reviewed head. These are workflow observations; production was
not promoted and no new direct local public-endpoint check is claimed.

Both environments reported Google OAuth and AI enabled. Magic-link email,
password authentication, and public new-user signup were disabled. Access is
through the controlled cohort. Configuration readiness does not prove a fresh
Google sign-in, email delivery, or real-provider feedback quality.

- Staging: [web version](https://staging.vocanova.site/version),
  [API version](https://api-staging.vocanova.site/version).
- Production: [web version](https://production.vocanova.site/version),
  [API version](https://api-production.vocanova.site/version).

Staging deploys automatically when pushes to `main` match the workflow's path
allowlist; a docs-only change under `docs/` does not itself trigger deployment.
Production requires manual dispatch. Inspect the actual triggers in
[staging](../../.github/workflows/deploy-staging.yml) and
[production](../../.github/workflows/deploy-production.yml); old comments still
mention former branch/promotion arrangements. Use the current operator section
in the [DevOps guide](../operations/11-devops-and-ci-cd.md#current-operator-process--2-october-2026)
and [monitoring runbook](../operations/monitoring.md). Keep host configuration private.

## Remaining readiness work

| Need               | Evidence or decision required                                                                                                                                                   | Starting point                                                                                                                                                         |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Real account entry | Verify controlled Google login/session recovery with a real provider session. If email/password is activated, prove delivery, verification, reset, and revoked-session behavior | [Staging cohort](../operations/staging-controlled-signup.md), [email activation](../development/account-and-release-operations.md)                                     |
| Learning quality   | Evaluate real-provider feedback against the documented rubric using representative A2–B1 sentences; inspect corrections, failures, and reporting                                | [AI requirements](../engineering/09-ai-features.md), [live evaluation command](../../apps/api/cmd/eval-live/), [feedback service](../../apps/api/business/aifeedback/) |
| Curriculum         | Editorially review and expand coherent situations based on learner needs; retain provenance                                                                                     | [Seed](../../apps/api/cmd/seed/), [content service](../../apps/api/business/content/)                                                                                  |
| Learner value      | Run an observed cohort; measure first-session friction, completed sessions, and return visits with cohort size and observation window                                           | [Evidence needs](mature-learning-and-account-experience.md#product-evidence-after-this-delivery)                                                                       |
| Recovery           | Synthetic dump/restore and failure controls pass locally; still need real backup inventory, isolated production recovery, retention and alert evidence                          | [Recovery rehearsal](../operations/postgres-recovery-rehearsal.md), [monitoring](../operations/monitoring.md)                                                          |
| Wider rollout      | Verify core flows on the chosen release with real accounts and physical/mobile devices; resolve failures and explicitly decide cohort expansion                                 | [Browser guidance](../development.md), [fixtures](../../apps/web/tests/e2e/)                                                                                           |
| Scaling            | Decide on a shared limiter before relying on multiple API replicas; current auth limits are process-local                                                                       | [Auth limiter](../../apps/api/business/auth/rate.go)                                                                                                                   |

Browser fixtures establish interface behavior with synthetic data. Disposable
database tests establish integration under test conditions. Neither proves live
email deliverability, Google account behavior, learner retention, or feedback
quality. Health checks and deployments also do not establish those claims.

The feedback evaluator now records meaning-aware observations and explicit
acceptance gaps. The v3 fixtures retain all 336 cases and 91 golden members;
six ambiguity cases remain excluded from status scoring. Curated regional
and phrase forms improve lexical validation without awarding semantic credit.
See the [evaluation guide](../engineering/feedback-evaluation.md) and
[fixture migration](../engineering/feedback-evaluation-fixtures.md) for the
meaning-aware cases, recorded evidence and checks that remain unmeasured.
An adapter-only run cannot establish full feedback acceptance.

Device pronunciation and optional calendar reminders are now part of the
authorized delivery scope. Speech recording/scoring, server-sent reminders,
monetization and native applications remain separate product decisions. Deliver
the connected learning experience reliably and prioritize additions using learner
evidence; historical Control Plane plans are not the current product roadmap.
