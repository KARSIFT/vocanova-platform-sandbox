# VOC-085 — Seed canonical P1 content in production and gate real learning routes

**Status: draft, not adopted.** Nothing in this package is implementation-authorized.
It is a draft response to
[issue #702](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/702),
prepared for plan review and adoption under **active A-004** (proposed **R3**).

## Identity and lifecycle

- Package ID: VOC-085
- Title: Seed canonical P1 content in production and gate real learning routes
- Canonical path:
  `specs/changes/VOC-085-seed-canonical-p1-content-in-production-and-gate`
- Lifecycle state: `draft` (not adopted, not authorized for implementation)
- Proposed risk: `R3` (draft proposal only — see `change.yaml`'s
  `planned_implementation_risk_floor`; measured path floor at drafting is
  **R3** for `.github/workflows/deploy-production.yml` and production smoke
  scripts)
- Owner: unassigned (see `change.yaml`'s `owners` block)
- Approval evidence: none yet — `approval_status: not-approved`,
  `implementation_authorized: false`, `implementation.authorized: false`,
  `repository_adoption_status: not-adopted`
- Target branch: `develop`
- Linked GitHub issues:
  - [#702](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/702)
    (this package's requirement source)
- Related but distinct packages:
  - [VOC-052](../VOC-052-staging-has-never-had-its-canonical-p1-content)
    — staging P1 seed (predecessor; production mirror deferred)
  - [VOC-050](../VOC-050-run-core-loop-e2e-against-real-staging-and-gate)
    — synthetic smoke identity + session mint (reused)
  - [VOC-067](../VOC-067-production-outage-root-cause-consider-unifying)
    — staging/production isolation + shared-edge (must be preserved)

## Why this exists

Canonical P1 learning content is present on staging but absent in production,
while production deploy smoke still passes. Verified 2026-08-16 evidence in
issue #702:

1. Current promoted revision is
   `62e381575d2585d2f02d69031b8e0119987d3757`.
2. A read-only production database query after the latest successful
   deployment reports `journey_situations=0`; staging reports 7.
3. Production public health and the latest deployment smoke suite are green,
   proving present smoke can pass with zero canonical learning content.
4. `deploy-staging.yml` builds and runs the existing idempotent
   `apps/api/cmd/seed` P1 content seed after migrations and the synthetic
   account seed, before application convergence.
5. `deploy-production.yml` bundles migrations and the synthetic smoke-user
   seed but does **not** bundle or run the canonical P1 content seed.
6. Production smoke only checks HTTP 200 for `/api/v1/journey-situations`; it
   does not assert a non-empty response, validate situation/word details, or
   sweep real page routes.
7. The synthetic-user SQL already deterministically refreshes
   `onboarding_status='completed'`, so the reserved account can be used
   without changing a real user.

Root cause: production content seeding was intentionally deferred (VOC-052-T02
closed out-of-scope-for-now), and the production gate validates transport/
status rather than usable content.

## What this package does

1. **Production P1 seed bundling and run** (`VOC-085-T00`): mirror staging —
   build a static Linux `p1-content-seed` with the repository Go toolchain
   pin (`go-version-file: apps/api/go.mod`), bundle it, run it against the
   production-private database after migrations and the synthetic-user seed,
   and abort before `docker compose up -d` on any seed failure. Add
   deterministic ordering/bundling/fail-closed tests.
2. **Content-aware production smoke** (`VOC-085-T01`): fail closed on HTTP
   200 with an empty situation list; with the workflow-minted synthetic
   session, validate at least one canonical situation detail and one word
   detail from identifiers returned by production. Extend self-tests.
3. **Authenticated non-mutating route sweep + live proof** (`VOC-085-T02`):
   cover the ten fixed routes plus one real `/discover/[situation]` and one
   real `/discover/[situation]/[word]` route derived from the API; deploy
   through the protected workflow; verify non-empty content through
   Cloudflare; confirm topology/isolation unchanged.

## What this package deliberately does NOT do

- Manual server or database edits, destructive reseeding, or replacing the
  canonical dataset.
- Cloudflare configuration changes.
- Real-user test activity, magic-link requests, OAuth completion, or
  state-changing learning actions.
- Creating naive unauthenticated page monitors (later monitoring-inventory
  package owns stable synthetic-check IDs).
- Adopting, authorizing, implementing, or merging itself.

## Open questions for the reviewing human

See `specification.md`. The most important at adoption:

1. Accept proposed **R3** (path floor R3), or raise to R4 given production
   database seed writes during deploy.
2. Exact route-sweep harness shape (`VOC-085-DEP-02`) — extend
   `smoke-test-production.sh`, sibling script, or narrowly scoped browser
   gate — as long as AC coverage and non-mutating constraints hold.

## Verification, approvals, release, and closure

See `test-plan.md`, `release-plan.md`, and `implementation-plan.md`.
Under **active A-004**, engineering-workflow gates do not wait on a founder
`approved` comment; R3 still requires strengthened evidence, independent
verification, and rollback credibility. `automatic_merge_allowed: true` is
set per AGENTS.md. This draft still carries no adoption or implementation
authority.
