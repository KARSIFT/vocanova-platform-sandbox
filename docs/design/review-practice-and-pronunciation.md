# Review, personal practice and clearer pronunciation

## Learner outcome

Home offers two independent activities: **Review words** recalls due saved
meanings one at a time; **Today’s practice** shows short mixed lessons along a
personal path. Practice completion is recorded study, not proof of proficiency.
Difficult free-form writing remains optional. Search and the existing three
primary destinations remain available. The practice read streams independently,
so a slow secondary request cannot delay Home's review card. Loading and failure
states do not invent zero progress.

The owner's latest seven-minute recording was visually sampled across its full
duration. It showed the activities being conflated, repeated review attempts,
and discovery losing a selected range on exit. This supplements the earlier
WordUp/Duolingo review; it is not continuous playback or an audio assessment.

## Review words

The owner's 18 October **Vocanova 3** recording (6:14) was visually sampled
throughout at ten-second intervals. At 0:10–0:30, the native-language choices
omit Persian. At 4:50–5:10, an incorrect answer offers another attempt in the
same review; at 5:50, that word returns. At 5:20–5:40, the plain open dropdown
and Continue have mismatched alignment. These are visual observations, not an
assessment of the recording's audio.

The server grades the answer and records the real Leitner result. Feedback shows
the resulting level and persisted next-review date in the device's local calendar.
Beside Continue, the learner
can keep automatic scheduling, choose 1, 3, 7, 14, 30, 60 or 120 days, or select
**I know this word**. An explicit date changes scheduling only. It does not change
the answer grade, advance the box again, add an attempt or award extra credit.
Self-reported knowledge stays distinct from tested recall.

Each canonical word finishes review for the learner's local day after any
correct, incorrect or explicitly skipped attempt. Other meanings of that word
can appear on later days; mixed personal practice remains independent. A wrong
answer moves back exactly one level with a floor of zero: a three-day interval
becomes one day. There is no consecutive-error reset. All automatic outcomes
wait at least 24 hours, including the lowest box. Existing audit and due dates
are preserved; eligibility uses server-created audit time and separate IANA
local-midnight boundaries, including daylight-saving transitions. Re-saving a
word, another tab, a restored old session or a client timestamp cannot create a
second review that day. Exact retry receipts remain immutable.

Continue and Next review share one aligned action row and matching control
heights. The schedule opens a styled menu with date choices and a checked
selection. Selection remains a draft until Continue. Arrow keys, Home/End,
Escape, Tab, outside click and focus return are supported, including narrow
mobile and dark mode. The closed control shows the automatic date without
repeating the menu's longer label.

Queue advancement, explicit scheduling and the retry receipt commit together.
Changes from another tab are rejected rather than overwritten. Legacy feedback
without scheduling metadata continues with its original automatic behavior.
Completion waits for the authoritative remaining queue before offering another
set or discovery. Practice is available after review too.

## Today’s practice

The path displays durable completed, current and upcoming lessons, resumes the
active lesson and uses exact retry receipts. Its daily count is independent of
reviewed words. Listening/meaning choices combine with authored gap filling and
tap-to-build sentences where reviewed examples exist. Familiar words come from
explicit knowledge, successful review or completed study; occasional new words
have a teaching step before questions.

The full catalog supports definition/listening exercises. Rich context and
sentence exercises depend on the smaller reviewed example catalog; this does
not establish 10,000 editorially certified lessons or pictures. Feedback, token
selection and first-answer accuracy survive reloads. Assisted or revealed answers
do not become unaided first correct answers. Practice never writes SRS grades,
review credit or mastery. It uses immediate correction/retry; the review queue
defers mistakes to a later day. Discovery's native history update preserves
opaque browser/router state and notifies the focused exit link, including initial
mount and Back/Forward. The selected 100-word range survives leaving the session.
Teaching steps reuse the existing original meaning pictures; graded questions
exclude those pictures and their target identifiers.

## Registration language

The native-language picker offers 40 named languages, with Persian / فارسی
among the initial choices. Search matches English names, native names and the
Farsi alias, storing the established language code. Languages outside the list
remain valid text entries. Going Back and the final summary show the readable
name rather than an opaque code. The suggestions are keyboard-operable 44px
buttons with a labelled group, including the empty search result state. This is
a self-reported native-language preference, not a promise of translated app UI.

## Pronunciation

Speech controls share one owned playback controller. Default speed is 0.85,
extra slow is 0.7, and normal speed/device voice selection are under **Voice &
speed**. Only these preferences persist locally. Natural/enhanced device voices
rank above basic voices when available.

