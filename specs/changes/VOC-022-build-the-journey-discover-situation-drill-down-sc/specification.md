# VOC-022 — Build the Journey/Discover Situation Drill-Down Screen: Specification

## Objective and requirement source

Add the new dynamic route
`apps/web/src/app/(app)/discover/[situation]/page.tsx`, showing a scannable
list of words for the matching situation, built with **static UI only
against local mocked/placeholder data** — `apps/api` has no real endpoints
yet, so this task introduces no network call, API client import,
authentication, or database/migration work. A word already saved is
visually marked using **mock saved/not-saved state only** — there is no real
save action, since there is no backend to persist it. Additionally, each of
the seven situation cards on `apps/web/src/app/(app)/discover/page.tsx`
(built by VOC-021) must link to its own `/discover/[slug]` route — the exact
follow-up VOC-021 deliberately deferred (`VOC-021-D03`, `VOC-021-DEP-06`).

Requirement source:
- DOC-03 (`docs/design/03-ui-ux-design.md`) §5 — "Discovery is organized by
  real-life situation … Within a situation, words are shown one at a time or
  as a short scannable list; the backend controls ordering … A word already
  in the learner's saved list is visually marked and excluded from 'new'
  recommendations. Saving must succeed against the backend before the UI
  reflects it as saved — no optimistic-only save state that could desync
  from the backend."
- DOC-01 (`docs/product/01-mvp-prd.md`) §2 — "**Journey** — situation-based
  discovery … word detail, save/unsave" (this task implements only the
  within-situation word-list portion of that bundle; word detail and real
  save/unsave remain out of scope here, per the free-text request).
- DOC-05 (`docs/engineering/05-database-design.md`) §7 —
  `word_meanings` (`short_definition`, `learner_definition`, `part_of_speech`)
  and §8 — `journey_words` (join table linking situations to word meanings,
  ordered by `is_core desc, display_order, relevance_score desc`) — used
  here only as naming inspiration for mock field shape, not a real
  dependency on either table.

All three are `status: approved`, `owner: founder`. Per `AGENTS.md`, an
approved document describes target UX but is not itself implementation
authority for a specific change; a founder-approved, implementation-ready
state must still be recorded at adoption. This document is a draft, not an
approved specification.

## Scope and non-goals

In scope:
1. **A new dynamic route** `apps/web/src/app/(app)/discover/[situation]/page.tsx`
   rendering, for a known situation slug, a short scannable `<ul>`/`<li>`
   list of mocked words for that situation (`VOC-022-D01`).
2. **Per-situation mock word data**, colocated in the new route file (a
   local constant, not shared with `discover/page.tsx`), with each entry
   carrying a word/phrase, a short meaning, and a mock `isSaved` boolean
   (`VOC-022-D02`).
3. **Visual "saved" marking that does not rely on color alone** — an
   icon-plus-text label (e.g. a checkmark glyph plus the word "Saved"),
   satisfying DOC-03 §10's "no information conveyed by color alone"
   (`VOC-022-D03`).
4. **Graceful handling of an unknown/invalid situation slug** via Next.js's
   built-in `notFound()` — not a custom backend lookup, not a client-side
   thrown error, not a blank page (`VOC-022-D04`).
5. **Exactly one change to `discover/page.tsx`**: wrapping each situation
   card's existing content in a `next/link` `<Link href="/discover/{slug}">`
   so the card becomes a real navigation target to its own drill-down route
   (`VOC-022-D05`). This directly reverses VOC-021's explicit non-goal
   (`VOC-021-D03`, "no navigation into a situation") — a disclosed,
   deliberate scope progression, not a silent contradiction: VOC-021 named
   this exact follow-up (`VOC-021-DEP-06`) as the reason cards were left
   non-interactive.
6. **A "Back to Journey" link** on the drill-down page back to `/discover`,
   since the fixed `BottomNav` "Journey" tab's active-state highlighting
   (`pathname === item.href` in `bottom-nav.tsx`) does not recognize a
   nested `/discover/[situation]` path as "on the Journey tab" — a known,
   disclosed limitation of a file this task must not touch (`VOC-022-D08`).

Explicitly excluded, per the free-text request:
- **Word detail** — no separate page or modal per word. Word entries in the
  drill-down list are plain, non-interactive list items: no `href`,
  `onClick`, `<Link>`, or `<button>` on an individual word (`VOC-022-D06`).
- **Save/unsave functionality** — the `isSaved` mock flag is static,
  read-only mock data; no button, toggle, or click handler changes it, and
  no backend call of any kind is introduced (`VOC-022-D06`).
