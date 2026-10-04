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
All these learning mutations require authentication and CSRF protection.

Lists, story completion and independent practice do not manufacture SRS reviews,
mastery, mission completion or reward points. Saving for scheduled review remains
an explicit learner action. Story/lesson first-answer results describe those
activities; device pronunciation is not a speech assessment.

Account export is schema **1.5**, including list membership, frozen practice list
metadata, mixed lesson version/typed answer history and story activity. Internal
replay keys, client action IDs, fingerprints and private grading snapshots are
excluded. Account purge deletes the new requester-linked rows in foreign-key-safe
order. Forward migration **39** is
[private lists](../../apps/api/migrations/20261004100000_word_lists.sql); **40** is
[original stories](../../apps/api/migrations/20261004140000_original_stories.sql).
The source tree currently contains 40 forward migrations. Deployment/Atlas
revision-history acceptance must be checked separately.

## Verification checkpoint and remaining acceptance

| Evidence layer               | Current boundary                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Source and focused checks    | Independent root consolidated checks passed: 258 foundation, 52 API-client and 135 web-helper tests plus Go package suites. Formatting, lint/Go vet, client/E2E types and production web/API builds passed. The generated API description exactly matches its generator. **Final `pnpm run validate` passed (exit 0)** after the last persistence fix. The new actual-database regression reproduced typed mistakes missing before the fix and passed afterward; root reran all five PostgreSQL packages successfully.                                         |
| Disposable PostgreSQL        | Root independently reproduced all five integration suites for accounts, lessons, lists, practice and stories after applying 40 real forward migrations and canonical seed to isolated PostgreSQL 16. Ownership, concurrent replay, rollback, historical snapshots, retired-catalog start replay, frozen list practice and schema 1.5 export/purge passed. Direct SQL application does not verify deployed Atlas revision history.                                                                                                                              |
| Browser                      | **177/177 passed** across 360px, 430px and desktop on the production build: personal lists, meaning teaching, original stories, varied lessons, topic writing, explicit lesson saving, all 17 guides and affected existing lesson/pronunciation flows. Keyboard interaction, light/dark themes, accessibility scans, overflow, empty/error/conflict and exact-retry paths are covered. Root inspected light/dark desktop/mobile story screenshots. Mock-backed browser passes cannot establish live authentication, provider quality or PostgreSQL durability. |
| Live and learning acceptance | No new deployment, real-provider login/feedback run, physical-device audio validation or observed learner acceptance is recorded for these additions. The interactive Chrome connector failed with a kernel reset; browser verification used the local Chromium harness. Earlier release evidence applies only to its recorded revision.                                                                                                                                                                                                                       |

The next acceptance journey should create a list, add one specific meaning without
saving it for SRS, practise the selected list, complete a mixed lesson and explicitly
save a chosen meaning, read/resume a story, use its guide, and write/rewrite a sentence.
Repeat this journey against staging after PR CI/review and deployment; the local
browser and database checks cover reload, lost-response replay, conflicting writes,
deletion and account export/purge independently.

Topic writing reuses `word_detail` feedback: the provider checks the **selected
meaning and language**, **not adherence to the topic prompt**. It requires a saved
meaning and does not introduce unrestricted free-writing grading. Optional audio
uses the device engine with readable content and unavailable-audio recovery. This
addition includes no speech recording/grades, automatic translation or generated
teaching media, social system or subscription. Broader feedback quality, learner
usefulness and later polish still require their own evidence.
