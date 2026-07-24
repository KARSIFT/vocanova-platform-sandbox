# VOC-021 — Implementation Plan

## Preconditions and protected areas

Do not begin until this package is adopted and implementation is authorized:
a human sets `status`, `approval_status`, and `implementation.authorized` in
`change.yaml` (all at unadopted defaults in this draft), records a
founder-approved implementation-ready state (`VOC-021-DEP-01`), and resolves
the two open scope decisions — situation set/order (`VOC-021-D01`,
`VOC-021-DEP-04`) and mock-copy adequacy (`VOC-021-D08`, `VOC-021-DEP-05`).
Depends on the VOC-017 shell and VOC-018's wired tokens already on `develop`
(`VOC-021-DEP-03`, present at drafting). No protected areas are touched —
every target is under `apps/web/src/app/(app)/discover/**`.

## File reconciliation and implementation sequence

Existing state (do not disturb):
- `apps/web/src/app/(app)/discover/page.tsx` — the VOC-017 placeholder
  (`<h1>Journey</h1>` + one filler line). This is the file being
  **replaced** by this task; it is the only existing target.
- `apps/web/src/app/(app)/_components/bottom-nav.tsx`,
  `apps/web/src/app/(app)/layout.tsx`,
  `apps/web/src/app/(app)/home/page.tsx`,
  `apps/web/src/app/(app)/progress/page.tsx` — untouched; the Discover
  screen does not link to or modify any of these files.
- `apps/web/src/app/tokens.generated.css` — the VOC-018 token layer; reused
  as-is via Tailwind utility classes, not edited.

New targets:
- `apps/web/src/app/(app)/discover/page.tsx` (rewritten in place, not a new
  path).
- Optionally, `apps/web/src/app/(app)/discover/_lib/mock-data.ts` (or the
  mock data inline in `page.tsx`) and
  `apps/web/src/app/(app)/discover/_components/*.tsx` if the implementer
  chooses to extract presentational pieces — private, non-routable folders,
  matching the `_components` convention `bottom-nav.tsx` and VOC-019/
  VOC-020's screens already established at the `(app)` level.

Not created by this task: `apps/web/src/app/(app)/discover/[situation]/`
(or any similarly named dynamic-segment route) — explicitly out of scope
(`VOC-021-D03`).

Ordered steps (single task, `VOC-021-T00`):

1. Add the local mock data structure (`VOC-021-D00`, `VOC-021-D01`),
   commented as placeholder pending real API wiring, with exactly seven
   entries (`slug`, `title`, `shortDescription`) in the order specified in
   `tasks.md` (`VOC-021-D02`).
2. Rewrite `discover/page.tsx` as a server component whose root element is
   a `<div>` (never `<main>` — `VOC-021-D05`), rendering, in order: the
   `<h1>Journey</h1>` heading (`VOC-021-D07`), a short intro line, and a
   `<ul>` of seven `<li>` situation entries, each with an `<h2>` title and a
   plain `<p>` description, with no interactive element on any entry
   (`VOC-021-D03`, `VOC-021-D04`).
3. Use only the wired token utilities for color/spacing/radius/shadow/
   easing (`VOC-021-D06`) and no charting/animation/interaction library.
4. Run the deterministic checks and confirm no other file changed, and no
   `[situation]` (or similarly named) route was added.

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
`build:web` proves `/discover` still compiles and resolves; `typecheck:web`
covers `next typegen && tsc --noEmit`; `format:check` includes `apps/web` in
its prettier target set.)

Independent verification (exact-SHA, per `CLAUDE.md`): the reviewer confirms
exactly seven situation entries render from local mock data only, each with
a label and short description and nothing else (`VOC-021-AC-00`,
`VOC-021-AC-01`); confirms no entry carries an `href`, `onClick`, `<Link>`,
`<button>`, or any hover/focus affordance implying tappability, and that no
`/discover/[situation]` route exists (`VOC-021-AC-02`); confirms no API
client import, `fetch`, or `TanStack Query` usage was introduced, and that
`apps/api`, auth, and any migration file are untouched (`VOC-021-AC-03`);
confirms every color/spacing/radius/shadow/easing reference resolves to one
of VOC-018's 64 wired custom properties with no raw hex and no
`feedback`-scale utility (`VOC-021-AC-04`); confirms the `<ul>`/`<li>`
structure, exactly one `<h1>`, seven `<h2>`s, and no second `<main>`
(`VOC-021-AC-05`); confirms
`bottom-nav.tsx`/`(app)/layout.tsx`/`home/page.tsx`/`progress/page.tsx`/
`package.json`/lockfile are byte-for-byte unchanged; and confirms WCAG 2.2 AA
contrast for every token color pairing actually used (`VOC-021-R01`). The
verifier binds its verdict to the exact reviewed commit SHA and confirms the
implementer did not self-approve or self-merge.

## Deployment and rollback

`release.deployment: prohibited`. Merging the implementation to `develop` is
the entire scope; nothing here authorizes a production deployment. Rollback
is a plain `git revert` of the merge commit — the change replaces one static
placeholder file with another static file, and nothing depends on this
screen at runtime beyond ordinary navigation to `/discover`. Last-known-good
reference is `develop` at this package's (adoption-time re-pinned)
`base_sha`.
