# Review and discovery feedback — 15 November 2026

Status: approved by the owner on 15 November 2026 for one PR, merge and deployment.
Implementation and acceptance are tracked in the linked plan and product record;
this specification alone does not establish release success.

## Intent and decisions

Implement the owner's ten annotated production comments as one coordinated
delivery. Preserve the Warm guided coach direction, existing components/tokens,
three main destinations, server grading, durable sessions and idempotent actions.

The owner explicitly chose:

- Pictures and other meanings appear after checking an answer.
- Questions also vary: four picture options, text options, and sentence gaps.
- Today's review goal counts distinct reviewed words, including wrong answers.

The following details are proposed builder decisions, not additional owner quotes:
sentence gaps use four word choices; explicit skips do not earn goal credit;
discovery position persists per account across devices; success-dependent answer
XP remains separate from goal/streak completion.

## Approaches

1. **Presentation-only repair:** correct Home labels, move attribution and improve
   discovery controls. Small, but cannot deliver goal-sized sessions, effort-based
   completion or varied review questions.
2. **Extend the existing daily-review domain (recommended):** add an authoritative
   daily progress view, compatible mixed question snapshots, post-answer teaching
   and durable discovery position. Reuse practice's reviewed gap-building patterns
   without making practice and SRS review the same activity. This covers the
   owner's requests while preserving scheduling ownership.
3. **Replace review with the practice player:** requires reconciling incompatible
   retry, credit and scheduling behavior. No replacement is proposed.

The owner's responses confirm the requested mixed-review/effort-based behavior;
this document specifies approach 2 for review before implementation.

## Entry and attribution

Authenticated visits to `/` resolve through the existing identity and onboarding
rules: completed onboarding goes to Home; unfinished onboarding goes to setup.
Anonymous visitors retain the marketing page. Cookie presence alone is not proof
of authentication. Expired sessions retain sign-in recovery; API outages must not
cause redirect loops or expose cached private content. Sign-out/error bypasses
on the existing login routes remain intact.

Remove the full Word sources disclosure from the review player. Preserve catalog
attribution in API responses and its existing Journey location; provide a compact
route to a stable source/notice location from Help if needed. Do not delete source
licenses or provenance from distributed content.

## Daily goal and completion

The API supplies one daily-review view to Home, entry, player and completion:
learner-local day, frozen daily target, counted words, remaining allowance,
active-session status, eligible due count and next actual scheduled review time
when available. The exact field layout is left to implementation.

A committed, server-graded answer counts one canonical word toward the local-day
goal whether correct or wrong. Correctness, first-answer accuracy, SRS box changes
and success-dependent answer rewards remain truthful and separate. A wrong answer
must not become a successful recall record. Explicit skip, removed/known records,
duplicates, hints/reveals and invalid/conflicting actions do not create effort
credit. Optional pronunciation is not a hint. Multiple meanings of the same word
cannot earn multiple daily credits. Commit attempt, schedule, new goal credit,
mission completion and immutable receipt atomically.

Meeting the effort-based goal completes the review mission and its existing
completion/streak behavior once; answer-specific XP still follows successful
recall. New credit must not be inferred from client counters or a lesson's total.
Use a separate durable, uniquely constrained effort record rather than changing
the meaning of historical success records. Existing daily eligibility prevents a
word from being reviewed repeatedly on the same day.

Historical completed days/rewards/streaks are not rewritten. For the current day
at rollout, reconcile qualifying existing server-verified answer attempts by
canonical word using the established learner-local day; ensure reconciliation
cannot duplicate an already awarded mission reward. The implementation must
document and test its cutover behavior.

New sessions contain up to the remaining daily allowance, replacing LIMIT 10.
If today's target is 15 with none counted and enough due words, start with 15.
If 5 are counted, start with at most 10. Existing snapshots retain their question
identity and frozen denominator. Reconcile already open sessions against today's
allowance without losing pending feedback or replay receipts; excess work is
optional, never presented as mandatory daily work.

Once the target is reached, Home and completion show a tick and “Today's review
complete.” Discovery/practice becomes the primary next action. Additional review
is an explicit secondary choice and never restarts the required goal. Returning
to `/review` must respect that state rather than silently opening another set.

If fewer words are due, state “All available reviews finished” and the actual
count; do not fabricate target completion or re-review a word to fill the gap.
The next opportunity is described using the server's actual due time; do not
promise words tomorrow when all are scheduled later. Today's goal still resets
at learner-local midnight. Refresh overnight progress/allowance without replaying
credit or discarding an unresolved action. Preserve the existing frozen daily
target policy when settings change mid-day.

## Relative schedule wording

