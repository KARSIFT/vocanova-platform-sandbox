# VOC-019 — Acceptance Criteria

All criteria are observable and bidirectionally traceable to decisions
`VOC-019-D00`..`D07`, task `VOC-019-T00`, tests `VOC-019-TEST-00`..`TEST-05`,
and evidence `VOC-019-EV-00`..`EV-05`.

## VOC-019-AC-00 — Today's Mission shows review target and progress toward it

- Requirement source: `VOC-019-D00`, `VOC-019-D05`
- Tasks: `VOC-019-T00`
- Tests: `VOC-019-TEST-00`
- Evidence: `VOC-019-EV-00`
- Result: pending

`apps/web/src/app/(app)/home/page.tsx` renders a "Today's Mission" section
showing the mock review target (e.g. "10 words") and progress toward it (e.g.
"3 of 10 words reviewed today"), expressed in **both** a visual element (e.g. a
progress bar) and explicit text — never by color alone. The progress value is
computed from local mock data only; no API/network call is present.

## VOC-019-AC-01 — Current streak is shown

- Requirement source: `VOC-019-D00`
- Tasks: `VOC-019-T00`
- Tests: `VOC-019-TEST-01`
- Evidence: `VOC-019-EV-01`
- Result: pending

The page renders the mock current-streak count as a plain text stat (e.g.
"5-day streak"), sourced from local mock data only.

## VOC-019-AC-02 — Due-review count is shown, display-only

- Requirement source: `VOC-019-D02`
- Tasks: `VOC-019-T00`
- Tests: `VOC-019-TEST-02`
- Evidence: `VOC-019-EV-02`
- Result: pending

The page renders the mock due-review count as a plain text stat (e.g. "8 words
due today"). The due-review count is **not** wrapped in a link, button, or any
other control — it has no destination, because no `/review` route exists on
`develop` at this task's scope.

## VOC-019-AC-03 — Exactly one accessible route into Journey/Discover

- Requirement source: `VOC-019-D03`
- Tasks: `VOC-019-T00`
- Tests: `VOC-019-TEST-03`
- Evidence: `VOC-019-EV-03`
- Result: pending

The page renders exactly one navigational link targeting `/discover` (the
route VOC-017 resolved for the Journey tab), with:
1. a visible text label (not icon-only);
2. a ≥44×44px touch target;
3. a visible keyboard focus indicator (no unreplaced `outline-none`).

## VOC-019-AC-04 — Static UI only: no backend, no protected-area touch

- Requirement source: `VOC-019-D00`
- Tasks: `VOC-019-T00`
- Tests: `VOC-019-TEST-04`
- Evidence: `VOC-019-EV-04`
- Result: pending

`apps/web/src/app/(app)/home/page.tsx` (and any colocated subcomponent under
`(app)/home/_components/`, if the implementer adds one) contains no import of
an API client, `fetch`, `TanStack Query`, or any network call. `apps/api`,
any authentication code path, and any database/migration file are unchanged.
`apps/web/src/app/(app)/_components/bottom-nav.tsx`,
`apps/web/src/app/(app)/layout.tsx`,
`apps/web/src/app/(app)/discover/page.tsx`, and
`apps/web/src/app/(app)/progress/page.tsx` are byte-for-byte unchanged.

## VOC-019-AC-05 — Only wired design tokens are used

- Requirement source: `VOC-019-D05`
- Tasks: `VOC-019-T00`
- Tests: `VOC-019-TEST-05`
- Evidence: `VOC-019-EV-05`
- Result: pending

Every color, spacing, radius, shadow, and easing/duration utility or
`var(--…)` reference in the new markup resolves to one of the 64 custom
properties VOC-018 emitted (`spacing`, `neutral`, `primary`, `secondary`,
`fontSize`/`text-*`, `radius`, `elevation`/`shadow-*`, `easing`/`ease-*`,
`duration`). No raw hex color literal and no `feedback`/`success`/`warning`/
`error` Tailwind utility class appears anywhere in the new code.

## VOC-019-AC-06 — Deterministic checks pass

- Requirement source: `VOC-019-D00`..`D07`
- Tasks: `VOC-019-T00`
- Tests: `VOC-019-TEST-06`
- Evidence: `VOC-019-EV-06`
- Result: pending

`pnpm run lint:web`, `pnpm run typecheck:web`, and `pnpm run build:web` all
exit zero, and `pnpm run format:check` reports the `apps/web` changes clean,
with no new lint/type/format errors introduced elsewhere in the workspace. No
new runtime dependency, and no change to `apps/web/package.json` or the
lockfile.
