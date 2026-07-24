# VOC-022 — Release Plan

## Release and deployment authorization

Not applicable, and not authorized by this draft. `release.deployment` is
`prohibited` — merging the implementation to `develop` is the entire scope.
A merged package does not itself authorize any production deployment. No
release authority is claimed here; adoption and merge decisions belong to a
human.

## Preconditions, monitoring, and outcome

Exact revision: the implementation PR's head commit, bound in the reviewer's
verdict per `CLAUDE.md`. Preconditions: package adopted; both open scope
decisions resolved (`VOC-022-D02`/`DEP-04` per-situation mock word
content/count, `VOC-022-D05`/`DEP-05` `discover/page.tsx` navigation-diff
shape); implementation authorized against a founder-approved
implementation-ready state (`VOC-022-DEP-01`); `base_sha` re-pinned to the
then-current `develop` head (`VOC-022-DEP-02`); CI green
(`lint:web`/`typecheck:web`/`build:web`/`format:check`). No runtime
monitoring applies — a static mock drill-down screen has no backend,
telemetry, or dynamic state to monitor; the only "monitoring" is CI on the
PR. Outcome owner: founder. Because this draft leaves
`automatic_merge_allowed: false`, no automatic merge is asserted; the merge
decision is made by a human/CI at implementation time under the merge-gate
policy then in force, not by this document.

## Rollback

Trigger: post-merge discovery of a broken `/discover/{slug}` build, a
missing or incorrectly-structured "saved" marker (e.g. color-only), a
reintroduced interactive affordance on a word entry that shouldn't exist yet
(`VOC-022-AC-02`), an unhandled unrecognized-slug case
(`VOC-022-AC-03`), an unauthorized change to `discover/page.tsx` beyond the
navigation edit (`VOC-022-AC-04`), a `feedback`-scale/raw-hex/bare-duration
utility usage that should have used a wired token (`VOC-022-AC-06`), a
reintroduced second `<main>` landmark (`VOC-022-AC-07`), or any inadvertent
backend/API import (`VOC-022-AC-05`). Mechanism: `git revert` of the merge
commit — safe and complete, since the change adds one static route and one
minimal edit to another static file, and nothing depends on this screen
beyond ordinary navigation between `/discover` and `/discover/{slug}`.
Owner: founder. Last-known-good reference: `develop` at this package's
(adoption-time re-pinned) `base_sha`.

## Independent verification, human approvals, and closure

Independent verification: exact-SHA reviewer verdict per `CLAUDE.md`,
covering the per-situation word-list content and non-interactivity of word
entries (`VOC-022-AC-00`..`AC-02`), the not-found handling
(`VOC-022-AC-03`), the confined `discover/page.tsx` diff
(`VOC-022-AC-04`), scope discipline against backend/protected-area touch
(`VOC-022-AC-05`), token-usage restriction (`VOC-022-AC-06`), the
semantic-structure/no-second-`<main>` rule (`VOC-022-AC-07`), and the
deterministic checks (`VOC-022-AC-08`), plus confirmation that
`bottom-nav.tsx`/`(app)/layout.tsx`/`home/page.tsx`/`progress/page.tsx`/
`package.json`/lockfile are unchanged and that WCAG 2.2 AA contrast holds
for every token color pairing actually rendered (`VOC-022-R01`).

Required human approvals: a founder-approved implementation-ready state at
adoption (`VOC-022-DEP-01`, including both open scope decisions,
`VOC-022-D02` and `VOC-022-D05`), plus the merge decision at implementation
time. Proposed class is R1; under active A-003 no standing
technical-steward approval is required merely for the class, and this is
below R3 regardless — but a human still adopts the package, approves the
requirement, and authorizes the merge, and the independent verifier must
still review the saved-marker accessibility question, the
`discover/page.tsx` scope-boundary discipline, and token-usage discipline
as genuine semantic dimensions, not formalities.

Do not conflate repository merge, release, activation, and closure. Closure:
no originating issue exists to auto-close (this package originates from a
free-text request plus a source document, not a GitHub issue); closure is
the reviewer's exact-SHA PASS plus the same manual `develop`-merge closure
step the recent VOC-0xx packages used. Production deployment and autonomous
release remain disabled.
