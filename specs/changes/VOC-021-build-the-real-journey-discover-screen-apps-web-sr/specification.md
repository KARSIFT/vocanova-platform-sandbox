# VOC-021 — Build the Real Journey/Discover Screen: Specification

## Objective and requirement source

Replace the placeholder `apps/web/src/app/(app)/discover/page.tsx` (a bare
`<h1>Journey</h1>` plus one filler line, added as part of VOC-017's
navigation shell) with the real top-level Journey/Discover screen, built with
**static UI only against local mocked/placeholder data** — `apps/api` has no
real endpoints yet, so this task introduces no network call, API client
import, authentication, or database/migration work.

Requirement source:
- DOC-01 (`docs/product/01-mvp-prd.md`) §2 — "**Journey** — situation-based
  discovery (Airport, Restaurant, Hotel Check-in, Job Interview, Daily
  Conversation, Work Meeting, University Class, etc.), word detail,
  save/unsave."
- DOC-03 (`docs/design/03-ui-ux-design.md`) §5 — "Discovery is organized by
  real-life situation … not by grammar topic or difficulty tier alone …
  Within a situation, words are shown one at a time or as a short scannable
  list …"
- DOC-05 (`docs/engineering/05-database-design.md`) §8 — the
  `journey_situations` model: `slug` (unique), `title`, `short_description`,
  `level_band`, `category`, `status`, `display_order`.

All three are `status: approved`, `owner: founder`. Per `AGENTS.md`, an
approved document describes target UX but is not itself implementation
authority for a specific change; a founder-approved, implementation-ready
state must still be recorded at adoption. This document is a draft, not an
approved specification.

## Scope and non-goals

In scope (all under `apps/web/src/app/(app)/discover/`):
1. **A static list of seven situation cards** — Airport, Restaurant, Hotel
   Check-in, Job Interview, Daily Conversation, Work Meeting, University
   Class — from local mock data (`VOC-021-D00`, `VOC-021-D01`).
2. **Each card shows a label (the situation title) and a short description**
   only (`VOC-021-D02`).
3. **Cards are non-interactive** — no link, button, or click handler, and no
   `/discover/[situation]` route is created (`VOC-021-D03`).
4. **Presented as a semantic list** (`<ul>`/`<li>`) in a responsive layout —
   single column at mobile widths, a multi-column grid at wider viewports
   (`VOC-021-D04`).

Explicitly excluded, per the free-text request:
- Any word-list-within-situation, word detail, or save/unsave control. DOC-01
  §2 bundles these into the full Journey tab; this task is the top-level
  situation list only (`VOC-021-D02`).
- Any navigation into a situation (no `/discover/[situation]` route, no
  `href`/`onClick`/`<Link>`/`<button>` on a card). Deferred to a follow-up
  package (`VOC-021-D03`).
- Any category/level-band badge, word-count figure, saved-word indicator,
  search box, or filter control on the situation list (`VOC-021-D02`).
- Any backend/API call, data fetching, `TanStack Query`, or import of an API
  client. `apps/api`, authentication, and any database/migration work are
  untouched (`VOC-021-D00`).
- Any bare `duration-<name>`/`ease-<name>` Tailwind utility class. Tailwind
  v4 auto-generates utilities only from the `spacing`/`color` (and similar
  recognized) `@theme` namespaces — not from `--duration-*`/`--ease-*`
  custom properties — so only the `var(--…)` arbitrary-value form (e.g.
  `duration-[var(--duration-fast)]`) resolves to anything.
  `scripts/foundation/check-tailwind-token-usage.mjs` (run by `lint:web`)
  deterministically rejects a bare `duration-<name>` class anywhere under
  `apps/web/src/app/(app)/**/*.tsx` (`VOC-021-D06`).
- A second `<main>` landmark. `apps/web/src/app/(app)/layout.tsx` already
  renders the group's one `<main>`; the same
  `check-tailwind-token-usage.mjs` script deterministically rejects any
  route `page.tsx` (other than `layout.tsx` itself) that renders its own
  `<main>` (`VOC-021-D05`).
- Any new runtime dependency, `package.json`, or lockfile change.
- Any change to `apps/web/src/app/(app)/_components/bottom-nav.tsx`,
  `apps/web/src/app/(app)/layout.tsx`, `apps/web/src/app/(app)/home/page.tsx`,
  or `apps/web/src/app/(app)/progress/page.tsx`.
- DOC-03 §9's loading/error states, since there is no dynamic data source to
  be loading or erroring (`VOC-021-D00`).

## Risk and protected areas

