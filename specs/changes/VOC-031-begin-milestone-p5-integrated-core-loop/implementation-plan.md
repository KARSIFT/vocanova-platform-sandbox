# VOC-031 — Implementation Plan

## Preconditions and protected areas

Do not begin until this draft is adopted, `D01`–`D04` are resolved (or explicitly deferred with a
recorded rationale), the adopted `develop` base and its repository commands are recorded, and the
A1→P1→P2→P3→P4 acceptance-chain status is reconfirmed (`VOC-031-DEP-01`). `apps/web/src/
middleware.ts` (`T01`) is R3-protected as an authentication/authorization path; no later task may
weaken what it authorizes while relaxing what it displays. If `D02`/`D03` activate new CI steps,
`.github/workflows/` is independently protected (deployment/rollback) and that sub-task needs its
own reviewed unit. Preserve existing compatible work; no A1/P1/P2/P3/P4 backend mechanic is
re-litigated, and no new backend business module, schema, migration, or route is introduced.

## File reconciliation and implementation sequence

First inventory the actual scaffold carried from VOC-025/026/027/028/030: confirm
`apps/web/src/app/(app)/{home,progress}` still have their `loading.tsx`/`error.tsx` pair and that
`discover`, `discover/[situation]`, `discover/[situation]/[word]`, and `reviews` still do not;
confirm `apps/web/src/middleware.ts` still conflates non-`401` failures with unauthenticated;
confirm `scripts/foundation/mock-inventory.mjs`'s `expectedMocks` is still empty (no P4-pending
mock reappeared); and confirm F3 staging still does not exist (repeat the `VOC-031-D00` inspection
at the adopted base SHA, since time may have passed since this draft). Then execute
`T00 → T01 → T02 → T03 → T04 → T05` in order. `T00` establishes the loading/error/empty pattern
that `T02`'s consistency audit and `T03`'s accessibility pass both depend on, so it must land
first. `T01`'s middleware change must be reviewed independently of its UI consequences — the
security property (fail-closed) and the UX property (recoverable error instead of silent
sign-out) are both required, and neither substitutes for the other. Do not add a new test harness,
CI dependency, or workflow step under `T03`/`T04` ahead of `D02`/`D03` actually resolving to
require it. Do not build onboarding or Settings screens/APIs ahead of `D01` resolving to require
them.

## Validation and independent verification

Run every installed relevant command discovered at implementation time: root `pnpm validate`
(format/lint/typecheck/test/build), the `scripts/governance/validate-governance.sh` and
`scripts/governance/classify-change-risk.sh` checks as applicable to the changed paths, `go vet`/
`go test`/`go build` for `apps/api` (run to confirm this package causes no backend regression, even
though it changes no backend code), the UI-state/reliability/consistency/accessibility/performance
tests this package adds, and the extended mock-inventory check (`T05`). Claude Code independently
reviews each exact final SHA for: scope (no new backend capability, no onboarding/Settings unless
`D01` requires it); the `T01` fail-closed property with concrete evidence, not an assertion
(`VOC-031-R00`); no idempotency/duplicate-reward regression in the retry-safety changes
(`VOC-031-R01`); design-token/consistency/accessibility/performance evidence quality against
whatever `D02`/`D03` actually resolve, distinguishing a genuine pass from a documented limitation;
contract/OpenAPI drift (should be none — `VOC-031-TEST-34`); and implementer separation. Missing
staging, open-decision, or tooling evidence remains a blocker or limitation, never a pass.

## Deployment and rollback

This draft authorizes no deployment. Future staging rollout (when F3 exists and `D01`–`D04` are
resolved) is ordered: adopted-baseline build/checks → deploy → health/smoke → verify the full
sign-in → discover → save → review → sentence-practice → mission-completes → progress-reads loop
across the `D04`-proposed supported layouts → verify the `T01` fail-closed property under a
simulated backend-outage condition → accessibility/performance evidence review → monitoring → a
rollback rehearsal of this package's frontend/middleware changes. Trigger rollback on: any
authorization regression traced to `T01`; a duplicate reward or lost-input incident traced to
`T01`'s retry-safety changes; a confirmed critical accessibility regression; or a regression in the
underlying A1/P1/P2/P3/P4 loop this package integrates. Roll back/recover under the approved
procedure: redeploy the previous known-good frontend/middleware artifact, confirm the pre-P5
A1–P4 loop still functions unaffected (this package touches no backend write path, so no backend
data state is at risk), validate with non-production identities, and record the last-known-good
revision; production activation remains separately governed.