- **Sentence practice** — no entry point into sentence practice from this
  screen (`VOC-022-D06`).
- Any backend/API call, data fetching, `TanStack Query`, or import of an API
  client. `apps/api`, authentication, and any database/migration work are
  untouched (`VOC-022-D00`).
- Any change to `apps/web/src/app/(app)/_components/bottom-nav.tsx`,
  `apps/web/src/app/(app)/layout.tsx`, `apps/web/src/app/(app)/home/page.tsx`,
  or `apps/web/src/app/(app)/progress/page.tsx`.
- Any change to `discover/page.tsx` beyond the single navigation addition
  described in `VOC-022-D05` — no restructuring of its mock data, heading,
  intro text, or situation set/order.
- Any bare `duration-<name>`/`ease-<name>` Tailwind utility class, matching
  the restriction VOC-018 through VOC-021 already established
  (`VOC-022-D07`).
- A second `<main>` landmark on the new route
  (`apps/web/src/app/(app)/layout.tsx` already renders the group's one
  `<main>`) (`VOC-022-D08`).
- Any new runtime dependency, `package.json`, or lockfile change.
- `generateStaticParams` is **not required** by this task — the dynamic
  segment may render per-request (no data-fetching cost applies to static
  mock data either way); whether an implementer adds it is discretionary and
  does not change any acceptance criterion (`VOC-022-D09`).
- DOC-03 §9's loading/error states beyond the not-found case, since there is
  no dynamic (network) data source that can be loading or erroring
  (`VOC-022-D00`).

## Risk and protected areas

Proposed risk: **R1**. Every target path is under
`apps/web/src/app/(app)/discover/**`, which `classify-change-risk.sh` maps to
its `*` default (R1) — no target matches an R2/R3/R4 pattern (no
`package.json`, no `*/auth/*`, no `*/migrations/*`, no governance or
CI-workflow path). This is a draft proposal — `classify-change-risk.sh` and a
human's judgment govern the actual class at implementation time.

Semantic note (does **not** change the path floor, but the independent
verifier must review it per `CLAUDE.md`): like VOC-019 through VOC-021, this
is a learner-facing content surface rendering real token colors; WCAG 2.2 AA
contrast (1.4.3 text, 1.4.11 non-text) against the `neutral`/`primary` token
pairs used here is a genuine review dimension. Additionally, because this
task turns VOC-021's previously non-interactive situation cards into real
navigation links, the verifier must confirm the new `<Link>` carries a
visible focus indicator and a hover affordance built only from wired tokens
(`VOC-022-D05`), and that the "saved" marker on a word entry is genuinely
non-color-only (`VOC-022-D03`) — both real accessibility/UX-affordance
risks, not formalities. No protected areas are touched.

## Decisions, contradictions, security, and privacy

`VOC-022-D00` — **Static UI, local mock data, no backend.** The
within-situation word list is rendered from a single, clearly-named local
mock data structure (e.g. `MOCK_SITUATION_WORD_LISTS`) colocated with the
new route, explicitly commented as placeholder pending real API wiring. No
`fetch`, `TanStack Query`, or API client import is introduced. This mirrors
VOC-019/VOC-020/VOC-021's precedent (`VOC-019-D00`, `VOC-020-D00`,
`VOC-021-D00`). Since there is no dynamic data source, no loading/error state
applies beyond the not-found case (`VOC-022-D04`).

`VOC-022-D01` — **New dynamic route, async server component, `params`
awaited.** `apps/web/src/app/(app)/discover/[situation]/page.tsx` is a new
file; no existing file occupies this path. Next.js 16 (this repo's pinned
`next` version, `16.2.10`) resolves App Router page `params` as a `Promise`
for server components, not a plain object — the component must be declared
`async` and `await` its `params` prop (e.g.
`async function SituationDiscoverPage({ params }: { params: Promise<{ situation: string }> })`).
Treating `params` as a synchronous object (a common carry-over mistake from
older Next.js versions) fails `pnpm run typecheck:web`'s `next typegen`
step, since the generated `PageProps` type for this route expects a
`Promise`.

