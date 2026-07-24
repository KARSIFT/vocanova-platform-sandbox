# VOC-020 — Acceptance Criteria

All criteria are observable and bidirectionally traceable to decisions
`VOC-020-D00`..`D08`, task `VOC-020-T00`, tests `VOC-020-TEST-00`..`TEST-07`,
and evidence `VOC-020-EV-00`..`EV-07`.

## VOC-020-AC-00 — Confidence Points total is shown

- Requirement source: `VOC-020-D00`, `VOC-020-D07`
- Tasks: `VOC-020-T00`
- Tests: `VOC-020-TEST-00`
- Evidence: `VOC-020-EV-00`
- Result: pending

`apps/web/src/app/(app)/progress/page.tsx` renders a "Confidence Points"
stat showing a single mock total (e.g. "1,240 Confidence Points"), sourced
from local mock data only; no API/network call is present, and no ledger or
idempotency logic is reimplemented.

## VOC-020-AC-01 — Current streak is shown

- Requirement source: `VOC-020-D00`, `VOC-020-D02`
- Tasks: `VOC-020-T00`
- Tests: `VOC-020-TEST-01`
- Evidence: `VOC-020-EV-01`
- Result: pending

The page renders the mock current-streak count as a plain text stat (e.g.
"12-day streak"), sourced from local mock data only.

## VOC-020-AC-02 — Longest streak is shown, distinct from current streak

- Requirement source: `VOC-020-D02`
- Tasks: `VOC-020-T00`
- Tests: `VOC-020-TEST-02`
- Evidence: `VOC-020-EV-02`
- Result: pending

The page renders the mock longest-streak count as its own plain text stat
(e.g. "Longest streak: 30 days"), visibly distinct from the current-streak
stat (`VOC-020-AC-01`) — not merged into a single combined number.

## VOC-020-AC-03 — Completion history view is shown, motivational not analytical

- Requirement source: `VOC-020-D01`, `VOC-020-D08`
- Tasks: `VOC-020-T00`
- Tests: `VOC-020-TEST-03`
- Evidence: `VOC-020-EV-03`
- Result: pending

The page renders a 7-entry, day-by-day completion strip covering the current
week from local mock data. Each entry's completed/not-completed state is
expressed in **both** an icon/text label and a visual (fill/marker)
difference — never by color alone. No chart, graph, sparkline, or
percentage-over-time figure is present, and no charting/graphing library is
imported.

## VOC-020-AC-04 — Static UI only: no backend, no protected-area touch

- Requirement source: `VOC-020-D00`
- Tasks: `VOC-020-T00`
- Tests: `VOC-020-TEST-04`
- Evidence: `VOC-020-EV-04`
- Result: pending

`apps/web/src/app/(app)/progress/page.tsx` (and any colocated subcomponent
under `(app)/progress/_components/`, if the implementer adds one) contains
no import of an API client, `fetch`, `TanStack Query`, or any network call,
and no charting/graphing library import. `apps/api`, any authentication code
path, and any database/migration file are unchanged.
`apps/web/src/app/(app)/_components/bottom-nav.tsx`,
`apps/web/src/app/(app)/layout.tsx`,
`apps/web/src/app/(app)/home/page.tsx`, and
`apps/web/src/app/(app)/discover/page.tsx` are byte-for-byte unchanged.

## VOC-020-AC-05 — Only wired design tokens are used, no bare duration/easing utility

- Requirement source: `VOC-020-D05`
- Tasks: `VOC-020-T00`
- Tests: `VOC-020-TEST-05`
- Evidence: `VOC-020-EV-05`
- Result: pending

Every color, spacing, radius, shadow, and easing/duration utility or
`var(--…)` reference in the new markup resolves to one of the 64 custom
properties VOC-018 emitted (`spacing`, `neutral`, `primary`, `secondary`,
`fontSize`/`text-*`, `radius`, `elevation`/`shadow-*`, `easing`/`ease-*`,
`duration`). No raw hex color literal and no `feedback`/`success`/`warning`/
`error` Tailwind utility class appears anywhere in the new code. No bare
`duration-<name>` or `ease-<name>` utility class appears — only the
`var(--…)` arbitrary-value form.

## VOC-020-AC-06 — No second `<main>` landmark

- Requirement source: `VOC-020-D04`
- Tasks: `VOC-020-T00`
- Tests: `VOC-020-TEST-06`
- Evidence: `VOC-020-EV-06`
- Result: pending

`progress/page.tsx` does not render its own `<main>` element; its root
element is a `<div>` or `<section>`, since
`apps/web/src/app/(app)/layout.tsx` already renders the group's one `<main>`.

## VOC-020-AC-07 — Deterministic checks pass

- Requirement source: `VOC-020-D00`..`D08`
- Tasks: `VOC-020-T00`
- Tests: `VOC-020-TEST-07`
- Evidence: `VOC-020-EV-07`
- Result: pending

`pnpm run lint:web` (including
`scripts/foundation/check-tailwind-token-usage.mjs`), `pnpm run typecheck:web`,
and `pnpm run build:web` all exit zero, and `pnpm run format:check` reports
the `apps/web` changes clean, with no new lint/type/format errors introduced
elsewhere in the workspace. No new runtime dependency, and no change to
`apps/web/package.json` or the lockfile.
