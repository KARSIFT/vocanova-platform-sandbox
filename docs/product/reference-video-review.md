# WordUp and Duolingo recording review

Reviewed 4 October 2026. The owner supplied **WordUp.mp4** (12:53) and
**Duolingo.mp4** (19:00) as important product role models. Use these primary
references alongside the [product bible](00-product-bible.md) and
[learning workspace](../design/learning-workspace.md) for subsequent design work.

## Evidence and limits

The review covers visual samples throughout both recordings at eight-second
intervals, plus close views around teaching, corrections and completion. It is
not continuous audiovisual playback or an audio-quality evaluation. Timestamps
identify observed screens; flow ranges are approximate. Brief transitions can
fall between samples. Reopen relevant segments before implementing details that
depend on animation, timing or sound.

Recordings, extracted frames and processing tools remain outside the public
repository because desktop/account information appears in them. The owner placed the source copies in the parent workspace
`Opponent/` folder. The private working index records paths, dimensions, durations
and SHA-256 hashes; the copies match the reviewed originals. Both
recordings are 1920 × 1080 with audio; audio was not transcribed or assessed.
Visible controls do not prove general availability, persistence, accessibility
or teaching effectiveness. Account identities and learner statistics are not
retained as product evidence.

The comparison uses local revision `0b2cee78` and the merged
[release PR #1483](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1483).
Fresh production web/API version reads on 4 October returned **0.3.1**, commit
`73846e2eac6596733bcca1b4a0a7c6a0e4cc2df0`, environment production. This verifies
release identity, not live AI quality or every learner flow.

## WordUp observations

| Timestamp                | Observed experience                                                                                                                                                               | Vocanova implication                                                                                                                                               |
| ------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 00:00; 07:28–07:52       | Home connects reviews, discovery, the next word and practice entry points.                                                                                                        | Keep one useful next action and connect discovery to learning/review. A large review backlog should not imply mandatory daily workload.                            |
| 00:08–01:20              | Friends, invitations, country/global league views and rankings. Invalid invitation feedback is visible.                                                                           | Social competition is a separate privacy/fairness decision, beyond the personal vocabulary loop.                                                                   |
| 01:36–02:24              | Daily goal choices, calendar/streak explanation and milestone details, including certificate promotions.                                                                          | Explain earned progress and the next milestone. Activity is not certified proficiency.                                                                             |
| 02:32–03:20              | Related search entries; illustrated word pages with pronunciation notation, part of speech, definition, translation, examples and usage guidance.                                 | Deepen meaning-specific teaching and visual context. Keep distinct meanings understandable.                                                                        |
| 03:28–04:16              | Note editor and favorite/list-assignment dialogs; share and image-report actions.                                                                                                 | Extend existing notes with personal named lists and simple content reporting. Opening a dialog does not prove persistence.                                         |
| 05:20–06:56              | Range-based knowledge overview; card, tile and list views; Should learn/Already knew choices update visible state. Map copy describes relevance ordering.                         | Connect overview to browsing and explain real personal states. Do not invent frequency ranks, vocabulary estimates or relevance evidence.                          |
| 07:04–07:20              | Known, to-learn and started collections.                                                                                                                                          | Distinguish self-reported knowledge, learning intention and recorded activity.                                                                                     |
| 07:53–08:54              | Learn/Skip and Learn now/Learn later/Already knew choices; definition, usage tips and common combinations; next-word action and optional other meanings. Some tips are paywalled. | Give agency and richer teaching. The sampled sequence shows teaching completion, not a recall assessment; preserve Vocanova grading and explicit saving.           |
| 09:04–09:12; 10:24       | Conversation/speaking entry points show a free-plan limit; microphone control is visible.                                                                                         | Full roleplay, recognition and pronunciation grading are not established by these samples.                                                                         |
| 09:20–10:16              | Topic writing returns successive tips: original text highlighted, suggested change and reason, with Previous/Next/Start over controls.                                            | Connect corrections to original wording and enable a rewrite. Topic writing complements target-word sentence practice. One correction is not a quality evaluation. |
| 10:32; 12:00             | Scenario, idiom, phrasal-verb and word-list entry points; custom lists, topics and exams.                                                                                         | Add useful organization and practical content in bounded increments. A tile does not prove its full course.                                                        |
| 11:12–11:44; 12:08–12:24 | Accent, translation, reminders, challenge types, appearance and sync controls.                                                                                                    | Preferences need visible effects and honest limits. A sync screen proves neither offline conflict handling nor alert delivery.                                     |

## Duolingo observations

| Timestamp                              | Observed experience                                                                                                                                                                 | Vocanova implication                                                                                                                                                  |
| -------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 00:08–00:24; 02:08; 07:36–07:52        | Current activity on a learning path; unit guidebook with useful phrases and audio.                                                                                                  | Keep situation grouping, current step and return-to-learning clear; add compact practical unit references.                                                            |
| 00:40–01:12                            | Sound, animation, motivation, listening and notification preferences.                                                                                                               | Accommodate audio-unavailable and reduced-motion use without penalties; reminder consent stays explicit.                                                              |
| 03:04–06:40                            | A session mixes context choices, missing-word listening, sentence assembly, typed missing words and typed listening. Progress and bottom Check/Continue controls remain consistent. | Bring curated variation into guided lessons; keep stable accessible task controls.                                                                                    |
| 03:28                                  | Prompt word help tooltip.                                                                                                                                                           | Offer contextual help and record assistance rather than claiming unaided recall.                                                                                      |
| 05:04; 06:44–06:56                     | Correction after an error; later review of missed exercises with Previous mistake labels before completion.                                                                         | Vocanova already requires a correct retry before continuing. An end-of-session retry pass is an alternative design needing durable attempts and clear accuracy rules. |
| 04:16; 07:04–07:20; 09:10              | Encouragement, completion, streak/quest updates; sounds lesson reports accuracy and XP.                                                                                             | Give a concise truthful summary and next action; separate reward activity from proficiency and limit compulsory celebration screens.                                  |
| 08:00–09:04                            | Word discrimination, same/different listening and audio-to-word matching pairs.                                                                                                     | A small curated sound course needs reliable audio and reviewed contrasts. Playback alone is not sound training.                                                       |
| 09:20–12:40                            | Practice hub opens a full repeatable listening session with sentence assembly; targeted practice and mistakes are visible.                                                          | Make modes discoverable; keep confirmed progress and first-attempt results.                                                                                           |
| 12:48–15:20; 16:40–18:56               | Story library and two dialogue flows reveal lines, expose audio, insert comprehension/phrase questions and return to the path.                                                      | Original short situation stories can connect vocabulary to longer language; require authored questions and resumable grading.                                         |
| 02:32–02:56; 07:28; 15:28; 16:00–16:16 | Leagues, status choices, statistics, achievements, friends and avatar editing.                                                                                                      | Social/avatar systems are later choices with operating costs; personal progress remains central.                                                                      |

## Existing foundations and genuine gaps

The [lesson catalog](../../apps/api/business/lessons/catalog.go) builds nine
steps per lesson: three teaching cards, three meaning recalls, then three context
choices. The [state machine](../../apps/api/business/lessons/lessons.go) requires
a correct answer before continuing and retains first-answer accuracy. It has no
end-of-session retry pass. Separate practice already offers typed recall,
listening choices and mistakes; varied modes are not wholly missing.

[Word detail](<../../apps/web/src/app/(app)/_components/word-detail-content.tsx>)
already provides definitions, part of speech, authored examples/usage notes,
device audio, explicit saving and per-meaning knowledge/notes. Saved-word sentence
feedback already offers corrections, tips, revision, another attempt and history.
Lesson completion already lists practised words and first-answer accuracy, with
lesson-specific practice, next-lesson and saved-review links. The knowledge map,
filtered collection, personal plan, starting check and private achievements also exist. Extend these foundations instead of duplicating them.
No implemented named custom-list, original story-course, minimal-pair sound-course
or speech-grading feature was found in the inspected routes and domains.

## Prioritized maturity work

These are recommendations derived from the recordings and code, not shipped
features. Follow the owner's feature-first direction before a dedicated polish
pass.

1. **Connect the vocabulary loop.** Add explicit per-meaning Save for review
   actions at lesson completion, preserving and clarifying its existing practice
   links. Add searchable
   requester-owned named lists usable as practice selection. List membership,
   known status and SRS saving stay separate. Acceptance: reload/device durability,
   exact retries, no silent restoration of removed words, export/deletion,
   empty/error/conflict recovery.
2. **Deepen word teaching.** Extend existing usage notes with reviewed common
   combinations, contrasts and mistakes. Add optional meaning-specific translation
   and original/licensed visual context where useful. Acceptance: editorial and
   licensing evidence, correct meaning alignment, missing-asset fallback,
   optional media and provider-independent readable content.
3. **Vary guided sessions and completion.** Bring existing typed/listening modes
   into lessons and add curated sentence assembly/matching where useful. Evaluate
   delayed mistake retry against current immediate retry. Extend the existing
   words/first-attempt completion summary with corrections and a clearer
   recommended next action. Acceptance: explicit
   help/reveal, audio-unavailable recovery, server grading, versioned snapshots
   preserving old sessions, resume/concurrency/replay checks and no fake credit.
4. **Add original mini-stories and unit references.** Start with a small complete
   A2–B1 dialogue set, line audio, vocabulary help, comprehension checks and
   guidebooks. Acceptance: reviewed questions/distractors, durable resume and
   completion, accessible transcript/audio, return to path and learner usefulness
   before enlarging the library.
5. **Broaden productive practice.** Add topic writing on existing correction,
   reason and rewrite controls. Later sound discrimination and bounded scenario
   conversation require independent content/audio/cost-quality gates. Acceptance:
   real provider availability and feedback quality, original draft preservation,
   safe replay, clear failures and predictable low cost. No speech-score promise
   without validated assessment.

Then polish the integrated journey: consistent controls, clear map legends,
comfortable mobile reading, keyboard access, dark mode, optional motion and short
celebrations. Verify 360px, 430px and desktop, deployed flows and learner feedback.
Technical checks alone do not prove learning value.

Friends/leagues, certificates, subscriptions, exams, avatars, extensions and
offline sync remain separately evaluated opportunities. Their visibility in a
reference does not automatically make them requirements. Do not copy competitor
text, characters, illustrations or video assets into Vocanova.

## Working rule

Before changing a relevant screen, revisit the reference segment and current
Vocanova behavior; identify the learner outcome being improved. Record changes
to priorities and evidence. Implement and verify on one delivery branch for a
consolidated PR under repository instructions. The full product goal remains
unfinished.
