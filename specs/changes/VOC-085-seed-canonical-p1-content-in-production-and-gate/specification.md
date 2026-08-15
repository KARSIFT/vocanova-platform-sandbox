# VOC-085 — Seed canonical P1 content in production and gate real learning routes: Specification

## Objective and requirement source

Close the defect reported in
[GitHub issue #702](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/702):
restore canonical P1 learning content to production through the normal
repository deployment path, and make production verification fail closed when
content or real learning routes are unavailable.

Authority context recorded in the issue: founder/user remediation directive
dated 2026-08-16 authorizes issue creation, package adoption/implementation,
PRs, workflow execution, deployment, live verification, remediation, and
closure without routine approval. **This draft package still does not adopt
or authorize itself**; adoption remains a separate A-004 plan-review / adopt
path.

Primary context (issue #702 + drafting-time repo read):

| Item | Value |
|------|--------|
| Promoted revision (evidence) | `62e381575d2585d2f02d69031b8e0119987d3757` |
| Production content | `journey_situations=0` (read-only query) |
| Staging content | `journey_situations=7` |
| Production health/smoke | Green despite empty content |
| Staging deploy | Builds/runs `apps/api/cmd/seed` as `p1-content-seed` after migrations + synthetic-user seed, before `up -d` |
| Production deploy | Migrations + synthetic-user seed only; no P1 content seed |
| Production smoke gap | `/api/v1/journey-situations` HTTP 200 only; no non-empty or detail/route assertions |
| Synthetic account | Idempotent seed refreshes `onboarding_status='completed'` |

**Objective:** after this package's implementation, every production deploy
upserts the existing canonical P1 seed before application convergence; smoke
fails on empty content; authenticated non-mutating checks prove real
situation/word APIs and the listed learning routes; repeated deploys remain
idempotent; and shared-edge / isolation invariants remain intact.

## Confirmed findings (issue #702 + drafting-time re-read)

- `apps/api/cmd/seed` embeds `voc026-p1.json` and is documented as safe to
  rerun: existing rows are updated by fixed primary keys in a single
  transaction (idempotent upsert; no DELETE of user-owned learning state).
- `deploy-staging.yml` already implements the production-target pattern:
  `actions/setup-go` with `go-version-file: apps/api/go.mod`,
  `go build -C apps/api -o .../p1-content-seed ./cmd/seed`, SCP of the
  binary, then
  `DATABASE_URL="$migration_database_url" /opt/vocanova/apps/api/bin/p1-content-seed`
  after `seed-synthetic-smoke-user.sh` and before `docker compose up -d`.
- `deploy-production.yml` copies migrate + synthetic-user scripts and runs
  them against the production-private Postgres bridge IP, then immediately
  `up -d` — no P1 content seed build, bundle, or run.
- `infra/scripts/smoke-test-production.sh` treats
  `GET /api/v1/journey-situations` as pass on HTTP 200 only (body discarded).
- Synthetic session mint already exists in `deploy-production.yml`
  (VOC-050-T04); this package reuses that cookie and must not invent magic
  links, OAuth completion, or real-user shortcuts.
- VOC-052-T02 explicitly deferred production seed mirroring; this package is
  the governed completion of that deferred work plus stronger gates.

## Scope and non-goals

In scope:

1. Build and bundle the existing canonical P1 seed in
   `deploy-production.yml`, using the repository Go toolchain pin and a
   static Linux binary as in staging.
2. Run it against the production-private database after migrations and the
   idempotent synthetic-user seed, and before `docker compose up -d`.
3. Preserve existing production data; rely on and test the seed's idempotent
   upsert semantics. Any seed failure must abort before application
   convergence.
4. Strengthen production smoke so a 200 response with an empty situation list
   fails.
5. With the reserved synthetic account, validate at least one canonical
   situation and one word-detail API response.
6. Add a non-mutating authenticated production route sweep or equivalent
   browser-level functional gate covering all fixed routes:
   `/`, `/signin`, `/auth/magic`, `/onboarding`, `/home`, `/discover`,
   `/reviews`, `/progress`, `/settings`, `/settings/account`, and at least
   one real `/discover/[situation]` plus one real
   `/discover/[situation]/[word]` route derived from the API response.
7. Reuse the workflow-minted synthetic session. Do not request magic links,
   complete OAuth, mutate a real user, or perform state-changing learning
   actions.
8. Keep the synthetic user's onboarding status deterministic through the
   repository seed; do not edit the live database manually.
9. Add deterministic self-tests/config tests for seed ordering/bundling,
   empty-content rejection, response parsing, route coverage, auth-cookie
   handling, and failure behavior.
10. Deploy through the protected workflow, verify non-empty content and
    dynamic details through Cloudflare, and confirm topology/isolation
    remain unchanged.

Non-goals / explicitly excluded:

- Manual server/database edits, destructive reseeding, replacing the
  canonical dataset.
- Cloudflare changes.
- Real-user test activity.
- Unrelated UI redesign.
- Naive unauthenticated page monitors / new monitoring-inventory IDs
  (deferred to the later monitoring-inventory package).
- Snapshot-then-recheck-drift promotion tasks (not applicable).
- Adopting or authorizing this package from within the draft.

## Risk and protected areas

Builder assessment: expected paths include
`.github/workflows/deploy-production.yml` and
`infra/scripts/smoke-test-production*.sh` (R3 floor), plus production
functional-gate tests. Drafting-time
`scripts/governance/classify-change-risk.sh --files-from` against the
expected list reported **Detected path-based risk floor: R3**.

This package **proposes R3** for the change as a whole. This is a **draft
proposal for the reviewing human at adoption time, not a determination**.
The independent verifier may raise to R4 if semantic review of production
database writes during deploy warrants it. Content upserts are limited to
repository-owned canonical seed records; user-owned learning state must not
be overwritten beyond the seed's existing upsert semantics for fixed seed
IDs.

Protected areas: `.github/workflows/deploy-production.yml`, production
smoke/synthetic checks, canonical seed invocation. Do not weaken staging vs
production isolation.

Under **active A-004**, engineering-workflow gates require no founder
`approved` comment. EHR is not triggered by this drafting pass.

## Decisions, contradictions, security, and privacy

`VOC-085-D00` (recorded for traceability; formal acceptance at adoption):
Production deploy must build, bundle, and run the existing
`apps/api/cmd/seed` P1 content seed after migrations and the synthetic-user
seed, and before application convergence, mirroring staging's fail-closed
shape.

`VOC-085-D01`: Seed failures exit non-zero before `docker compose up -d`;
partial success must not be reported.

`VOC-085-D02`: The seed remains the existing idempotent upsert tool. Repeated
deploys must not duplicate canonical situations/words or overwrite
user-owned learning state beyond fixed seed-ID upsert semantics.

`VOC-085-D03`: Production smoke must fail when
`GET /api/v1/journey-situations` returns HTTP 200 with an empty list.

`VOC-085-D04`: With the workflow-minted synthetic session, production gates
must validate at least one real situation detail and one word detail from
identifiers returned by production APIs.

`VOC-085-D05`: A non-mutating authenticated gate must cover the ten fixed
routes listed in scope and at least one real situation route and one real
word route derived from the API. No magic links, OAuth completion,
real-user mutation, or state-changing learning actions.

`VOC-085-D06`: Synthetic onboarding completion remains the responsibility of
the repository synthetic-user seed; no manual live DB edits.

`VOC-085-D07`: Topology/isolation invariants from VOC-067 remain intact:
single shared-edge nginx; database/secret/network/directory/deploy-user
isolation; absence of 8081/8443.

Security/privacy:

- Use only the reserved synthetic test identity and read-only route/API
  checks for verification.
- Do not read real users through privileged shortcuts or change real-user
  state.
- Do not log production secrets, session opaque values beyond what existing
  mint plumbing already redacts, or personal data.
- Canonical seed writes are repository-owned content only.

## Data, migrations, analytics, and accessibility

- **Data / migrations:** No new schema migration. Additive/upsert content
  seed only via existing `apps/api/cmd/seed`. Synthetic-user seed unchanged
  except as already idempotent.
- **Analytics:** None expected — evidence-backed non-applicability for product
  analytics instrumentation in this package.
- **Accessibility:** Route sweep is non-mutating reachability/functional
  gating, not an a11y redesign. Do not regress existing page accessibility
  while adding checks.

## Open questions

1. **Route-sweep harness shape (`VOC-085-DEP-02`):** Prefer extending
   `smoke-test-production.sh` and its selftest if curl-level authenticated
   GETs satisfy AC for all listed routes; use a sibling script or narrowly
   scoped browser harness only if SPA/auth cookie behavior requires it.
   Requirement coverage is fixed; mechanism is an implementer choice that
   must be recorded in evidence.
2. **Risk elevation:** Reviewing human / independent verifier may raise
   proposed R3 to R4 if production DB seed-during-deploy is judged to need
   the stronger R4 evidence class. Under A-004 that does not reintroduce a
   founder-comment merge gate.

## Contradictions

None between the issue body and drafting-time repository evidence. The
historical VOC-052 decision to defer production seeding is superseded by
issue #702's remediation directive for this new package; this draft does not
rewrite VOC-052's historical records.
