# VOC-022 — Acceptance Criteria

All criteria are observable and bidirectionally traceable to decisions
`VOC-022-D00`..`D09`, task `VOC-022-T00`, tests `VOC-022-TEST-00`..`TEST-08`,
and evidence `VOC-022-EV-00`..`EV-08`.

## VOC-022-AC-00 — New dynamic route renders a per-situation word list from local mock data

- Requirement source: `VOC-022-D00`, `VOC-022-D01`, `VOC-022-D02`
- Tasks: `VOC-022-T00`
- Tests: `VOC-022-TEST-00`
- Evidence: `VOC-022-EV-00`
- Result: pending

`apps/web/src/app/(app)/discover/[situation]/page.tsx` exists as an async
server component that `await`s its `params` prop. For each of the seven
known situation slugs (`airport`, `restaurant`, `hotel-check-in`,
`job-interview`, `daily-conversation`, `work-meeting`, `university-class`),
navigating to `/discover/{slug}` renders at least three word entries sourced
from a single local mock data structure; no API/network call is present.

## VOC-022-AC-01 — Each word entry shows the word and a short meaning; saved state uses both icon and text

- Requirement source: `VOC-022-D02`, `VOC-022-D03`
- Tasks: `VOC-022-T00`
- Tests: `VOC-022-TEST-01`
- Evidence: `VOC-022-EV-01`
- Result: pending

Each word entry renders the word/phrase and one short meaning as plain
text. Every situation's word list includes at least one mock-saved word and
at least one mock-not-saved word. A mock-saved word's marker combines a
decorative, `aria-hidden` icon with a visible text label (e.g. "Saved");
this marker is entirely absent — not merely recolored — on a
mock-not-saved word.

## VOC-022-AC-02 — No word detail, save/unsave, or sentence-practice affordance on any word entry

- Requirement source: `VOC-022-D06`
- Tasks: `VOC-022-T00`
- Tests: `VOC-022-TEST-02`
- Evidence: `VOC-022-EV-02`
- Result: pending

No word entry renders an `href`, `onClick`, `<Link>`, or `<button>`. No
control anywhere on the page can change the mock `isSaved` flag. No
sentence-practice entry point (button, link, or embedded component) appears
anywhere on the drill-down page.

## VOC-022-AC-03 — Unknown/invalid situation slug triggers Next's built-in not-found handling

- Requirement source: `VOC-022-D04`
- Tasks: `VOC-022-T00`
- Tests: `VOC-022-TEST-03`
- Evidence: `VOC-022-EV-03`
- Result: pending

Navigating to `/discover/{unknown-slug}` (a slug absent from the seven known
situations) calls `notFound()` from `next/navigation` and resolves to
Next.js's built-in not-found handling — not a thrown client-side error, not
a blank page, and not any custom lookup against a backend.

## VOC-022-AC-04 — `discover/page.tsx`'s only change is the navigation addition

- Requirement source: `VOC-022-D05`
- Tasks: `VOC-022-T00`
- Tests: `VOC-022-TEST-04`
- Evidence: `VOC-022-EV-04`
- Result: pending

Each of the seven situation cards on `apps/web/src/app/(app)/discover/page.tsx`
is wrapped in a `next/link` `<Link href="/discover/{slug}">` pointing at its
own drill-down route. Diffing this file against its pre-VOC-022 state shows
no change to `MOCK_DISCOVER_SITUATIONS`'s content/order, the `<h1>Journey</h1>`
heading, or the intro paragraph — the only substantive change is the
`next/link` import and the `<Link>` wrapper (plus, if needed, moving
existing className styling from `<li>` onto the new `<Link>` and adding
hover/focus-visible styling built only from wired tokens).

## VOC-022-AC-05 — Static UI only: no backend, no protected-area touch

- Requirement source: `VOC-022-D00`
- Tasks: `VOC-022-T00`
- Tests: `VOC-022-TEST-05`
- Evidence: `VOC-022-EV-05`
- Result: pending

`apps/web/src/app/(app)/discover/[situation]/page.tsx`,
`apps/web/src/app/(app)/discover/page.tsx`, and any colocated subcomponent
under `(app)/discover/_components/` (if the implementer adds one) contain no
import of an API client, `fetch`, `TanStack Query`, or any network call.
`apps/api`, any authentication code path, and any database/migration file
are unchanged. `apps/web/src/app/(app)/_components/bottom-nav.tsx`,
`apps/web/src/app/(app)/layout.tsx`, `apps/web/src/app/(app)/home/page.tsx`,
and `apps/web/src/app/(app)/progress/page.tsx` are byte-for-byte unchanged.

## VOC-022-AC-06 — Only wired design tokens are used, no bare duration/easing utility

- Requirement source: `VOC-022-D07`
- Tasks: `VOC-022-T00`
- Tests: `VOC-022-TEST-06`
- Evidence: `VOC-022-EV-06`
- Result: pending

Every color, spacing, radius, shadow, and easing/duration utility or
`var(--…)` reference in the new/changed markup resolves to one of the 64
custom properties VOC-018 emitted (`spacing`, `neutral`, `primary`,
`secondary`, `fontSize`/`text-*`, `radius`, `elevation`/`shadow-*`,
`easing`/`ease-*`, `duration`). No raw hex color literal and no
`feedback`/`success`/`warning`/`error` Tailwind utility class appears
anywhere in the new/changed code. No bare `duration-<name>` or
`ease-<name>` utility class appears — only the `var(--…)` arbitrary-value
form or an auto-generated color/spacing utility.

## VOC-022-AC-07 — Semantic structure: one `<h1>`, `<ul>`/`<li>` word list, no second `<main>`

- Requirement source: `VOC-022-D08`
- Tasks: `VOC-022-T00`
- Tests: `VOC-022-TEST-07`
- Evidence: `VOC-022-EV-07`
- Result: pending

`[situation]/page.tsx` renders exactly one `<h1>` (the situation title) and
no other heading level for individual word entries. The word list is
wrapped in a `<ul>` with one `<li>` per word. The page's root element is a
`<div>`, not `<main>`. A "Back to Journey" link to `/discover` is present.

## VOC-022-AC-08 — Deterministic checks pass

- Requirement source: `VOC-022-D00`..`D09`
- Tasks: `VOC-022-T00`
- Tests: `VOC-022-TEST-08`
- Evidence: `VOC-022-EV-08`
- Result: pending

`pnpm run lint:web` (including
`scripts/foundation/check-tailwind-token-usage.mjs`), `pnpm run typecheck:web`
(including `next typegen`, which validates the new route's `params: Promise`
shape), and `pnpm run build:web` all exit zero, and `pnpm run format:check`
reports the `apps/web` changes clean, with no new lint/type/format errors
introduced elsewhere in the workspace. No new runtime dependency, and no
change to `apps/web/package.json` or the lockfile.