`VOC-022-D02` — **Per-situation mock word content is placeholder text, not
reviewed product content (OPEN, `VOC-022-DEP-04`).** Each situation's word
list (word/phrase, short meaning, mock `isSaved` flag) is illustrative text
drafted for this package, not founder-reviewed product copy, and the word
count per situation (this draft proposes six per situation) is not a fixed
product requirement — DOC-05 §8's real `journey_words` relevance/ordering
model will ultimately determine the real count and order once `apps/api`
exists. `tasks.md` proposes specific content; a human should review it (or
explicitly accept it as adequate placeholder content) at adoption. To keep
acceptance criteria testable without overspecifying arbitrary mock content,
`acceptance-criteria.md` requires only: at least three words per known
situation, and — to demonstrate both the DOC-03 §5 "already saved" and
"new"/not-saved visual states — at least one mock-saved and at least one
mock-not-saved word in every situation's list.

`VOC-022-D03` — **"Saved" marking must not rely on color alone.** DOC-03
§10 ("no information conveyed by color alone") and DOC-03 §5 ("a word
already in the learner's saved list is visually marked") together require
that the saved indicator combine a decorative icon glyph
(`aria-hidden="true"`) with a visible text label (e.g. "Saved"), not merely
a background/border color change on the list item. A not-saved word renders
no such marker at all (absence of the marker, not a different color, is
what distinguishes it) — this is intentionally the same "don't invent a
second, unrequested state" caution DOC-03 §1's "one clear action per
screen" principle argues for.

`VOC-022-D04` — **Unknown/invalid slug: Next's built-in `notFound()`, not a
custom lookup or thrown error.** The route's local mock data is keyed by the
seven known situation slugs VOC-021 defined
(`airport`/`restaurant`/`hotel-check-in`/`job-interview`/
`daily-conversation`/`work-meeting`/`university-class`). If the resolved
`params.situation` does not match any key, the component calls
`notFound()` from `next/navigation`, which renders Next's default
not-found handling. No custom `not-found.tsx` is added anywhere in the repo
by this task (none exists today under `apps/web/src/app` — confirmed by
direct inspection) — building a styled not-found page is a reasonable
future enhancement but is not requested here and is out of scope; the
default Next.js not-found behavior is an acceptable, honest "this doesn't
exist" outcome for an invalid URL, not a silently-swallowed error.

`VOC-022-D05` — **Navigation added to `discover/page.tsx`, and only
navigation (OPEN, `VOC-022-DEP-05`).** The free-text request is explicit:
"this requires adding navigation to the existing situation cards — keep
this the only change to that file." This draft's proposed, minimal diff
(detailed in `tasks.md`): add one `import Link from "next/link";`, then for
each situation card, wrap the existing `<h2>`/`<p>` pair in
`<Link href={\`/discover/${situation.slug}\`}>`, moving the card's existing
`className` (border/background/padding/shadow) from the `<li>` onto the new
`<Link>` and adding a hover and `focus-visible` treatment built only from
wired tokens (`VOC-022-D07`). The mock `MOCK_DISCOVER_SITUATIONS` array,
the `<h1>Journey</h1>` heading, the intro paragraph, and the `<ul>` wrapper
are otherwise byte-identical to VOC-021's implementation. This is recorded
as an open decision, not silently resolved: a human should confirm this
diff shape is an acceptable interpretation of "the only change to that
file," or direct a narrower/different approach.

`VOC-022-D06` — **No word detail, no save/unsave, no sentence practice.**
Per the free-text request's explicit non-goals, each word entry in the
drill-down list is a plain, non-interactive list item: no `href`,
`onClick`, `<Link>`, or `<button>` anywhere on an individual word, and no
control that would toggle the mock `isSaved` flag. DOC-01 §2 bundles "word
detail, save/unsave" into the fuller Journey-tab description; this task
deliberately implements only the within-situation word-list surface DOC-03
§5 describes, leaving word detail and real save/unsave as required
follow-up work once `apps/api` has real endpoints (`VOC-022-DEP-06`).

`VOC-022-D07` — **Token usage restricted to what VOC-018 actually wired,
and no bare duration/easing utility class.** Only the eight scales VOC-018
emitted as Tailwind v4 `@theme` custom properties are available for use:
`spacing`, `neutral`, `primary`/`secondary`, `fontSize`, `radius`,
`elevation`, and `easing`/`duration` (via `var(--…)`/arbitrary-value only —
no named `duration-<name>` or `ease-<name>` utility class exists, since
Tailwind v4 does not generate utilities from those two custom-property
namespaces). No raw hex color and no `feedback`/`success`/`warning`/`error`
Tailwind utility may be used (the `feedback` scale remains unwired —
`VOC-018-DEP-05`, confirmed still true at this draft's authoring by
inspecting `apps/web/src/app/tokens.generated.css`). This applies equally to
the new `[situation]/page.tsx` route and the `<Link>` hover/focus styling
added to `discover/page.tsx` — e.g. `focus-visible:outline-primary-600` (an
auto-generated color utility from the wired `primary` scale) is permitted;
`focus-visible:outline-blue-600` (the non-token color `bottom-nav.tsx`
happens to use, a file this task does not touch or imitate) is not.
`scripts/foundation/check-tailwind-token-usage.mjs`, run by `lint:web`,
deterministically fails on any bare `duration-<name>` class or a second
`<main>`; there is no equivalent automated check for raw hex or the unwired
`feedback` scale, so those remain a structured-inspection item
(`VOC-022-AC-06`).

`VOC-022-D08` — **Heading hierarchy, semantic list, no second `<main>`, and
a "Back to Journey" link.** The new route's root element is a `<div>` (never
`<main>` — `apps/web/src/app/(app)/layout.tsx` already renders the group's
one `<main>` for every route, unchanged by this task). Content order: a
"Back to Journey" `<Link href="/discover">`, one `<h1>` (the situation
title), a short scannable intro line, and a `<ul>` with one `<li>` per word
(no heading level per word — a word entry is a data row, not a distinct
content section, matching DOC-03 §5's "short scannable list" framing; the
situation title is the page's only heading). The back-link is included
because `bottom-nav.tsx`'s active-tab highlighting
(`pathname === item.href`) will not mark "Journey" active on a nested
`/discover/[situation]` path — a known limitation of a file this task must
not touch, disclosed here rather than silently left unaddressed by any
in-scope means.

`VOC-022-D09` — **`generateStaticParams` is optional, not required.**
Because the route renders static mock data with no per-request cost, either
static generation via `generateStaticParams` (enumerating the seven known
slugs) or ordinary per-request dynamic rendering satisfies every acceptance
criterion in this package; the implementer may choose either without it
constituting a scope deviation.

### Contradictions

`VOC-022-D05` is a deliberate, disclosed reversal of VOC-021's non-goal
(`VOC-021-D03`, "no navigation into a situation … drilling into a situation
is a follow-up package's scope") — not a silent contradiction, since VOC-021
explicitly named this exact follow-up (`VOC-021-DEP-06`) as the reason those
cards were left non-interactive. DOC-01 §2's bundling of "word detail,
save/unsave" into one Journey-tab description remains only partially built
by this task, per the free-text request's explicit non-goals
(`VOC-022-D06`) — DOC-01's fuller Journey scope remains the target for
follow-up packages, same disclosure pattern VOC-021 used for its own
narrower slice. No other contradiction between DOC-01, DOC-03, DOC-05, and
the free-text request was found.

### Security and privacy

None. The new screen renders static mock text/markup only — the mock
"saved" flag is a fixed placeholder boolean, not derived from or associated
with any real learner's actual saved-word state, no input fields, no
network/API calls, no storage, cookies/tokens, and no authentication or
authorization surface. The one interactive element this task adds
(navigation `<Link>`s from `discover/page.tsx` to the new route, and the
"Back to Journey" link on the new route) are ordinary client-side page
navigations with no data submission of any kind. DOC-03 §5's "saving must
succeed against the backend before the UI reflects it as saved — no
optimistic-only save state" rule is not engaged at runtime because this
task has no save action at all (mock-only, read-only `isSaved`); the mock
data's placeholder nature is made explicit in the source so it is never
mistaken for a live integration by a future maintainer.

## Data, migrations, analytics, and accessibility

- **Data / migrations:** none. No database, schema, or migration; no
  `apps/api` change. Rollback is a plain `git revert`.
- **Analytics:** none. A static mock screen with only plain page-navigation
  links emits no analytics events.
- **Accessibility:** in scope. The word list uses a semantic `<ul>`/`<li>`
  structure so assistive technology announces the item count; the page
  retains exactly one `<h1>` (DOC-03 §10); the "saved" marker combines an
  `aria-hidden` icon with a visible text label so the state is never
  conveyed by color alone (`VOC-022-D03`); the new `<Link>`s (both the
  situation-card links and the "Back to Journey" link) carry a visible
  `focus-visible` outline built from wired tokens. Honest limitation,
  matching VOC-017/VOC-019/VOC-020/VOC-021 precedent (`VOC-017-D08`): with
  no `apps/web` component/a11y test runner yet, this is verified by
  construction and structured inspection, not automated axe/Playwright
  assertions.
