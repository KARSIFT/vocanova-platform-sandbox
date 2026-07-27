# VOC-031 — P5 Integrated Core Loop Staging and Rollback Evidence

## Purpose

This document records the evidence required by `VOC-031-AC-11` (via
`VOC-031-TEST-41`..`VOC-031-TEST-45`). It is drafted before adoption/
implementation; it is updated at `T11` implementation time to record the
in-repository evidence actually produced by the merged `T00`–`T11` PRs, and
to document the staged exercises and rollback rehearsal that can only run
once F3 staging exists — following the exact pattern
`VOC-025-DEP-01`/`staging-evidence.md` and `VOC-030-DEP-02`/`staging-evidence.md`
established for this repository.

## Current status

Execution of live staging evidence is **blocked** by `VOC-031-DEP-02`: the F3
staging gate and supporting staging environment do not yet exist. No P5-complete
declaration is made here, and none can be made until `T00`–`T11` are
implemented and F3 exists.

## Planned in-repository evidence (produced by `T00`–`T11`, recorded at `T11` time)

| Evidence | Requirement | Status / source |
| --- | --- | --- |
| `EV-00`..`EV-03` | Onboarding schema/migration and seed-rule domain logic | Not yet produced — `T00` |
| `EV-04`..`EV-07` | Onboarding API, gating, `/me` additive field | Not yet produced — `T01` |
| `EV-08`..`EV-11` | Settings read/write, upsert-race safety, non-retroactivity | Not yet produced — `T02` |
| `EV-12`..`EV-17` | Email-change verification: token mechanics, race safety, session/OAuth non-impact | Not yet produced — `T03` |
| `EV-18`..`EV-23` | Account deletion: deactivation, idempotency, anonymization sweep, scope enum | Not yet produced — `T04` |
| `EV-24`..`EV-27` | Settings/account frontend flows | Not yet produced — `T05` |
| `EV-28`..`EV-31` | Reliability/recovery pass, A1–P4 regression check | Not yet produced — `T06` |
| `EV-32`..`EV-35` | Accessibility automation install and CI wiring | Not yet produced — `T07` |
| `EV-36` | Full core-loop end-to-end suite | Not yet produced — `T08` |
| `EV-37`..`EV-39` | Performance automation install and CI wiring | Not yet produced — `T09` |
| `EV-40` | UX-consistency audit record | Not yet produced — `T10` |
| `EV-41`..`EV-42` | Installed suite pass, extended mock-inventory check | Not yet produced — `T11` |
| `EV-43` | Staging full-loop and cross-user-denial exercise | **Blocked by `VOC-031-DEP-02`** — F3 staging does not exist. Procedure is documented below; live execution is recorded as blocked. |
| `EV-44` | New-tables rollback rehearsal | **Blocked by `VOC-031-DEP-02`** — F3 staging does not exist. Procedure is documented below; live execution is recorded as blocked. |
| `EV-45` | Exact-SHA independent verification (per PR) | **Performed by Claude Code (different model binding) at each PR's exact final SHA** — this is not the implementer's evidence to record; it is produced by the independent-verification role and reported per-PR. |

## Staging exercise plan (blocked by F3 + the open decisions)

Once `VOC-031-DEP-02` is resolved and a non-production F3 environment is
available, the following exercises must be executed and their results
appended to this document or linked from PR evidence.

### `EV-43` — Full-loop and cross-user-denial exercise

1. With a non-production learner identity that has no pre-existing
   onboarding profile, complete the onboarding flow, verify redirection to
   `/home`, and confirm `daily_review_target` was seeded per `VOC-031-D04`
   (or preserved, if a prior customization already existed).
2. Exercise the full core loop (discover → save → review → sentence
   practice → progress) as the DOC-10 §7 flow describes.
3. On `/settings`, change each of the six editable fields and confirm the
   change is reflected on reload; confirm a `dailyReviewTarget` change does
   not alter the current local day's already-created mission target.
4. On `/settings/account`, request and confirm an email change with a
   non-production test-inbox address; confirm the old-email notification is
   received and the current session remains valid throughout.
5. On `/settings/account`, complete the account-deletion confirmation flow
   for a disposable non-production identity; confirm immediate logout,
   confirm the account can no longer authenticate, and (after the sweep
   runs, forced past `purge_after` for the test) confirm the DOC-05 §16
   disposition was applied per table.
6. Create a second non-production identity and confirm it can never reach
   the first identity's onboarding profile, settings, email-change state, or
   deletion-request state (404, not 403, per the existing A1 convention).
