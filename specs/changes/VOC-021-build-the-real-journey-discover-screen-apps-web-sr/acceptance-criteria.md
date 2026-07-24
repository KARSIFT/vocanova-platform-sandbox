# VOC-021 — Acceptance Criteria

All criteria are observable and bidirectionally traceable to decisions
`VOC-021-D00`..`D08`, task `VOC-021-T00`, tests `VOC-021-TEST-00`..`TEST-06`,
and evidence `VOC-021-EV-00`..`EV-06`.

## VOC-021-AC-00 — Seven situation cards render from local mock data

- Requirement source: `VOC-021-D00`, `VOC-021-D01`
- Tasks: `VOC-021-T00`
- Tests: `VOC-021-TEST-00`
- Evidence: `VOC-021-EV-00`
- Result: pending

`apps/web/src/app/(app)/discover/page.tsx` renders exactly seven situation
entries — Airport, Restaurant, Hotel Check-in, Job Interview, Daily
Conversation, Work Meeting, University Class, in that order — sourced from a
single local mock data structure; no API/network call is present.

## VOC-021-AC-01 — Each card shows a label and a short description only

- Requirement source: `VOC-021-D02`
- Tasks: `VOC-021-T00`
- Tests: `VOC-021-TEST-01`
- Evidence: `VOC-021-EV-01`
- Result: pending

Each situation entry renders its title as a heading and one short
descriptive sentence as plain text. No category badge, `level_band` badge,
word-count figure, saved-word indicator, search box, or filter control is
present anywhere on the page.

## VOC-021-AC-02 — Cards are non-interactive; no drill-down route exists

- Requirement source: `VOC-021-D03`
- Tasks: `VOC-021-T00`
- Tests: `VOC-021-TEST-02`
- Evidence: `VOC-021-EV-02`
- Result: pending

No situation entry renders an `href`, `onClick`, `<Link>`, or `<button>`. No
`apps/web/src/app/(app)/discover/[situation]` route (or any similarly named
dynamic segment) exists in the change set.

## VOC-021-AC-03 — Static UI only: no backend, no protected-area touch

- Requirement source: `VOC-021-D00`
- Tasks: `VOC-021-T00`
- Tests: `VOC-021-TEST-03`
- Evidence: `VOC-021-EV-03`
- Result: pending

`apps/web/src/app/(app)/discover/page.tsx` (and any colocated subcomponent
under `(app)/discover/_components/`, if the implementer adds one) contains
no import of an API client, `fetch`, `TanStack Query`, or any network call.
`apps/api`, any authentication code path, and any database/migration file
are unchanged. `apps/web/src/app/(app)/_components/bottom-nav.tsx`,
`apps/web/src/app/(app)/layout.tsx`, `apps/web/src/app/(app)/home/page.tsx`,
and `apps/web/src/app/(app)/progress/page.tsx` are byte-for-byte unchanged.

## VOC-021-AC-04 — Only wired design tokens are used, no bare duration/easing utility

- Requirement source: `VOC-021-D06`
- Tasks: `VOC-021-T00`
- Tests: `VOC-021-TEST-04`
- Evidence: `VOC-021-EV-04`
- Result: pending

Every color, spacing, radius, shadow, and easing/duration utility or
`var(--…)` reference in the new markup resolves to one of the 64 custom
properties VOC-018 emitted (`spacing`, `neutral`, `primary`, `secondary`,
`fontSize`/`text-*`, `radius`, `elevation`/`shadow-*`, `easing`/`ease-*`,
`duration`). No raw hex color literal and no `feedback`/`success`/`warning`/
`error` Tailwind utility class appears anywhere in the new code. No bare
`duration-<name>` or `ease-<name>` utility class appears — only the
`var(--…)` arbitrary-value form.

## VOC-021-AC-05 — Semantic list structure, single `<h1>`, no second `<main>` landmark

- Requirement source: `VOC-021-D04`, `VOC-021-D05`, `VOC-021-D07`
- Tasks: `VOC-021-T00`
- Tests: `VOC-021-TEST-05`
- Evidence: `VOC-021-EV-05`
- Result: pending

The seven situation entries are wrapped in a `<ul>` with one `<li>` per
situation. `discover/page.tsx` does not render its own `<main>` element; its
root element is a `<div>`. The page renders exactly one `<h1>` ("Journey"),
with each situation title as an `<h2>`.

## VOC-021-AC-06 — Deterministic checks pass

- Requirement source: `VOC-021-D00`..`D08`
- Tasks: `VOC-021-T00`
- Tests: `VOC-021-TEST-06`
- Evidence: `VOC-021-EV-06`
- Result: pending

`pnpm run lint:web` (including
`scripts/foundation/check-tailwind-token-usage.mjs`), `pnpm run typecheck:web`,
and `pnpm run build:web` all exit zero, and `pnpm run format:check` reports
the `apps/web` changes clean, with no new lint/type/format errors introduced
elsewhere in the workspace. No new runtime dependency, and no change to
`apps/web/package.json` or the lockfile.
