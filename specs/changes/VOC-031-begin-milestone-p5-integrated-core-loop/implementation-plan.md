# VOC-031 — Implementation Plan

## Preconditions and protected areas

Do not begin until this draft is adopted, `D06`–`D08` are resolved, the
adopted `develop` base and its repository commands are recorded, and the
P1→P2→P3→P4 in-repository evidence status is confirmed (`VOC-031-DEP-01`).
`D07` (onboarding-seeds-`user_settings`) is a hard gate on `T00`; `D06`
(settings/account write-field boundary) is a hard gate on `T02`; no later
task may proceed on a guessed resolution. Database schemas/migrations, Ent
schemas, `.github/workflows/*`, every already-shipped A1/P1/P2/P3/P4
transactional write path, requester-scoped authorization, and the committed
OpenAPI/client contract are R3/R4 protected. Preserve existing compatible
work; no P1–P4 backend business-logic mechanic is re-litigated by this
package — only its frontend/reliability wrapping and the new item-3/item-4
surface from specification.md.

## File reconciliation and implementation sequence

First inventory the actual scaffold carried from VOC-025/026/027/028/030:
the A1 `auth` module and `users`/`external_identities` tables, the P1–P4
business modules and their Ent schemas/migrations, the `gamification`
module's existing internal-only `user_settings` upsert
(`EnsureUserSettings`), the already-retired `MOCK_HOME_STATE`/
`MOCK_PROGRESS_STATE` state (confirm they remain retired), and the current
absence of `user_onboarding_profiles`, any `/onboarding`/`/settings`/
`/settings/account` route, and any accessibility/performance test tooling —
repeat the `VOC-031-D00` inspection at the adopted base SHA, since time may
have passed since this draft. Then execute `T00 → T01 → ... → T09` in
order; a task depending on an open decision waits for its resolution rather
than guessing. Keep `onboarding` transaction-scoped (no own-transaction
opens — DOC-06 §3), matching `missions`/`gamification`. Commit generated
OpenAPI and the matched client with their source changes. Do not wire a
frontend read/write to a real endpoint until its approved contract exists
(`T03` follows `T02`, `T01`'s frontend follows its own same-PR backend).
Introduce the shared loading/empty/error components (`T04`) before or in
the same PR series that starts adopting them broadly, so `T03`'s new
screens can use the final version rather than a throwaway one. Do not
invent account-deletion or email-change capability (`D06`), retroactive
onboarding backfill for pre-existing accounts, or any P4 mission/streak/
reward mechanic change.

## Validation and independent verification

Run every installed relevant command discovered at implementation time: root
`pnpm validate`/`pnpm test`/`pnpm build`, the
`scripts/governance/validate-governance.sh` and
`scripts/governance/classify-change-risk.sh` checks as applicable to the
changed paths, Go `gofmt`/`go vet`/`go test`/`go build`, web
lint/typecheck/build/format, the new Playwright/axe-core accessibility
sweep (`T06`) and Lighthouse CI job (`T07`) this package installs, the
migration/API/reliability/consistency tests this package adds, and the
extended mock-inventory check (`T09`). Claude Code independently reviews
each exact final SHA for: scope and the classifier floor; migration safety
and no A1/P1/P2/P3/P4 schema regression; for `T02` specifically, that the
new public write surface follows the exact existing
auth/CSRF/idempotency/requester-scoping pattern with no weakened check; for
`T04`/`T05`/`T08` specifically, that no P1–P4 backend transaction or
business-logic path is altered and that the pre-existing behavior of each
touched screen is unchanged where not explicitly in scope; that the new
accessibility/performance CI jobs are blocking, not advisory (`VOC-031-R06`);
the `D06`/`D07` resolutions actually implemented as recorded (not silently
altered); contract/OpenAPI/client drift; staging/rollback evidence; and
implementer separation. Missing staging, open-decision, or tooling evidence
remains a blocker or limitation, never a pass; a missing check is not
reported as passing.

## Deployment and rollback

This draft authorizes no deployment. Future staging rollout (when F3 exists
and `D06`–`D08` are resolved) is ordered: adopted-baseline build/checks →
apply the one-new-table migration under the approved procedure → deploy →
health/smoke → verify the full onboard → discover → save → review →
complete → sentence-practice → progress → settings loop at all three
supported layouts → cross-user/CSRF/idempotency validation on the two new
write surfaces → accessibility/performance CI job results reviewed →
monitoring → then a new-table rollback rehearsal. Trigger rollback on false
onboarding/settings/account state reaching a learner, suspected cross-user
exposure of onboarding/settings/account data, a confirmed CSRF or
duplicate-write defect on either new endpoint, a regression in any
underlying P1–P4 write path or screen this package's reliability/UX work
touches, an accessibility or performance threshold regression reaching
production, or migration/schema failure. Roll back/recover under the
approved procedure: preserve `user_onboarding_profiles`/`user_settings`/
`users.display_name` state, restore the pre-P5 P1–P4 behavior cleanly,
validate with non-production identities, and record the last-known-good
revision; production activation remains separately governed.
