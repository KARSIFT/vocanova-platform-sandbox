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

The server grades the answer and records the real Leitner result. Feedback shows
the resulting level and persisted next-review date in the device's local calendar.
Beside Continue, the learner
can keep automatic scheduling, choose 1, 3, 7, 14, 30, 60 or 120 days, or select
**I know this word**. An explicit date changes scheduling only. It does not change
the answer grade, advance the box again, add an attempt or award extra credit.
Self-reported knowledge stays distinct from tested recall.

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
retains its separate delayed mistakes. Discovery's native history update preserves
opaque browser/router state and notifies the focused exit link, including initial
mount and Back/Forward. The selected 100-word range survives leaving the session.
Teaching steps reuse the existing original meaning pictures; graded questions
exclude those pictures and their target identifiers.

## Pronunciation

Speech controls share one owned playback controller. Default speed is 0.85,
extra slow is 0.7, and normal speed/device voice selection are under **Voice &
speed**. Only these preferences persist locally. Natural/enhanced device voices
rank above basic voices when available.

Default authenticated playback uses the existing configured Cloudflare AI
service, Deepgram Aura 2 English (`luna`). Only exact active canonical word text
or reviewed starter examples, up to 500 characters, may be synthesized. Learner
drafts, notes and arbitrary input are rejected. Shared public MP3s are cached by
text/model/voice. A database-enforced cap reserves at most 1,000 new characters
per UTC day per environment before a paid call. Ambiguous failures are not
automatically retried or refunded. A cache-key lock prevents duplicate generation
across midnight. At the currently documented [Cloudflare price](https://developers.cloudflare.com/workers-ai/models/aura-2-en/),
the cap is roughly $1/month per environment.

Disabled AI, provider/quota/network failure or unsupported audio falls back to
the device voice. Explicit device voice selection bypasses generation. The UI
labels the actual source. Stop, replay, navigation and another control cancel
owned playback; late audio cannot interrupt the next word. Browser autoplay
denial is respected. Synthetic clips prove control behavior only; real provider
availability, decoding and perceived voice quality require staging evidence and
listening feedback.

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
