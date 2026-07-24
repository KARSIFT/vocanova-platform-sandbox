# VOC-022 — Build the Journey/Discover Situation Drill-Down Screen

**Draft package — not adopted, not approved, not implementation authority.**
Prepared by the planner role from a free-text request grounded in the
approved product and design/database documents DOC-01
(`docs/product/01-mvp-prd.md`), DOC-03 (`docs/design/03-ui-ux-design.md`),
and DOC-05 (`docs/engineering/05-database-design.md`). A human must review,
resolve the open decisions below, and adopt this package (recording a
founder-approved, implementation-ready state) before any implementation.

## Identity and lifecycle

- Package ID: `VOC-022`
- Canonical path:
  `specs/changes/VOC-022-build-the-journey-discover-situation-drill-down-sc/`
- Lifecycle state: `draft` (unadopted; see `change.yaml` — every
  adoption/authorization gate is left at its unadopted default:
  `approval_status: not-approved`, `implementation_authorized: false`,
  `automatic_merge_allowed: false`, `repository_adoption_status:
  not-adopted`)
- Proposed risk: **R1** (draft proposal only — the authoritative floor is
  whatever `scripts/governance/classify-change-risk.sh` computes at
  implementation time. Every target file lives under
  `apps/web/src/app/(app)/discover/**`, which matches no R2/R3/R4 path
  pattern in `classify-change-risk.sh`, so R1 is the expected path-detected
  floor — the same floor VOC-019, VOC-020, and VOC-021 used.)
- Owner (decision): founder
- Requirement source: the free-text planner request, grounded in DOC-03
  (`docs/design/03-ui-ux-design.md`) §5 ("Within a situation, words are shown
  one at a time or as a short scannable list … A word already in the
  learner's saved list is visually marked …"), DOC-01
  (`docs/product/01-mvp-prd.md`) §2 (Journey tab bundles situation-based
  discovery, word detail, save/unsave), and DOC-05
  (`docs/engineering/05-database-design.md`) §7–§8 (`word_meanings`,
  `journey_words` — used here only as naming inspiration for mock field
  shape, not a real dependency on those tables). All three are `status:
  approved`, `owner: founder`. Per `AGENTS.md`, an approved document
  describes target UX but is not itself implementation authority for a
  specific change; a founder-approved, implementation-ready state must be
  recorded at adoption. This document is a draft, not an approved
  specification.
- Target branch: `develop`

## Objective and requirement source

Add the new dynamic route
`apps/web/src/app/(app)/discover/[situation]/page.tsx`: a static, mocked
within-situation word list (DOC-03 §5's "a short scannable list"), where a
word already saved is visually marked using **mock saved/not-saved state
only** — no real save action, no backend/API wiring. Additionally, make each
of the seven situation cards on the existing
`apps/web/src/app/(app)/discover/page.tsx` (built by VOC-021) link to its own
`/discover/[slug]` route, since VOC-021 deliberately left those cards
non-interactive pending this exact follow-up
(`VOC-021-DEP-06`/`VOC-021-D03`).

## Scope, non-goals, risk, and protected areas

In scope (all under `apps/web/src/app/(app)/discover/`):
- A new `apps/web/src/app/(app)/discover/[situation]/page.tsx` dynamic route
  rendering a scannable list of mocked words for the matching situation,
  each visually marked as saved or not-saved from local mock data only.
- Exactly one change to the existing `discover/page.tsx`: wrapping each
  situation card's content in a `next/link` `<Link>` to `/discover/{slug}`.
  Per the free-text request, this is deliberately **the only change** to
  that file — the mock situation array, heading, and intro text are
  otherwise untouched.
- Reuse of the existing wired design tokens (`packages/design-tokens` via
  VOC-018's Tailwind v4 `@theme` layer,
  `apps/web/src/app/tokens.generated.css`) through ordinary Tailwind utility
  classes and `var(--…)` arbitrary values or auto-generated color/spacing
  utilities — exactly as VOC-018 through VOC-021 did. No bare
  `duration-<name>`/`ease-<name>` utility class.

Non-goals, explicitly per the free-text request: word detail (a separate
page/modal per word), save/unsave functionality (no backend to persist it —
the "saved" marker is mock/visual only), and sentence practice. Also
excluded: any backend/API call or data-fetching code; any change to
`apps/api`, authentication, home/progress pages, `bottom-nav.tsx`, or any
database/migration work.

No protected areas touched (no `auth`, `payments`, `migrations`, governance,
or CI-workflow paths). No production impact (merge-to-`develop` only;
deployment remains prohibited).

## Open decisions for the human adopter (see `specification.md`)

1. **Per-situation mock word list content and count (`VOC-022-D02`,
   OPEN).** Each situation's word list (word + short meaning + mock
   saved/not-saved flag) is illustrative placeholder text drafted for this
   package, not founder-reviewed product copy, and the exact word count per
   situation is this draft's proposal (six), not a fixed product
   requirement. A human should review the exact content (or explicitly
   accept it as adequate placeholder content) at adoption.
2. **Navigation diff shape in `discover/page.tsx` (`VOC-022-D05`, OPEN).**
   The free-text request requires this file's *only* change to be adding
   navigation. This draft proposes a specific minimal diff (wrap each
   card's existing content in a `<Link>`, moving the card's existing
   className onto the `<Link>`) in `tasks.md`. A human should confirm this
   is an acceptable interpretation of "the only change," or direct a
   narrower/different approach.

## Verification, approvals, release, and closure

Deterministic evidence (at implementation time): `pnpm run lint:web`
(includes `scripts/foundation/check-tailwind-token-usage.mjs`'s bare-duration
and nested-`<main>` checks), `pnpm run typecheck:web`, `pnpm run build:web`,
and `pnpm run format:check`. `apps/web` has no component/a11y test runner yet
(same gap VOC-017/VOC-019/VOC-020/VOC-021 recorded), so verification also
relies on the structured inspection checklist in `test-plan.md`. Independent
verification is exact-SHA per `CLAUDE.md`, covering scope discipline (no
word-detail/save-unsave/sentence-practice, no API/data-fetching code, no
protected-area touch, `discover/page.tsx`'s diff confined to the navigation
addition), the accessibility checklist, and confirmation that no other
route/component file changed. Because this is a draft, no approval, merge,
or release authority is claimed here; those gates stay closed in
`change.yaml` and are a human's decision at adoption.
