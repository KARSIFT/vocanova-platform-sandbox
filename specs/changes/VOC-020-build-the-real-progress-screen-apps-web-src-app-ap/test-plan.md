# VOC-020 — Test Plan

Note on tooling: `apps/web` has no test runner yet (no Vitest/RTL/Playwright
— the root `test` script runs only foundation and Go tests), the same gap
VOC-017 and VOC-019 recorded. These tests therefore combine the available
deterministic build/type/lint/format checks (including
`scripts/foundation/check-tailwind-token-usage.mjs`, run by `lint:web`) with
a structured code-inspection checklist. No secrets or production data are
used by any test.

## VOC-020-TEST-00 — Confidence Points total is shown

- Covers: `VOC-020-AC-00`
- Preconditions: `VOC-020-T00` complete.
- Procedure: inspect `progress/page.tsx`; confirm a Confidence Points stat
  renders a mock total as plain text, sourced from local mock data only;
  confirm no `fetch`, API client import, or `TanStack Query` usage appears
  in the file.
- Expected result: Confidence Points total visible as text, sourced from
  local mock data only.
- Evidence: `VOC-020-EV-00`

## VOC-020-TEST-01 — Current streak is shown

- Covers: `VOC-020-AC-01`
- Preconditions: `VOC-020-T00` complete.
- Procedure: inspect `progress/page.tsx`; confirm the mock current-streak
  count renders as plain text.
- Expected result: current-streak stat visible as text.
- Evidence: `VOC-020-EV-01`

## VOC-020-TEST-02 — Longest streak is shown, distinct from current streak

- Covers: `VOC-020-AC-02`
- Preconditions: `VOC-020-T00` complete.
- Procedure: inspect `progress/page.tsx`; confirm the mock longest-streak
  count renders as its own plain-text stat, not merged with or
  indistinguishable from the current-streak stat.
- Expected result: two separately labeled stats — current streak and
  longest streak.
- Evidence: `VOC-020-EV-02`

## VOC-020-TEST-03 — Completion history is a motivational, accessible strip

- Covers: `VOC-020-AC-03`
- Preconditions: `VOC-020-T00` complete.
- Procedure:
  1. Inspect `progress/page.tsx`; confirm a 7-entry completion-history strip
     renders from local mock data, one marker per day of the current week.
  2. Confirm each marker's completed/not-completed state is expressed in
     both an icon/text label and a visual fill difference — disabling color
     must still leave every marker's state legible from text/icon alone.
  3. Confirm no chart/graph/sparkline component, percentage-over-time
     figure, or charting/graphing library import is present.
- Expected result: a 7-day completion strip, motivational in tone, with
  every marker's state legible without color and no analytical/chart
  visualization present.
- Evidence: `VOC-020-EV-03`

## VOC-020-TEST-04 — Static UI only: no backend, no protected-area touch

- Covers: `VOC-020-AC-04`
- Preconditions: `VOC-020-T00` complete.
- Procedure:
  1. Run `git diff --stat` against the base SHA; confirm every changed path
     is under `apps/web/src/app/(app)/progress/`.
  2. Confirm `bottom-nav.tsx`, `(app)/layout.tsx`, `home/page.tsx`, and
     `discover/page.tsx` are byte-for-byte unchanged.
  3. Confirm no file under `apps/api/**` changed, and no auth or migration
     file changed.
  4. Grep the new/changed files for `fetch(`, an
     `@vocanova/api-client`-style import, `TanStack Query` usage, or a
     charting/graphing library import; confirm none are present.
- Expected result: change set confined to
  `apps/web/src/app/(app)/progress/`; no backend, auth, or migration file
  touched; no data-fetching or charting code introduced.
- Evidence: `VOC-020-EV-04`

## VOC-020-TEST-05 — Only wired design tokens are used, no bare duration/easing utility

- Covers: `VOC-020-AC-05`
- Preconditions: `VOC-020-T00` complete.
- Procedure:
  1. List every Tailwind color/spacing/radius/shadow/easing utility class
     and every `var(--…)` reference in the new/changed files.
  2. Confirm each resolves to one of the 64 custom properties
     `apps/web/src/app/tokens.generated.css` emits (`spacing`, `neutral`,
     `primary`, `secondary`, `fontSize`/`text-*`, `radius`,
     `elevation`/`shadow-*`, `easing`/`ease-*`, `duration`).
  3. Confirm no raw hex color literal and no `feedback`/`success`/`warning`/
     `error` utility class appears.
  4. Run `pnpm run lint:web` and confirm
     `scripts/foundation/check-tailwind-token-usage.mjs` reports no bare
     `duration-<name>`/`ease-<name>` utility finding.
- Expected result: every color/spacing/radius/shadow/easing reference
  resolves to a VOC-018-wired custom property; no raw hex, no unwired
  `feedback` scale usage, no bare duration/easing utility class.
- Evidence: `VOC-020-EV-05`

## VOC-020-TEST-06 — No second `<main>` landmark

- Covers: `VOC-020-AC-06`
- Preconditions: `VOC-020-T00` complete.
- Procedure: run `pnpm run lint:web` and confirm
  `scripts/foundation/check-tailwind-token-usage.mjs` reports no nested-`<main>`
  finding for `progress/page.tsx`; separately inspect the file to confirm
  its root element is a `<div>` or `<section>`.
- Expected result: no `<main>` element in `progress/page.tsx`; the
  deterministic check passes.
- Evidence: `VOC-020-EV-06`

## VOC-020-TEST-07 — Deterministic check suite passes

- Covers: `VOC-020-AC-07`
- Preconditions: `VOC-020-T00` complete.
- Procedure: run, from repo root:
  ```bash
  pnpm run lint:web
  pnpm run typecheck:web
  pnpm run build:web
  pnpm run format:check
  ```
  Then confirm `git diff` shows no change to `apps/web/package.json` or the
  lockfile and no new runtime dependency.
- Expected result: all four commands exit zero with no new findings
  elsewhere in the workspace; no dependency/lockfile change.
- Evidence: `VOC-020-EV-07`

No migration or rollback test is applicable (a static front-end file
replacement with no persisted state; rollback is a plain `git revert`). No
authorization/security test is applicable (no auth, input, network, or
storage surface — see `impact-analysis.md`). No automated a11y/component
test is applicable (`apps/web` has no test runner yet — VOC-017's
already-flagged follow-up, `VOC-017-D08`/`DEP-05`, is not reopened or
resolved by this task).
