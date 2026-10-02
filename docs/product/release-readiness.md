# Vocanova release readiness

This is the working completion plan for the product requested on 2 October 2026. Vocanova already implements its main learning loop. Finishing it requires reliable behavior, enough useful content and evidence that real learners can use and trust the deployed service. Historical delivery records remain useful context; this checklist requires fresh evidence for a public launch.

Read [the current state](current-state.md), [product bible](00-product-bible.md), [learning workspace design](../design/learning-workspace.md) and [development guide](../development.md) before extending scope.

## Definition of a usable first release

A target learner can sign in through a supported method, choose sensible learning preferences, find relevant words, save them, finish a short review, write an original sentence, understand useful feedback and return later without losing confirmed progress. The interface works on small phones and desktop with keyboard access, readable themes and understandable recovery paths. Production data can be restored, failures are detected, and the release can be identified and rolled back.

Points and streaks support this experience. They must not imply a proficiency score or reward unconfirmed activity. Expanding content or gamification should follow evidence about learning value rather than the number of screens shipped.

## Ordered delivery work

| Priority | Outcome                                 | Acceptance evidence                                                                                                                                 | Current position                                                                                                 |
| -------- | --------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| 1        | Trustworthy review and mission behavior | Reproduced defects fixed; regression tests; stable reloads and retries; target/timezone boundaries exercised                                        | PR #1470 merged; staging release and synthetic journey passed                                                    |
| 1        | Clear project instructions              | Current source map, commands, deployment process and handoff; retired automation clearly historical                                                 | Source documentation and two repository skills reconciled and merged                                             |
| 1        | Repeatable integration verification     | Full Go database tests, production build and browser matrix on the reviewed revision                                                                | Final local checks and required CI passed; merged staging revision verified                                      |
| 2        | Consistent daily learning experience    | Mobile 360/430px and desktop walkthroughs; keyboard and light/dark checks; empty/error/long-content states                                          | Browser matrix, theme/width screenshots and Lighthouse passed; live-device acceptance remains open               |
| 2        | Sufficient practical content            | Inventory by situation and level; editorial check of meanings/examples/distractors; pilot learners can find useful vocabulary for repeated sessions | 7 situations, 51 words, 54 meanings; reviewed Daily Conversation expansion; pilot usefulness remains open        |
| 2        | Reliable live sentence feedback         | Synthetic evaluation set against the configured provider; correctness, helpfulness, failures, latency and measured cost documented                  | Evaluator repair merged in PR #1471 and verified on staging; live quality and human/service evidence remain open |
| 2        | Durable operations                      | Documented backup schedule, retention, separate storage and successful isolated restore; release rollback rehearsal; alert delivery proof           | Synthetic restore and failure controls pass locally; production recovery and alert evidence remain open          |
| 3        | Working intended signup path            | Real provider sign-in and email lifecycle verified; configuration accurately reflected in the UI; owner selects when to expand access               | Controlled Google signup live; email/password disabled                                                           |
| 3        | Learner validation                      | Small consented pilot of A2–B1 learners; task completion, misunderstandings and return visits inform the next iteration                             | No fresh pilot evidence                                                                                          |
| 3        | Public launch review                    | Accurate privacy/terms and support/contact arrangements; remaining release blockers resolved                                                        | Requires owner/business decisions and appropriate review                                                         |

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

## Pending edits and recovery — current slice

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
findings. These checks use synthetic API state; they do not establish real
provider behavior or production data recovery. Full browser and release
verification remain pending at this update.

## Deliver work in bounded slices

Each slice should identify the learner problem, state the intended behavior, change the smallest relevant surface, include meaningful regression coverage and document remaining limits. Use parallel agents for independent areas and independent review; coordinate shared files, builds and servers. Follow [repository instructions](../../AGENTS.md) for draft PRs, CI, merging and deployment.

Prioritize demonstrated failures in the learning loop over a new tutor, leaderboard, social system or native app. Those features require a separate product case and are not necessary to complete the current first release.

## Evidence boundaries

- A healthy container or database connection is availability evidence; it does not prove learner-task success or recoverable backups.
- Mock-backed browser tests verify interface behavior; they do not establish live OAuth, email delivery or provider feedback quality.
- Historical test counts apply to the revision and date recorded, not every subsequent change.
- Do not mark a milestone complete until its acceptance evidence is linked and outstanding limitations are stated.
- Keep operational identifiers and private recovery material out of public documentation unless they are already intentional public product information.

## Decisions still needed before launch

The existing baseline is global A2–B1 English learning and three primary destinations: Home, Journey and Progress. Continue within that direction. Before unrestricted launch, establish the first intended audience/cohort, support owner and contact, email/provider configuration, acceptable recurring provider budget and backup retention/recovery targets. These decisions do not block local implementation or synthetic testing.
