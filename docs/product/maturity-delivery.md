# Integrated maturity delivery — 4 October 2026

This document maps the connected additions on `codex/vocanova-maturity`.
They extend the existing practical A2–B1 vocabulary product; they do not establish
that the whole product is finished, deployed or accepted by learners. The owner
requests one consolidated delivery, followed by an integrated polish and quality
pass. The [reference review](reference-video-review.md), [current state](current-state.md)
and [release readiness](release-readiness.md) remain the wider context.

## Learner experience and navigation

Home, Journey and Progress remain the three primary destinations. The additions
live inside Journey, saved vocabulary and Practice.

| Addition              | Implemented experience                                                                                                                                                                                                                                                                                           | Entry and authoritative source                                                                                                                                                                                                                          |
| --------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Personal lists        | Create, search, rename and delete named lists; search list meanings; explicitly add/remove a meaning from discovered or saved word details. List membership, saved SRS words and self-reported knowledge are separate.                                                                                           | `/lists`, `/lists/{listId}`; [list service](../../apps/api/business/wordlists/), [UI](<../../apps/web/src/app/(app)/lists/>)                                                                                                                            |
| List practice         | Choose a list for typed recall or listening; use a frozen list name/revision and session content. Short sessions contain at most six targets. Membership accepts active canonical meanings; availability counts identify the supported subset of the **90 reviewed practice targets**, not all catalog meanings. | `/practice?list={listId}`, `/practice/session/{sessionId}`; [practice catalog](../../apps/api/business/practice/catalog.go), [snapshot builder](../../apps/api/business/practice/snapshot.go)                                                           |
| Varied guided lessons | New sessions keep nine steps: three teaching cards, one meaning choice, one typed recall, one listening choice, then three context choices. Wrong answers require a correct retry; first-answer totals remain distinct from successful retries. Completion has explicit per-meaning Save for review choices.     | `/learn/{lessonKey}`; [mixed builder](../../apps/api/business/lessons/varied.go), [lesson state machine](../../apps/api/business/lessons/lessons.go)                                                                                                    |
| Meaning teaching      | Shared authored examples and usage guidance in canonical/saved details; collocation phrase pronunciation; an index linking actual other meanings of the same word. No guessed related-word graph is added.                                                                                                       | `/vocabulary/{wordSlug}`, `/words/{userWordId}`, canonical `#meaning-{meaningId}` anchors; [teaching](<../../apps/web/src/app/(app)/_components/meaning-teaching.tsx>), [comparison](<../../apps/web/src/app/(app)/_components/meaning-comparison.tsx>) |
| Original stories      | Six original situation dialogues, each with eight lines and two checks: one comprehension choice and one phrase completion. The ten-step sessions reveal lines progressively, persist answers and resume. Full reading, vocabulary help and optional device audio are also available.                            | `/stories`, `/stories/{storyKey}`, `/stories/session/{sessionId}`; [authored catalog](../../apps/api/business/stories/catalog.go), [story state machine](../../apps/api/business/stories/stories.go)                                                    |
| Situation guides      | Seventeen compact guidebooks with 51 original practical phrases, purposes, linked vocabulary/examples, and lesson/story/writing destinations where available.                                                                                                                                                    | `/discover/{situationSlug}#unit-guide`; [guide content](<../../apps/web/src/app/(app)/discover/[situation]/_components/unit-guide-content.ts>)                                                                                                          |
| Topic writing         | Seventeen original situation prompts connect a selected saved meaning to existing correction, explanation, improvement tip, rewrite and history controls.                                                                                                                                                        | `/writing?situation={situationSlug}&meaning={meaningId}`; [prompts](<../../apps/web/src/app/(app)/writing/_components/writing-prompts.ts>), [page](<../../apps/web/src/app/(app)/writing/page.tsx>)                                                     |

## State and compatibility contracts

List APIs are requester-owned `GET /api/v1/word-lists` and `GET/PUT/DELETE
/api/v1/word-lists/{listId}`, with meaning membership under
`/{listId}/members/{meaningId}`. Creation uses a caller-generated UUID and
`expectedRevision: 0`; later writes require the current revision and an
`Idempotency-Key`. Names allow 80 trimmed Unicode characters, with 50 active lists
and 500 meanings per list. Exact retries return current membership/name state;
they never restore a subsequently removed meaning. Deleted IDs retain tombstones
and cannot be recreated. Uncertain UI writes retain their exact request identity;
conflicts require a current-state reload before another edit.

