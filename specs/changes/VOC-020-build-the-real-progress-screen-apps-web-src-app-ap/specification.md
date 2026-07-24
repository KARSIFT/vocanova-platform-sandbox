# VOC-020 — Build the Real Progress Screen: Specification

## Objective and requirement source

Replace the placeholder `apps/web/src/app/(app)/progress/page.tsx` (a bare
`<h1>Progress</h1>` plus one filler line, added as part of VOC-017's
navigation shell) with the real Progress screen, built with **static UI only
against local mocked/placeholder data** — `apps/api` has no real endpoints
yet, so this task introduces no network call, API client import,
authentication, or database/migration work.

Requirement source:
- DOC-01 (`docs/product/01-mvp-prd.md`) §2 — "**Progress** — Confidence
  Points, streak, completion history, motivation-focused summary."
- DOC-03 (`docs/design/03-ui-ux-design.md`) §4 item 5 — "**Progress** shows
  Confidence Points total, streak, and a simple day-by-day or week-by-week
  completion view — motivational, not analytical."
- DOC-03 §8 — "Progress screen is intentionally simple: Confidence Point
  total, current/longest streak, and a completion history view. Celebration
  moments … should be lightweight … not a full-screen interruption …"

Both DOC-01 and DOC-03 are `status: approved`, `owner: founder`. Per
`AGENTS.md`, an approved document describes target UX but is not itself
implementation authority for a specific change; a founder-approved,
implementation-ready state must still be recorded at adoption. This document
is a draft, not an approved specification.

## Scope and non-goals

In scope (all under `apps/web/src/app/(app)/progress/`):
1. **Confidence Points total** — a single running total, shown as a text
   stat, from local mock data (`VOC-020-D07`).
2. **Current streak** — a text stat (e.g. "12-day streak").
3. **Longest streak** — a distinct text stat from current streak (e.g.
   "Longest streak: 30 days"), per DOC-03 §8's explicit "current/longest"
   pairing (`VOC-020-D02`).
4. **Completion history view** — a simple, motivational (not analytical)
   day-by-day strip covering the current week, from local mock data
   (`VOC-020-D01`, `VOC-020-D08`).

Explicitly excluded:
- Any backend/API call, data fetching, `TanStack Query`, or import of an API
  client. `apps/api`, authentication, and any database/migration work are
  untouched (`VOC-020-D00`).
- A charting/graph library, sparkline, or any analytical/statistical
  visualization (percentages-over-time, bar/line charts). DOC-03 §8 is
  explicit that the completion view is "motivational, not analytical"
  (`VOC-020-D08`).
- A celebration-moment animation on streak milestones (DOC-03 §8). Deferred
  because a fixed mock state has no real milestone-crossing event to animate
  (`VOC-020-D03`).
- The true zero-history/first-mission empty state DOC-03 §9 names for this
  exact screen (OPEN, `VOC-020-D06`, `VOC-020-DEP-05`).
- The `feedback` design-token scale (`success`/`warning`/`error`, added by
  VOC-016). VOC-018 explicitly declined to wire it into `apps/web`'s Tailwind
  `@theme` layer (`VOC-018-DEP-05`), and it remains unwired as of this draft
  (`apps/web/src/app/tokens.generated.css` has no `--color-success-*`,
  `--color-warning-*`, or `--color-error-*` properties). Wiring it is a
  follow-up package's scope, not this one.
- Any bare `duration-<name>`/`ease-<name>` Tailwind utility class. Tailwind
  v4 auto-generates utilities only from the `spacing`/`color` (and similar
  recognized) `@theme` namespaces — not from `--duration-*`/`--ease-*`
  custom properties — so only the `var(--…)` arbitrary-value form (e.g.
  `duration-[var(--duration-fast)]`) resolves to anything.
  `scripts/foundation/check-tailwind-token-usage.mjs` (run by `lint:web`)
  deterministically rejects a bare `duration-<name>` class anywhere under
  `apps/web/src/app/(app)/**/*.tsx` (`VOC-020-D05`).
- A second `<main>` landmark. `apps/web/src/app/(app)/layout.tsx` already
  renders the group's one `<main>`; the same
  `check-tailwind-token-usage.mjs` script deterministically rejects any
  route `page.tsx` (other than `layout.tsx` itself) that renders its own
  `<main>` (`VOC-020-D04`).
