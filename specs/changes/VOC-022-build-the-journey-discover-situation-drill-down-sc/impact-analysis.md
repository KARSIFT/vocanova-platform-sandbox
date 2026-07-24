# VOC-022 — Impact Analysis

## Security and privacy

None. The new drill-down screen renders static mock markup only — the mock
`isSaved` flag is a fixed placeholder boolean, not derived from or
associated with any real learner's actual saved-word state; no input
fields, no network/API calls, no data fetching, no storage, cookies/tokens,
and no authentication or authorization surface. The one interactive
behavior this task adds is ordinary client-side page navigation
(`next/link` `<Link>`s from `discover/page.tsx` to `/discover/{slug}`, and a
"Back to Journey" link on the new route) — no data submission of any kind.
DOC-03 §5's "saving must succeed against the backend before the UI reflects
it as saved" rule is not engaged at runtime because this task implements no
save action at all. `apps/api` is untouched. No secrets are introduced.

## Data and migrations

None. No database, schema, migration, or persisted state of any kind.
Purely a front-end addition (`[situation]/page.tsx`) plus a minimal,
disclosed edit to `discover/page.tsx` under
`apps/web/src/app/(app)/discover/`. Rollback is a plain `git revert` of the
merge commit with no data implications.

## Analytics and accessibility

- **Analytics:** none — a static mock screen with only plain page-navigation
  links emits no analytics events.
- **Accessibility:** in scope. The word list uses a semantic `<ul>`/`<li>`
  structure so assistive technology announces the item count; one `<h1>`
  per page (DOC-03 §10); the "saved" marker must combine an `aria-hidden`
  icon with a visible text label so the state is never conveyed by color
  alone (`VOC-022-D03`) — a genuine review item, since DOC-03 §5 requires
  the saved state to be "visually marked" and DOC-03 §10 forbids relying on
  color alone to do it. The new `<Link>`s must carry a visible
  `focus-visible` outline built only from wired tokens. Honest limitation,
  same as VOC-017/VOC-019/VOC-020/VOC-021 (`VOC-017-D08`/`DEP-05`):
  `apps/web` has no component/a11y test runner yet, so this is verified by
  construction and structured inspection, not automated axe/Playwright
  assertions — that automated coverage remains the named follow-up VOC-017
  already flagged, not a gap this package reopens or resolves.

## Risks, dependencies, and evidence

- `VOC-022-R00` (Low, scoping): "static UI only, mock saved/not-saved state"
  is easy to overrun into building a real save/unsave toggle, word detail,
  or sentence practice. Mitigated by the hard non-goals in
  `specification.md` and the explicit `VOC-022-D06` decision; the verifier
  should reject any such creep.
- `VOC-022-R01` (Medium, semantic): like VOC-019 through VOC-021, this
  renders token colors across a learner-facing content surface. WCAG 2.2 AA
  contrast (1.4.3 text, 1.4.11 non-text) for every `neutral`/`primary`
  text-on-background pairing used must be verified by the implementer and
  re-checked by the independent reviewer.
- `VOC-022-R02` (Medium, UX-affordance / accessibility): the "saved" marker
  must not rely on color alone (`VOC-022-D03`) — the inverse failure mode
  from VOC-021's non-interactive cards (which needed to *avoid* looking
  tappable); here a marker that looks decorative-only, with no text label,
  would fail DOC-03 §10. The verifier must specifically check for an
  `aria-hidden` icon paired with visible text, not color/background alone.
- `VOC-022-R03` (Medium, scope-boundary discipline): the free-text request's
  constraint that `discover/page.tsx`'s only change is the navigation
  addition is easy to violate unintentionally (e.g. reformatting unrelated
  lines, reordering the mock array, or renaming a field while "just adding
  a Link"). The verifier must diff this file specifically against its
  pre-VOC-022 state and confirm only the navigation-related lines changed
  (`VOC-022-D05`, `VOC-022-AC-04`).
- `VOC-022-R04` (Low, correctness): treating the new route's `params` prop
  as a synchronous object instead of a `Promise` (an easy carry-over mistake
  from pre-Next-15 conventions) would fail `next typegen`/`typecheck:web`
  deterministically — a real but cheaply-caught risk (`VOC-022-D01`).
- `VOC-022-R05` (Low): reusing `bottom-nav.tsx`'s existing non-token
  `outline-blue-600` focus style (rather than a wired `primary` token
  utility) on the new `<Link>`s would violate `VOC-022-D07`'s token
  restriction; the spec calls this out explicitly precisely because that
  exact pattern already exists elsewhere in the codebase as something to
  avoid, not imitate.
- `VOC-022-R06` (Low): a bare `duration-<name>`/`ease-<name>` utility class
  is an easy mistake to reintroduce (previously found live on
  VOC-019-T00's first implementation attempt, per
  `scripts/foundation/check-tailwind-token-usage.mjs`'s own header comment)
  — caught deterministically by `lint:web`, not just by inspection.
- `VOC-022-DEP-01`: Requirement must reach a founder-approved,
  implementation-ready state at adoption. DOC-01, DOC-03, and DOC-05 are
  approved documents, but an approved document is not itself implementation
  authority (`AGENTS.md`); this draft is not authority on its own.
- `VOC-022-DEP-02`: Base state (`base_sha`) is pinned to the drafting-time
  `develop` head (`044d42941ec1eee68fb66e53d59bd2e89fc51011`) and must be
  re-pinned to the then-current `develop` head at adoption.
- `VOC-022-DEP-03`: Depends on VOC-021's seven-situation `/discover` list
  (`apps/web/src/app/(app)/discover/page.tsx`, exact slugs confirmed by
  direct inspection: `airport`, `restaurant`, `hotel-check-in`,
  `job-interview`, `daily-conversation`, `work-meeting`,
  `university-class`) and VOC-018's wired Tailwind v4 `@theme` tokens
  (`apps/web/src/app/tokens.generated.css`, 64 custom properties), both
  present on `develop` at drafting time.
- `VOC-022-DEP-04`: Open scope decision — per-situation mock word content
  and count, this draft's proposed six-word lists (`tasks.md`) vs. a
  human-directed alternative (`VOC-022-D02`). Pending founder confirmation
  at adoption.
- `VOC-022-DEP-05`: Open scope decision — whether this draft's proposed
  minimal `discover/page.tsx` diff shape (wrap card content in `<Link>`,
  move existing className onto it) is an acceptable interpretation of "the
  only change to that file," or a human directs a narrower/different
  approach (`VOC-022-D05`). Pending founder confirmation at adoption.
- `VOC-022-DEP-06`: Required follow-up, not part of this task — word
  detail (a separate page/modal per word) and real, backend-persisted
  save/unsave (DOC-01 §2's fuller Journey-tab scope), once `apps/api` has
  real `journey_situations`/`journey_words`/`user_words` endpoints to wire
  against.
- `VOC-022-EV-00`..`VOC-022-EV-08`: CI run output (lint/typecheck/build/format)
  plus the independent reviewer's exact-SHA verdict, produced at
  implementation time, not now.
