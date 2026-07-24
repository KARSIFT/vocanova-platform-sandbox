# VOC-021 — Build the Real Journey/Discover Screen (Situation List)

**Draft package — not adopted, not approved, not implementation authority.**
Prepared by the planner role from a free-text request grounded in the approved
product and design/database documents DOC-01 (`docs/product/01-mvp-prd.md`),
DOC-03 (`docs/design/03-ui-ux-design.md`), and DOC-05
(`docs/engineering/05-database-design.md`). A human must review, resolve the
open decisions below, and adopt this package (recording a founder-approved,
implementation-ready state) before any implementation.

## Identity and lifecycle

- Package ID: `VOC-021`
- Canonical path:
  `specs/changes/VOC-021-build-the-real-journey-discover-screen-apps-web-sr/`
- Lifecycle state: `draft` (unadopted; see `change.yaml` — every
  adoption/authorization gate is left at its unadopted default)
- Proposed risk: **R1** (draft proposal only — the authoritative floor is
  whatever `scripts/governance/classify-change-risk.sh` computes at
  implementation time. Every target file lives under
  `apps/web/src/app/(app)/discover/**`, which matches no R2/R3/R4 path
  pattern in `classify-change-risk.sh`, so R1 is the expected path-detected
  floor — the same floor VOC-019's Home screen and VOC-020's Progress screen
  used.)
- Owner (decision): founder
- Requirement source: the free-text planner request, grounded in DOC-01
  (`docs/product/01-mvp-prd.md`) §2 ("**Journey** — situation-based discovery
  (Airport, Restaurant, Hotel Check-in, Job Interview, Daily Conversation,
  Work Meeting, University Class, etc.), word detail, save/unsave"), DOC-03
  (`docs/design/03-ui-ux-design.md`) §5 ("Discovery is organized by real-life
  situation … not by grammar topic or difficulty tier alone"), and DOC-05
  (`docs/engineering/05-database-design.md`) §8 (the `journey_situations`
  model — `slug`, `title`, `short_description`, `level_band`, `category`,
  `status`, `display_order` — used here only as naming inspiration for mock
  field shape, not a real dependency on that table). All three are
  `status: approved`, `owner: founder`. Per `AGENTS.md`, an approved document
  describes target UX but is not itself implementation authority for a
  specific change; a founder-approved, implementation-ready state must be
  recorded at adoption. This document is a draft, not an approved
  specification.
- Target branch: `develop`

## Objective and requirement source

Replace the placeholder `apps/web/src/app/(app)/discover/page.tsx` (currently
a bare heading and one line of filler text) with the real top-level
Journey/Discover screen: a list/grid of situation-based discovery categories
(Airport, Restaurant, Hotel Check-in, Job Interview, Daily Conversation, Work
Meeting, University Class), each a simple, non-interactive card/entry with a
label and a short description, from **local mocked/placeholder data only** —
`apps/api` has no real endpoints yet, so no network call, API client import,
or data-fetching code is introduced.

## Scope, non-goals, risk, and protected areas

In scope (all under `apps/web/src/app/(app)/discover/`):
- Real Journey/Discover screen markup replacing the placeholder, rendering
  the seven named situations above as a static, non-interactive list from
  local mock data.
- Reuse of the existing wired design tokens (`packages/design-tokens` via
  VOC-018's Tailwind v4 `@theme` layer,
  `apps/web/src/app/tokens.generated.css`) through ordinary Tailwind utility
  classes and `var(--…)` arbitrary values — no new token wiring.

Non-goals, explicitly per the free-text request: any word-list-within-
situation, word detail, save/unsave, or navigation into a situation (no
`/discover/[situation]` route) — drilling into a situation is a follow-up
package's scope, not this one. Also excluded: any backend/API call or
data-fetching code; any change to `apps/api`, authentication, home/progress
pages, `bottom-nav.tsx`, or any database/migration work; a category/level-band
badge, word-count, saved-word indicator, search, or filter control on the
situation list (DOC-01 §2 bundles these into the full Journey tab; this task
is the top-level situation list only); any bare
`duration-<name>`/`ease-<name>` Tailwind utility class (Tailwind v4 does not
generate utilities from those custom-property namespaces — only `var(--…)`
arbitrary-value form works, enforced by
`scripts/foundation/check-tailwind-token-usage.mjs`); a second `<main>`
landmark (`apps/web/src/app/(app)/layout.tsx` already renders the one
`<main>` for every route in this group — also enforced by the same script);
new runtime dependencies or any `package.json`/lockfile change; any change to
`bottom-nav.tsx`, `(app)/layout.tsx`, `home/page.tsx`, or `progress/page.tsx`.

No protected areas touched (no `auth`, `payments`, `migrations`, governance,
or CI-workflow paths). No production impact (merge-to-`develop` only;
deployment remains prohibited).

## Open decisions for the human adopter (see `specification.md`)

1. **Situation set and order (`VOC-021-D01`, OPEN).** DOC-01 §2 and DOC-03 §5
   both list the seven situations named in the free-text request but qualify
   them with "etc." — an illustrative, not exhaustive, set. This draft treats
   exactly those seven, in the order given, as the definitive mock content
   for this task. A human should confirm this set/order is adequate for a
   static mock, or direct a different set (the real, larger set will
   ultimately come from the backend `journey_situations` table, not this
   task).
2. **Mock copy is placeholder text, not reviewed product content
   (`VOC-021-D08`, OPEN).** Each situation's short description is illustrative
   text drafted for this package, not founder-reviewed product copy. A human
   should review the exact wording (or explicitly accept it as adequate
   placeholder text) at adoption.

## Verification, approvals, release, and closure

Deterministic evidence (at implementation time): `pnpm run lint:web`
(includes `scripts/foundation/check-tailwind-token-usage.mjs`'s bare-duration
and nested-`<main>` checks), `pnpm run typecheck:web`, `pnpm run build:web`,
and `pnpm run format:check`. `apps/web` has no component/a11y test runner yet
(same gap VOC-017/VOC-019/VOC-020 recorded), so verification also relies on
the structured inspection checklist in `test-plan.md`. Independent
verification is exact-SHA per `CLAUDE.md`, covering scope discipline (no
API/data-fetching code, no drill-down navigation, no protected-area touch),
the accessibility checklist, and confirmation that no other route/component
file changed. Because this is a draft, no approval, merge, or release
authority is claimed here; those gates stay closed in `change.yaml` and are a
human's decision at adoption.