- Any new runtime dependency, `package.json`, or lockfile change.
- Any change to `apps/web/src/app/(app)/_components/bottom-nav.tsx`,
  `apps/web/src/app/(app)/layout.tsx`, `apps/web/src/app/(app)/home/page.tsx`,
  or `apps/web/src/app/(app)/discover/page.tsx`.
- DOC-03 §9's loading/error states, since there is no dynamic data source to
  be loading or erroring (`VOC-020-D07`).

## Risk and protected areas

Proposed risk: **R1**. Every target path is under
`apps/web/src/app/(app)/progress/**`, which `classify-change-risk.sh` maps to
its `*` default (R1) — no target matches an R2/R3/R4 pattern (no
`package.json`, no `*/auth/*`, no `*/migrations/*`, no governance or
CI-workflow path). This is a draft proposal — `classify-change-risk.sh` and a
human's judgment govern the actual class at implementation time.

Semantic note (does **not** change the path floor, but the independent
verifier must review it per `CLAUDE.md`): like VOC-019's Home screen, this is
a large learner-facing content surface rendering real token colors; WCAG 2.2
AA contrast (1.4.3 text, 1.4.11 non-text) against the `primary`/`secondary`/
`neutral` token pairs used here is a genuine review dimension. The
per-day completion markers additionally raise DOC-03 §10's "no information
conveyed by color alone" rule directly — each marker needs a text/icon label,
not just a fill-color difference between "completed" and "missed" days. No
protected areas are touched.

## Decisions, contradictions, security, and privacy

`VOC-020-D00` — **Static UI, local mock data, no backend.** All four Progress
elements are rendered from a single, clearly-named local mock data structure
(e.g. `MOCK_PROGRESS_STATE`) colocated with the page, explicitly commented as
placeholder pending real API wiring. No `fetch`, `TanStack Query`, or API
client import is introduced. This mirrors VOC-019's Home-screen precedent
(`VOC-019-D00`) and keeps the frontend from "inventing" progress state in a
way indistinguishable from real backend-authoritative data (DOC-03 §1's hard
rule is about runtime behavior, not a static-mock development task, but the
mock must still be visibly a mock in the source so a future implementer does
not mistake it for a live integration).

`VOC-020-D01` — **Completion history is a day-by-day strip over the current
week (OPEN, `VOC-020-DEP-04`).** DOC-03 §4 item 5 offers "day-by-day **or**
week-by-week" without settling which. This draft proposes the simpler,
smaller-surface option: a fixed 7-entry strip, one marker per day of the
current week, each with an explicit completed/not-completed state and a
text/icon label (never color alone, per DOC-03 §10). This is recorded as an
open scope decision, not silently resolved: a human should confirm this
granularity or direct a week-by-week (multi-week) summary instead.

`VOC-020-D02` — **Current and longest streak are shown as two distinct
stats.** DOC-03 §8 explicitly pairs "current/longest streak" as two related
but separate numbers (mirroring `streak_states.current_streak_count` /
`longest_streak_count` in
[05](../../../docs/engineering/05-database-design.md) §12, used here only as
naming inspiration for the mock field names — this task has no dependency on
that table or any backend streak-calculation logic). Rendering only one
combined figure would under-deliver DOC-03 §8's explicit requirement.

`VOC-020-D03` — **No celebration-moment animation.** DOC-03 §8 describes
lightweight celebration moments on streak milestones. Because this task's
mock data represents one fixed, arbitrary point in a streak (not a real
milestone-crossing event a learner just triggered), implementing a
celebration animation here would be presentational invention with nothing
real to react to — the same reasoning VOC-019 used to defer Home's
celebration moment (`VOC-019-D07`). Deferred to the real API-wiring package,
where a genuine streak-milestone transition exists to animate.

