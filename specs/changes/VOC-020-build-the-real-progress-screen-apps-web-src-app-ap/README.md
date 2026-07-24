# VOC-020 — Build the Real Progress Screen (Confidence Points, Streak, Completion History)

**Draft package — not adopted, not approved, not implementation authority.**
Prepared by the planner role from a free-text request grounded in the approved
product and design documents DOC-01 (`docs/product/01-mvp-prd.md`) and DOC-03
(`docs/design/03-ui-ux-design.md`). A human must review, resolve the open
decisions below, and adopt this package (recording a founder-approved,
implementation-ready state) before any implementation.

## Identity and lifecycle

- Package ID: `VOC-020`
- Canonical path:
  `specs/changes/VOC-020-build-the-real-progress-screen-apps-web-src-app-ap/`
- Lifecycle state: `draft` (unadopted; see `change.yaml` — every
  adoption/authorization gate is left at its unadopted default)
- Proposed risk: **R1** (draft proposal only — the authoritative floor is
  whatever `scripts/governance/classify-change-risk.sh` computes at
  implementation time. Every target file lives under
  `apps/web/src/app/(app)/progress/**`, which matches no R2/R3/R4 path
  pattern in `classify-change-risk.sh`, so R1 is the expected path-detected
  floor — the same floor VOC-019's Home screen and VOC-017's shell used.)
- Owner (decision): founder
- Requirement source: the free-text planner request, grounded in DOC-01 §2
  ("**Progress** — Confidence Points, streak, completion history,
  motivation-focused summary") and DOC-03 §4 item 5 and §8 (Progress screen
  detail: "Confidence Point total, current/longest streak, and a simple
  day-by-day or week-by-week completion view — motivational, not
  analytical"). Both DOC-01 and DOC-03 are `status: approved`,
  `owner: founder`. Per `AGENTS.md`, an approved document describes target UX
  but is not itself implementation authority for a specific change; a
  founder-approved, implementation-ready state must be recorded at adoption.
  This document is a draft, not an approved specification.
- Target branch: `develop`

## Objective and requirement source

Replace the placeholder `apps/web/src/app/(app)/progress/page.tsx` (currently
a bare heading and one line of filler text) with the real Progress screen:
**Confidence Points total**, **current streak**, **longest streak**, and a
**simple completion history view** (day-by-day within the current week, per
`VOC-020-D01`). Scope is deliberately narrow per the free-text request:
**static UI only, backed by local mocked/placeholder data** — `apps/api` has
no real endpoints yet, so no network call, API client import, or
data-fetching code is introduced.

## Scope, non-goals, risk, and protected areas

In scope (all under `apps/web/src/app/(app)/progress/`):
- Real Progress screen markup replacing the placeholder, rendering the four
  elements above from local mock data.
- Reuse of the existing wired design tokens (`packages/design-tokens` via
  VOC-018's Tailwind v4 `@theme` layer,
  `apps/web/src/app/tokens.generated.css`) through ordinary Tailwind utility
  classes and `var(--…)` arbitrary values — no new token wiring.

Non-goals: any backend/API call or data-fetching code; any change to
`apps/api`, authentication, home screen, discover/Journey screen, or any
database/migration work; a charting/graph library or component (DOC-03 §8
calls the Progress screen "motivational, not analytical" — see
`VOC-020-D08`); a celebration-moment animation (`VOC-020-D03`); a true
zero-history/first-mission empty state (`VOC-020-D06`, open); the `feedback`
token scale (`success`/`warning`/`error`), which VOC-018 deliberately left
unwired (`VOC-018-DEP-05`) and which remains unwired on `develop` as of this
draft; any bare `duration-<name>`/`ease-<name>` Tailwind utility class
(Tailwind v4 does not generate utilities from those custom-property
namespaces — only `var(--…)` arbitrary-value form works, enforced by
`scripts/foundation/check-tailwind-token-usage.mjs`); a second `<main>`
landmark (`apps/web/src/app/(app)/layout.tsx` already renders the one
`<main>` for every route in this group — also enforced by the same script);
new runtime dependencies or any `package.json`/lockfile change; any change to
`bottom-nav.tsx`, `(app)/layout.tsx`, `home/page.tsx`, or `discover/page.tsx`.

No protected areas touched (no `auth`, `payments`, `migrations`, governance,
or CI-workflow paths). No production impact (merge-to-`develop` only;
deployment remains prohibited).

## Open decisions for the human adopter (see `specification.md`)

1. **Completion history granularity (`VOC-020-D01`, OPEN).** DOC-03 §4 item 5
   offers "day-by-day **or** week-by-week" without settling which. This draft
   proposes a 7-entry, day-by-day strip covering the current week (a common,
   simple habit-tracker pattern) rather than a multi-week, week-by-week
   summary. A human should confirm this choice or direct a different
   granularity.
2. **True empty/first-mission state deferred (`VOC-020-D06`, OPEN).** DOC-03
   §9 names the Progress screen specifically as its example of a required
   empty state ("empty Progress screen invites the learner to complete their
   first mission"). This task renders one fixed, populated mock state only
   (an established learner with points, streak, and history already
   present) and defers the true zero-state to the real API-wiring follow-up
   package, the same honest-deferral pattern VOC-019 used for Home's dynamic
   states (`VOC-019-D06`). A human should confirm this deferral is acceptable
   for a static-mock task or direct that a second, illustrative empty-state
   render be added now.

## Verification, approvals, release, and closure

Deterministic evidence (at implementation time): `pnpm run lint:web`
(includes `scripts/foundation/check-tailwind-token-usage.mjs`'s bare-duration
and nested-`<main>` checks), `pnpm run typecheck:web`, `pnpm run build:web`,
and `pnpm run format:check`. `apps/web` has no component/a11y test runner yet
(same gap VOC-017 and VOC-019 recorded), so verification also relies on the
structured inspection checklist in `test-plan.md`. Independent verification
is exact-SHA per `CLAUDE.md`, covering scope discipline (no API/data-fetching
code, no protected-area touch), the accessibility checklist, and confirmation
that no other route/component file changed. Because this is a draft, no
approval, merge, or release authority is claimed here; those gates stay
closed in `change.yaml` and are a human's decision at adoption.
