# Vocanova feature-complete delivery

## Active consolidated maturity delivery — 4 October 2026

The connected additions are implemented on `codex/vocanova-maturity` in
[PR #1484](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1484): named
lists and list practice, explicit lesson completion saving, shared authored
meaning teaching, mixed guided exercises, original stories, situation guidebooks
and topic writing. Read the source-linked
[maturity delivery contracts and acceptance](maturity-delivery.md) for navigation,
counts, versioning and privacy contracts. These additions extend the
recording-derived priorities without importing competitor assets.

Fresh follow-up `pnpm run validate` passed (exit 0), as did six actual PostgreSQL
packages including `app/api`. The preceding affected browser checkpoint passed
177/177 cases plus three screenshot-only checks. The latest three-spec follow-up
passed **53 cases with one existing desktop skip** across all three layouts,
including all six previously failing route loops and all 24 lesson-save cases.
Application revision `deacbf8c` passed all applicable hosted CI, including
[full accessibility](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/37229666520)
with **767 passed and 37 existing skips**. The previous six route-loop failures
were reproduced and corrected without weakening application authentication or
accessibility assertions. Independent review accepted the export, story and
save-recovery fixes.

The release candidate is **0.4.0**, kept in the same consolidated PR. The opt-in
real-staging maturity journey is authored and independently reviewed. A deliberate
failure probe used fake session/CSRF markers: two expected sanitized errors, eight
retained artifacts inspected, zero marker leaks. This is credential-report evidence,
not a live journey pass. Merge, deployed release identity and real-staging acceptance
remain separate from the local/hosted source-check checkpoint recorded here. See
[release readiness](release-readiness.md) for acceptance boundaries.

New migrations 39–40 and export schema 1.5 cover the additional learner data.
Source/build, local browser and isolated database evidence do not establish live
authentication/feedback, physical-device audio or learner usefulness. The full
product goal is not completed by opening this consolidated PR.

## Historical baseline release and reference checkpoint — 4 October 2026

This checkpoint records the preceding release, **not the maturity additions in
PR #1484**. The earlier connected expansion merged through
[PR #1483](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1483) and
released as **0.3.1**, commit `73846e2eac6596733bcca1b4a0a7c6a0e4cc2df0`.
Production web/API version reads and the successful deployment workflow were
checked on 4 October; see [current state](current-state.md). The older integration
checkpoint retained below predates that release. Its undeployed and pending
implementation statements are historical; live quality and learner acceptance
limits remain open.

The owner's WordUp/Duolingo recordings are primary design references. Read the
[timestamped comparison and maturity requirements](reference-video-review.md)
before subsequent feature work. At this reference checkpoint, richer teaching,
personal lists, varied guided sessions, original stories and productive practice
were future priorities. Their current local implementation is now recorded in
the active inventory and maturity document above; shipment and learner acceptance
still require separate evidence.

Original owner direction, 2 October 2026: the current MVP is too narrow. Build the connected feature set first, then conduct a dedicated design, usability and quality pass. Keep the work on one delivery branch and open one consolidated PR. Existing reliability fixes are supporting work, not the product milestone.

## Product target

Vocanova helps an individual English learner decide what to learn, understand a word in context, recall it, use it, and return with a clear next step. WordUp informs useful vocabulary and a personal knowledge map; Duolingo informs short guided sessions, varied practice and sustainable habits. Vocanova has its own content and interface. It does not need every feature of either reference product to provide this complete experience.

The first complete offering remains free and focused on practical English for A2–B1 learners. Existing A1/B2 self-reported preferences must remain valid, but the interface must accurately describe available content. No subscription, employer dashboard or external social network is required for this delivery.

## Active required feature inventory

The rows retain the accepted product scope and reflect the integrated local
implementation for PR #1484. The [maturity delivery map](maturity-delivery.md)
links the current counts to authored content and implementation sources. Earlier
baseline checks are identified separately where relevant; a screen or passing
local check does not establish deployment or learner acceptance. Remaining gates
belong in [release readiness](release-readiness.md).

| Area                         | Required experience                                                                                      | Current implementation and remaining acceptance                                                                                                                                                                                                                                                                                                                                                                                                   |
| ---------------------------- | -------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Personal learning plan       | Use goals, focus, daily pace and current vocabulary to recommend a next lesson; allow changing direction | Optional self-check, editable goal/focus and personal recommendations use unfinished sessions, focus and known/mastered target coverage, preserving onboarding history. Earlier focused preference/recommendation browser checks passed. No CEFR or vocabulary-size estimate is claimed.                                                                                                                                                          |
| Guided lessons               | Short lessons that teach, test recall and practise context; exit, resume and revisit completed lessons   | Thirty lessons support new nine-step snapshots: three teaching cards, meaning choice, typed recall, listening choice and three context choices. Original lesson definitions and historical sessions remain readable. Completion offers explicit per-meaning Save for review; wrong answers require correct retry while first-answer totals remain distinct. Current PostgreSQL and mock-browser checks pass; live/learner acceptance remains.     |
| Vocabulary discovery         | Search words and meanings, browse topics and levels, see saved state, choose useful words                | Canonical search, situation/level/knowledge filters, pagination and saved/known overlays remain available. Journey connects to lists, guides, stories and topic writing without adding primary navigation destinations. Earlier focused search/filter browser checks passed.                                                                                                                                                                      |
| Rich word pages              | Meaning, word type, useful examples, pronunciation, usage cues and clear next practice actions           | Canonical and saved details share authored examples, grouped usage guidance and collocation phrase audio. Comparison links point to actual other meanings of the same word through canonical `#meaning-{meaningId}` anchors. Current meaning/keyboard/light-dark browser checks pass. Device playback is not pronunciation scoring; no invented related-word graph is claimed.                                                                    |
| Personal vocabulary          | Search/filter the collection, see learning stages, find due words and revisit meanings                   | Saved-collection literal search, stage/due filters, full filtered counts, bound cursors and empty/error recovery remain implemented. Explicit saving, private notes and self-assessment stay distinct from list membership and SRS mastery. Earlier database/browser checks and current affected save/detail flows support behavior; learner usefulness remains open.                                                                             |
| Personal lists               | Organize selected meanings into named collections and practise an explicit collection                    | Create/search/rename/delete lists and add/remove membership from discovered or saved details. Limits are 50 active lists, 80 trimmed Unicode characters per name and 500 meanings per list. Revision checks, exact retry and current-state recovery protect newer edits. Membership does not automatically save for SRS or award points. Current PostgreSQL and mock-browser checks pass.                                                         |
| Varied practice              | Meaning recall, context choices, typed recall, listening, sentence writing and spaced reviews            | Typed/listening sessions can select a list with its current revision and freeze its name/content for resume. Availability counts expose the supported subset of 90 reviewed targets; invalid, stale or empty lists have no course fallback. Existing review, mistake and writing flows remain connected. Current PostgreSQL/mock-browser checks cover replay and list changes/deletion; physical-device audio remains unverified.                 |
| Mistake review               | Return to difficult words and incorrect lesson answers with targeted practice                            | Recorded qualifying mistake sources are resolved without clearing newer mistakes. Typed first-answer failures persist as mistakes; the reproduced PostgreSQL regression now passes. Empty states do not fabricate errors. Current independent database verification supports persistence; live journey acceptance remains.                                                                                                                        |
| Writing feedback             | Useful affordable feedback, understandable corrections, history and safe retry                           | Seventeen original situation prompts connect a selected saved meaning to existing feedback/rewrite/history controls. `word_detail` checks the selected meaning and language, **not topic adherence**; this is not unrestricted free-writing grading. Current mock-browser flows pass. The historical provider gates below do not resolve staging availability or establish broad real-provider quality; neither optional candidate was activated. |
| Stories and situation guides | Practise practical language through connected reading, checks and learning destinations                  | Six original stories each contain eight lines and two checks, with ten-step durable/resumable sessions. Seventeen guidebooks contain 51 original phrases and connect vocabulary, lessons, available stories and writing. Authored-source checks plus current PostgreSQL/mock-browser verification pass. Real learner usefulness, live journey and device audio remain open.                                                                       |
| Knowledge and progress       | Show learned activity, vocabulary stages, lesson progress and practice history                           | Whole-dataset knowledge summary, vocabulary map, lesson/practice history and private per-meaning state remain implemented. New list/story activity participates in schema 1.5 export and account purge. No proficiency score is inferred from activity or one page of words. Current PostgreSQL privacy checks pass; deployed acceptance remains.                                                                                                 |
| Habit and rewards            | Daily challenges, streaks, earned achievements and clear celebrations                                    | Existing points/streaks remain server-confirmed; eight milestones derive from qualifying history. Guided/independent practice, list membership and story completion do not manufacture SRS reviews or reward points. Learner usefulness still needs observation.                                                                                                                                                                                  |
| Reminders and return         | Optional daily reminders with explicit setup, understandable time behavior and a way to stop             | Calendar-file download and manual import were released in the preceding baseline. Its six exporter, nine browser and 30 Settings checks are historical evidence. Calendar controls local time, DST and alerts; actual import/alert delivery remains unverified. Email/push is not enabled and legacy preferences are not consent.                                                                                                                 |
| Practical curriculum         | Enough reviewed lessons, words, examples and exercises for sustained use                                 | The starter curriculum contains 30 lessons/90 targets within 17 situations, 89 words and 92 meanings. Original IDs and seven lesson definitions remain unchanged. The additional stories/guides/prompts are counted separately above. Editorial and current persistence checks support the content; real learner usefulness is unproven.                                                                                                          |

## Current integrated acceptance boundary — PR #1484

Fresh follow-up `pnpm run validate` passed with exit 0: formatting, lint/Go vet,
type checks, 258 foundation tests, 52 API-client tests, 135 web-helper tests, Go
package suites and application builds. Root independently reran six actual
PostgreSQL packages: accounts, lessons, lists, practice, stories and `app/api`,
against 40 forward migrations and the canonical seed in isolated PostgreSQL 16.
Direct SQL application does not verify deployed Atlas revision history.

Export transport now explicitly preserves all six learning fields: lists, stories,
learning preferences, practice sessions, word knowledge and guided lessons.
Legacy schema versions and absent/null fields remain intact. Story current-step
and feedback references correctly accept null; library reads select the newest
session per story key. Lesson-save 409 recovery rereads canonical status; an
unavailable read requires a status check and does not automatically save.

The preceding production-build browser checkpoint passed 177/177 affected cases
across 360px, 430px and desktop, plus three screenshot-only checks. The latest
follow-up ran the complete lesson-save, primary-navigation and Home mobile
accessibility specs at all three layouts: **53 passed, one existing desktop skip**.
All six formerly failing route loops and all 24 lesson-save cases pass. The route
loops lacked a fixture session cookie; the correction seeds authentication and
retains navigation, skip-link keyboard and overflow assertions. These mock-backed
results do not establish live authentication, provider quality or database durability.

Core CI, performance and automated review passed on the preceding pushed revision.
Hosted full accessibility previously failed (six failed, 752 passed, 37 skipped);
a new hosted passing run remains required after the pending follow-up push.
PR #1484 remains ready with `hold`, not merged or deployed. The opt-in staging
maturity journey is authored; credential safety and fake artifact verification are
ongoing, with no live pass. Execute the connected journey after deployment and
record its revision/results in [release readiness](release-readiness.md).
Live feedback availability/quality, physical-device audio, calendar import/alerts
and learner usefulness remain separate acceptance work. No speech grades,
automatically translated teaching media, social system or subscription is claimed.

## Historical pre-release integration checkpoint — superseded

The following paragraphs preserve the earlier delivery-branch checkpoint before
PR #1483's release and PR #1484's additions. **The 38-migration/schema 1.4 counts,
undeployed descriptions, smaller test totals and pending implementation statements
below are historical, not the current integration status.** Use the active sections
above for the current 40-migration/schema 1.5 delivery and acceptance. The recorded
limits on provider and learner evidence still apply where no later evidence exists.

The [starter curriculum](starter-curriculum.md), [calendar reminder](calendar-reminders.md)
and preference persistence are implemented on the delivery branch, not deployed.
Learning preferences add migration 38 and account export schema 1.4. A disposable
PostgreSQL 16 run applied all 38 forward migrations and passed preference,
export/anonymization and current/historical practice checks; this is not live
durability or an Atlas revision-history verification.

Earlier mock-browser checks passed for the connected lesson/search/practice
surfaces and 27/27 self-check, plan and achievement cases. Later checks passed
42/42 learning-direction/practice-selection/search cases, all nine calendar and
30 Settings cases, and 48 self-check/recommendation/collection-filter cases. The
three existing library save/detail/practice/remove checks passed separately.
The subsequent recommendation run passed 15/15, including Home's expired-session
redirect and resume/unavailable states across the three layouts.
These are focused runs, not a fresh complete browser matrix. Saved-collection
queries also passed two PostgreSQL regressions against all 38 migrations and the
real canonical seed. Calendar export/browser tests do not establish a calendar
application's alarm delivery. The separately preregistered nano service gate used synthetic
inputs and isolated memory; its failed/stopped result does not accept provider
quality or resolve staging availability. A later GPT-4o mini gate passed ten
reused development cases with 14 POSTs and bounded independent AI acceptance.
Its source/binary manifest was captured after the first two POSTs; internal
case/rubric hash guards ran before network, and the same process continued after
disclosure. Memory-only persistence, no live repair or human review, and wording/
tone limitations remain. Neither candidate was activated or deployed. Keep these
limits explicit during the combined release and later polish pass.

The integrated local baseline passed `pnpm run validate` with exit 0, covering
formatting, lint, Go vet, type checks, 256 foundation tests, 50 API-client tests,
133 web helper tests, Go package suites and application builds. Bounded
mock-inventory predicates/tests now cover the approved expansion. The practice
migration received whitespace-only formatting for the inventory scanner;
whitespace-stripped SQL stayed identical, and Atlas checksum validation plus the
Go migration-scanner tests passed. The 38-migration database evidence remains
applicable to that unchanged SQL behavior.

Pending self-check/saved-empty interface changes and the AI input
case-preservation fix are later work and are not covered by this baseline. They
need focused verification, followed by final combined and deployed acceptance.

## Original delivery sequence — retained scope

This sequence records the accepted build direction, not an inventory of remaining
implementation gaps. Current integration and acceptance status appears above.

1. Build reusable lesson persistence, catalog search and knowledge summaries in parallel with pronunciation controls.
2. Connect them through Home, Journey and Progress. Home gives a useful next lesson and due reviews; Journey contains the learning path and word discovery; Progress explains actual learning and vocabulary coverage.
3. Add assessment/preferences, additional exercise modes, mistake practice, achievements, working reminders and broader curriculum on those foundations.
4. Run the integrated learner journey. Resolve functional, persistence and availability failures as part of feature implementation.
5. Conduct the dedicated polish pass: visual hierarchy, mobile ergonomics, accessibility, interaction consistency, speed and feedback quality. Validate the combined delivery before its single PR.

Checks during implementation establish that the new feature works and retains learner data. They should not turn into a separate cosmetic project or repeated broad validation before the connected feature set exists.

## State and learning rules

- Lesson steps are graded by the API and resume from confirmed state. Browser reloads, retries and another device cannot duplicate completion or overwrite newer progress.
- Teaching, lesson recall and scheduled SRS reviews are distinct activities. A lesson must not manufacture a completed review or language mastery.
- Save/removal remains an explicit learner action. Starting a lesson must not silently restore removed vocabulary.
- Canonical lesson content is versioned and validated; active sessions retain the content they began with. AI availability must not prevent curated lessons.
- New learner data participates in account export/deletion, follows existing authentication/CSRF boundaries, and remains isolated by learner.
- Audio starts only when requested. Device pronunciation is identified accurately, and unsupported browsers have a useful fallback. Speech recognition or pronunciation scoring must not be implied by playback.
- New aggregates cover the entire relevant dataset. Unknown, absent and zero are distinct.
- Updating current learning direction preserves original onboarding answers, daily settings and confirmed learning history. Stale conflicting edits must not overwrite a newer choice.
- A calendar download requires explicit import; changing a form does not change imported events. Stored legacy reminder preferences do not authorize email or push delivery.
- Reference product descriptions inform requirements, not evidence that Vocanova has implemented them.

## Historical reference observations — 2 October 2026

- [WordUp support](https://www.wordupapp.co/support) describes vocabulary assessment, goal/interest suggestions, a knowledge map and varied tests.
- [Duolingo's learning path](https://blog.duolingo.com/new-duolingo-home-screen-design/) explains guided sequencing with review integrated into the path.
- [Duolingo practice overview](https://blog.duolingo.com/ways-to-practice-in-duolingo/) describes distinct practice formats and personalized practice.

These references were checked on 2 October 2026. The implementation map remains [current state](current-state.md); final acceptance evidence belongs in [release readiness](release-readiness.md).

The owner also provided an authenticated Duolingo Chrome tab for direct inspection.
The walkthrough covered its learning path, Practice hub, Quests and Sounds overview.
Observed patterns include mixed lesson/story/listening/review activities within a
unit, a current-step entry point, dedicated mistake practice, sound groups with
example words, and daily/monthly challenges. These observations reinforce the
feature inventory above. Account details, learner history and other people's
profile information are not product reference material and are not retained here.

The owner also provided the installed Windows WordUp application. A read-only
walkthrough on 2 October 2026 covered its home, knowledge map and full word detail.
Observed features include explicit already-known/should-learn choices, a visual
word grid, scheduled reviews, word and example audio, accent choices, translations,
personal notes and usage tips. Its home also offers separate vocabulary, writing
and speaking practice, plus AI roleplay. Practice entry points were observed;
their grading quality and full flows were not tested. These observations support
explicit self-assessment and private notes in Vocanova, kept separate from SRS
mastery, saved words and earned progress. Vocanova's map describes its actual
catalog, without claiming WordUp's frequency ranks or copying its content.
