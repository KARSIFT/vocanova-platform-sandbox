# VOC-076 — Staging Core-Loop E2E: Review-Queue Answer Button Stays Disabled

**Status: draft, not adopted.** Nothing in this package is implementation-authorized.
It is a draft response to
[issue #575](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/575),
prepared for founder/steward review at adoption time.

## Identity and lifecycle

- Package ID: VOC-076
- Title: Staging Core-Loop E2E — Review-Queue Answer Button Stays Disabled
- Canonical path:
  `specs/changes/VOC-076-staging-core-loop-e2e-review-queue-answer-button`
- Lifecycle state: `draft` (not adopted, not authorized for implementation)
- Proposed risk: `R2` (draft proposal only — see `change.yaml`'s
  `planned_implementation_risk_floor`, not a determination; path floor measured
  at drafting time is `R1` for review UI + staging E2E; `R3` if
  `deploy-staging.yml` enters scope)
- Owner: unassigned (see `change.yaml`'s `owners` block)
- Approval evidence: none yet — `approval_status: not-approved`,
  `implementation_authorized: false`
- Target branch: `develop`
- Linked GitHub issues:
  - [#575](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/575)
    (this package's requirement source)
- Related packages and runs:
  - [VOC-050](specs/changes/VOC-050-run-core-loop-e2e-against-real-staging-and-gate)
    — introduced the staging core-loop gate
  - [VOC-074](specs/changes/VOC-074-voc-065-t01-s-fix-does-not-resolve-the-reviews)
    — hardened step 5/7 against vacuous passes; distinct from this disabled-button
    timeout (`VOC-076-DEP-02`)
  - Failing run:
    [31748423831](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31748423831)
    (develop deploy after VOC-075 merge)

## Why this exists

The staging `deploy-staging` core-loop journey
(`apps/web/tests/staging-e2e/core-loop.staging.spec.ts`) failed on the
`develop` deploy triggered by VOC-075's merge. In step 5 ("work the review
queue"), `reviewOneCard` waited the full **240s** test timeout to click the
first multiple-choice answer button. Playwright's call log shows the locator
resolved to a **disabled** button (`aria-pressed="false"`) that never became
enabled, then was **detached from the DOM** (card re-render or replacement)
before the click could land.

Drafting-time code grounding (not a confirmed root cause): in
`apps/web/src/app/(app)/reviews/_components/review-session.tsx`, multiple-choice
option buttons use `disabled={isLoading || phase === "feedback"}` where
`isLoading = isSubmitting || isRefetching`. A stuck loading flag, a missing
enabled wait in the E2E helper, or staging-only API latency are the leading
candidates named in issue #575 and `VOC-076-DEP-00`.

## What this package does

1. **Confirm the root cause** (`VOC-076-T00`) with direct evidence against the
   failing run and/or a reproducible staging/local path — product hang vs E2E
   race vs staging latency.
2. **Fix the confirmed cause** (`VOC-076-T01`) narrowly (review-session and/or
   staging E2E wait conditions), with regression coverage appropriate to the
   cause.
3. **Verify on real staging** (`VOC-076-T02`) that step 5 completes without the
   disabled-button 240s timeout on a real `deploy-staging.yml` run.

## What this package deliberately does NOT do

- Not a re-litigation of VOC-074's `reviews_completed` increment bug.
- Not an assumed edit to `deploy-staging.yml` unless adoption/`VOC-076-DEP-01`
  expands after T00.
- Not weakening VOC-050's staging gate or VOC-074's `reviewedCards >= 1`
  hardening — this package restores the gate's ability to run past the MC
  answer click.
- Does not adopt itself. `change.yaml` leaves every adoption/authorization
  field at its unadopted default.

## Open questions for the reviewing human

See `specification.md`. The most important at adoption:

1. Accept proposed `R2` semantic elevation above the measured `R1` path floor
   (and the R3 raise if `deploy-staging.yml` is in scope).
2. Investigation priority among the `VOC-076-DEP-00` candidates.
3. Whether VOC-074 open tasks need explicit merge-order coordination
   (`VOC-076-DEP-02`).

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`. This package
carries no standing approval; adoption, implementation authorization, independent
verification, and any required human approval remain to be recorded against the
exact implemented revision, per AGENTS.md and CLAUDE.md.
