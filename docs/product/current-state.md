# Vocanova current product state

Reviewed on 2026-10-02 against repository baseline
`f53c30211045ea6d15a6114250cd6b250d0e0732` plus the staging-fixture follow-up
in this working revision. This is a dated inventory and readiness
guide. Update it when behavior, deployment, or verified evidence changes; it does
not establish that the final product is complete.

Use the [release-readiness plan](release-readiness.md) to turn the gaps below
into bounded delivery work with explicit acceptance evidence.

## Product and source of truth

Vocanova serves A2–B1 learners through a practical vocabulary loop: discover a
word in a real situation, save it, recall it, write an original sentence, receive
focused feedback, and return for the next daily mission. The
[product bible](00-product-bible.md) and [MVP PRD](01-mvp-prd.md) define the
baseline; later owner-authorized delivery documents describe additions.
[AGENTS.md](../../AGENTS.md) governs implementation and release work.

A separate ChatGPT Site prototype describes a three-word lesson, recall,
sentence practice, discovery, collection, progress, and mobile navigation with
sample data. It is design input; it does not establish backend functionality,
real learner data, or deployment of this repository. Recovery of the full
ChatGPT project was incomplete during this review; this inventory does not
claim to preserve every conversation or attachment.

## Implemented surfaces

These links locate existing implementation. They do not imply each flow was
freshly tested live on the review date.

| Surface                 | Current implementation                                                           | Source                                                                                                                                                                                               |
| ----------------------- | -------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Entry                   | Landing, capability-driven sign-in/signup, onboarding                            | [Web routes](../../apps/web/src/app/), [auth capabilities](../../apps/web/src/lib/auth-capabilities.ts)                                                                                              |
| Home                    | Daily mission, next action, due reviews, saved words, sentence practice          | [Home](<../../apps/web/src/app/(app)/home/page.tsx>), [missions](../../apps/api/business/missions/)                                                                                                  |
| Journey                 | Situation browsing, situation/word detail, saved vocabulary                      | [Discovery](<../../apps/web/src/app/(app)/discover/>), [content](../../apps/api/business/content/), [learning](../../apps/api/business/learning/)                                                    |
| Recall                  | Focused review sessions and deterministic spaced scheduling                      | [Review routes](../../apps/web/src/app/review/), [reviews](../../apps/api/business/reviews/)                                                                                                         |
| Sentence practice       | Feedback, recoverable drafts, retries, sentence history                          | [Practice](<../../apps/web/src/app/(app)/_components/sentence-feedback.tsx>), [history](<../../apps/web/src/app/(app)/progress/sentences/>), [feedback service](../../apps/api/business/aifeedback/) |
| Progress                | API-backed activity, missions/streaks, Confidence Points                         | [Progress](<../../apps/web/src/app/(app)/progress/>), [gamification](../../apps/api/business/gamification/)                                                                                          |
| Account                 | Learning preferences, profile, email change, password security, export, deletion | [Settings](<../../apps/web/src/app/(app)/settings/>), [accounts](../../apps/api/business/accounts/), [passwords](../../apps/api/business/password/)                                                  |
| Appearance and identity | Light/Dark/System; web/API release identity                                      | [Theme](../../apps/web/src/lib/theme-preference.ts), [web version](../../apps/web/src/app/version/), [release operations](../development/account-and-release-operations.md)                          |

Home, Journey, and Progress remain the three primary destinations. Additions are
documented in [product maturity](product-maturity-delivery.md),
[learning continuity](mature-learning-and-account-experience.md), and
[password/profile/theme scope](password-profile-and-theme.md). These contain
dated evidence or acceptance plans, rather than a fresh test report.

The canonical [seed](../../apps/api/cmd/seed/voc026-p1.json) contains seven
situations, 51 words, 54 meanings and 72 examples: a bounded starter curriculum,
not broad curriculum coverage. [Daily Conversation](daily-conversation-curriculum.md)
has 18 ordered meanings and two examples each, with editorial level rationales
and preserved existing identities. Word pages suppress repeated short/full
definitions while retaining distinct fuller guidance. Generated feedback is
distinct from canonical vocabulary.