Practice start accepts `listId` and `listRevision` together, excluding `lessonKey`
and mistake mode. There is no silent full-course fallback for invalid, deleted,
stale or empty selections. Typed list practice excludes inactive/unsupported
members; zero usable members returns a conflict. Listening still requires the
three reviewed alternatives for a target's lesson and reports unavailable content
if those alternatives are missing. Existing sessions keep their frozen content
and list metadata after later list edits or deletion. Practice content uses
`starter-90-v2`; the supported previous `starter-21-v1` snapshots remain readable.

New guided snapshots use `exerciseVersion: mixed-recall-v1` while original lesson
definitions remain version `1`. Historical snapshots with no exercise version keep
their original steps and action fingerprints. Typed grading normalizes Unicode
NFC, whitespace and case. Story snapshots use `original-dialogues-v1` and
`curated-choice-v1`; exact start/action replay returns current saved progress,
including a start replay after a story is retired from the current catalog.
Story current-step and feedback schema references accept null for completion or
no feedback. Library reads select the newest session per story key instead of
loading every repeated attempt. All learning mutations require authentication
and CSRF protection.

Lists, story completion and independent practice do not manufacture SRS reviews,
mastery, mission completion or reward points. Saving for scheduled review remains
an explicit learner action. Story/lesson first-answer results describe those
activities; device pronunciation is not a speech assessment.

Account export is schema **1.5**, including list membership, frozen practice list
metadata, mixed lesson version/typed answer history and story activity. HTTP
transport explicitly preserves lists, stories, learning preferences, practice
sessions, word knowledge and guided lessons. Older schema versions and absent/null
learning fields remain intact. Learner-visible deleted-list history includes its
`deletedAt` timestamp; interactive list DTOs omit it. Internal replay keys, client action IDs, fingerprints and private grading snapshots are
excluded. Account purge deletes the new requester-linked rows in foreign-key-safe
order. Forward migration **39** is
[private lists](../../apps/api/migrations/20261004100000_word_lists.sql); **40** is
[original stories](../../apps/api/migrations/20261004140000_original_stories.sql).
The source tree currently contains 40 forward migrations. Deployment/Atlas
revision-history acceptance must be checked separately.

## Verification checkpoint and remaining acceptance

- **Fresh local validation:** follow-up `pnpm run validate` passed (exit 0),
  including 258 foundation, 52 API-client and 135 web-helper tests, Go suites,
  formatting, lint/Go vet, types and production builds. Generated OpenAPI matches
  its generator. Explicit export transport/legacy nulls, nullable story references,
  newest-per-key story history and lesson-save conflict recovery are covered.
- **Actual PostgreSQL:** root independently reran six packages—accounts, lessons,
  lists, practice, stories and `app/api`—against isolated PostgreSQL 16 with 40
  forward migrations and canonical seed. Ownership, concurrent replay, rollback,
  historical snapshots, frozen list practice and schema 1.5 export/purge pass.
  This does not verify deployed Atlas revision history.
- **Browser checkpoints:** the preceding production-build affected matrix passed
  177/177 cases across 360px, 430px and desktop, plus three screenshot-only checks.
  The latest complete lesson-save, primary-navigation and Home mobile accessibility
  specs passed **53 cases with one existing desktop skip** at all three layouts.
  All six formerly failing route loops and all 24 save cases pass. Cookie seeding
  corrects the old authenticated-fixture assumption while retaining keyboard and
  navigation assertions. These mock-backed runs do not establish live login,
  provider quality or database durability.
- **PR evidence:** [PR #1484](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1484)
  is ready with `hold`, not merged or deployed; the follow-up push is pending.
  Core CI, performance and automated review passed before these follow-ups.
  Hosted full accessibility previously failed (six failed, 752 passed, 37 skipped);
  a fresh hosted run is required after pushing the fixes.
- **Live and learning acceptance:** the opt-in staging maturity journey is authored,
  with credential safety and fake artifact verification ongoing. No live pass,
  new deployment, real-provider feedback run, physical-device audio validation or
  learner acceptance is recorded. The earlier release applies only to its revision.
  The Chrome connector failed with a kernel reset; local browser verification used
  Chromium.

The staging journey should create a list, add one specific meaning without saving
it for SRS, practise that list, complete a mixed lesson and explicitly save a chosen
meaning, read/resume a story, use its guide, and write/rewrite a sentence. Execute
it after PR CI/review and deployment, recording revision and results in
[release readiness](release-readiness.md). Lesson-save conflict recovery reads
canonical state; failed status reads offer a usable check without saving automatically.

Topic writing reuses `word_detail` feedback: the provider checks the **selected
meaning and language**, **not adherence to the topic prompt**. It requires a saved
meaning and does not introduce unrestricted free-writing grading. Optional audio
uses the device engine with readable content and unavailable-audio recovery. This
addition includes no speech recording/grades, automatic translation or generated
teaching media, social system or subscription. Broader feedback quality, learner
usefulness and later polish still require their own evidence.