Meaning-specific authenticated playback uses OpenAI's pinned
`gpt-4o-mini-tts-2025-12-15` model and `marin` voice. Server-owned guidance includes
the actual selected part of speech, definition and reviewed context. The input
contains only the word or exact approved example to speak. Nouns, verbs,
adjectives, adverbs and other catalogued uses share this contract; present/past
`read` have distinct meanings and guidance. It does not invent forms absent from
the catalog. See the [official speech guide](https://developers.openai.com/api/docs/guides/text-to-speech).

The initial supplement adds 19 original uses across 12 existing words. Canonical
word and primary meaning IDs, the 10,000-entry rank map and learner state stay
unchanged. Personal practice randomizes equally recent eligible meanings once
when saving the session, preserving least-practised priority. Unfamiliar sibling
uses enter explicit teaching at the existing teaching interval. Review continues
to follow its due schedule and grades the exact selected meaning.

Requests carry a meaning ID. Saved lesson questions also carry a revision of the
frozen spelling, POS and full definition. Changed live content fails closed before
generation; the client rejects a mismatched meaning/revision before playback.
Historical incomplete snapshots retain compatible optional metadata. Examples
must belong to the requested meaning and pass the reviewed-content boundary.
Learner drafts, notes, arbitrary text, client instructions and voice overrides
are rejected. A noun `resume` may use the reviewed `résumé` spelling; the verb
does not inherit that alias.

Shared public MP3s persist by exact meaning/context, provider, model, voice and
instruction version. Database reservations bound new generation to 20 calls and
1,000 input characters per UTC day per environment. Provider reference guidance
is separately bounded. Failures consume the reservation and are not automatically
retried or refunded. A cache-key lock prevents duplicate generation across
midnight. These are request limits, not a promise of a particular monthly bill.

Fresh output must pass a bounded complete MP3 frame/decode check and a conservative
signal-presence threshold before it becomes a reusable clip. The check caps
compressed input at 1 MiB and decoded content at 60 seconds, processing PCM with
an 8 KiB buffer. It rejects malformed, truncated, silent and near-silent output;
it does not establish correct words, stress or naturalness. Decoding occurs once
before persistence, so cached replay does not repeat that work. The v2 instruction
identity separates these validated clips from earlier outputs. The historical
Cloudflare hash also changes with the new meaning-aware identity; old rows are
retained but do not warm the new provider's cache. No cache deletion is needed.

Meaning-specific requests cannot fall back to a guessed device pronunciation or
the legacy text-only provider. Failures offer retry without changing learning
progress. Context-free speech retains device fallback and voice selection. The UI
identifies AI pronunciation. Stop, replay, navigation and another control cancel
owned playback; late audio cannot interrupt the next word. Browser autoplay
denial is respected. Synthetic clips prove control behavior only; real provider
availability, decoding and perceived voice quality require staging evidence and
owner listening feedback. Decoding alone cannot establish correct stress or tense.

Staging receives the existing GitHub `OPENAI_API_KEY` as a dedicated speech key;
text feedback configuration stays independent. Missing credentials stop deployment
before rewriting configuration. The manual production workflow has the same
speech binding for a future authorized promotion; this task does not dispatch it.
An opt-in same-repository PR label
`audio-audition` generates at most six fixed public samples, stops on the first
failure, and preserves the first results in a seven-day artifact. Routine CI,
fork PRs and merge-queue runs do not call the paid speech provider.

## Sign-in and data

**Keep me signed in** defaults to the existing persistent behavior. Unchecking
creates a browser-session cookie and a server session lasting at most 24 hours.
Browser session restoration can preserve the cookie, so closing a browser is not
promised to sign out. The choice is bound to single-use OAuth state; callback
query parameters cannot override it. No provider/tenant configuration changes.

Additive migrations retain historical sessions. Export schema 1.7 exposes new
practice fields and daily-review history through the explicit download DTO,
excluding private snapshots, grading maps, receipts and credentials. Historical
session `creditedToday` is its saved snapshot, not a newly calculated current-day
total. A journey with no new words exports zero; unrelated historical practice
retains nullable journey fields. Shared
pronunciation cache contains public content and no customer identities.

## Delivery

One consolidated PR targets main with normal review, protected checks and merge
queue. A merge increments the release version and deploys staging. Production
promotion is outside this phase. Local database, synthetic browser, hosted and
live staging evidence are recorded separately in [release readiness](../product/release-readiness.md).