Closed schedule and options use “Tomorrow,” “In 3 days,” “In 1 week,” “In 2 weeks,”
“In 1 month,” etc. Calculate from the committed next-review timestamp and stored
learner timezone, including DST/month boundaries. Exact dates remain available
as supplementary accessible detail. Preserve existing allowed intervals and
automatic scheduling; selection stays a draft until Continue commits it.

## Mixed review questions

Select and freeze question kind, opaque option IDs/order, content revisions and
private answer once when creating a session. Reload/retry must not reroll it.
Old snapshots remain valid text-definition questions. All kinds still assess
one due meaning and produce one attempt and one scheduling outcome.

- **Text definition:** existing headword and four definition options.
- **Picture choice:** headword and four equally presented pictures representing
  four distinct candidate meanings. Use approved, exact-revision assets with
  useful alternatives and an unambiguous correct option. A single target picture
  above text choices is not this mode. Use meaningful option descriptions for
  screen readers, opaque choice IDs, visible keyboard focus and text alternatives.
  Do not expose a correct-option label in pre-answer UI. Each choice also freezes
  an equivalent definition so a failed image can offer an explicit switch to all
  text options, preserving option IDs and grading. No image is required to finish.
- **Sentence gap:** an existing reviewed example with the target replaced by a
  blank and four candidate words. Do not display or pronounce the missing answer
  separately before grading. Candidate construction must rule out multiple
  defensible answers, duplicates and synonymous aliases.

Use eligible kinds in a varied sequence without promising a fixed distribution.
Do not add duplicate questions solely to reach a mode quota. If no vetted
sentence or suitable four-picture set exists, use text definitions. Content
eligibility must be established in source-backed tests/editorial fixtures, not
assumed from the presence of any image/example. No new live AI-generated
questions or bulk image generation is part of this delivery. Include real
approved catalog examples of each new kind in release evidence; synthetic-only
mode coverage is insufficient to claim production availability.

## Post-answer teaching

After a matching server-confirmed correct or wrong answer, show the frozen graded
definition, its approved matching picture when available, and optional “Other
meanings of [word]” with actual canonical meanings and their applicable pictures.
Distractors are never relabelled as alternate meanings. If only one meaning is
available, say so without inventing others.

Reuse MeaningPicture/revision validation and existing canonical detail APIs.
Current alternate content may differ from the frozen question; it must not
overwrite the answer originally graded. Fetch optional details after feedback,
key them by session/step/meaning and ignore stale responses after Continue.
Image/detail failures never block Continue, lose feedback or alter grading.
Keep one concise feedback announcement, not a live announcement of the whole
asynchronous gallery. Removed/skipped/changed-elsewhere feedback must not teach
an unrelated current card.

## Discovery

Keep range and word selection above the card. Move Previous word beneath the
card as a quiet secondary control, visually separate from Learn/Already know.
Preserve cross-band movement, explicit revisits, keyboard order and navigation
locks during uncertain saves. Do not add knowledge mutations to navigation.

Each range-picker row shows the numeric range, a miniature 100-word status strip
and readable aggregate progress. Reuse existing band wordStates/counts and the
map's compact drawing logic. Each row is one accessible selection, not 100 tiny
tab stops; text conveys states without relying on color. Preserve legacy-data
fallbacks and self-contained SVG pattern definitions. Search, selected-range
scrolling, Escape and focus return continue working.

Persist last successfully displayed discovery position as a bounded account
preference: catalog version, band/rank and active/revisit intent. Generic discovery
entry resumes it; explicit valid links take priority. Preserve exact revisits of
already saved/known words. Re-read current server state and validate catalog/rank
before resuming. Isolate accounts, include export/deletion implications, and
debounce/coalesce navigation writes so opening discovery stays responsive.
Persistence failure leaves navigation usable with honest retry behavior, without
writing learning progress. Concurrent tabs need deterministic last-confirmed
position handling and stale writes must not override newer confirmations.

## Acceptance and delivery

One delivery branch and consolidated draft/hold PR; normal review, CI and merge
queue, then strict staging and separate manual production deployment. No release
is claimed by this design. Root coordinates builds, Git and deployment; independent
implementation/review owns bounded non-overlapping files after approval.

Required evidence includes PostgreSQL transaction/replay/concurrency/local-midnight
checks for effort credit and mission/streak/reward behavior; goal 15 vs batch 10;
wrong/skipped/removed cases; fewer due words; overnight/legacy sessions; all three
question kinds and content fallbacks; no pre-answer teaching leaks; correct/wrong
post-answer teaching; optional fetch failures/stale responses; authenticated root
entry; discovery exact-word resume across accounts/devices and explicit links;
mini-grid/count agreement. Verify affected views at 360px, 430px and desktop,
both themes, keyboard, 200% text and error/empty states. Use actual stage/provider
and live identity/health evidence for release; synthetic browser results do not
prove learner effectiveness, provider quality or actual screen-reader speech.