Daily Conversation includes optional context practice: three
independently reviewed message/dialogue examples contrast invite/join,
sounds good/keep in touch and reschedule/cancel. Explanations follow the choice;
learners can retry, restart or exit locally. Completion uses existing word-page
links and explicit saving before sentence practice. All six canonical IDs, text
and slugs must match the situation response; missing or drifted references omit
the activity. It adds no points, mission/progress changes, API/schema or provider
calls. Full workspace validation passed (246 foundation, 32 API-client and 98
web helper tests, Go checks and builds). All 21 focused browser checks passed
across 360px, 430px and desktop in both themes. The full browser matrix passed
314 tests with 37 existing skips. The change merged in
[PR #1476](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1476), with
staging release checks passing as recorded below. That release's deployed
synthetic journey did not exercise this specific activity; its interaction
evidence came from the mock-backed browser suite.

[PR #1478](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1478) added a focused context phase to the existing staging
journey, before its save/review/submission actions. It checks the Daily
Conversation inventory, the three examples and explanations, keyboard
retry/continuation, completion/restart/exit, reload reset and canonical word
destinations. It does not save words or submit sentences; the rest of the
existing staging journey still changes the reserved synthetic account and may
call the evaluator. The exact helper is also exercised locally. Its first
[staging run](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36973540873)
passed identity/health/OAuth checks but failed an ambiguous example-text lookup
on the sounds-good page after completing the three examples. That is not a full
journey pass. The follow-up checks one visible example within the main landmark,
with local controls for hidden and visibly duplicated markup.
[PR #1480](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1480) shipped
that correction, and [staging36977504685](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36977504685)
passed the complete context phase. The existing review phase then failed because
the persistent synthetic account had completed its daily target. The full journey
has not passed on this revision. The fixture follow-up prepares a fresh guarded
staging account while preserving retired history and normal learner behavior.
No real-account, physical-device, screen-reader speech or learning-effectiveness
claim follows from synthetic checks. See
[curriculum guidance](daily-conversation-curriculum.md#optional-context-practice).

Sentence practice protects pending drafts and retry identities from older
feedback actions. Current session errors remain visible alongside earlier
feedback, and late report responses cannot mark a newer result as reported.
Settings preserve edits made during an earlier save and distinguish those
unsaved edits from confirmed changes. A retry explicitly resends fields whose
previous save response was lost. Reminder controls store preferences only;
the interface does not claim reminder delivery is available.

The consolidated delivery also protects the separate profile editor: an earlier
save cannot replace newer typing, overlapping saves are ignored while pending,
and the status distinguishes confirmed changes from a newer unsaved name.
Review choices exclude other meanings of the current canonical word because the
word-only prompt cannot distinguish those valid answers. When fewer than three
safe alternatives remain, the existing self-check mode is used. Sign-in recovery
uses known capabilities when suggesting another method; unknown capabilities
fall back to retry guidance. These working changes require the combined release
verification recorded in [release readiness](release-readiness.md).

## Live deployment observation

Public checks on 2026-10-02 at approximately 03:24 UTC observed staging web and API
serving version `0.3.1`, commit `56bbff54a7ac5872a996af9d6b5c29f94e0ffe5f`, after
[PR #1473](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1473).
Production remained at `5b16186c6471cbbc1c935a703bb5a8a33bf28797`, version
`0.3.1`. Both APIs reported database health `ok` and correct environment labels.
The [staging deployment](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36959705987)
passed release identity, OAuth initiation and the reserved synthetic learner
journey. These establish availability and the tested synthetic flow; they do
not establish real account acceptance, live feedback quality or recoverable
backups. No production release was performed for this milestone.

Later, [PR #1474](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1474)
merged as `ea6ec479b9cb2923430944d51224dfd00f339786` and its
[staging deployment](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36962911361)
passed release identity, health, OAuth initiation and the reserved learner
journey. A fresh local public-endpoint request returned HTTP 403; this later
release observation therefore comes from the deployment checks.

[PR #1475](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1475) then
delivered pending-edit recovery as `bb0241cb2ace5f848a1e28bbd4b3cb9b3d3118db`.
Required CI and Codex/Claude reviews cleared; local browser verification passed
299 tests with 37 existing skips. Its
[staging deployment](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36965361774)
succeeded, verified at 04:41:55 UTC, including strict release identity, web/API
health, OAuth initiation and the reserved synthetic learner core loop. This is
workflow verification, not a new direct public-endpoint observation after the
earlier HTTP 403. Production was not promoted.

[PR #1476](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1476) delivered
context practice as `e4cc4236f1f1c18df0059b005a4ade3b07addc31` at 05:16:59 UTC.
Applicable CI, inspected Codex code/security and Claude reviews, and all seven
merge-queue checks passed. [Staging deployment 36968236895](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36968236895)
succeeded, verified at 05:21 UTC, including exact release identity, web/API health,
OAuth initiation and the reserved synthetic learner core loop. The merged source
tree matches the reviewed head. These are workflow observations; production was
not promoted and no new direct local public-endpoint check is claimed.

Both environments reported Google OAuth and AI enabled. Magic-link email,
password authentication, and public new-user signup were disabled. Access is
through the controlled cohort. Configuration readiness does not prove a fresh
Google sign-in, email delivery, or real-provider feedback quality.

- Staging: [web version](https://staging.vocanova.site/version),
  [API version](https://api-staging.vocanova.site/version).
- Production: [web version](https://production.vocanova.site/version),
  [API version](https://api-production.vocanova.site/version).

Staging deploys automatically when pushes to `main` match the workflow's path
allowlist; a docs-only change under `docs/` does not itself trigger deployment.
Production requires manual dispatch. Inspect the actual triggers in
[staging](../../.github/workflows/deploy-staging.yml) and
[production](../../.github/workflows/deploy-production.yml); old comments still
mention former branch/promotion arrangements. Use the current operator section
in the [DevOps guide](../operations/11-devops-and-ci-cd.md#current-operator-process--2-october-2026)
and [monitoring runbook](../operations/monitoring.md). Keep host configuration private.

## Remaining readiness work

| Need               | Evidence or decision required                                                                                                                                                   | Starting point                                                                                                                                                         |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Real account entry | Verify controlled Google login/session recovery with a real provider session. If email/password is activated, prove delivery, verification, reset, and revoked-session behavior | [Staging cohort](../operations/staging-controlled-signup.md), [email activation](../development/account-and-release-operations.md)                                     |
| Learning quality   | Evaluate real-provider feedback against the documented rubric using representative A2–B1 sentences; inspect corrections, failures, and reporting                                | [AI requirements](../engineering/09-ai-features.md), [live evaluation command](../../apps/api/cmd/eval-live/), [feedback service](../../apps/api/business/aifeedback/) |
| Curriculum         | Editorially review and expand coherent situations based on learner needs; retain provenance                                                                                     | [Seed](../../apps/api/cmd/seed/), [content service](../../apps/api/business/content/)                                                                                  |
| Learner value      | Run an observed cohort; measure first-session friction, completed sessions, and return visits with cohort size and observation window                                           | [Evidence needs](mature-learning-and-account-experience.md#product-evidence-after-this-delivery)                                                                       |
| Recovery           | Synthetic dump/restore and failure controls pass locally; still need real backup inventory, isolated production recovery, retention and alert evidence                          | [Recovery rehearsal](../operations/postgres-recovery-rehearsal.md), [monitoring](../operations/monitoring.md)                                                          |
| Wider rollout      | Verify core flows on the chosen release with real accounts and physical/mobile devices; resolve failures and explicitly decide cohort expansion                                 | [Browser guidance](../development.md), [fixtures](../../apps/web/tests/e2e/)                                                                                           |
| Scaling            | Decide on a shared limiter before relying on multiple API replicas; current auth limits are process-local                                                                       | [Auth limiter](../../apps/api/business/auth/rate.go)                                                                                                                   |

Browser fixtures establish interface behavior with synthetic data. Disposable
database tests establish integration under test conditions. Neither proves live
email deliverability, Google account behavior, learner retention, or feedback
quality. Health checks and deployments also do not establish those claims.

The feedback evaluator now records meaning-aware observations and explicit
acceptance gaps. The v3 fixtures retain all 336 cases and 91 golden members;
six ambiguity cases remain excluded from status scoring. Curated regional
and phrase forms improve lexical validation without awarding semantic credit.
See the [evaluation guide](../engineering/feedback-evaluation.md) and
[fixture migration](../engineering/feedback-evaluation-fixtures.md) for the
meaning-aware cases, recorded evidence and checks that remain unmeasured.
An adapter-only run cannot establish full feedback acceptance.

Voice/pronunciation, reminders, monetization, and native applications need
separate product decisions. Deliver the learning loop reliably and prioritize
additions using learner evidence; historical Control Plane plans are not the
current product roadmap.