`VOC-020-D04` — **No second `<main>` landmark.**
`apps/web/src/app/(app)/layout.tsx` already renders the one `<main>` for
every route in this group (confirmed by direct inspection — this task does
not touch `layout.tsx`). `progress/page.tsx` must render a `<div>` or
`<section>` at its root, not `<main>`, matching the rule
`scripts/foundation/check-tailwind-token-usage.mjs` enforces (added after a
live VOC-019-T00 finding — see that script's own header comment). Component
structure otherwise: `progress/page.tsx` remains a **server component** (no
client state or interactivity is needed for static mock content); if the
implementer finds the file grows unwieldy, presentational subcomponents may
be colocated under `(app)/progress/_components/`, mirroring the colocation
convention `bottom-nav.tsx` and VOC-019's Home screen already established.
Introducing `"use client"` is **not** warranted by anything in this task's
scope.

`VOC-020-D05` — **Token usage restricted to what VOC-018 actually wired, and
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
those remain a structured-inspection item (`VOC-020-AC-05`).

`VOC-020-D06` — **True empty/first-mission state deferred (OPEN,
`VOC-020-DEP-05`).** DOC-03 §9 names the Progress screen specifically:
"empty Progress screen invites the learner to complete their first mission."
This is a stronger, more specific callout than the general "empty states are
required" rule VOC-019 deferred for Home (`VOC-019-D06`) — DOC-03 uses
Progress as its own worked example. This draft nonetheless renders **one**
fixed, populated mock state (an established learner with points, a streak,
and completion history already present), consistent with "static UI only,
using mocked/placeholder data" from the free-text request, and defers the
true zero-state render to the real API-wiring follow-up package — the
natural point where a genuine "no history yet" condition exists to render
accurately. This is recorded as an open decision, not a silent drop: a human
should confirm this deferral or direct that a second, illustrative
empty-state render be added to this task now.

`VOC-020-D07` — **Confidence Points total is a static figure, not a ledger
computation.** `docs/engineering/05-database-design.md` §12 models
`confidence_point_ledger` as an **append-only ledger** (never a mutable
balance field) — this task does not reimplement, simulate, or reference
ledger/idempotency logic; the mock "Confidence Points total" is a single
static number standing in for whatever the real ledger's running balance
would be. "Confidence Points" is the settled product term for this figure
(`docs/product/00-product-bible.md` §"Confidence Points";
`docs/product/README-migration-notes.md` §3 confirms the name was retitled
from "XP" in an earlier draft).

`VOC-020-D08` — **Motivational, not analytical: no chart/graph
component.** DOC-03 §8 states the completion view must be "motivational, not
analytical." The completion history is rendered as a plain strip of
day-labeled completion markers (icon/text + fill, not a numeric chart axis,
percentage-over-time figure, or trend line), and no charting/graphing library
is introduced (also enforced structurally by `VOC-020-AC-06`'s no-new-dependency
requirement). This also keeps visual tone consistent with DOC-03 §11's
"avoid grading/exam-like patterns" guidance — completed/missed days are
distinguished by icon and label, not a red/green pass-fail treatment.

### Contradictions

DOC-03 §4 item 5's "day-by-day or week-by-week" leaves the completion-history
granularity genuinely open; this draft picks day-by-day and records that
choice as open, not silently resolved (`VOC-020-D01`). DOC-03 §9's
Progress-specific empty-state example vs. this task's static-mock scope is
likewise recorded as open, not silently resolved (`VOC-020-D06`). No other
contradiction between DOC-01, DOC-03, and the free-text request was found.

### Security and privacy

None. The screen renders static mock text/markup only — no links, buttons,
input fields, network/API calls, storage, cookies/tokens, authentication or
authorization surface, and no personal data (the mock points/streak/history
values are fixed placeholder numbers, not derived from or associated with any
real learner). DOC-03 §1's "backend is authoritative" rule is not engaged at
runtime because nothing here fetches or persists progress/streak/completion
state; the mock data's placeholder nature is made explicit in the source so
it is never mistaken for a live integration by a future maintainer.

## Data, migrations, analytics, and accessibility

- **Data / migrations:** none. No database, schema, or migration; no
  `apps/api` change. Rollback is a plain `git revert`.
- **Analytics:** none. A static mock screen emits no analytics events.
- **Accessibility:** in scope. Each completion-history day marker expresses
  its completed/not-completed state in both an icon/text label and a fill
  difference — never color alone (DOC-03 §10). Confidence Points, current
  streak, and longest streak are plain text, inherently accessible to screen
  readers. Honest limitation, matching VOC-017/VOC-019 precedent
  (`VOC-017-D08`): with no `apps/web` component/a11y test runner yet, this is
  verified by construction and structured inspection, not automated
  axe/Playwright assertions.
