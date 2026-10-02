# Vocanova current product state

Reviewed on 2026-10-02 against repository baseline
`18f34a56e85e823b253b3809fe3c540c05534305` plus the synthetic recovery tooling in
this revision. This is a dated inventory and readiness
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

## Live deployment observation

Public checks on 2026-10-02 at 02:01 UTC observed staging web and API serving
version `0.3.1`, commit `18f34a56e85e823b253b3809fe3c540c05534305`, after
[PR #1472](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1472).
Production remained at `5b16186c6471cbbc1c935a703bb5a8a33bf28797`, version
`0.3.1`. Both APIs reported database health `ok` and correct environment labels.
The [staging deployment](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/36953460367)
passed release identity, OAuth initiation and the reserved synthetic learner
journey. These establish availability and the tested synthetic flow; they do
not establish real account acceptance, live feedback quality or recoverable
backups. No production release was performed for this milestone.

Both environments reported Google OAuth and AI enabled. Magic-link email,
password authentication, and public new-user signup were disabled. Access is
through the controlled cohort. Configuration readiness does not prove a fresh
Google sign-in, email delivery, or real-provider feedback quality.

- Staging: [web version](https://staging.vocanova.site/version),
  [API version](https://api-staging.vocanova.site/version).
- Production: [web version](https://production.vocanova.site/version),
  [API version](https://api-production.vocanova.site/version).

Staging deploys automatically from pushes to `main`; production requires manual
dispatch. Inspect the actual triggers in
[staging](../../.github/workflows/deploy-staging.yml) and
[production](../../.github/workflows/deploy-production.yml); old comments still
mention former branch/promotion arrangements. Keep host configuration private.

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