Proposed risk: **R1**. Every target path is under
`apps/web/src/app/(app)/discover/**`, which `classify-change-risk.sh` maps to
its `*` default (R1) — no target matches an R2/R3/R4 pattern (no
`package.json`, no `*/auth/*`, no `*/migrations/*`, no governance or
CI-workflow path). This is a draft proposal — `classify-change-risk.sh` and a
human's judgment govern the actual class at implementation time.

Semantic note (does **not** change the path floor, but the independent
verifier must review it per `CLAUDE.md`): like VOC-019's Home screen and
VOC-020's Progress screen, this is a learner-facing content surface rendering
real token colors; WCAG 2.2 AA contrast (1.4.3 text, 1.4.11 non-text) against
the `neutral` token pairs used here is a genuine review dimension.
Additionally, because each card visually resembles a tappable list item
while being deliberately non-interactive in this task (`VOC-021-D03`), the
verifier must confirm no interactive-looking affordance (hover/focus style,
cursor change, implicit link/button semantics) is present that would suggest
tappability where none exists — a genuine UX-affordance risk, not just an
accessibility technicality. No protected areas are touched.

## Decisions, contradictions, security, and privacy

`VOC-021-D00` — **Static UI, local mock data, no backend.** The situation
list is rendered from a single, clearly-named local mock data structure
(e.g. `MOCK_DISCOVER_SITUATIONS`) colocated with the page, explicitly
commented as placeholder pending real API wiring. No `fetch`,
`TanStack Query`, or API client import is introduced. This mirrors VOC-019's
Home-screen precedent (`VOC-019-D00`) and VOC-020's Progress-screen
precedent (`VOC-020-D00`), and keeps the frontend from "inventing" discovery
content in a way indistinguishable from real backend-authoritative data
(DOC-03 §1's hard rule is about runtime behavior, not a static-mock
development task, but the mock must still be visibly a mock in the source so
a future implementer does not mistake it for a live integration). Since
there is no dynamic data source, no loading or error state applies (DOC-03
§9 non-goal).

`VOC-021-D01` — **Situation set and order (OPEN, `VOC-021-DEP-04`).** DOC-01
§2 and DOC-03 §5 both name Airport, Restaurant, Hotel Check-in, Job
Interview, Daily Conversation, Work Meeting, and University Class, followed
by "etc." — an illustrative, not exhaustive, list. This draft treats exactly
those seven, in the order given in the free-text request, as the definitive
mock content, using DOC-05 §8's `journey_situations.display_order` concept
only as naming inspiration for the mock's fixed array order — this task has
no dependency on that table or any backend ordering logic. This is recorded
as an open scope decision, not silently resolved: a human should confirm
this set/order or direct a different one.

`VOC-021-D02` — **Card fields limited to label and short description only.**
DOC-01 §2 bundles "situation-based discovery, word detail, save/unsave" as
one Journey-tab description; the free-text request is explicit that this
task is the top-level situation list only. Accordingly, each card renders
only a title (label) and one short descriptive sentence, sourced from local
mock data with field names inspired by (but not dependent on) DOC-05 §8's
`journey_situations.title`/`short_description` columns. No category badge,
`level_band` badge, word-count figure, saved-word indicator, search box, or
filter control is rendered — those either belong to a within-situation word
list (out of scope) or to a fuller discovery-index experience not requested
here.