7. Repeat step 4's email-change flow with a replayed request (no
   `Idempotency-Key` applicable there, so instead: replay the same
   consumption token) and step 5's deletion with a replayed
   `Idempotency-Key`, and confirm neither produces a duplicate effect.

### `EV-44` — New-tables rollback rehearsal

1. Record current `user_onboarding_profiles`, `email_change_links`, and
   `account_deletion_requests` state for a test learner.
2. Apply the VOC-031 build and migrations in staging.
3. Run several onboarding/settings/email-change/deletion-request flows, then
   perform a rollback to the previously known-good revision.
4. Verify that:
   - No account is left in a state where it was deactivated by a
     deletion request but the pre-rollback code cannot resume its sweep.
   - `user_settings`/onboarding state committed before the rollback is
     preserved.
   - The A1–P4 write paths remain functional after rollback (nothing this
     package adds is on their critical path — `T02`–`T04` are purely
     additive new surfaces).

## Rollback triggers

Per `VOC-031` implementation-plan §Deployment and rollback / release-plan
§Rollback, initiate rollback on:

- A wrong-account or unauthorized account deletion or email change reaching
  production.
- An account-deletion sweep leaving inconsistent state.
- An onboarding gate locking a learner out of the loop with no recovery
  path.
- A regression in any A1–P4 screen or write path.
- Migration or schema failure.
- A new accessibility/performance automation gate that was bypassed rather
  than honestly reported as failing.

## Rollback procedure

1. Preserve `user_onboarding_profiles`/`user_settings` state committed
   before the rollback.
2. Ensure no `account_deletion_requests` row is left in a state neither the
   pre-rollback nor post-rollback code can resolve (complete or safely pause
   any in-flight sweep before rolling back its owning code).
3. Restore the pre-P5 A1–P4 behavior cleanly (verify no A1–P4 write path
   depended on anything this package added).
4. Revert the deployment to the last-known-good revision.
5. Validate with non-production identities.
6. Record the last-known-good revision and the rollback reason.

## Limitations / open dependencies

- **`VOC-031-DEP-02`**: F3 staging does not exist, so `EV-43`/`EV-44` cannot
  be run live. Procedures are documented; live execution is recorded as
  blocked.
- **`VOC-031-D02`/`D03`/`D05`–`D07`/`D09`**: open or this-draft-proposed
  decisions must be resolved or confirmed at adoption before the affected
  tasks' evidence can be treated as final.
- **`VOC-031-DEP-03`**: account-deletion production enablement additionally
  requires a separate founder go/no-go and the DOC-05 §16 legal review; this
  is not resolved by staging evidence alone.
- **`VOC-031-DEP-04`**: new Playwright/axe-core/Lighthouse CI tooling
  reconciliation against the adopted base's CI runner is required before
  `T07`/`T09` can be considered complete.

## P5 gate readiness

Per the DOC-12 §5 P5 gate wording: **"the full loop works coherently in
staging across supported layouts with no critical
product/security/data/accessibility/reliability defect."**

This document cannot yet report gate readiness, because `T00`–`T11` are not
yet implemented (this is a draft package). Once implemented, this section
must state, with evidence:

1. Whether the full A1–P5 loop, including the new onboarding and
   Settings/account surfaces, works coherently across the three supported
   layouts (360px, 430px, ≥1024px desktop).
2. Whether the new accessibility (`T07`) and performance (`T09`) automation
   passes on every core-loop screen, or exactly which shortfall remains
   open.
3. Whether any critical product/security/data/accessibility/reliability
   defect remains open.

The P5 gate **cannot be declared complete by this work alone** even once
implemented, because:

1. Live `EV-43`/`EV-44` staging/rollback evidence is blocked by the missing
   F3 staging environment (`VOC-031-DEP-02`).
2. Exact-SHA independent verification of every PR (`T00`..`T11`) is Claude
   Code's responsibility (`EV-45`); this implementer does not self-approve.
3. Production deployment is separately governed (A-003 §11/12); R3/R4
   founder authority, RL1/RL2 technical activation, and autonomous
   production release are all still disabled. Account-deletion production
   enablement specifically requires the additional founder go/no-go and
   legal review noted in `VOC-031-DEP-03`.

## Follow-up work

- Execute `EV-43`/`EV-44` once F3 staging exists.
- R1: fold the full A1–P5 loop into staging-readiness validation under
  production-like conditions.
- A future package may expand `appLanguage` beyond `en`-only once real i18n
  infrastructure exists (`VOC-031-D06`).
- A future package may address `VOC-031-D08`'s informational routing-drift
  note (`/reviews` vs. documented `/review`, missing `/words` routes) if the
  founder ever chooses to reconcile DOC-08 with the shipped app — not
  addressed by this package.
