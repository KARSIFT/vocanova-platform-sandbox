# VOC-019 — Build the Real Home Screen: Specification

## Objective and requirement source

Replace the placeholder `apps/web/src/app/(app)/home/page.tsx` (a bare `<h1>Home</h1>`
plus one filler line, added as part of VOC-017's navigation shell) with the real
Home screen, built with **static UI only against local mocked/placeholder data**
— `apps/api` has no real endpoints yet, so this task introduces no network call,
API client import, authentication, or database/migration work.

Requirement source:
- DOC-01 (`docs/product/01-mvp-prd.md`) §2 — "**Home** — Today's Mission,
  streak, due-review count, quick entry to discovery and sentence practice."
- DOC-03 (`docs/design/03-ui-ux-design.md`) §4 item 1 — "**Home** shows the
  day's mission (review target, progress toward it), current streak, and a
  route into Journey/Discover if the learner wants new words."
- DOC-03 §8 — Progress/streak/Confidence-Point framing, cited by the free-text
  request; note this section's primary subject is the separate **Progress**
  tab, not Home. The only part of §8 applicable to Home is the general
  "encouraging, not gamified for its own sake" tone (§1) and the "lightweight,
  not a full-screen interruption" celebration-moment guidance, which this task
  does not need to implement because a fixed mock state has no completion event
  to celebrate (`VOC-019-D07`).

Both DOC-01 and DOC-03 are `status: approved`, `owner: founder`. Per
`AGENTS.md`, an approved document describes target UX but is not itself
implementation authority for a specific change; a founder-approved,
implementation-ready state must still be recorded at adoption. This document is
a draft, not an approved specification.

## Scope and non-goals

In scope (all under `apps/web/src/app/(app)/home/`):
1. **Today's Mission** — the review target for the day and progress toward it,
   shown both visually and as explicit text (not color alone), from local mock
   data.
2. **Current streak** — a text stat (e.g. "5-day streak").
3. **Due-review count** — a text stat (e.g. "8 words due today"), display-only
   (`VOC-019-D02`).
4. **Route into Journey/Discover** — exactly one link into the existing
   `/discover` route (the route VOC-017 built for the Journey tab, per
   `VOC-017-DEP-04`'s founder-resolved choice), meeting the same accessibility
   conventions `BottomNav` already established.

Explicitly excluded:
- Any backend/API call, data fetching, `TanStack Query`, or import of an API
  client. `apps/api`, authentication, onboarding, and any database/migration
  work are untouched (`VOC-019-D00`).
- A `/review` route or a "start review" action — no such route exists on
  `develop` yet (VOC-017 built only `/home`, `/discover`, `/progress`); adding
  one is out of this task's scope (`VOC-019-D02`).
- The reusable sentence-practice component and any quick-entry point into it —
  DOC-03 §2 defines it as a component invoked from Home, Word Detail, and
  Review Completion, but building or wiring it is a separate, larger scope than
  a static mock Home screen; see the open scope note (`VOC-019-D01`).
- The `feedback` design-token scale (`success`/`warning`/`error`, added by
  VOC-016). VOC-018 explicitly declined to wire it into `apps/web`'s Tailwind
  `@theme` layer (`VOC-018-DEP-05`), and it remains unwired as of this draft
  (`apps/web/src/app/tokens.generated.css` has no `--color-success-*`,
  `--color-warning-*`, or `--color-error-*` properties). Wiring it is a
  follow-up package's scope, not this one (`VOC-019-D05`).
- Any new runtime dependency, `package.json`, or lockfile change.
- Any change to `apps/web/src/app/(app)/_components/bottom-nav.tsx`,
  `apps/web/src/app/(app)/layout.tsx`, `apps/web/src/app/(app)/discover/page.tsx`,
  or `apps/web/src/app/(app)/progress/page.tsx`.
- DOC-03 §9's empty/loading/error states, since there is no dynamic data source
  to be empty, loading, or erroring (`VOC-019-D06`).
- Celebration-moment animation on mission completion (DOC-03 §8) — deferred
  because a fixed mock state has no real completion event to animate
  (`VOC-019-D07`).

## Risk and protected areas

Proposed risk: **R1**. Every target path is under
`apps/web/src/app/(app)/home/**`, which the path classifier maps to its `*`
default (R1) — no target matches an R2/R3/R4 pattern (no `package.json`, no
`*/auth/*`, no `*/migrations/*`, no governance or CI-workflow path). This is a
draft proposal — `classify-change-risk.sh` and a human's judgment govern the
actual class at implementation time.

Semantic note (does **not** change the path floor, but the independent verifier
must review it per `CLAUDE.md`): this is the first screen to render real
learner-facing token colors as a large content surface (VOC-017's `BottomNav`
was a thin persistent bar); WCAG 2.2 AA contrast (1.4.3 text, 1.4.11 non-text)
against the `primary`/`neutral` token pairs used here is a genuine review
dimension, as VOC-018's impact analysis flagged would fall to "the consuming
change." No protected areas are touched.

## Decisions, contradictions, security, and privacy

`VOC-019-D00` — **Static UI, local mock data, no backend.** All four Home
elements are rendered from a single, clearly-named local mock data structure
(e.g. `MOCK_HOME_STATE`) colocated with the page, explicitly commented as
placeholder pending real API wiring. No `fetch`, `TanStack Query`, or API
client import is introduced. This keeps the frontend from "inventing" progress
state in a way indistinguishable from real backend-authoritative data (DOC-03
§1's hard rule is about runtime behavior, not a static-mock development task,
but the mock must still be visibly a mock in the source so a future
implementer does not mistake it for a live integration).

`VOC-019-D01` — **Sentence-practice entry and review-start action excluded
(OPEN, `VOC-019-DEP-04`).** DOC-01 §2 and DOC-03's information architecture
(§2) both describe Home as including "quick entry to discovery and sentence
practice" / "quick sentence-practice entry." The free-text request, however,
explicitly grounds Home's content in DOC-03 **§4 item 1**, which lists only
mission + progress, streak, and "a route into Journey/Discover" — no mention of
sentence practice or a review-start action — and the request's own scope
sentence names exactly those four elements. This draft follows the narrower,
explicitly-requested scope rather than the broader IA list, for two concrete
reasons beyond the request's own wording: (a) the sentence-practice component
does not exist yet as a reusable component in this codebase, so a Home
"quick-entry" link would have nothing real to point to; (b) DOC-03 §2 lists
Review Completion as sentence practice's primary MVP entry point, with Home as
one of three, so deferring Home's entry point does not block the sentence
feature's own MVP completion criterion (DOC-01 §3 item 5). This is recorded as
an open scope decision, not silently resolved: a human should confirm the
narrower scope or direct a follow-up task once `/review` and the
sentence-practice component exist.

`VOC-019-D02` — **Due-review count is a text stat, not a link.** No `/review`
route exists on `develop` (VOC-017 built `/home`, `/discover`, `/progress`
only). Rendering the due-review count as an actionable "Start review" link
would point at a route that 404s. This task therefore renders it as a plain
text stat (e.g. "8 words due today"); adding the link is the natural first step
of the follow-up package that builds `/review` (`VOC-019-DEP-04`).

`VOC-019-D03` — **Journey/Discover route target is `/discover`.** VOC-017
resolved the Journey tab's route to `/discover` (`VOC-017-DEP-04`,
founder-resolved), confirmed by the current `BottomNav`
(`apps/web/src/app/(app)/_components/bottom-nav.tsx`), which links the
"Journey" label to `/discover`. Home's route-in link therefore targets
`/discover` — the same destination the bottom nav's Journey tab already
resolves to — rather than a new or different path.

`VOC-019-D04` — **Component structure.** `home/page.tsx` remains a **server
component** (no client state or interactivity is needed for static mock
content — no button triggers a state change, the only interactive element is a
plain navigational `<Link>`). If the implementer finds the file grows
unwieldy, presentational subcomponents may be colocated under
`(app)/home/_components/` (a Next.js `_`-prefixed private, non-routable
folder), mirroring the colocation convention `bottom-nav.tsx` already
established for `(app)/_components/`. Introducing `"use client"` is **not**
warranted by anything in this task's scope.

`VOC-019-D05` — **Token usage restricted to what VOC-018 actually wired.** Only
the eight scales VOC-018 emitted as Tailwind v4 `@theme` custom properties are
available for use: `spacing`, `neutral`, `brand.primary`/`brand.secondary`
(as `primary`/`secondary` utilities), `fontSize`, `radius`, `elevation`, and
`easing`/`duration` (via `var()`/arbitrary-value only, per VOC-018's `D01`/`D04`
— no named duration utility class exists). No raw hex color and no
`feedback`/`success`/`warning`/`error` Tailwind utility may be used (the
`feedback` scale is unwired — `VOC-018-DEP-05`, confirmed still true at this
draft's authoring by inspecting `apps/web/src/app/tokens.generated.css`).
Mission progress and streak visuals use `primary`/`secondary`/`neutral` only,
which also happens to match DOC-03 §11's "avoid grading colors" guidance.

`VOC-019-D06` — **Empty/loading/error states deferred (OPEN,
`VOC-019-DEP-05`).** DOC-03 §9 requires empty/loading/error states "for every
screen with dynamic content." This task's Home screen has no dynamic data
source — a single fixed local mock state, no network call — so there is no
loading or error condition to render. The real API-wiring package (a required
follow-up, not part of this task) must add DOC-03 §9's states when Home starts
fetching live mission/streak/due-count data from `apps/api`. This mirrors the
pattern VOC-017 used to defer automated a11y testing as a named follow-up
(`VOC-017-D08`) rather than silently skipping a requirement.

`VOC-019-D07` — **No celebration-moment animation.** DOC-03 §8 describes
lightweight celebration moments on mission completion or streak milestones.
Because this task's mock data represents one fixed, arbitrary point of
progress (not a real completion event a learner just triggered), implementing
a celebration animation here would be presentational invention with nothing
real to react to. Deferred to the real API-wiring package, where a genuine
mission-completion transition exists to animate.

### Contradictions

The DOC-01 §2 / DOC-03 IA §2 vs. DOC-03 §4-item-1 scope tension over
sentence-practice/review-start on Home is recorded, not silently resolved
(`VOC-019-D01`). No other contradiction between DOC-01, DOC-03, and the
free-text request was found.

### Security and privacy

None. The screen renders static mock text/markup and one navigational link. No
input fields, no network/API calls, no storage, no cookies/tokens, no
authentication or authorization surface, and no personal data (the mock streak
and due-count numbers are not derived from or associated with any real
learner). DOC-03 §1's "backend is authoritative" rule is not engaged at
runtime because nothing here fetches or persists progress/mission/scheduling
state; the mock data's placeholder nature is made explicit in the source so it
is never mistaken for a live integration by a future maintainer.

## Data, migrations, analytics, and accessibility

- **Data / migrations:** none. No database, schema, or migration; no
  `apps/api` change. Rollback is a plain `git revert`.
- **Analytics:** none. A static mock screen emits no analytics events.
- **Accessibility:** in scope. Mission progress is expressed both visually
  (e.g. a progress bar) and in explicit text (e.g. "3 of 10 words reviewed
  today"), so progress is never conveyed by color alone (DOC-03 §10). The
  due-review count and streak are plain text, inherently accessible to screen
  readers. The Journey/Discover link is a real `<Link>` with a visible text
  label, a ≥44×44px touch target, and a visible focus indicator, consistent
  with the conventions `BottomNav` established in VOC-017
  (`VOC-017-D03`/`AC-02`). Honest limitation, matching VOC-017's precedent
  (`VOC-017-D08`): with no `apps/web` component/a11y test runner yet, this is
  verified by construction and structured inspection, not automated
  axe/Playwright assertions.
