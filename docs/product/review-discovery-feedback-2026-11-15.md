# Review and discovery feedback delivery — 15 November 2026

Implements the owner's ten annotated production comments and subsequent choices
for mixed review questions and effort-based goals. The approved
[design](../superpowers/specs/2026-11-15-review-discovery-feedback-design.md) and
[plan](../superpowers/plans/2026-11-15-review-discovery-feedback.md) describe the
contracts. This is implementation evidence; exact hosted CI, merged identity and
deployment outcomes belong to the consolidated delivery PR and release handoff.

## Learner behavior

| Owner feedback                              | Result                                                                                                                                                                                           |
| ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Signed-in landing page still says Sign in   | Root identity routing sends verified learners to Home or unfinished onboarding; guests retain the landing page.                                                                                  |
| Word sources interrupt review               | Full attribution stays on Journey rather than the exercise; source/license data is retained.                                                                                                     |
| Teach meanings and pictures after an answer | Checked feedback retains its frozen definition and loads exact-revision artwork plus actual optional alternate meanings without blocking Continue.                                               |
| Show relative review intervals              | Tomorrow, days, weeks and months use the saved schedule and stored learner timezone; exact dates remain accessible.                                                                              |
| Review repeats after the daily goal         | A canonical word counts once per local day after a checked answer, including wrong answers. Daily completion shows a tick, makes discovery primary and keeps additional review optional.         |
| Goal 15 still produces a 10-word set        | New required sets use the remaining confirmed allowance; fewer eligible words are described honestly. Legacy sessions retain frozen identities and denominators.                                 |
| Previous word is mixed with range selection | It is beneath the word card as a secondary control, retaining cross-range navigation and save locks.                                                                                             |
| Range picker is too sparse                  | Rows reuse a compact 100-word preview and readable counts; each row is one selectable control with aggregate fallback for older responses.                                                       |
| Discovery forgets the last word             | Requester-owned CAS position storage resumes exact words on generic entry, including known/saved revisits. Explicit links win; failures/conflicts expose retry without false saved confirmation. |
| Vary question types                         | Frozen four-picture, definition and sentence-gap choices share one server-graded attempt/schedule; unavailable content falls back to text.                                                       |

## Assessment and content boundaries

Effort credit is separate from successful recall. Wrong answers count toward the
goal and retain their actual SRS result; successful-answer XP and accuracy remain
success-dependent. Skips/hints/removed words do not create effort credit.
Canonical identity, unique constraints, server timestamps and transactional
receipts prevent duplicate credit across meanings, retries and tabs. Only the
current day is reconciled at rollout; completed historical days are not rewritten.
Existing one-time mission rewards and streak behavior follow the effort goal.

All three question kinds preserve their private answer and choice ordering once
created. Choice placement uses private cryptographic randomness, not the public
step ID. Picture/gap pre-answer projections omit target pronunciation identity;
gap prompts also omit the missing headword/slug. Feedback restores the frozen
identity only after grading. Failed artwork offers equivalent text options.

New mixed modes use bounded, explicitly reviewed, revision-pinned content:
four concrete picture meanings and two reviewed sentence contexts. The rest of
the catalog retains definition questions unless another eligible set is authored.
This is not 10,000 reviewed sentence exercises or a promise of a fixed mode mix.
The selected public picture bytes were freshly fetched and SHA-256 matched to
committed approved WebPs. Metadata/revision/semantic eligibility checks still
govern whether a live session uses those modes.

## Verification record

- Full local API suite with disposable PostgreSQL16: **2,753 passing test/subtest
  events, zero failures/skips**, 33 test-bearing packages; explicit local Go1.26.10.
  Hosted Go1.26.5 subsequently passed 33 package checks and 360 targeted PostgreSQL
  checks, with zero failures/skips in the verbose targeted group.
- API client **63** and web/helper **352** tests passed, as did application/browser
  types, lint, formatting and the production build. Linux hosted foundation checks
  passed **337/337**; the local macOS harness retains its four documented
  GNU-sed/Bash-coproc/browser-fixture limitations.
