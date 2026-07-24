# VOC-022 — Implementation Plan

## Preconditions and protected areas

Do not begin until this package is adopted and implementation is
authorized: a human sets `status`, `approval_status`, and
`implementation.authorized` in `change.yaml` (all at unadopted defaults in
this draft), records a founder-approved implementation-ready state
(`VOC-022-DEP-01`), and resolves the two open scope decisions — per-situation
mock word content/count (`VOC-022-D02`, `VOC-022-DEP-04`) and the exact
`discover/page.tsx` navigation-diff shape (`VOC-022-D05`, `VOC-022-DEP-05`).
Depends on VOC-021's seven-situation `/discover` list and VOC-018's wired
tokens already on `develop` (`VOC-022-DEP-03`, present at drafting). No
protected areas are touched — every target is under
`apps/web/src/app/(app)/discover/**`.

## File reconciliation and implementation sequence

Existing state (do not disturb beyond the single disclosed edit):
- `apps/web/src/app/(app)/discover/page.tsx` — VOC-021's real Discover
  situation list (seven non-interactive cards from
  `MOCK_DISCOVER_SITUATIONS`). This is the **only** existing target, and
  only for the single navigation edit described in `VOC-022-D05`/`tasks.md`
  step 5 — no other line may change.
- `apps/web/src/app/(app)/_components/bottom-nav.tsx`,
  `apps/web/src/app/(app)/layout.tsx`,
  `apps/web/src/app/(app)/home/page.tsx`,
  `apps/web/src/app/(app)/progress/page.tsx` — untouched.
- `apps/web/src/app/tokens.generated.css` — the VOC-018 token layer; reused
  as-is via Tailwind utility classes, not edited.

New targets:
- `apps/web/src/app/(app)/discover/[situation]/page.tsx` (new file).
- Optionally, `apps/web/src/app/(app)/discover/[situation]/_lib/mock-data.ts`
  (or the mock data inline in `page.tsx`) if the implementer chooses to
  extract it — a private, non-routable folder, matching the `_components`/
  `_lib` colocation convention `bottom-nav.tsx` and VOC-019/VOC-020/VOC-021's
  screens already established at the `(app)` level. This mock data structure
  must remain **local to the new route** — it must not be imported by or
  exported from `discover/page.tsx`, since that would exceed the single
  permitted navigation edit to that file.

Not created by this task: any word-detail route or modal, any save/unsave
API route or client mutation, any sentence-practice component wiring — all
explicitly out of scope (`VOC-022-D06`).

Ordered steps (single task, `VOC-022-T00`):

1. Add the new `[situation]/page.tsx` dynamic route as an `async` server
   component that `await`s `params` (`VOC-022-D01`).
2. Add the local mock word-list data structure (`VOC-022-D00`,
   `VOC-022-D02`), commented as placeholder pending real API wiring, keyed
   by the seven slugs VOC-021 already defined.
3. Implement the `notFound()` branch for an unrecognized slug
   (`VOC-022-D04`).
4. Render the drill-down markup: back link, `<h1>`, intro line, `<ul>` of
   words with the icon+text "saved" marker only where `isSaved` is `true`
   (`VOC-022-D03`, `VOC-022-D08`).
5. Make the single disclosed edit to `discover/page.tsx`: add the
   `next/link` import and wrap each card's content in a `<Link>` to
   `/discover/{slug}`, moving existing styling onto the `<Link>` and adding
   token-based hover/focus-visible styling (`VOC-022-D05`).
6. Use only the wired token utilities for color/spacing/radius/shadow/
   easing (`VOC-022-D07`) across both files, and no charting/animation/
   interaction library.
7. Run the deterministic checks and confirm no other file changed, and that
   `discover/page.tsx`'s diff is confined to the navigation edit.

## Validation and independent verification

Deterministic commands (run from repo root):

```bash
pnpm run lint:web
pnpm run typecheck:web
pnpm run build:web
pnpm run format:check
```

(`lint:web` runs `scripts/foundation/check-tailwind-token-usage.mjs`, which
deterministically rejects a bare `duration-<name>`/`ease-<name>` utility and
a second `<main>` landmark under `apps/web/src/app/(app)/**/*.tsx`.
`build:web` proves both `/discover` and `/discover/{slug}` still compile and
resolve; `typecheck:web` covers `next typegen && tsc --noEmit`, which
validates the new route's `params: Promise` shape; `format:check` includes
`apps/web` in its prettier target set.)

Independent verification (exact-SHA, per `CLAUDE.md`): the reviewer confirms
the new route renders at least three words per known situation slug from
local mock data only, with each word showing a word/phrase and short
meaning and no other field (`VOC-022-AC-00`, `VOC-022-AC-01`); confirms
every situation's word list includes at least one mock-saved and one
mock-not-saved word, and that the saved marker combines an `aria-hidden`
icon with visible text — never color alone (`VOC-022-AC-01`); confirms no
word entry carries an `href`, `onClick`, `<Link>`, `<button>`, or any
control that could change `isSaved` (`VOC-022-AC-02`); confirms an
unrecognized slug resolves to Next's built-in not-found handling
(`VOC-022-AC-03`); diffs `discover/page.tsx` specifically and confirms only
the navigation-related lines changed (`VOC-022-AC-04`); confirms no API
client import, `fetch`, or `TanStack Query` usage was introduced anywhere,
and that `apps/api`, auth, and any migration file are untouched
(`VOC-022-AC-05`); confirms every color/spacing/radius/shadow/easing
reference resolves to one of VOC-018's 64 wired custom properties with no
raw hex and no `feedback`-scale utility, including on the new `<Link>`
hover/focus styling (`VOC-022-AC-06`); confirms exactly one `<h1>`, the
`<ul>`/`<li>` word-list structure, no second `<main>`, and the presence of
the "Back to Journey" link (`VOC-022-AC-07`); confirms
`bottom-nav.tsx`/`(app)/layout.tsx`/`home/page.tsx`/`progress/page.tsx`/
`package.json`/lockfile are byte-for-byte unchanged; and confirms WCAG 2.2 AA
contrast for every token color pairing actually used (`VOC-022-R01`). The
verifier binds its verdict to the exact reviewed commit SHA and confirms the
implementer did not self-approve or self-merge.

## Deployment and rollback

`release.deployment: prohibited`. Merging the implementation to `develop` is
the entire scope; nothing here authorizes a production deployment. Rollback
is a plain `git revert` of the merge commit — the change adds one static
route and one minimal, disclosed edit to another static file, and nothing
depends on this screen at runtime beyond ordinary navigation between
`/discover` and `/discover/{slug}`. Last-known-good reference is `develop`
at this package's (adoption-time re-pinned) `base_sha`.
