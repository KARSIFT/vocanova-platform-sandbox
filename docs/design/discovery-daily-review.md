# Discovery and today's lesson

The owner’s 17 October 2026 direction replaces the everyday reveal/rate review
and difficult writing prompts with a connected word-discovery and recall loop.
One integrated delivery will be verified and deployed to staging. Production
promotion is a separate decision.

## Recording evidence

The private `Opponent/vocanova.mp4` recording lasts 2:15. Visual samples every
six seconds show the current dark mobile-width experience. Around 0:30–1:00,
review reveals a definition and asks for Again/Hard/Good/Easy. Around 1:18–1:48,
Start lesson opens a separate situation lesson with small instructions and
controls. Journey and the saved-word spotlight still advertise writing. These
observations do not establish physical sound quality or every transition.

Fresh inspection of WordUp at 5:20–6:56 confirms **100 words per map range**,
with displayed ranges extending to 10,000 words, plus cards/list views and
Should Learn / Already Knew decisions. The recording claims personal relevance
but does not reveal a ranking formula. Duolingo’s question flow uses one task,
visible progress, Check, feedback, Continue, and a later mistake retry. These
interaction patterns inform original Vocanova screens; competitor assets and
account information remain private and are not copied.

## Intended experience

Home leads with the active daily lesson or words due today. When none are due,
it offers discovery. The unrelated static lesson recommendation and writing
prompts no longer compete with this daily action. Situation courses and stories
remain supplementary learning routes.

The knowledge map presents the complete, real ranked inventory in 100-word
ranges, each with confirmed known/learning/new coverage. Selecting a range shows
its words; discovery presents one word, pronunciation, a short meaning and two
decisions: Learn or Already know. Decisions persist before advancing. Revisit,
changed decisions, retry and browser return must work. An assessment is not
graded mastery. Existing notes, saved references and review history survive.

Guided discovery skips assessed words and continues to an unfinished range.
When every word is assessed, it shows completion rather than restarting the
first range. Deliberate links from map tiles still open individual words for
revisiting or changing a choice.

The broad inventory must come from reproducible licensed sources, with stable
word/meaning identities and independently reviewed common-word senses. Existing
authored teaching and phrases retain their IDs and content. Source frequency is
distinct from the contiguous map position; the product must not imply personal
relevance, universal frequency or A2–B1 grading for general reference entries.
Missing pictures remain absent rather than using an unrelated illustration.

Today's lesson selects actual due saved words, in bounded resumable batches,
and asks for meanings one at a time. The server owns choices, correctness,
progress and scheduling. Mistakes receive a correction and a later retry. A
correct answer counts a meaning once for the learner's local day. Skips,
reveals, wrong answers, replay and another tab must not inflate daily credit.
First-answer accuracy remains separate from eventual success.

The intended eight-box schedule is 10 minutes for relearning, then 1, 3, 7, 14,
30, 60 and 120 days. Correct answers wait at least a day; existing stored due
timestamps are not rewritten. Daily sessions use the existing scheduler and
reward transaction rather than creating a second progress system.

Automatic pronunciation follows the discovery Start action and each displayed
word. Replay, mute, unavailable voices and browser playback restrictions have
visible recovery. Sound failure does not block text learning or award credit.
No free-form writing is required in the primary daily loop.

## Visual direction

Keep the existing blue identity: primary `#1d4ed8`, pale blue `#eff6ff`, text
`#171717`, secondary text `#525252`, surface `#ffffff`, background `#fafafa`;
use the existing theme tokens for dark equivalents. System typography keeps
body/choice text at a comfortable 16px or larger. The map is the distinctive
overview; the lesson is quiet, with a large word, full-width choice targets and
a stable bottom action area. Preserve Home, Journey and Progress navigation;
focused learning hides unrelated navigation and search. Avoid duplicate help,
tiny captions, decorative panels and color-only feedback.

## Acceptance gates

- Every map range is backed by real catalog entries; all entries are reachable
  and ordered reproducibly, with source notices and honest scope.
- Discovery decisions, notes/history preservation and concurrent changes are
  verified against PostgreSQL; uncertain writes do not advance the card.
- Daily selection uses due words, including a single-word queue and backlogs.
  Correct/incorrect/skip/retry, lost responses, concurrency, midnight and stale
  saved state have server tests. Export and deletion cover new learner records.
- Actual rendered 360px, 430px and desktop screens are inspected in both themes,
  with keyboard, long content, loading/error and unavailable-audio states.
- Relevant unit, database, browser, accessibility and performance checks pass;
  mock fixtures do not establish real database or live-provider behavior.
- One consolidated reviewed PR passes applicable hosted checks and the protected
  queue; staging verifies the merged identity and real discovery/daily credit.

This document states the intended delivery. Code and fresh acceptance results,
not this plan, establish completion.