- The final integrated local Chrome selection passed **453/453** with zero
  failures/skips/retries across 360px, 430px and desktop. It includes the complete
  set of observed journey compatibility failures, both themes, keyboard/200% text,
  unfiltered axe scans, fallback fonts and eight window heights. The compiled
  source manifest stayed unchanged during the run.
- Independent task, whole-branch and scoped follow-up reviews cleared the repaired
  findings. Current-head hosted browser counts, retry/scope dispositions,
  Lighthouse reports, merge identity and real release checks are maintained in the
  [consolidated PR](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1527).

Synthetic fixtures establish UI interaction, not provider quality, real account
authentication delivery, learner effectiveness or actual screen-reader speech.
Database tests establish the exercised transaction properties on disposable
PostgreSQL, not production durability. A successful PR alone is not a claim that
the entire product is finished.

## Integrated review follow-up

Final whole-branch review identified and verified repair targets for canonical
no-due recovery, cross-tab account identity, and DST draft wording. Discovery PUT
now requires the independently verified rendered account ID and rejects a
mismatched authenticated requester before any mutation, including retries after
an optional initial preference-read failure. Current and legacy recognized
no-due rejections recover through confirmed reads without reusing a rejected
Start key; uncertain starts retain their exact keys. Draft interval labels now
use the prospective saved duration in the learner timezone, including today/two
local dates later around DST, without changing scheduling values.

The staging counter gate now compares Home with the API's uncapped effort count,
checks the mission's capped count separately, and requires fresh correctly graded
cards even after goal completion. Fourteen seed helpers and five progress helpers
cover strict canonical resolution and an 18-to-20 effort increase with target15.
A real-link regression observes an acknowledged hover prefetch before a second
tab confirms a newer position, then requires a fresh navigation response and the
new word with a confirmed save. It does not establish completed cache
materialization before that update or universal absence of stale-cache cases.

The first hosted synthetic recovery rehearsal exposed a schema-dump comparison
failure in the new discovery constraint: PostgreSQL expands a mixed BETWEEN/AND
expression and flattens its grouping on restore. The new, unmerged migration now
uses equivalent explicit inclusive comparisons; the constraint name, semantics
and strict schema comparator are retained. Isolated RED reproduction captured the
sole two-parenthesis difference. The fresh GREEN dump/restore has identical schema
hashes, matching rows/sequences and integrity checks, and verified cleanup. Extra
source/target checks accepted ranks 1/100/101/10000 with correct bands and rejected
0/10001/mismatched bands under the original named constraint. Discovery/migration
Go follow-up passed 129 test/subtest events with no failures/skips. Hosted recovery
subsequently matched all54 tables/schema/sequences and rejected seven injected
faults at their expected stages, with verified cleanup.

The full browser inventory also required older journey tests to adopt exact
reviewed/target counts, goal-aware terminal states, authenticated root routing,
Journey-only attribution and explicit extra-review intent. Their interaction,
count, geometry and accessibility assertions remain active. An acknowledged RSC
prefetch makes the warm navigation test's setup explicit while retaining its
700ms loading criterion; it is not a claim of a new application performance fix.
A selected prefetch response may remain unfinished while a later request succeeds,
so HTTP EOF is not used as the learner-readiness oracle. Discovery checks the
actual fresh navigation, displayed word and confirmed save; cached diagnostics
cannot mask an original timeout.

Actual presentation regressions were repaired too: feedback focus now respects
both persistent bars and complete preceding click targets within real scroll
limits, and text-choice correctness markers retain their own row so 200% text
does not split ordinary answer words. Exact completion dates are text content
inside native time elements, alongside the visible relative label, rather than
an ignored author name; the [W3C naming guidance](https://www.w3.org/WAI/ARIA/apg/practices/names-and-descriptions/)
marks the time role as unsuitable for author naming. A regression checks the
server timestamp, independently formatted exact date and exposed accessibility
tree. No axe rules, relevant assertions or thresholds were disabled.
