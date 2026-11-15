# Review and discovery feedback delivery — 15 November 2026

Implements the owner's ten annotated production comments and subsequent choices
for mixed review questions and effort-based goals. The approved
[design](../superpowers/specs/2026-11-15-review-discovery-feedback-design.md) and
[plan](../superpowers/plans/2026-11-15-review-discovery-feedback.md) describe the
contracts. This is implementation evidence; exact hosted CI, merged identity and
deployment outcomes belong to the consolidated delivery PR and release handoff.

## Learner behavior

| Owner feedback | Result |
| --- | --- |
| Signed-in landing page still says Sign in | Root identity routing sends verified learners to Home or unfinished onboarding; guests retain the landing page. |
| Word sources interrupt review | Full attribution stays on Journey rather than the exercise; source/license data is retained. |
| Teach meanings and pictures after an answer | Checked feedback retains its frozen definition and loads exact-revision artwork plus actual optional alternate meanings without blocking Continue. |
| Show relative review intervals | Tomorrow, days, weeks and months use the saved schedule and stored learner timezone; exact dates remain accessible. |
| Review repeats after the daily goal | A canonical word counts once per local day after a checked answer, including wrong answers. Daily completion shows a tick, makes discovery primary and keeps additional review optional. |
| Goal 15 still produces a 10-word set | New required sets use the remaining confirmed allowance; fewer eligible words are described honestly. Legacy sessions retain frozen identities and denominators. |
| Previous word is mixed with range selection | It is beneath the word card as a secondary control, retaining cross-range navigation and save locks. |
| Range picker is too sparse | Rows reuse a compact 100-word preview and readable counts; each row is one selectable control with aggregate fallback for older responses. |
| Discovery forgets the last word | Requester-owned CAS position storage resumes exact words on generic entry, including known/saved revisits. Explicit links win; failures/conflicts expose retry without false saved confirmation. |
| Vary question types | Frozen four-picture, definition and sentence-gap choices share one server-graded attempt/schedule; unavailable content falls back to text. |

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

- Baseline: 302 web/helper tests and focused API packages passed before changes.
- New behavior was reproduced RED before the relevant fixes: authenticated root,
  additional-review intent, position transport, export DTO, review/discovery UI,
  midnight Home refresh and persistence confirmation races.
- Independent task review found and verified fixes for picture identity leakage,
  answer-position bias/public-ID derivation, false discovery save confirmation
  and Home's stale overnight snapshot.
- Full local API suite with disposable PostgreSQL 16: **2,753 test/subtest
  events passed, zero failures/skips**, 33 packages; explicit local Go 1.26.10.
  Exact repository Go 1.26.5 remains a separate hosted CI gate.
- Final production web build passed. Initial final Chrome matrix: **60 passed**,
  zero failures/skips/retries across 360px, 430px and desktop, including both
  themes, keyboard/200% text, unfiltered axe scans and screenshots.
- Earlier compatibility run: 189 passed, nine obsolete label assertions failed;
  those exact assertions were updated to the approved labels and all nine passed
  in a 15-case follow-up that also passed six visual/axe walkthroughs. No timeout
  increases or rule exclusions were introduced.

- Final client tests: **63 passed**; web/helper tests after staging-harness extensions: **352 passed**; application
  and browser TypeScript checks passed. Workspace, formatting and web/package lint
  checks passed. Foundation checks: **333/337**, with the four documented macOS
  GNU-sed/Bash-coproc/browser-fixture failures; the new bounded inventory checks
  passed 13/13 and Linux CI must pass the full harness.
- Firefox initially passed 58/60 cases and exposed focused feedback partly behind
  the fixed action rail at 360px. A geometry regression reproduced the defect;
  feedback now focuses without native scrolling and centers in view. Both failing
  themes passed the focused regression with unfiltered axe and native keyboard
  clearance checks. This is a source repair, not an axe exclusion or scan-position
  normalization. The full repaired Firefox matrix subsequently passed 60/60 with zero skips or retries. Other final engine checks remain separate evidence.

Remaining exact-head engine, hosted CI/review and rollout results are recorded by
the consolidated PR. Synthetic fixtures establish UI interaction only. They do
not prove provider quality, real account authentication delivery, learner
effectiveness or actual screen-reader speech. Database tests establish the
tested persistence/transaction properties on a disposable local database, not
production durability.

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
A real-link prefetch regression completed three local Chrome repeats: after a
second tab confirms a newer position, clicking the prefetched generic link reads
fresh data. This bounded negative reproduction does not establish all cache cases.

Focused rebuilt browser verification of this repair wave passed 17/17; full
post-wave engine and hosted checks remain separate release evidence.
