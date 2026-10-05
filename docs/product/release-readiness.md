# Vocanova release readiness

## Connected learner experience — integrated verification

The current integrated redesign is described in
[learner experience improvement](../design/learner-experience.md). A fresh web/API
build passes. Formatting, lint, types and 258 foundation, 52 client and 146 web
helper tests pass. Full local validation stops at three existing PostgreSQL tests
because the local Docker engine is unavailable; it is not claimed as a full pass.
The final affected browser matrix passes 149 cases with one existing scope skip
at all three widths, including both themes. The broad run recorded 888 passes,
37 skips and two obsolete Home-label assertions, subsequently repaired and
verified in the affected run. Required hosted database, full accessibility,
performance and container acceptance is recorded on the consolidated PR; local
fixture checks do not substitute for those gates. Prior release evidence does
not accept this new branch. Production promotion remains a separate manual action.

[PR #1487](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1487) contains
the integrated delivery. Its actual automated review prompted presentation
memoization, a disabled-button cue and a working recommendation-recovery shortcut.
A fresh build and 84 focused browser cases pass after those refinements. All
required hosted checks accepted `ea0a9d8`; the full hosted browser suite passed
893 cases with 37 intentional scope skips, and all 12 Lighthouse screen/layout
audits passed. All seven merge-group workflows passed and the normal queue merged
the delivery as `36361f19`. Staging run `37341576653` passed release identity,
core-loop and maturity journeys on rerun. Its first attempt stopped before the
journeys because runner package downloads timed out; it is not counted as a pass.

A late review found a resume presentation defect in correctly graded choice
questions. The separate narrow correction and its fresh evidence are recorded in
the learner-experience design note; earlier staging acceptance does not accept
that newer correction.

## Visual learning phase — prior accepted verification

See [visual learning delivery](../design/visual-learning-delivery.md) for the
artwork phase's acceptance record. Its consolidated PR #1486 merged and passed
applicable hosted checks, including container asset delivery. That evidence does
not accept the newer connected learner experience branch above. A new production
deployment remains separate.

## Consolidated maturity verification — 5 October 2026

[PR #1484](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1484) adds the
connected list, teaching, mixed-lesson, story, guidebook and writing capabilities
in [maturity delivery](maturity-delivery.md). The release candidate is **0.4.0**. It merged on 4 October as
`af3eacb44edacedd4ee0fec0fa9d1f3026e9f742`. The source-check evidence below
precedes real-staging acceptance; the later promotion evidence is recorded below.
Independent root evidence:

- **Fresh validation:** follow-up `pnpm run validate` passed (exit 0): 258
  foundation, 52 client and 135 web-helper tests plus Go suites, formatting,
  lint/Go vet, types and production web/API builds. Generated OpenAPI matches
  its generator. These checks do not accept the live feedback provider.
- **Actual database:** six PostgreSQL packages passed with `-tags=integration`
  and `-count=1`: accounts, lessons, wordlists, practice, stories and `app/api`.
  Forty actual forward migrations and canonical seed were applied to isolated
  PostgreSQL 16. Atlas v1.2.0 library checksum was regenerated/read back;
  historical migration hashes remain unchanged. Direct SQL application is not
  proof of deployed Atlas revision history.
- **Privacy and compatibility:** explicit HTTP export DTOs preserve all six
  learning fields, typed wording and visible false feedback, retain legacy schema
  versions/null collections and exclude private snapshots/receipts. Nullable story
  references, newest session per story key, ownership, concurrent replay, frozen
  list practice and schema 1.5 export/purge are regression-covered. Lesson-save
  conflicts now reread canonical state; an unknown status cannot trigger automatic
  saving. Earlier fixed persistence/replay findings remain covered.
- **Browser checkpoints:** the preceding affected matrix passed 177/177 cases
  across desktop, 360px and 430px, plus three screenshot-only checks. The latest
  full lesson-save, primary-navigation and Home mobile accessibility specs passed
  **53 cases with one existing desktop skip** across all three layouts. All six
  formerly failing route loops and all 24 save cases pass. The route loops lacked
  a fixture session cookie; seeding authenticated state preserves auth guards,
  navigation assertions and keyboard skip-link proof. These fixtures verify built
  interface/transport behavior, not live login, provider grading or production
  durability. The Chrome connector failed twice with a kernel reset; the local
  Chromium harness remains usable.
- **Hosted checks:** application revision `deacbf8c` passed all applicable CI,
  including web/API, controlled sign-in, container, architecture, workflow,
  performance and synthetic recovery checks. Full accessibility
  [run37229666520](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/37229666520)
  passed **767 cases with 37 existing skips**. This supersedes the preceding
  six-failure run; the corrected route-loop tests retain their assertions.
- **Staging credential safety:** the opt-in maturity journey is authored and
  independently reviewed. Direct API checks use bounded Node requests with
  sanitized errors. A deliberate fake-token GET/DELETE failure probe inspected
  eight retained artifacts, including embedded HTML data, and found zero markers.
  Browser types/discovery and 29 deployment/install contracts pass. The new
  source-check checkpoint preceded the real-staging journey; its later successful
  run is recorded below. Static/failure-probe evidence alone is not deployed
  acceptance.

[Staging run 37232617616](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/37232617616)
deployed the merged 0.4.0 revision and passed the exact identity gate, then failed
the core journey on a hidden guide word link selected by an outdated test. The
feedback phase was not reached. Independent review identified another test
selector that can toggle audio speed instead of saving a meaning. Correct these
selectors before promotion; the original maturity journey was skipped. The
focused correction in [PR #1485](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1485)
passed 18 browser regressions across desktop, 360px and 430px, including actual
fixture API state/readback and reload persistence.
[Staging retry 37234463076](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/37234463076)
passed exact identity, the real core journey and the real maturity journey at
`2d222621ac721e1d81dd2d329c08a97c971ffc76`. Its differences from merged `af3eacb44`
are limited to three tests and two docs; deployment inputs and application source
match. This accepts these live flows while retaining the wider gates below.
Production promotion is verified below. Earlier 0.3.1 evidence alone does not
accept PR #1484. Live AI feedback quality/availability,
actual reminder alarms, physical speech playback and observed learner acceptance
remain separate open gates. The older 38-migration/schema 1.4 and smaller local
checkpoints below are historical; the active maturity inventory is 40/schema 1.5.

## Latest release evidence — verified 5 October 2026

[Production deployment 37235094589](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/37235094589)
completed successfully from merged `af3eacb44edacedd4ee0fec0fa9d1f3026e9f742`.
Fresh production web/API `/version` probes agree on **0.4.0**, that commit and
**production**; API health reports service/database **ok**. Exact release identity,
readiness and the normal synthetic smoke gates passed. Independent release review
confirmed that merged application/deployment inputs match accepted staging
`2d222621`; its differences are limited to three tests and two docs. This is source
equivalence, not an assertion that staging and production image digests match.

The consolidated features are deployed with controlled signup. Broad real-provider
feedback quality, physical-device playback, actual calendar alarms and learner
acceptance remain separate evidence requirements. Passing these deployment gates
does not declare the full product goal complete.

## Previous release evidence — verified 4 October 2026

The learning expansion described below was merged in
[PR #1483](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1483) and
[deployed successfully](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/37055589721)
on 2 October. Fresh production web/API version endpoints on 4 October agree on
version **0.3.1**, environment **production** and commit
`73846e2eac6596733bcca1b4a0a7c6a0e4cc2df0`. The final integrated accessibility
workflow also reports success. This supersedes the undeployed status in the
historical checkpoints below.

Real AI availability/quality, actual calendar import/alerts, physical-device audio,
learner usefulness and public legal/support operating decisions remain separate
acceptance work. Deployed version identity does not prove these outcomes. The
[recording review](reference-video-review.md) adds timestamped design evidence and
prioritized maturity requirements; it does not accept Vocanova runtime behavior.

## Historical completion plan

This is the working completion plan for the product requested on 2 October 2026. Vocanova already implements its main learning loop. Finishing it requires reliable behavior, enough useful content and evidence that real learners can use and trust the deployed service. Historical delivery records remain useful context; this checklist requires fresh evidence for a public launch.

Read [the current state](current-state.md), [product bible](00-product-bible.md), [learning workspace design](../design/learning-workspace.md) and [development guide](../development.md) before extending scope.

The owner's later feature-first direction expands this baseline. Use
[feature-complete delivery](feature-complete-delivery.md) for the required
connected feature set and build order. The previous consolidated acceptance
below applies only to that earlier revision; it does not accept the ongoing
lesson, knowledge-map and varied-practice expansion. That work remains on one
delivery branch for a single consolidated PR after integration and verification.

## Active feature expansion checkpoint — 2026-10-02

The delivery branch implements the connected learning expansion described in
[current state](current-state.md#active-feature-expansion--delivery-branch-not-deployed).
It has not been accepted on staging or promoted to production. The earlier
consolidated PR evidence below remains historical and does not cover this work.

- **Curriculum and durable sessions:** the [30-lesson course](starter-curriculum.md)
  targets 90 distinct meanings. All 400 original seed rows and seven original
  lesson definitions are preserved. The inventory is 17 situations, 89 words/
  phrases, 92 meanings, 148 examples and 200 notes. Disposable PostgreSQL 16
  checks passed repeated seed execution, preserved saved references, all guided
  lessons and independent practice, including actual stored version-1 session
  read/list/replay/answer/continue behavior. The expanded catalog uses
  `starter-90-v2`; supported `starter-21-v1` sessions retain their snapshots.
  Unknown content/grading versions still fail closed. The same run checked the
  recovery fixture's invariants after reseeding; it was not a new dump/restore.
- **Learning direction and privacy:** authenticated GET/PATCH
  `/api/v1/learning-preferences` separates current goal/focus from immutable
  onboarding answers. Writes require CSRF and an expected revision; an
  already-applied intent returns current state without another write, and a
  conflicting stale intent is rejected. Migration
  [20261002233000](../../apps/api/migrations/20261002233000_learning_preferences.sql)
  brings the inventory to 38. All 38 forward SQL migrations applied to a fresh
  disposable database, followed by passing preference concurrency/preservation,
  account export/anonymization and current/historical practice checks. This was
  forward SQL verification, not an Atlas revision-history test. Export schema
  1.4 includes the new data. Editing is connected to the plan and starting-word
  check. The later 42-check mock-browser set passed learning-direction
  retry/conflict/CSRF recovery, lesson-specific practice selection and canonical
  search. This does not establish real-account or deployment acceptance.
- **Calendar reminders:** the [optional calendar export](calendar-reminders.md)
  is implemented with validated date/time, a daily floating-local event and a
  display alarm. Six exporter tests pass and a bounded static review found no
  actionable issue. All nine calendar browser checks and 30 affected Settings
  checks passed across 360px, 430px and desktop. They inspect actual downloaded
  bytes, unchanged settings, both themes, keyboard access and preparation failures.
  A real calendar import, correct first
  event/timezone and observed alert are still required for calendar delivery
  acceptance. Email/push reminders are not enabled; retained legacy preferences
  do not authorize messages. Imported reminders are edited or stopped in the
  learner's calendar, and repeat imports can create duplicates.
- **Earlier local UI evidence:** guided lessons, vocabulary search, device
  pronunciation, notes/knowledge and repeatable practice have focused mock-backed
  browser passes recorded in current state. Self-check, learning plan and
  achievements passed 27/27 checks across 360px, 430px and desktop. Those runs
  predate the later learning-direction and calendar checks above. They are not
  live-provider, real-calendar, learner-outcome or final combined-build evidence.

The collection now supports literal word/short-definition search and separate
stage/due filters over the full requester-owned saved set. List items expose
learner-facing review state and actual scheduling eligibility; raw persisted
status stays unchanged. A single PostgreSQL query supplies filtered totals and
page data, including exhausted pages. Filter/requester-bound cursors reject stale
or mismatched queries rather than silently mixing result sets. Both focused
PostgreSQL regressions passed against all 38 migrations and the real canonical
seed, including more than 50 rows, archived canonical compatibility, legacy
review state, literal `%`/`_`, isolation and deleted pagination boundaries.

Home and the learning plan now read a personal lesson recommendation based on
unfinished sessions, current focus and known/mastered target coverage. Reading it
does not complete lessons or award progress. A 48-check mock-browser set passed
self-check, recommendation and collection filters in all three layouts; the
three existing save/library/detail/practice/remove checks passed in a separate
follow-up. A later recommendation run passed 15/15 checks across the three
layouts, including Home's expired-session redirect, resume after known-status
changes and unavailable-state recovery. These are focused checks, not a fresh
full browser matrix.

The integrated baseline passed `pnpm run validate` with exit 0: formatting,
lint, Go vet, type checks, all 256 foundation tests, 50 API-client tests, 133 web
helper tests, Go package suites and application builds. Earlier mock-inventory
failures were resolved with bounded predicates and regression tests for the
approved additions. The practice-migration scanner's false positive was fixed by
formatting one column per line; whitespace-stripped SQL stayed identical. Its
Atlas checksum was updated, and Atlas validation and Go migration-scanner tests
passed. The existing 38-migration database evidence therefore predates formatting
only, not a SQL behavior change.

This accepted baseline does not cover the pending self-check/saved-empty
interface changes or AI input case-preservation fix. Those changes need focused
verification before the final combined acceptance. Local validation also does
not substitute for the affected browser matrix, staging/provider acceptance or
production release requirements.

The optional OpenAI runtime adapters and `sentence-feedback-v4` prompt are
implemented locally. The semantic contract stays `feedback-schema-v3`, and the
canonical output allowance remains 300 tokens. OpenAI requires explicit provider
selection; there is no automatic fallback or deployment activation from the
evaluation. Local safety checks precede provider moderation, and both OpenAI
stages use zero transport retries within the service request deadline.

The new preregistered GPT-5 nano service gate is **FAILED / STOPPED**. Four of ten
planned synthetic cases were exercised: missing-target and local urgent-safety
controls made zero provider calls; an ordinary invitation and an agreement-error
sentence each made one moderation and one feedback POST. All four POSTs returned
complete HTTP 200 responses. The gate stopped at the agreement case because
`naturalness=natural` differed from the frozen `understandable` reference. The
model did identify the agreement problem and supplied a faithful retry tip;
valid output and those correct observations do not waive the preregistered
failure. Six cases remain unrun, without relabeling or tuning to continue.

The run used the real service with synthetic targets and isolated memory
repositories. It observed history and replay behavior for the ordinary case,
not PostgreSQL durability, real learners or live constrained repair. Usage was
returned, but billing was not reconciled. Independent coordinating AI review
confirmed the stopped result; this is not human learner review or a general
ranking of models. No runtime provider was activated and no deployment changed.
Live writing-feedback availability, representative quality and controlled
provider/service acceptance remain release blockers.

A separately approved GPT-4o mini follow-up completed the ten reused development
cases with 14 POSTs (eight moderation and six feedback), no transport retries and
no constrained repair. All frozen automated references, safety outcomes and
memory-backed lifecycle/replay checks passed. Independent coordinating AI review
read all ten vectors and six feedback outputs and accepted **only the bounded
development gate**. Technical headings, a cheerful grief-related headline and a
lowercase correction remain pedagogical/style limitations, not evidence of
general learner usefulness.

The source/binary hash manifest was recorded after the first two POSTs because
its initial file lookup failed. The binary's exact case/rubric hash guards ran
before network access; the error was disclosed and the same process continued,
without rerunning paid cases. Preserve that provenance gap with this result.
Returned usage was 9,416 input and 704 output tokens; the recorded list-price
estimate was $0.0018348, not reconciled billing. The longest observed service case
was 4.459 seconds, which is not a production latency estimate. The run used
synthetic memory repositories, not PostgreSQL service persistence, and exercised
no live repair. It is not held-out, full-golden-set or human learner evidence.
The failed/stopped nano result remains intact. No provider default, activation or
deployment changed, and the live availability/quality release requirements remain.

## Consolidated delivery acceptance — 2026-10-02

Implementation revision `5b074584cb1799d6b8da7ee02909c703382a3ee5` in
[PR #1482](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1482) passed
the complete [pre-merge staging deployment and learning journey](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36989699136)
at 09:29 UTC. A second [staging-only scheduled journey](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36990179266)
passed at 09:31 UTC. Both ran the frozen implementation revision; the second
exercised fresh account preparation and repeated the full learning loop on the
same day. The deployment also verified release identity, health and OAuth
initiation. No overlapping old scheduled run was observed during either check.

The consolidated fixes cover review distractors, pending profile edits,
capability-aware authentication guidance, completed onboarding without stored
questionnaire answers and repeatable staging fixtures. Local validation, hosted CI and
independent review passed; automated code and security review of this revision
completed without remaining valid findings. The final merge still follows normal
CI and queue checks. Production promotion and the human, provider, recovery and
launch requirements below remain open. Earlier dated records document the
failures and corrections that led to this acceptance.

## Live feedback quality remains open — 2026-10-02

The v3 evidence below precedes the v4 service gate recorded above; it retains its
original scope and does not accept the newer optional provider implementation.

A separate v3 development pilot compared `gpt-5-nano` and `gpt-4.1-nano` on the same nine synthetic
cases through evaluator-only adapters. All 18 responses were complete and schema-valid, but AI
review rejected both candidates for meaning or original-sentence diagnostic errors despite status
agreement of 9/9 and 8/9. This inspected sample informed prompt revision; it is not held-out quality
or human learner evidence. The full 91-case evaluation and moderation, repair, persistence and
learner-service checks remain unproven by this pilot. No runtime provider changed and no deployment
resulted. See the [dated pilot evidence and limits](../engineering/feedback-evaluation.md#v3-development-pilot--2026-10-02).
The prior consolidated acceptance above does not close this live-quality blocker.

The delivery branch also corrects a deployment configuration mismatch: both
workflows wrote `@cf/meta/llama-3.1-8b-instruct-fp8-fast`, while the feedback
adapter and production constructor use `@cf/meta/llama-3.3-70b-instruct-fp8-fast`.
The workflows now match that existing default. Both models exist in
[Cloudflare's current pricing list](https://developers.cloudflare.com/workers-ai/platform/pricing/).
The latter has documented [JSON-mode support](https://developers.cloudflare.com/workers-ai/features/json-mode/);
the former is absent from that feature's supported-model list, checked 2 October 2026.
This does not establish the cause of the live failure. A configuration parity regression passes. This is a local
correction, not proof that the staging moderation failure is resolved: no
deployment or live provider request has verified it, and safe failure diagnostics
remain necessary. The later optional OpenAI implementation and failed/stopped
service gate are recorded above; they do not close this availability or quality
gap.

## Definition of a usable first release

A target learner can sign in through a supported method, choose sensible learning preferences, find relevant words, save them, finish a short review, write an original sentence, understand useful feedback and return later without losing confirmed progress. The interface works on small phones and desktop with keyboard access, readable themes and understandable recovery paths. Production data can be restored, failures are detected, and the release can be identified and rolled back.

Points and streaks support this experience. They must not imply a proficiency score or reward unconfirmed activity. Expanding content or gamification should follow evidence about learning value rather than the number of screens shipped.

## Ordered delivery work

| Priority | Outcome                                 | Acceptance evidence                                                                                                                                 | Current position                                                                                                                                    |
| -------- | --------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1        | Trustworthy review and mission behavior | Reproduced defects fixed; regression tests; stable reloads and retries; target/timezone boundaries exercised                                        | PR #1470 merged; staging release and synthetic journey passed                                                                                       |
| 1        | Clear project instructions              | Current source map, commands, deployment process and handoff; retired automation clearly historical                                                 | Source documentation and two repository skills reconciled and merged                                                                                |
| 1        | Repeatable integration verification     | Full Go database tests, production build and browser matrix on the reviewed revision                                                                | Final local checks and required CI passed; merged staging revision verified                                                                         |
| 2        | Consistent daily learning experience    | Mobile 360/430px and desktop walkthroughs; keyboard and light/dark checks; empty/error/long-content states                                          | Browser matrix, theme/width screenshots and Lighthouse passed; live-device acceptance remains open                                                  |
| 2        | Sufficient practical content            | Inventory by situation and level; editorial check of meanings/examples/distractors; pilot learners can find useful vocabulary for repeated sessions | Local 30-lesson course and persistence checks passed; 17 situations, 89 words, 92 meanings; deployed and learner usefulness acceptance remains open |
| 2        | Reliable live sentence feedback         | Synthetic evaluation set against the configured provider; correctness, helpfulness, failures, latency and measured cost documented                  | Nano gate failed/stopped; later 4o-mini ten-case development gate accepted only within its stated limits; no activation or live-quality acceptance  |
| 2        | Durable operations                      | Documented backup schedule, retention, separate storage and successful isolated restore; release rollback rehearsal; alert delivery proof           | Synthetic restore and failure controls pass locally; production recovery and alert evidence remain open                                             |
| 3        | Working intended signup path            | Real provider sign-in and email lifecycle verified; configuration accurately reflected in the UI; owner selects when to expand access               | Controlled Google signup live; email/password disabled                                                                                              |
| 3        | Learner validation                      | Small consented pilot of A2–B1 learners; task completion, misunderstandings and return visits inform the next iteration                             | No fresh pilot evidence                                                                                                                             |
| 3        | Public launch review                    | Accurate privacy/terms and support/contact arrangements; remaining release blockers resolved                                                        | Requires owner/business decisions and appropriate review                                                                                            |

## Delivery evidence — 2026-10-02

The learning-reliability and guidance slice merged in
[PR #1470](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1470) as
`43e58080718f746f1d0052c528c4ea00d345ea60`.
[Staging deployment 36945712699](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36945712699)
passed the required deployment, public health, release-identity, OAuth-initiation
and reserved synthetic learner checks. Fresh web/API checks at 00:28 UTC matched
the merged revision. Production remained at `5b16186c`; production release and
public-launch acceptance are still separate work. The historical local checks
below apply to the stated revisions, with final-head evidence in the follow-ups.

Four fixes are implemented with regression coverage:

- Review answer ordering now stays consistent between server rendering and
  browser hydration, avoiding a changed question on first interaction. See the
  [option builder](<../../apps/web/src/app/(app)/reviews/_components/review-session-options.ts>)
  and [hydration regression](../../apps/web/tests/e2e/review-hydration.spec.ts).
- A learner's saved daily review target survives UTC and request-time timezone
  fallback. See [settings resolution](../../apps/api/business/gamification/timezone.go)
  and [regressions](../../apps/api/business/gamification/timezone_test.go).
- One naturally missed day can use available grace when its snapshot is absent
  or still open. Protection and the grace debit occur on confirmed completion;
  reading the mission does not spend grace. See
  [streak reconciliation](../../apps/api/business/gamification/streak.go),
  [mission transitions](../../apps/api/business/missions/service.go), and
  [database regressions](../../apps/api/business/reviews/key_postgres_integration_test.go).
- Sentence target matching accepts noun variants supported by the canonical
  curriculum, including regular hyphenated compounds and the documented
  `syllabi` form. See [target matching](../../apps/api/business/aifeedback/target.go)
  and [seed-derived regressions](../../apps/api/business/aifeedback/target_seed_test.go).

Independent reviews reported no actionable findings. The local checks recorded
for this slice are:

- Final `pnpm validate` exited successfully on all four fixes, including the
  complete Go suite against a migrated disposable database and both production
  builds. The affected feedback package also passed its focused checks.
- The full browser suite passed 242 tests with 37 existing skips.
- Twenty-four screenshots covered 360px, 430px, and 1280px widths in light and
  dark themes. Overflow checks found none; representative screenshots were
  visually inspected.
- All 12 Lighthouse audits passed the configured thresholds: performance 100,
  accessibility 100, and best practices 96.
- Repository entry points, the current-state map, development guidance, and two
  repository skills were reconciled with the current workflow. The documentation
  links and formatting were checked.

The browser suite, screenshots, and Lighthouse audits use synthetic/mock-backed
data. They establish rendering and interaction under test conditions, not live
Google sign-in, email delivery, real-provider feedback quality, physical-device
behavior, learner retention, or recoverable production backups. The migrated
disposable database verifies integration paths, not production durability. No
public-launch milestone is complete from these results alone.

### Dependency security follow-up

A fresh registry audit identified newer advisories beyond the initial GitHub
alert. The follow-up updates Next.js to 16.3.6, fast-uri to 3.1.8, and
brace-expansion to 5.0.12. Independent review found no unrelated lockfile changes,
and the updated lockfile audit reported no known vulnerabilities. The application
does not use the affected Node.js ImageResponse path or directly import fast-uri;
this source inspection does not establish the contents of deployed images.

Frozen installation, full workspace validation, all 242 browser tests and all
applicable GitHub checks passed on dependency revision `2739e3da`. The database
integration evidence above applies to `0008ba70`, whose API code was unchanged
by the dependency update. Track later revision evidence in
[PR #1470](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1470).

### Review follow-up: truthful activity history

The automated review found two additional history concerns. A reconstructed
missed day now uses today's established mission target instead of a hardcoded
twenty. Later settings changes cannot overwrite that established target.
Existing historical targets, counters and timezones remain unchanged. Because
no historical settings record exists for an absent day, the reconstructed target
is a fallback, not proof of yesterday's original goal.

Progress now distinguishes completed missions from grace-protected streak days.
The API adds the actual mission `status`, while retaining the legacy `completed`
boolean for compatibility. The client accepts older API responses and labels
ambiguous legacy history "Completed or protected"; current responses show
"Completed", "Streak protected" or "Not complete" with distinct existing theme
colors.

Both failures were reproduced before fixing them. Independent review found no
actionable issues. Full workspace validation passed again, including 244
foundation, 32 API-client and 93 web helper tests, the full Go suite against a
migrated disposable database, and both production builds. Database regressions
also check the learner's account export, settings drift, preserved history,
duplicate recovery and rollback. The full browser suite passed 248 tests with
37 existing skips. Six new cases cover activity labels, distinct colors,
accessibility and overflow at 360px, 430px and desktop in both themes. Six
refreshed Progress screenshots had no horizontal overflow; representative mobile
and desktop views were visually inspected. Fresh OpenAPI generation matched the
committed contract. Release checks remain separate from this local evidence.

### Review follow-up: answer positions across sessions

A separate review identified that permanent card identities alone keep each
word's answer in the same position across later sessions. The server now
generates a fresh session seed after reading request data and serializes it to
the client. The mounted session retains that seed through queue updates, so
choices agree during hydration and remain steady during practice while varying
across later sessions. Both review entry routes use the same page.

A deterministic regression reproduced the old behavior: 64 supplied session
seeds produced only one answer position. The corrected helper exercises all four
positions for the same card. Browser coverage also checks unchanged choices
after an authoritative queue refresh, without requiring random sessions to
produce different positions every time.

Independent review found no issues. Full workspace validation passed with 244
foundation, 32 API-client and 94 web helper tests, Go tests and both production
builds. The complete browser suite passed 248 tests with 37 existing skips. The
API is unchanged from the database-backed verification at `48ff7dca`; this
frontend follow-up did not rerun the optional PostgreSQL integration suite.

## Feedback evidence repair — PR #1471

The evaluator now supplies the selected meaning and retains one observation per
input, including failures and returned feedback. The versioned fixture migration
preserves 308 old identities, expands the full set to 336 cases and retains the
56 old golden members in a 91-case subset covering all nine categories. Paired
A2/B1 cases preserve language correctness while testing explanation level.

Status agreement no longer substitutes for correction quality, safety or
service-side intervention. Missing measurements and human review produce explicit
acceptance gaps; observed failures take precedence. CLI reports use a new private
file, preserve existing evidence, hide environment defaults in help, validate
settings before calls and distinguish incomplete acceptance with exit 3.
See the [evaluation guide](../engineering/feedback-evaluation.md) and
[fixture migration](../engineering/feedback-evaluation-fixtures.md).

Focused fake-provider and CLI regressions pass, including the reproduced
historical label errors, credential-bearing help defaults, report truncation,
file permissions and invalid settings. Independent review corrected two further
fixture ambiguities and found no remaining actionable issues. Final local
`pnpm validate` passed formatting, lint/vet, type checks, 244 foundation tests,
32 API-client tests, 94 web helper tests, the Go suite and both production
builds. The optional PostgreSQL integration environment and browser matrix were
not rerun for this evaluator/CLI/documentation change; learner routes and
database behavior are unchanged. The final changes merged as
`1601c67b0c18d21eaf513164ce2bca2758b1ace0` in
[PR #1471](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1471).
Required CI and [staging deployment 36949720029](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36949720029)
passed. Public web/API checks at 01:15 UTC matched that revision; production
remained at `5b16186c`. These checks establish deployment of the evaluator repair,
not passing model quality.
No paid/live provider run or production-provider configuration change was made.

Review follow-up corrected absent-provider accounting: no configured provider
now records `provider_unavailable` with zero calls consistently. A reproduced
overflow in the old numeric formatter could corrupt a large finite cost; the
report now uses standard numeric formatting and a string builder. Fixture
assembly avoids absolute mutation indices, preserves Unicode text when forming
clauses and explicitly checks expanded golden membership. Report-file cleanup
still closes on early failures and now avoids a redundant close after success.

## Daily Conversation and natural forms — PR #1472

The reviewed [Daily Conversation curriculum](daily-conversation-curriculum.md)
expands from six to 18 ordered meanings, with two examples and distinct fuller
guidance per meaning. All 298 existing seed row identities and their relationships
remain intact. Item levels are editorial A2/B1 judgments, not certification.

Sentence validation accepts the documented regional spellings and inflections of
`catch up`, `meet up`, `keep in touch` and `sounds good` through exact curated
word/type/part-of-speech entries. Matching remains contiguous and token bounded;
lexical presence still requires the provider to judge the selected meaning. The
minimum three-word practice rule remains unchanged. Evaluation fixture v3 retains
all 336 case IDs and 91 golden members, with only the three resolved regional
exclusions removed; six ambiguity exclusions remain.

Both Journey and saved-word detail pages suppress blank or repeated fuller
definitions after whitespace/case normalization. Distinct explanations, including
ones that begin with the short definition, remain visible.

Implementation `679f192c` in
[PR #1472](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1472) passed
full workspace validation, including 244 foundation, 32 API-client and 94 web
helper tests, Go tests against a migrated disposable PostgreSQL database, and
both builds. The seed rerun test exercised actual content constraints and a
surrogate saved-learning reference; it does not stand in for a production
restore. All 72 shipped examples and nine approved variant/meaning pairs passed
the repository-loaded target regression.

The corrected complete [CI browser matrix](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36951969577)
passed 266 tests with 37 existing skips. Local verification first exposed an
incorrect test expectation of B1 for the intentionally A2 `meet up` entry;
correcting that expectation required no curriculum change. Eighteen screenshots
cover the curriculum and word pages at 360px, 430px and desktop in both themes.
They passed overflow and critical/serious accessibility checks; representative
views were visually inspected. [Lighthouse](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36951969591)
and all other applicable CI passed on that implementation.

Independent editorial, backend, interface and release-path reviews found no
remaining actionable issues. Automated review requested two minor cleanups:
clarify that the ninth variant check comes from the second `follow-up` meaning
(not a second `cancel` meaning), and share the identical usage-note formatter.
The follow-up preserves rendered behavior. Final revision `922086d6` passed the
[266-test browser matrix](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36952671883)
with 37 existing skips, [Lighthouse](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36952671794),
workspace validation and applicable reviews. It merged as
`18f34a56e85e823b253b3809fe3c540c05534305`.
[Staging deployment 36953460367](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36953460367)
passed the release, health, OAuth-initiation and reserved synthetic journey checks.
Public web/API identities matched at 02:01 UTC; production remained at `5b16186c`.
Synthetic evidence does not establish live provider quality or observed learner value.

## Synthetic recovery tooling — PR #1473

The [recovery rehearsal](../operations/postgres-recovery-rehearsal.md) creates
fresh isolated local PostgreSQL 16 databases, applies the actual migrations and
canonical seed, adds a synthetic learner and restores a custom-format archive.
The local successful round trip compared 30 tables and 415 rows, including
nonempty learning records, with matching schema and content digests. Fourteen
independent checks also passed on both databases, including rolled-back writes
and enforcement of selected constraints and the append-only ledger trigger.

The five-case local acceptance suite passed in 41.7 seconds on PostgreSQL 16.15:
success, corrupt archive, truncated archive, missing activity record and SIGTERM.
Every case verified cleanup. The first run exposed a startup retry that consumed
an interruption; a failing regression reproduced it, and the corrected suite
confirmed prompt failure and cleanup. The complete run's elapsed time is test
evidence, not a production recovery target. Reports identify the base revision,
dirty checkout, exact input hashes and local image ID; final revision CI evidence
must be tracked separately.

Independent static review found no remaining actionable issues. Local workspace
validation passed formatting, lint/vet, type checks, 245 foundation tests, 32
API-client tests, 94 web helper tests, the Go suite and both production builds.
The new foundation wrapper includes 15 offline recovery contracts. All 21
workflow contracts and immutable action-reference checks also passed. Browser
routes and API implementation are unchanged; the browser matrix and optional
Go PostgreSQL environment were not rerun for this operations-only slice.

Review follow-up reproduced a second cancellation edge: the first signal during
cleanup could leave a provisional PASS unchanged. A single phase-aware handler
now records failure while allowing cleanup to finish, including repeated signals
and temporary-file removal. Nineteen offline contracts and the five-case real
Docker suite passed again (42.9 seconds). A cached-tool capability check also
rejects old `pg_dump` binaries without `--restrict-key` before migrations; current
PostgreSQL 16.15 supports the comparison option, as the real runs demonstrate.

A further review reproduced interruption during report serialization, which
could terminate the process with an empty report, and missing deletion protection
on either ledger, which the old checker accepted. Reports now publish atomically
under deferred interruption handling; a failed corrective write removes only its
own stale publication. Both ledger triggers must retain unconditional row-level
update/delete protection, and a rolled-back deletion probe verifies enforcement.
Twenty-three offline contracts pass. The real acceptance suite now has seven
cases, including both intentionally weakened trigger definitions, and passes;
the historical checker failed those new regressions by incorrectly returning PASS.

Terminal-level Ctrl-C testing then reproduced a child-command interruption that
could abort cleanup even while the parent deferred its signal. Commands now run
in separate process sessions while retaining bounded timeout/cancellation. All
six named constraints also have specific invalid-write probes and post-rollback
checks. A real negative case replaces the feedback guard with `CHECK (true)`;
the historical checker incorrectly accepted it. The final suite comprises 24
offline contracts and eight real database cases, including these regressions.

The final local eight-case suite passed in 57.697 seconds. Recovery
[CI 36958781566](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36958781566)
passed all eight cases on a clean checkout with the same tree as reviewed head
`36ef9f6f`. Independent and automated reviews had no remaining actionable
findings. [PR #1473](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1473)
merged as `56bbff54a7ac5872a996af9d6b5c29f94e0ffe5f` after required checks passed.
[Staging deployment 36959705987](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36959705987)
passed identity, health, OAuth initiation and the reserved learner journey.
Public checks at approximately 03:24 UTC matched the merged revision; production
remained at `5b16186c`.

No production data, credentials, existing container or provider is used. This
does not establish production backup schedule, retention, separate storage,
point-in-time recovery, application acceptance after restoration or alert receipt.

## Backup discovery — delivered in PR #1474

The [read-only discovery workflow](../operations/backup-discovery.md) gathers
bounded system timer metadata and fixed tool-presence observations through the
existing production SSH boundary. It accepts no remote command, path or branch
input and uses strict supplied host-key verification. Sanitized reports preserve
unknown production association, artifacts, storage, restore and alert claims.
Offline checks and actual dispatch evidence must be recorded separately; a
successful collection cannot establish backup or public-launch readiness.

Local validation passed all 44 focused collector/transport tests, including
actual SIGINT/SIGTERM tests against synthetic local child processes. Full
workspace validation passed 246 foundation, 32 API-client and 94 web helper
tests, Go checks and both builds before the final cancellation/freshness
follow-ups; the final focused wrapper and all 22 workflow contracts passed
afterward. Independent static review found no remaining actionable issues.
The browser matrix and optional Go database environment were not rerun because
application behavior was unchanged. All configured PR and merge-group checks
passed, followed by the
[staging deployment](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36962911361).

One [manual discovery](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36963233434)
on reviewed main `ea6ec479b9cb2923430944d51224dfd00f339786` completed on
2026-10-02. The validated artifact reported complete, nontruncated collection of
the supported system-timer metadata and one generic backup-name candidate.
This establishes access to that bounded metadata, not a production backup:
Vocanova association, retained artifacts, separate storage, restore record and
alert receipt remain unknown. Cron, user timers, containers and provider backups
were outside the collection scope. The report is retained privately; identify
the actual backup mechanism before expanding inspection or running a restore.

## Pending edits and recovery — delivered in PR #1475

Delayed browser requests reproduced settings edits being overwritten, a
server-applied save whose lost response made a reverted retry appear saved,
older sentence actions erasing the current draft, and current session errors
being hidden by retained feedback. Settings now preserve newer edits against
the confirmed response and explicitly resend fields with uncertain outcomes.
Sentence controls protect pending text and retry identity; current errors and
late report responses stay associated with the appropriate request.

Local workspace validation passed 246 foundation, 32 API-client and 94 web
helper tests, Go checks and both builds. Focused browser checks passed 26
settings cases (four existing viewport skips) and 18 sentence cases across
360px, 430px and desktop, including light/dark error and unsaved-change states.
Twelve screenshots were retained and representative mobile/desktop captures
were inspected. Independent static review found no remaining actionable
findings. The full local browser matrix passed 299 tests with 37 existing skips.
These checks use synthetic API state; they do not establish real provider
behavior or production data recovery.

[PR #1475](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1475) merged
as `bb0241cb2ace5f848a1e28bbd4b3cb9b3d3118db` after required CI and Codex/Claude
reviews cleared. [Staging deployment 36965361774](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36965361774)
succeeded, verified at 04:41:55 UTC, including strict identity, web/API health,
OAuth initiation and the reserved synthetic learner core loop. Production was
not promoted; no new direct public-health observation is claimed after the
earlier local HTTP 403.

## Optional Daily Conversation context activity — delivered in PR #1476

Three independently reviewed authored examples use message/dialogue completion
to contrast invite/join, sounds good/keep in touch and reschedule/cancel.
Explanations follow the choice, with local retry/restart/exit and existing
word-page links for explicit saving and sentence practice. All six canonical
meaning IDs, text and slugs are checked against the actual situation response;
missing or drifted references omit the optional activity. It changes no points,
missions or progress and introduces no API/schema or provider calls.

Full workspace validation passed: 246 foundation, 32 API-client and 98 web
helper tests, Go checks and builds. All 21 focused browser checks passed at
360px, 430px and desktop in light/dark themes, including keyboard navigation,
feedback text, retry/restart/exit, reload, missing-content fallback and explicit
saving before sentence practice. The full browser matrix passed 314 tests with
37 existing skips in 4.3 minutes. Thirty-six screenshots and six additional
360px viewport captures were retained; selected activity views were inspected.
The viewport captures verified that apparent navigation overlap in an oversized
element capture was a capture artifact, with retry visible and keyboard reachable.
Independent static review found no remaining actionable issues.
Actual screen-reader speech and physical-device behavior remain unverified.
Its [curriculum guidance](daily-conversation-curriculum.md#optional-context-practice)
defines the authored scope; observed learner value and learning effectiveness
remain unverified.

[PR #1476](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1476) merged
at 05:16:59 UTC as `e4cc4236f1f1c18df0059b005a4ade3b07addc31`. Applicable CI
and inspected current-head Codex/Claude reviews cleared; all seven merge-queue
checks passed. [Staging deployment 36968236895](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36968236895)
succeeded, verified at 05:21 UTC, including exact release identity, web/API health,
OAuth initiation and the reserved synthetic learner core loop. Production was
not promoted. The deployed browser journey chooses a situation and word and
tests the general learning loop; it does not exercise this context activity or
establish the complete deployed Daily Conversation curriculum. The activity's
specific UI/content checks above use the local mock-backed suite and actual seed
file. Real-account access, physical-device behavior and learner usefulness remain
separate acceptance gaps.

## Deployed context coverage — first run and lookup follow-up

PR1476's staging journey did not visit the new activity. PR1478 added a
bounded phase inside the existing authenticated journey before its learning
mutations, using the same reserved session and unchanged workflows/time budget.
The phase checks the Daily Conversation inventory, all three examples and both
explanations, keyboard wrong/retry/correct transitions, completion/restart/exit,
reload reset and canonical word destinations. A local wrapper runs the exact
same helper against the existing mock-backed build; staging uses its real API
and seeded content. No new workflow, test account or mock-auth override is added
to staging.

The added phase does not intentionally save, review or submit sentences and
checks that it emitted no browser API mutation requests. That observation is
not a server-wide audit. The existing full journey still mutates reserved
synthetic learning state and may call the evaluator. All 18 targeted local
browser checks passed across desktop, 360px and 430px; E2E typechecking, the
production build and 21 existing scheduling/dispatch/path checks passed.
This does not replace a consented learner
pilot, real Google-account acceptance, physical-device checks, screen-reader
speech or feedback-quality evaluation.

Staging trace capture is disabled because the test uses live session cookies;
runner log masking does not sanitize trace archives. Screenshots, video, error
context and HTML reports still contain synthetic UI and test diagnostics and
must not receive credentials through logs or attachments. An isolated local
failure probe used a fake cookie: the former trace setting retained the marker,
while the current configuration produced no trace and no marker in the retained
results or embedded HTML report. This does not sanitize earlier artifacts or
establish protection against future explicit secret logging.

[PR #1478](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1478) merged
as `65c34b04` after applicable CI and inspected reviews cleared. Its first
[staging run 36973540873](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36973540873)
passed release identity, health and OAuth checks, then **failed** the added phase
at a sounds-good example lookup with two exact DOM matches. It reached all three
activity examples and completion but did not complete the full journey. No
deployed acceptance pass is claimed from that run.

The follow-up scopes canonical content to the main landmark and requires exactly
one globally visible copy of the example. Local hidden-copy and visible-copy
controls verify that hidden markup does not cause an ambiguous lookup while a
visible duplicate still fails. This is a robust visibility assertion, not proof
that hidden streaming markup caused the original live failure. The corrected
journey must pass against the real staging content before closing this gate.
All 24 focused checks passed locally across the three layouts, including both
new controls; E2E typechecking and the 21 workflow contracts also passed.
The helper also follows the actual completion link in the same tab after its
restart/exit/reload checks. This avoids an observed mobile-emulation failure
where a modified click did not produce the expected new-page event.
The three shared-journey checks also passed three consecutive repetitions per
layout (27 checks, no retries).

## Staging daily-target fixture follow-up — working revision

[PR #1480](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1480) merged
as `f53c3021` after current-head CI, inspected code/security reviews and queue
checks passed. Its [staging run36977504685](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36977504685)
passed release identity, health, OAuth initiation and the entire context phase.
It then failed the existing review phase at the required minimum of one card.
The synthetic failure snapshot explicitly showed today's target was complete.
This establishes the context correction, not a complete staging acceptance pass.

The existing seed restored a due word but retained the account's completed daily
mission. The working correction prepares a fresh, guarded staging fixture while
retaining retired synthetic identities and their complete history. Shared job
concurrency protects preparation through browser completion. Default production
seeding, learner daily targets and the minimum-one-card gate remain intact.
See [the operator guidance](../operations/monitoring.md#scheduled-synthetics).

Disposable PostgreSQL integration passed: an old completed mission, its rewards
and learning history survived retirement; the fresh account completed a real
review the same day. Refusal, forced-failure rollback, session revocation,
repeated preparation and normal-seed idempotency controls passed. Atlas applied
the migration set to an empty database and confirmed a second apply was a no-op.
All 37 focused workflow contracts and the normal migration suite passed.
At this checkpoint the hosted staging journey was still pending. Its later
completion is recorded in the opening acceptance section; production promotion
remains separate.

## Consolidated learner reliability — working revision

The remaining implementation is kept on one delivery branch for one final PR.
Three reproduced gaps are addressed together with the staging fixture correction:

- Review choices exclude the current word's alternate meanings before selecting
  distractors. Actual seeded pairs for reservation, deadline and follow-up
  reproduce the ambiguity. A short queue retains the existing self-check fallback;
  answer identities, deterministic ordering and learning mutations are unchanged.
- The separate profile editor holds its pending guard through the request,
  preserves newer typing, applies server normalization to the submitted draft
  when appropriate, and distinguishes earlier saved changes from newer edits.
  Errors retain the current name for an explicit retry.
- Google recovery suggests email/password or an email link only when that method
  is enabled. Magic-link and signup guidance likewise follow known capabilities.
  Unknown or unavailable alternatives get retry guidance. No authentication
  method is enabled by this change.

The focused review-option regression changed from six failures to twelve passing
tests. Authentication baselines reproduced nine failures; profile baselines
reproduced lost pending state, misleading success and overlapping submissions
across all three layouts. Final local workspace validation passed formatting,
lint/vet, type checks, 250 foundation tests, 32 API-client tests, 114 web helper
tests, the Go suite against a fresh migrated PostgreSQL 16 database and both
production builds. E2E typechecking passed. The full browser matrix passed 383
checks with 37 existing skips in 5.2 minutes, covering the corrected profile and
authentication flows in light/dark themes at all three layouts. Captured profile
and recovery states were visually inspected; authentication overflow checks passed.
Independent review found no remaining actionable findings, and the fresh
dependency audit reported zero known vulnerabilities. Hosted acceptance was
pending at this checkpoint and subsequently passed as recorded above. These
checks cannot establish real provider access or learner acceptance.

Automated review found a further signup-specific recovery case: an enabled
sign-in method need not have a form on the signup page. Six unit regressions and
four desktop browser cases reproduced the misleading suggestions. Signup now
recommends creating an account only when its password form is shown and does not
advertise an absent magic-link form. Final local validation passed again with
250 foundation, 32 API-client and 121 web helper tests, database-backed Go checks
and both builds. All 84 affected authentication browser checks passed with 12
existing skips across the three layouts and both themes; captured signup states
were inspected. The full 383-check local matrix above predates this follow-up.

The disposable PostgreSQL readiness probe now supplies its fixture password
explicitly, and the test container enforces SCRAM for host connections. Atlas,
fixture-retirement and default-seed integration checks passed under that stricter
configuration. The earlier default-image runs passed; no universal timeout on
the former readiness probe is claimed.

A subsequent review found that fresh staging preparation depended on optional
Sentry configuration. Deployment now persists `ENVIRONMENT=staging` in its core
configuration step, and both wrapper and SQL require that explicit environment.
Missing, blank or production values fail closed; monitoring settings cannot
authorize or block preparation. Two isolated configuration regressions and six
wrapper cases failed before correction. All 253 foundation checks and the
disposable Atlas, fixture-history, refusal, rollback and default-seed integrations
passed after correction. The configuration tests also verify stale-value
replacement, preservation of unrelated values and repeat-run idempotence without
a monitoring DSN. Independent review found no further issue in this delta.
The preceding revision's complete hosted CI passed; hosted verification of this
correction then exposed the onboarding defect described next.

Pre-merge [staging run36988013273](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36988013273)
successfully deployed the reviewed revision and passed configuration, fixture
preparation, health, identity and OAuth initiation. The browser then returned to
onboarding. Source tracing found that the current-user lookup discarded an
authoritative completed status when a grandfathered account had no questionnaire
profile, a state explicitly supported by the onboarding migration. The shared
lookup now preserves known status for that expected missing-answers condition;
nil profiles, unknown statuses and genuine read errors remain conservative.
No questionnaire answers are fabricated.

Six unit cases and two PostgreSQL-backed current-user responses reproduced the
status defect; all fourteen unit cases and three real-database response cases
passed after correction. Two browser controls reproduced the staging helper's
premature navigation and loss of failed-save feedback. The shared helper now
waits for successful submission and the application's own Home navigation.
All twelve affected onboarding/core-journey browser checks passed across the
three layouts. Full workspace validation passed again with 253 foundation,
32 API-client and 121 web helper tests, database-backed Go checks, both builds
and E2E typechecking. Independent review found no further actionable issue.
The corrected implementation subsequently passed both hosted journeys recorded
in the opening acceptance section.

## Deliver work in bounded slices

Each slice should identify the learner problem, state the intended behavior, change the smallest relevant surface, include meaningful regression coverage and document remaining limits. Use parallel agents for independent areas and independent review; coordinate shared files, builds and servers. Follow [repository instructions](../../AGENTS.md) for draft PRs, CI, merging and deployment.

The owner requests one consolidated PR for the remaining product-completion
work. Keep bounded implementation tasks and local verification on the delivery
branch, then review and validate the combined result before opening that PR.

Prioritize demonstrated failures in the learning loop over a new tutor, leaderboard, social system or native app. Those features require a separate product case and are not necessary to complete the current first release.

## Evidence boundaries

- A healthy container or database connection is availability evidence; it does not prove learner-task success or recoverable backups.
- Mock-backed browser tests verify interface behavior; they do not establish live OAuth, email delivery or provider feedback quality.
- Historical test counts apply to the revision and date recorded, not every subsequent change.
- Do not mark a milestone complete until its acceptance evidence is linked and outstanding limitations are stated.
- Keep operational identifiers and private recovery material out of public documentation unless they are already intentional public product information.

## Decisions still needed before launch

The existing baseline is global A2–B1 English learning and three primary destinations: Home, Journey and Progress. Continue within that direction. Before unrestricted launch, establish the first intended audience/cohort, support owner and contact, email/provider configuration, acceptable recurring provider budget and backup retention/recovery targets. These decisions do not block local implementation or synthetic testing.
