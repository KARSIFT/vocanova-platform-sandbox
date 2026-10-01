# Vocanova release readiness

This is the working completion plan for the product requested on 2 October 2026. Vocanova already implements its main learning loop. Finishing it requires reliable behavior, enough useful content and evidence that real learners can use and trust the deployed service. Historical delivery records remain useful context; this checklist requires fresh evidence for a public launch.

Read [the current state](current-state.md), [product bible](00-product-bible.md), [learning workspace design](../design/learning-workspace.md) and [development guide](../development.md) before extending scope.

## Definition of a usable first release

A target learner can sign in through a supported method, choose sensible learning preferences, find relevant words, save them, finish a short review, write an original sentence, understand useful feedback and return later without losing confirmed progress. The interface works on small phones and desktop with keyboard access, readable themes and understandable recovery paths. Production data can be restored, failures are detected, and the release can be identified and rolled back.

Points and streaks support this experience. They must not imply a proficiency score or reward unconfirmed activity. Expanding content or gamification should follow evidence about learning value rather than the number of screens shipped.

## Ordered delivery work

| Priority | Outcome                                 | Acceptance evidence                                                                                                                                 | Current position                                                                                   |
| -------- | --------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| 1        | Trustworthy review and mission behavior | Reproduced defects fixed; regression tests; stable reloads and retries; target/timezone boundaries exercised                                        | Four fixes implemented, independently reviewed and locally validated; PR/CI pending                |
| 1        | Clear project instructions              | Current source map, commands, deployment process and handoff; retired automation clearly historical                                                 | Source documentation and two repository skills reconciled; PR/CI pending                           |
| 1        | Repeatable integration verification     | Full Go database tests, production build and browser matrix on the reviewed revision                                                                | Final workspace validation, database-backed suite, builds and browser matrix passed; PR/CI pending |
| 2        | Consistent daily learning experience    | Mobile 360/430px and desktop walkthroughs; keyboard and light/dark checks; empty/error/long-content states                                          | Browser matrix, theme/width screenshots and Lighthouse passed; live-device acceptance remains open |
| 2        | Sufficient practical content            | Inventory by situation and level; editorial check of meanings/examples/distractors; pilot learners can find useful vocabulary for repeated sessions | Existing seed is small; inventory and editorial plan needed                                        |
| 2        | Reliable live sentence feedback         | Synthetic evaluation set against the configured provider; correctness, helpfulness, failures, latency and measured cost documented                  | Provider enabled; quality not yet revalidated                                                      |
| 2        | Durable operations                      | Documented backup schedule, retention, separate storage and successful isolated restore; release rollback rehearsal; alert delivery proof           | Not established by current read-only audit                                                         |
| 3        | Working intended signup path            | Real provider sign-in and email lifecycle verified; configuration accurately reflected in the UI; owner selects when to expand access               | Controlled Google signup live; email/password disabled                                             |
| 3        | Learner validation                      | Small consented pilot of A2–B1 learners; task completion, misunderstandings and return visits inform the next iteration                             | No fresh pilot evidence                                                                            |
| 3        | Public launch review                    | Accurate privacy/terms and support/contact arrangements; remaining release blockers resolved                                                        | Requires owner/business decisions and appropriate review                                           |

## Delivery evidence — 2026-10-02

This slice improves existing learning behavior and project guidance. It is local
implementation evidence; the draft PR, required CI, merge, staging deployment,
and any production release remain separate steps.

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
