# VOC-085 — Implementation Plan

## Preconditions and protected areas

Do not begin implementation until this package is adopted (`status:
adopted`, `approval_status` per house adoption convention,
`implementation_authorized: true` / `implementation.authorized: true` in
`change.yaml`). Under **active A-004**, adoption does not wait on a founder
`approved` comment; exact-revision plan review and deterministic gates still
apply.

Additionally:

- Treat issue #702's verified root cause as the starting point
  (`VOC-085-DEP-00`). If implementer evidence shows a different primary cause,
  stop and record it before expanding scope.
- Reuse `apps/api/cmd/seed` as-is (`VOC-085-DEP-01`); do not replace the
  canonical dataset.
- Record the chosen route-sweep harness shape (`VOC-085-DEP-02`) in evidence.
- Path floor **R3**; proposed package risk **R3** — strengthened evidence and
  rollback credibility required; no founder-comment merge gate under A-004.
- Preserve staging/production isolation and shared-edge topology (VOC-067).
- Never commit secrets, mint tokens, or session values into evidence files.

## File reconciliation and implementation sequence

1. **`VOC-085-T00` — Production P1 seed build, bundle, run, ordering tests**
   - Mirror the staging pattern in `.github/workflows/deploy-production.yml`:
     - Set up Go with `go-version-file: apps/api/go.mod`.
     - Build a static Linux binary
       (`go build -C apps/api -o .../p1-content-seed ./cmd/seed`).
     - Bundle/SCP it with the existing production deploy artifacts.
     - On the host, after migrations and
       `seed-synthetic-smoke-user.sh`, run
       `DATABASE_URL="$migration_database_url" .../p1-content-seed`
       before `docker compose up -d`.
   - Fail closed on seed failure (no `continue-on-error`).
   - Add deterministic tests for bundling presence, order relative to
     migrations/synthetic seed/`up -d`, and fail-closed abort behavior.
   - Do not redesign `apps/api/cmd/seed` or edit live databases manually.

2. **`VOC-085-T01` — Content-aware production smoke + detail checks**
   - Update `infra/scripts/smoke-test-production.sh` so
     `GET /api/v1/journey-situations` with the synthetic session requires
     HTTP 200 **and** a non-empty parsed list.
   - From returned identifiers, perform non-mutating situation-detail and
     word-detail API checks (at least one each).
   - Extend `smoke-test-production.selftest.sh` (or equivalent) for
     empty-content rejection, response parsing, and positive non-empty
     fixtures. Never use production secrets in fixtures.

3. **`VOC-085-T02` — Authenticated route sweep + live production proof**
   - Implement the non-mutating authenticated route sweep covering the ten
     fixed routes and one real situation + one real word route derived from
     the API (`VOC-085-DEP-02` shape choice).
   - Wire it into the production deploy path after session mint, reusing
     `SMOKE_TEST_SESSION_COOKIE` / workflow-minted session.
   - Add self-tests for route coverage, auth-cookie handling, and failure
     behavior.
   - After merge/promotion through the protected path, record live
     Cloudflare verification of non-empty content and dynamic details, plus
     topology/isolation confirmation (shared-edge only; no 8081/8443;
     staging/production isolation intact).

Preserve compatible existing work (VOC-050 synthetic mint, VOC-052 staging
seed, VOC-038 smoke suite structure). Prefer minimal diffs.

## Validation and independent verification

Deterministic commands before claiming tasks complete (adjust to
`docs/development.md` and package filters):

```bash
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
# Production smoke self-tests (exact path may expand with T01/T02):
bash infra/scripts/smoke-test-production.selftest.sh
# Any new foundation/deploy ordering tests introduced by T00:
# bash scripts/foundation/<new-or-existing-test>   # as implemented
```

Independent verifier (per `CLAUDE.md`) must bind each report to the exact
reviewed commit SHA, confirm the implementer did not approve/merge its own
work, identify active authority model **A-004**, and report remaining R3
evidence obligations. EHR is not expected.

## Deployment and rollback

- Authorization boundary: package adoption authorizes implementation PRs
  only; production deployment remains the normal repository
  `deploy-production.yml` path after develop→main promotion (or explicit
  workflow_dispatch fallback), not a side-channel SSH procedure.
- Rollout: T00 → T01 → T02 in order. Live proof is T02's closing evidence.
- Rollback trigger: seed aborts mid-deploy incorrectly after convergence;
  empty content still green; destructive content behavior; real-user
  mutation; isolation/topology breakage.
- Rollback mechanism: revert responsible task commit(s) and redeploy via
  repository workflow. Canonical seed rows are not destructively removed by
  rollback; workflow/check strictness returns to the reverted revision.
- Accountable owner: named in each task's evidence file (unassigned at
  drafting).
- Last-known-good: tree preceding the first merged VOC-085 task (known-empty
  production content with green status-only smoke per issue #702).
