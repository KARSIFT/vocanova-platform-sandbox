# VOC-021 — Impact Analysis

## Security and privacy

None. The screen renders static mock markup only — no links, buttons, input
fields, network/API calls, data fetching, storage, cookies/tokens,
authentication or authorization surface, and no personal data — the mock
situation titles/descriptions are fixed placeholder content, not derived
from or associated with any real learner. DOC-03 §1's "backend is
authoritative" rule is not engaged at runtime because nothing here fetches
or persists discovery/situation state. No secrets are introduced. `apps/api`
is untouched.

## Data and migrations

None. No database, schema, migration, or persisted state of any kind. Purely
a front-end file replacement under `apps/web/src/app/(app)/discover/`.
Rollback is a plain `git revert` of the merge commit with no data
implications.

## Analytics and accessibility

- **Analytics:** none — a static mock screen with no interactive elements
  emits no analytics events.
- **Accessibility:** in scope. The situation list uses a semantic `<ul>`/
  `<li>` structure so assistive technology announces the item count; each
  situation title is a real `<h2>`, giving screen-reader users
  heading-level navigation; one `<h1>` per page (DOC-03 §10). Because cards
  are deliberately non-interactive (`VOC-021-D03`), they must carry no
  implicit link/button semantics, `tabindex`, or hover/focus affordance
  styling that would misleadingly suggest tappability — a genuine review
  item, not a formality, since a static card that merely *looks* tappable is
  a worse UX outcome than one that plainly reads as informational content.
  Honest limitation, same as VOC-017/VOC-019/VOC-020
  (`VOC-017-D08`/`DEP-05`): `apps/web` has no component/a11y test runner
  yet, so this is verified by construction and structured inspection, not
  automated axe/Playwright assertions — that automated coverage remains the
  named follow-up VOC-017 already flagged, not a gap this package reopens or
  resolves.

## Risks, dependencies, and evidence

- `VOC-021-R00` (Low, scoping): "static UI only" against DOC-01 §2's fuller
  Journey-tab description (situation discovery + word detail + save/unsave)
  is easy to overrun into building within-situation word lists, word detail,
  save/unsave, or a `/discover/[situation]` route. Mitigated by the hard
  non-goals in `specification.md` and the explicit `VOC-021-D02`/`D03`
  decisions; the verifier should reject any such creep.
- `VOC-021-R01` (Medium, semantic): like VOC-019's Home screen and VOC-020's
  Progress screen, this renders token colors across a learner-facing content
  surface. WCAG 2.2 AA contrast (1.4.3 text, 1.4.11 non-text) for every
  `neutral` text-on-background pairing used must be verified by the
  implementer and re-checked by the independent reviewer.
- `VOC-021-R02` (Medium, UX-affordance / accessibility): a card-like list
  item that is deliberately non-interactive (`VOC-021-D03`) risks looking
  tappable without being tappable — the opposite failure mode from
  VOC-020's per-day markers (which needed to convey state without color
  alone), but a genuine review dimension of its own. The verifier must
  specifically check that no hover/focus/cursor styling implies
  interactivity on a situation entry.
- `VOC-021-R03` (Low, forward-compatibility): the mock data shape (e.g.
  `MOCK_DISCOVER_SITUATIONS`) is not itself a contract — when the real
  API-wiring and drill-down-navigation follow-up package lands, it is
  expected to reshape or discard this structure freely (including adding
  `href`s once `/discover/[situation]` exists). No downstream code should be
  written against the mock's exact shape as if it were a DTO.
- `VOC-021-R04` (Low): using the `feedback` token scale (unwired,
  `VOC-018-DEP-05`) would be an easy way to add emphasis color to a card;
  the spec forbids it (`VOC-021-D06`) precisely because it is not wired into
  `apps/web`'s Tailwind `@theme` layer yet.
- `VOC-021-R05` (Low): a bare `duration-<name>`/`ease-<name>` utility class
  is an easy mistake to reintroduce (it was found live on VOC-019-T00's
  first implementation attempt, per
  `scripts/foundation/check-tailwind-token-usage.mjs`'s own header comment)
  — now caught deterministically by `lint:web`, not just by inspection.
- `VOC-021-DEP-01`: Requirement must reach a founder-approved,
  implementation-ready state at adoption. DOC-01, DOC-03, and DOC-05 are
  approved documents, but an approved document is not itself implementation
  authority (`AGENTS.md`); this draft is not authority on its own.
- `VOC-021-DEP-02`: Base state (`base_sha`) is pinned to the drafting-time
  `develop` head (`e31575e06dd3397ff0bc218f5e34089b0bdf7fc5`) and must be
  re-pinned to the then-current `develop` head at adoption.
- `VOC-021-DEP-03`: Depends on the VOC-017 shell (`/discover` route,
  `BottomNav`, `(app)/layout.tsx`) and VOC-018's wired Tailwind v4 `@theme`
  tokens (`apps/web/src/app/tokens.generated.css`, 64 custom properties),
  both present on `develop` at drafting time (confirmed by direct
  inspection).
- `VOC-021-DEP-04`: Open scope decision — situation set/order, exactly the
  seven named entries in the free-text request's order (this draft's
  proposal) vs. a different or larger illustrative set (`VOC-021-D01`).
  Pending founder confirmation at adoption.
- `VOC-021-DEP-05`: Open scope decision — whether this draft's proposed
  mock copy (the seven short descriptions in `tasks.md`) is adequate
  placeholder content or needs different wording (`VOC-021-D08`). Pending
  founder confirmation at adoption.
- `VOC-021-DEP-06`: Required follow-up, not part of this task — the real
  `/discover/[situation]` drill-down route, within-situation word list, word
  detail, and save/unsave (DOC-01 §2's fuller Journey-tab scope), once
  `apps/api` has real `journey_situations`/`journey_words`/`user_words`
  endpoints to wire against.
- `VOC-021-EV-00`..`VOC-021-EV-06`: CI run output (lint/typecheck/build/format)
  plus the independent reviewer's exact-SHA verdict, produced at
  implementation time, not now.
