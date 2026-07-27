# VOC-031 — Implementation Plan

## Preconditions and protected areas

Do not begin until this draft is adopted, `VOC-031-D02`/`D03`/`D05`–`D07`/`D09`
are resolved or confirmed (D03's resolution specifically gates `T01` —
`VOC-031-DEP-05`), the adopted `develop` base and its repository commands are
recorded, and the A1→P1→P2→P3→P4 acceptance chain status is confirmed
(`VOC-031-DEP-01`). Database schemas/migrations, Ent schemas, the new
`users`/`accounts` modules, the already-shipped `GET /api/v1/me` response
this package additively extends, `apps/api/business/auth`'s existing
token/session/rate-limit primitives (reused, never modified), and the
real, irreversible account-deletion procedure are R3/R4 protected. Preserve
existing compatible work; no A1–P4 mechanic outside `GET /api/v1/me`'s
additive field and the `T06` reliability fixes is re-litigated.

## File reconciliation and implementation sequence

First inventory the actual scaffold carried from VOC-025/026/027/028/030: the
A1 auth/session foundation (`apps/api/business/auth`), the
`content`/`learning`/`reviews`/`aifeedback`/`missions`/`gamification`
modules, the `user_settings` table (schema-complete, only
`timezone`/`daily_review_target` wired), and the actual shipped frontend
route tree (`/signin`, `/auth/magic`, `/home`, `/discover`, `/progress`,
`/reviews`) — and confirm `user_onboarding_profiles`,
`email_change_links`, `account_deletion_requests`, any `settings`/`accounts`
route, and any Playwright/axe-core/Lighthouse CI tooling still do not exist
immediately before starting (repeat the `VOC-031-D00` inspection at the
adopted base SHA, since time may have passed since this draft). Then execute
`T00 → T01 → T02 → T03 → T04 → T05 → T06 → T07 → T08 → T09 → T10 → T11` in
order; `T01` waits for `VOC-031-D03`'s resolution rather than guessing. Keep
onboarding/settings domain logic pure and unit-tested in isolation (mirrors
the P2 `scheduling.go` / P4 `gamification` pattern). Reuse
`apps/api/business/auth`'s existing token-hash, rate-limiter, and
session-revocation primitives for the email-change and account-deletion
flows rather than duplicating them. Commit generated OpenAPI and the matched
client with their source changes. Do not wire a frontend read/write to a
real endpoint until the approved contract exists (`T05` follows
`T02`–`T04`; `T08` follows `T07`). Do not invent a Settings field beyond the
founder-directed six, a working `appLanguage` switch, onboarding
resumability, or any R1/R2/L1 behavior.

## Validation and independent verification

Run every installed relevant command discovered at implementation time: root
`pnpm validate`/`pnpm test`/`pnpm build`, the
`scripts/governance/validate-governance.sh` and
`scripts/governance/classify-change-risk.sh` checks as applicable to the
changed paths, Go `gofmt`/`go vet`/`go test`/`go build`, web
lint/typecheck/build/format, the domain/migration/transaction/contract/
reliability tests this package adds, and — once installed by `T07`/`T09` —
the new Playwright/axe-core accessibility suite and Lighthouse CI budgets.
Claude Code independently reviews each exact final SHA for: scope and the
classifier floor; migration safety and no A1–P4 schema regression; for
`T03`/`T04` specifically, that the email-change and account-deletion flows
correctly reuse (and never weaken) `auth`'s existing token/session
primitives; the account-deletion mechanism's correctness against the exact
DOC-05 §16 disposition (with concrete non-production evidence, not an
assertion); requester scope and the 404-private-resource rule on every new
read; the `VOC-031-D02`/`D04`/`D05`/`D06`/`D07`/`D09` resolutions as actually
implemented (not silently altered); contract/OpenAPI/client drift;
accessibility and performance automation correctness and CI wiring;
staging/rollback evidence; and implementer separation. Missing staging,
open-decision, or tooling evidence remains a blocker or limitation, never a
pass; a missing check is not reported as passing.

## Deployment and rollback

This draft authorizes no deployment. Future staging rollout (when F3 exists
and the open decisions are resolved) is ordered: adopted-baseline
build/checks → apply the three new-table migrations under the approved
procedure → deploy → health/smoke → verify onboarding → settings
read/write → email-change request/confirm (non-production addresses only)
→ account-deletion request/sweep (non-production identity only) → the full
core-loop E2E suite → the accessibility/performance automation suites →
cross-user/CSRF/idempotency validation → monitoring → then a new-tables
rollback rehearsal. Trigger rollback on: any account deletion or email
change affecting a wrong or unauthorized account; a stuck/incomplete
account-deletion sweep leaving a deactivated-but-never-purged or
purged-without-deactivation state; a broken onboarding gate locking learners
out of the core loop; a regression in any A1–P4 screen; migration/schema
failure; or an accessibility/performance regression the new automation
would have caught but did not (a gap in the automation itself, not only a
gap in the product). Roll back/recover under the approved procedure:
preserve `user_settings`/onboarding state, ensure no account is left
partially deactivated, validate with non-production identities, and record
the last-known-good revision; production activation remains separately
governed, and account-deletion production enablement additionally requires
the founder go/no-go and legal review noted in `VOC-031-DEP-03`.
