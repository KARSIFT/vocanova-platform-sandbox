# VOC-019 — Build the Real Home Screen (Today's Mission, Streak, Due Reviews, Journey Entry)

**Draft package — not adopted, not approved, not implementation authority.**
Prepared by the planner role from a free-text request grounded in the approved
product and design documents DOC-01 (`docs/product/01-mvp-prd.md`) and DOC-03
(`docs/design/03-ui-ux-design.md`). A human must review, resolve the open
decisions below, and adopt this package (recording a founder-approved,
implementation-ready state) before any implementation.

## Identity and lifecycle

- Package ID: `VOC-019`
- Canonical path:
  `specs/changes/VOC-019-build-the-real-home-screen-apps-web-src-app-app-ho/`
- Lifecycle state: `draft` (unadopted; see `change.yaml` — every
  adoption/authorization gate is left at its unadopted default)
- Proposed risk: **R1** (draft proposal only — the authoritative floor is
  whatever `scripts/governance/classify-change-risk.sh` computes at
  implementation time. Every target file lives under `apps/web/src/app/(app)/home/**`,
  which matches no R2/R3/R4 path pattern, so R1 is the expected path-detected
  floor, the same floor VOC-017's shell used.)
- Owner (decision): founder
- Requirement source: the free-text planner request, grounded in DOC-01 §2 (MVP
  core screens) and DOC-03 §4 item 1 (daily session flow — Home) and §8
  (Progress/streak framing referenced by the request; Home itself is specified in
  §4 item 1, not §8, which covers the separate Progress tab — see the open note
  in `specification.md`). Both DOC-01 and DOC-03 are `status: approved`,
  `owner: founder`. Per `AGENTS.md`, an approved document describes target UX but
  is not itself implementation authority for a specific change; a
  founder-approved, implementation-ready state must be recorded at adoption. This
  document is a draft, not an approved specification.
- Target branch: `develop`

## Objective and requirement source

Replace the placeholder `apps/web/src/app/(app)/home/page.tsx` (currently a bare
heading and one line of filler text) with the real Home screen: **Today's
Mission** (review target and progress toward it), the learner's **current
streak**, the **due-review count**, and a **route into Journey/Discover**. Scope
is deliberately narrow per the free-text request: **static UI only, backed by
local mocked/placeholder data** — `apps/api` has no real endpoints yet, so no
network call, API client import, or data-fetching code is introduced.

## Scope, non-goals, risk, and protected areas

In scope (all under `apps/web/src/app/(app)/home/`):
- Real Home screen markup replacing the placeholder, rendering the four
  elements above from local mock data.
- Reuse of the existing wired design tokens (`packages/design-tokens` via
  VOC-018's Tailwind v4 `@theme` layer, `apps/web/src/app/tokens.generated.css`)
  through ordinary Tailwind utility classes — no new token wiring.

Non-goals: any backend/API call or data-fetching code; any change to
`apps/api`, authentication, onboarding, or any database/migration work; a
`/review` route or a "start review" action (no such route exists yet — see
`VOC-019-D02`); the reusable sentence-practice component (a separate, explicit
DOC-03 §2 component with its own entry points); the `feedback` token scale
(`success`/`warning`/`error`) which VOC-018 deliberately left unwired
(`VOC-018-DEP-05`) and which remains unwired on `develop` as of this draft; new
runtime dependencies or any `package.json`/lockfile change; any change to
`bottom-nav.tsx`, `(app)/layout.tsx`, `discover/page.tsx`, or
`progress/page.tsx`.

No protected areas touched (no `auth`, `payments`, `migrations`, governance, or
CI-workflow paths). No production impact (merge-to-`develop` only; deployment
remains prohibited).

## Open decisions for the human adopter (see `specification.md`)

1. **Sentence-practice entry / review-start action excluded (`VOC-019-D01`).**
   DOC-01 §2 and DOC-03's information architecture (§2) both list a
   sentence-practice/quick-entry affordance as part of the Home tab, but the
   free-text request explicitly cites DOC-03 **§4 item 1** — the daily-session-flow
   description of Home — which lists only mission, streak, and a route into
   Journey/Discover. This draft follows the request's narrower citation and
   excludes both the sentence-practice entry point and any "start review" action
   (there is also no `/review` route yet to link to). A human should confirm this
   narrower scope or direct a follow-up package to add the remaining Home
   affordances once `/review` and the sentence-practice component exist.
2. **Due-review count is display-only, not a link (`VOC-019-D02`).** Since
   `/review` does not exist on `develop` (VOC-017 built only `/home`,
   `/discover`, `/progress`), the due-review count is rendered as a text stat
   with no destination, avoiding a link to a non-existent route.
3. **Empty/loading/error states deferred (`VOC-019-D06`).** DOC-03 §9 requires
   empty/loading/error states for every screen with dynamic content. This task
   has no dynamic data source (fixed local mock state, no network call), so
   there is nothing to load or fail; the required follow-up is the real
   API-wiring package, which must add those states when Home starts fetching
   live data.

## Verification, approvals, release, and closure

Deterministic evidence (at implementation time): `pnpm run lint:web`,
`pnpm run typecheck:web`, `pnpm run build:web`, and `pnpm run format:check`.
`apps/web` has no component/a11y test runner yet (same gap VOC-017 recorded,
`VOC-017-D08`/`DEP-05`), so verification also relies on the structured
inspection checklist in `test-plan.md`. Independent verification is exact-SHA
per `CLAUDE.md`, covering scope discipline (no API/data-fetching code, no
protected-area touch), the accessibility checklist, and confirmation that no
other route/component file changed. Because this is a draft, no approval,
merge, or release authority is claimed here; those gates stay closed in
`change.yaml` and are a human's decision at adoption.