`VOC-021-D03` — **No navigation or drill-down into a situation.** Cards are
plain, non-interactive elements: no `href`, `onClick`, `<Link>`, or
`<button>`, and no `/discover/[situation]` (or similarly named) route is
created. This is an explicit non-goal in the free-text request ("do NOT
build … navigation into a situation") — drilling into a situation is a
follow-up package's scope, once `apps/api` has a real
`journey_situations`/`journey_words` endpoint to navigate toward. Rendering a
non-functional link or button here would create a dead-end tap target with
no destination, which DOC-03 §1's "one clear action per screen" principle
and ordinary UX practice both argue against.

`VOC-021-D04` — **Semantic list, responsive grid layout.** The seven cards
are wrapped in a `<ul>`, one `<li>` per situation, so assistive technology
announces "list of 7 items" rather than seven unrelated blocks. Layout is a
single-column stack at mobile widths (360–430px, DOC-03 §1's mobile-first
target) expanding to a multi-column grid at wider viewports, using ordinary
Tailwind responsive-variant utilities (e.g. `sm:grid-cols-2`) — these are
core Tailwind breakpoint utilities, not part of any `@theme` token scale, so
they are unaffected by the token-usage restriction in `VOC-021-D06`.

`VOC-021-D05` — **No second `<main>` landmark.**
`apps/web/src/app/(app)/layout.tsx` already renders the one `<main>` for
every route in this group (confirmed by direct inspection — this task does
not touch `layout.tsx`). `discover/page.tsx` must render a `<div>` at its
root (containing the `<h1>` and the `<ul>`), not `<main>`, matching the rule
`scripts/foundation/check-tailwind-token-usage.mjs` enforces. Component
structure otherwise: `discover/page.tsx` remains a **server component** (no
client state or interactivity is needed for a static, non-interactive mock
list); if the implementer finds the file grows unwieldy, presentational
subcomponents may be colocated under `(app)/discover/_components/`,
mirroring the colocation convention `bottom-nav.tsx` and VOC-019/VOC-020's
screens already established. Introducing `"use client"` is **not** warranted
by anything in this task's scope.

`VOC-021-D06` — **Token usage restricted to what VOC-018 actually wired, and
no bare duration/easing utility class.** Only the eight scales VOC-018
emitted as Tailwind v4 `@theme` custom properties are available for use:
`spacing`, `neutral`, `primary`/`secondary`, `fontSize`, `radius`,
`elevation`, and `easing`/`duration` (via `var(--…)`/arbitrary-value only —
no named `duration-<name>` or `ease-<name>` utility class exists, since
Tailwind v4 does not generate utilities from those two custom-property
namespaces). No raw hex color and no `feedback`/`success`/`warning`/`error`
Tailwind utility may be used (the `feedback` scale is unwired —
`VOC-018-DEP-05`, confirmed still true at this draft's authoring by
inspecting `apps/web/src/app/tokens.generated.css`).
`scripts/foundation/check-tailwind-token-usage.mjs`, run by `lint:web`,
deterministically fails on any bare `duration-<name>` class; there is no
equivalent automated check for raw hex or the unwired `feedback` scale, so
those remain a structured-inspection item (`VOC-021-AC-04`).

`VOC-021-D07` — **Heading hierarchy.** The page root retains a single
`<h1>Journey</h1>` (matching the existing placeholder and the `BottomNav`
tab label, confirmed as `"Journey"` in
`apps/web/src/app/(app)/_components/bottom-nav.tsx`), followed by one `<h2>`
per situation card (its title) and a plain `<p>` for the short description —
never itself a heading. Seven `<h2>` siblings under one `<h1>` is a standard,
valid heading structure.

`VOC-021-D08` — **Mock copy is placeholder text, not reviewed product copy
(OPEN, `VOC-021-DEP-05`).** Each situation's short description is
illustrative text drafted for this package to satisfy "short description"
from the free-text request; it is not founder-reviewed product copy. This is
recorded as an open decision, not silently presented as final: a human
should review the exact wording (`tasks.md` proposes specific text) or
explicitly accept it as adequate placeholder content at adoption.

### Contradictions

DOC-01 §2 describes the Journey tab as bundling situation-based discovery,
word detail, and save/unsave as one MVP-complete surface; this task
intentionally implements only the top-level situation list, per the
free-text request's explicit non-goals (`VOC-021-D02`, `VOC-021-D03`). This
is a disclosed scope narrowing agreed by the requester, not a silently
resolved contradiction — DOC-01's fuller Journey scope remains the target
for follow-up packages. DOC-01 §2's/DOC-03 §5's "etc." after the seven named
situations leaves the exact situation set genuinely open; this draft picks
exactly those seven and records that choice as open, not silently resolved
(`VOC-021-D01`). No other contradiction between DOC-01, DOC-03, DOC-05, and
the free-text request was found.

### Security and privacy

None. The screen renders static mock text/markup only — no links, buttons,
input fields, network/API calls, storage, cookies/tokens, authentication or
authorization surface, and no personal data (the mock situation
titles/descriptions are fixed placeholder content, not derived from or
associated with any real learner). DOC-03 §1's "backend is authoritative"
rule is not engaged at runtime because nothing here fetches or persists
discovery/situation state; the mock data's placeholder nature is made
explicit in the source so it is never mistaken for a live integration by a
future maintainer.

## Data, migrations, analytics, and accessibility

- **Data / migrations:** none. No database, schema, or migration; no
  `apps/api` change. Rollback is a plain `git revert`.
- **Analytics:** none. A static mock screen with no interactive elements
  emits no analytics events.
- **Accessibility:** in scope. The situation list uses a semantic `<ul>`/
  `<li>` structure so assistive technology announces the item count; each
  card's title is a real `<h2>`, not a styled `<div>`, so screen-reader users
  can navigate by heading; because cards are deliberately non-interactive
  (`VOC-021-D03`), they carry no implicit link/button semantics, no
  `tabindex`, and no hover/focus affordance styling that would misleadingly
  suggest tappability. Honest limitation, matching VOC-017/VOC-019/VOC-020
  precedent (`VOC-017-D08`): with no `apps/web` component/a11y test runner
  yet, this is verified by construction and structured inspection, not
  automated axe/Playwright assertions.
