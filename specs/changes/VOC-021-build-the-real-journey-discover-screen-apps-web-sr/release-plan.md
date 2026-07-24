# VOC-021 — Release Plan

## Release and deployment authorization

Not applicable, and not authorized by this draft. `release.deployment` is
`prohibited` — merging the implementation to `develop` is the entire scope.
A merged package does not itself authorize any production deployment. No
release authority is claimed here; adoption and merge decisions belong to a
human.

## Preconditions, monitoring, and outcome

Exact revision: the implementation PR's head commit, bound in the reviewer's
verdict per `CLAUDE.md`. Preconditions: package adopted; both open scope
decisions resolved (`VOC-021-D01`/`DEP-04` situation set/order,
`VOC-021-D08`/`DEP-05` mock-copy adequacy); implementation authorized
against a founder-approved implementation-ready state (`VOC-021-DEP-01`);
`base_sha` re-pinned to the then-current `develop` head (`VOC-021-DEP-02`);
CI green (`lint:web`/`typecheck:web`/`build:web`/`format:check`). No runtime
monitoring applies — a static mock Discover screen has no backend,
telemetry, or dynamic state to monitor; the only "monitoring" is CI on the
PR. Outcome owner: founder. Because this draft leaves
`automatic_merge_allowed: false`, no automatic merge is asserted; the merge
decision is made by a human/CI at implementation time under the merge-gate
policy then in force, not by this document.

## Rollback

Trigger: post-merge discovery of a broken `/discover` build, incorrect or
missing mock content, a reintroduced interactive affordance or drill-down
route that shouldn't exist yet (`VOC-021-AC-02`), a `feedback`-scale/raw-hex/
bare-duration-utility usage that should have used a wired token
(`VOC-021-AC-04`), a reintroduced second `<main>` landmark
(`VOC-021-AC-05`), or any inadvertent backend/API import (`VOC-021-AC-03`).
Mechanism: `git revert` of the merge commit — safe and complete, since the
change replaces one static placeholder file with another and nothing
depends on this screen beyond ordinary navigation to `/discover`. Owner:
founder. Last-known-good reference: `develop` at this package's
(adoption-time re-pinned) `base_sha`.

## Independent verification, human approvals, and closure

Independent verification: exact-SHA reviewer verdict per `CLAUDE.md`,
covering the situation-list content and its non-interactivity
(`VOC-021-AC-00`..`AC-02`), scope discipline against backend/protected-area
touch (`VOC-021-AC-03`), token-usage restriction (`VOC-021-AC-04`), the
semantic-structure/no-second-`<main>` rule (`VOC-021-AC-05`), and the
deterministic checks (`VOC-021-AC-06`), plus confirmation that
`bottom-nav.tsx`/`(app)/layout.tsx`/`home/page.tsx`/`progress/page.tsx`/
`package.json`/lockfile are unchanged and that WCAG 2.2 AA contrast holds
for every token color pairing actually rendered (`VOC-021-R01`).

Required human approvals: a founder-approved implementation-ready state at
adoption (`VOC-021-DEP-01`, including both open scope decisions,
`VOC-021-D01` and `VOC-021-D08`), plus the merge decision at implementation
time. Proposed class is R1; under active A-003 no standing
technical-steward approval is required merely for the class, and this is
below R3 regardless — but a human still adopts the package, approves the
requirement, and authorizes the merge, and the independent verifier must
still review the non-interactivity/affordance question and token-usage
discipline as genuine semantic dimensions, not formalities.

Do not conflate repository merge, release, activation, and closure. Closure:
no originating issue exists to auto-close (this package originates from a
free-text request plus a source document, not a GitHub issue); closure is
the reviewer's exact-SHA PASS plus the same manual `develop`-merge closure
step the recent VOC-0xx packages used. Production deployment and autonomous
release remain disabled.
