# VOC-031 — Impact Analysis

## Security and privacy

`VOC-031-R00`: onboarding hard-gate could wall off every pre-existing account
(ties to `VOC-031-D03`). Every account in this repository today —
including non-production test/staging identities created during VOC-025's
through VOC-030's evidence collection — has `onboarding_status='not_started'`,
because onboarding has never existed until this package. Gating every
`(app)` route on `onboardingStatus=='completed'` (`T01`'s literal default)
means those existing identities must complete onboarding again before any
prior milestone's staging procedure can be repeated against them. Mitigate
by surfacing this explicitly (done, `VOC-031-D03`) rather than silently
resolving it, and by not implementing `T01` past the point this draft
records until the founder confirms which resolution applies.

`VOC-031-R01`: email-change account-takeover risk. Requesting an email
change needs only a valid session (not re-authentication), so a hijacked
session could redirect future magic-link logins to an attacker's address.
Mitigate with the mandatory old-email notification (`VOC-031-D05`,
`VOC-031-AC-03`) as a security control, not an optional nicety, plus the same
rate-limiting, hashed-single-use-token, and environment-scoping discipline
already proven for magic links. The current session is deliberately left
intact through both steps (changing a login address is not equivalent to
revoking a session already in legitimate — or illegitimate — use); if the
session itself is compromised, the appropriate remedy is the existing
logout/session-revocation mechanism, not this flow.

`VOC-031-R02`: duplicate-email confirmation race. Two users could request a
change to the same new email concurrently; only one confirm may succeed.
Mitigate by relying on `users`' existing unique partial index
(`lower(email) where deleted_at is null`) as the authoritative guard and
handling its violation at confirm time as a stable conflict response, not an
unhandled 500 or a silently-accepted second write.

`VOC-031-R03`: account deletion is a destructive, irreversible action
against real learner data. Risks: wrong-account deletion, insufficient
anonymization leaving PII linkable, incomplete session revocation leaving
continued access. Mitigate with an explicit multi-step confirmation (not a
single click), synchronous and tested deactivation + full session/token
revocation at request time, the exact DOC-05 §16 per-table disposition
implemented and tested (not approximated), and no `ON DELETE CASCADE`
anywhere (each table's disposition is explicit code, not an implicit
database cascade that could over- or under-delete). Production enablement
specifically requires a separate founder go/no-go and the DOC-05 §16 legal
review (`VOC-031-DEP-03`) — this package builds and tests the mechanism
against non-production data only.

## Data and migrations

`VOC-031-R04`: new CI tooling (Playwright, axe-core, Lighthouse CI)
introduces build-time cost and potential flakiness that must not be silently
treated as non-blocking. Mitigate by targeting a fixed local production
build (not a live network, not the dev server) for Lighthouse, documenting
CI runner/browser-dependency reconciliation (`VOC-031-DEP-04`), and reporting
any threshold not yet met honestly as a recorded limitation rather than
lowering the threshold or skipping the check.

`VOC-031-R05`: settings-write vs. lazy-creation race. `T02`'s first-ever
`PATCH /api/v1/settings` write and `gamification`'s existing lazy
`user_settings` row creation (VOC-030) could race for the same user on their
first-ever interaction with either surface. Mitigate with an upsert
(`ON CONFLICT DO UPDATE`) pattern on both paths rather than a plain `INSERT`
that could violate the row's unique constraint under concurrency.

`VOC-031-R06`: touching many already-shipped screens in the `T06`
reliability/recovery pass risks regressing already-accepted A1–P4 behavior
(mirrors `VOC-030-R01`). Mitigate with regression tests proving every
touched existing screen's pre-existing behavior is unchanged, not merely a
visual/manual check.

`VOC-031-R07`: `idempotency_keys.scope` enum gap (`VOC-031-D09`) — a genuine
contradiction between DOC-05 §13 (no `account_deletion` scope value) and
DOC-06 §9 / DOC-07 (idempotency required for account deletion). Left
unresolved, `T04`'s idempotency handling has no valid `scope` to write.
Mitigate by surfacing the contradiction explicitly (done, in
specification.md) and applying the minimal reconciliation (add the enum
value) rather than guessing an unapproved value into production schema.

`VOC-031-R08`: the no-queue account-deletion sweep (`VOC-031-D07`) depends
on some existing periodic-invocation mechanism (the same one that already
runs session/magic-link cleanup) actually being scheduled in a real
deployment; this repository has no cron/scheduler infrastructure of its own
(DOC-06 §15 — lightweight cleanup only, no queue system). Mitigate by
documenting the sweep's invocation dependency explicitly as an operational
precondition (the same precondition the existing `Cleanup()` job already
has), not silently assuming it runs.

## Analytics and accessibility

Analytics: onboarding answers, Settings values, and account-lifecycle events
are never duplicated into `daily_activity_summaries` or any ledger; only
`feature_audit_logs`-style operational events (e.g. "email changed",
"account deletion requested") are recorded, minimized, and never include the
old/new email in plaintext beyond what the transactional flow itself
requires. Accessibility and performance are this milestone's own named
scope, not incidental: `VOC-031-R09` is a regression risk specific to
introducing the first automated accessibility/performance gates this
repository has had — a newly-installed, initially-strict gate could surface
a backlog of pre-existing violations across A1–P4 screens that were never
checked before. Mitigate by running the new `T07`/`T09` suites against the
full existing core loop (not only the new P5 screens) and fixing what they
find as part of `T06`/`T10`, rather than scoping the new automation only to
the screens this package adds and leaving older screens unchecked.

## Risks, dependencies, and evidence

- `VOC-031-R10`: the genuine new product ambiguity in `VOC-031-D03`
  (onboarding-gate retroactivity) must be resolved by the founder before
  `T01` proceeds (`VOC-031-DEP-05`); this draft does not guess it.
- `VOC-031-R11`: this is the first milestone to introduce two entirely new
  business modules (`users`, `accounts`) and two entirely new automated CI
  gates (accessibility, performance) in one package. A task-ordering mistake
  (e.g. wiring frontend before its backend contract exists) risks rework.
  Mitigate with the fixed `T00 → T11` order and each task's own independent
  Claude Code review.
- `VOC-031-DEP-01`..`DEP-05`: dependencies recorded in `change.yaml`.
- `VOC-031-EV-00`..`EV-45`: migration, persistence, domain-logic, API,
  frontend, reliability, accessibility, performance, UX-consistency,
  contract, mock-inventory, staging, rollback, and exact-SHA review evidence
  referenced by the acceptance criteria.
