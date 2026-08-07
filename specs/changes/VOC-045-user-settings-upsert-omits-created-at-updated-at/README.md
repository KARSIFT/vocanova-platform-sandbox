# VOC-045 — user_settings Upsert Omits created_at/updated_at from the INSERT Branch, Violating NOT NULL and Blocking Every New User's Onboarding

**Status: proposed, not adopted.** Nothing in this package is
implementation-authorized. It is a draft response to
[issue #341](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/341),
prepared for founder/steward review at adoption time.

## Why this exists

`user_settings`'s schema
(`apps/api/migrations/20260725130000_voc030_p4_user_settings.sql`) declares
`created_at` and `updated_at` as `NOT NULL` with no `DEFAULT`. Two separate
`INSERT ... ON CONFLICT (user_id) DO UPDATE` upsert call sites —
`apps/api/business/users/postgres.go`'s `CompleteOnboarding` and
`apps/api/business/gamification/repository.go`'s `UpsertUserSettings` — both
omit these two columns from their INSERT column list, setting `updated_at`
only inside the `ON CONFLICT DO UPDATE ... SET` clause, which never runs on a
genuine first insert. Since onboarding is the very first code path to
actually exercise a real INSERT (rather than the `ON CONFLICT` update
branch) for a brand-new user, this pre-existing defect — present since
gamification's own upsert was first written, per the onboarding code's own
comment that it deliberately copied that pattern — was latent until real user
traffic hit it.

The issue reports this reproduced live in production during VOC-038-T03's
core-loop validation: a real, brand-new user completed the onboarding form
and got a `400`, with the underlying Postgres error
(`pq: null value in column "created_at" of relation "user_settings" violates
not-null constraint`) misleadingly surfaced to the user as a generic
validation failure by `mapOnboardingError`'s fallback branch
(`apps/api/app/api/onboarding.go`). **Every real new user is currently
blocked from completing onboarding** — this is also the last step of
VOC-038-T03's own non-AI core-loop validation.

## What this package deliberately does NOT do

- It does not propose a schema migration (e.g. adding `DEFAULT now()` to
  `created_at`/`updated_at`) as its primary fix. The issue's own suggested
  fix targets the two known-broken call sites directly; a schema-level
  `DEFAULT` is flagged as an open question for the reviewing human to decide
  is in or out of scope, not assumed here (see `specification.md`).
- It does not touch `apps/api/app/api/onboarding.go`'s `mapOnboardingError`
  fallback-message behavior itself, even though the issue notes it
  misleadingly presents this as a validation error — that is a separate,
  narrower UX/error-mapping concern the issue raises in passing, not its
  reported root cause or requested fix; flagged as an open question rather
  than silently folded into this package's scope.
- It does not adopt itself. `change.yaml` leaves every adoption/authorization
  field at its template default. No task in `tasks.md` may be dispatched
  until a real adoption decision is recorded.

## Open questions flagged for the reviewing human

`specification.md`'s "Open questions" section flags: (1) how each call site
should consistently source its INSERT-branch timestamp value, given that
`postgres.go`'s `CompleteOnboarding` already has a `now time.Time` parameter
in scope while `gamification/repository.go`'s `UpsertUserSettings` does not;
(2) whether a schema-level `DEFAULT now()` fix (closing off this whole class
of omission at any future call site, not just these two) is also in scope for
this package or deliberately deferred; and (3) whether
`mapOnboardingError`'s misleading fallback message is in scope for this
package or a separate, narrower follow-up.

## Structure

Mirrors recent packages' convention (e.g. VOC-043, VOC-042, VOC-041, VOC-039,
VOC-040): `specification.md`, `acceptance-criteria.md`, `impact-analysis.md`,
`implementation-plan.md`, `tasks.md`, `test-plan.md`, `release-plan.md`.

## Recommended next action for the reviewing human

1. Confirm the proposed `R3` risk classification in `change.yaml` (this is a
   confirmed hard outage of the core new-user onboarding path in production,
   touching two upsert call sites against a `NOT NULL`-constrained table, but
   requires no schema migration in its primary fix).
2. Decide open question 2 (whether a schema-level `DEFAULT now()` fix is also
   in scope) and open question 3 (whether the `mapOnboardingError` messaging
   issue is in scope or a separate follow-up) before or during adoption.
3. Adopt (or request changes to) this package, then dispatch `VOC-045-T00`
   through `VOC-045-T02` individually, as prior packages' tasks were
   dispatched one at a time.
