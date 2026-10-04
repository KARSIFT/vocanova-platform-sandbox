# Vocanova feature-complete delivery

## Consolidated maturity feature set — 4 October 2026

The owner renewed the feature-first, one-PR direction. The connected additions are
implemented on `codex/vocanova-maturity`: named lists and their practice selection,
lesson completion saving, authored meaning teaching, mixed guided exercises,
original stories, situation guidebooks and topic writing. Read the
[maturity delivery contracts and acceptance](maturity-delivery.md). These additions
extend the recording-derived priorities without importing competitor assets.

New migrations 39–40 and export schema 1.5 cover the additional learner data.
Combined source/build and isolated PostgreSQL evidence are separate from live
feedback, physical-device audio and learner usefulness. Required release/learner
evidence remains in [release readiness](release-readiness.md); the full product
goal is not completed by opening this consolidated PR.

## Release and new reference checkpoint — 4 October 2026

The connected expansion was merged through
[PR #1483](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1483) and
released as **0.3.1**, commit `73846e2eac6596733bcca1b4a0a7c6a0e4cc2df0`.
Fresh production web/API version reads and the successful deployment workflow
were checked on 4 October; see [current state](current-state.md). The local
implementation/checkpoint wording below was written before that release and is
historical. Its undeployed statements are superseded here, while live quality
and learner acceptance limits remain open.

The owner's WordUp/Duolingo recordings are primary design references. Read the
[timestamped comparison and maturity requirements](reference-video-review.md)
before subsequent feature work. They extend the direction with richer teaching,
personal lists, varied guided sessions, original stories and productive practice.
These are prioritized future improvements, not shipped or accepted capabilities.

Owner direction, 2 October 2026: the current MVP is too narrow. Build the connected feature set first, then conduct a dedicated design, usability and quality pass. Keep the work on one delivery branch and open one consolidated PR. Existing reliability fixes are supporting work, not the product milestone.

## Product target

Vocanova helps an individual English learner decide what to learn, understand a word in context, recall it, use it, and return with a clear next step. WordUp informs useful vocabulary and a personal knowledge map; Duolingo informs short guided sessions, varied practice and sustainable habits. Vocanova has its own content and interface. It does not need every feature of either reference product to provide this complete experience.

The first complete offering remains free and focused on practical English for A2–B1 learners. Existing A1/B2 self-reported preferences must remain valid, but the interface must accurately describe available content. No subscription, employer dashboard or external social network is required for this delivery.

## Required feature inventory

The rows below retain the required scope and record the delivery branch's current
position. Local implementation is distinct from integrated browser verification,
deployment and learner acceptance. A feature is not complete merely because its
screen exists. See [release readiness](release-readiness.md#active-feature-expansion-checkpoint--2026-10-02)
for the evidence boundaries and remaining gates.

| Area                   | Required experience                                                                                      | Current local implementation and remaining acceptance                                                                                                                                                                                                                                                            |
| ---------------------- | -------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Personal learning plan | Use goals, focus, daily pace and current vocabulary to recommend a next lesson; allow changing direction | Optional self-check, editable goal/focus and a personal plan are implemented. Recommendations use unfinished sessions, focus and known/mastered target coverage; changes preserve onboarding history. Focused preference and recommendation browser checks pass. No CEFR or vocabulary-size estimate is claimed. |
| Guided lessons         | Short lessons that teach, test recall and practise context; exit, resume and revisit completed lessons   | Versioned 30-lesson catalog and durable teaching/recall/context sessions implemented. Repeat seeding, all lessons and preserved old practice snapshots pass disposable-database checks; representative mock-browser flows pass. Integrated release acceptance remains.                                           |
| Vocabulary discovery   | Search words and meanings, browse topics and levels, see saved state, choose useful words                | Canonical search, situation/level/knowledge filters, pagination and saved/known overlays implemented, with focused mock-browser checks.                                                                                                                                                                          |
| Rich word pages        | Meaning, word type, useful examples, pronunciation, usage cues and clear next practice actions           | Canonical and saved detail pages expose device pronunciation for words/examples, useful usage headings and learning destinations. Device playback is not pronunciation scoring.                                                                                                                                  |
| Personal vocabulary    | Search/filter the collection, see learning stages, find due words and revisit meanings                   | Saved-collection literal search, stage/due filters and full filtered counts are implemented with bound cursors and meaningful empty/error recovery. Database and browser checks pass. Private notes and self-assessment remain distinct from saving and SRS mastery.                                             |
| Varied practice        | Meaning recall, context choices, typed recall, listening, sentence writing and spaced reviews            | Practice hub, repeatable typed/listening sessions and existing review/writing flows connected. API grading, explicit reveal/assistance, exact retries and resume are implemented. Playback starts only on request; no graded speaking claim.                                                                     |
| Mistake review         | Return to difficult words and incorrect lesson answers with targeted practice                            | Practice uses recorded mistake sources and resolves the specific qualifying source without clearing newer mistakes. Empty states do not fabricate errors; database and mock-browser checks cover the behavior.                                                                                                   |
| Writing feedback       | Useful affordable feedback, understandable corrections, history and safe retry                           | Observed staging moderation failure remains unresolved by live acceptance. Optional OpenAI/v4 is local only. Nano failed/stopped; a later ten-case 4o-mini development gate passed within the limited evidence below. No provider activation or broad quality acceptance.                                        |
| Knowledge and progress | Show learned activity, vocabulary stages, lesson progress and practice history                           | Whole-dataset knowledge summary, vocabulary map, lesson/practice history and private per-meaning state implemented. No proficiency score is inferred from activity or one page of words.                                                                                                                         |
| Habit and rewards      | Daily challenges, streaks, earned achievements and clear celebrations                                    | Existing points/streaks remain server-confirmed; eight milestones derive from qualifying recorded history. Guided/independent practice does not manufacture SRS reviews or reward points. Learner usefulness still needs observation.                                                                            |
| Reminders and return   | Optional daily reminders with explicit setup, understandable time behavior and a way to stop             | Calendar-file download and manual import implemented; six exporter and nine browser checks pass, alongside 30 Settings checks. Calendar controls local time, DST and alerts; actual import/alert delivery is unverified. Email/push is not enabled and legacy preferences are not consent.                       |
| Practical curriculum   | Enough reviewed lessons, words, examples and exercises for sustained use                                 | Expanded to 30 lessons/90 targets within 17 situations, 89 words and 92 meanings. Original IDs and seven lesson definitions remain unchanged. Editorial and persistence checks support the content; real learner usefulness is unproven.                                                                         |

## Current integration boundary

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

## Build order

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

## Reference observations

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
