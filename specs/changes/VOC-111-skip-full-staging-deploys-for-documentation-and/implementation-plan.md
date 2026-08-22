# VOC-111 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `.github/workflows/deploy-staging.yml`, staging SSH deploy semantics,
  `tests/staging-e2e/core-loop.staging.spec.ts`, repository secrets used by
  deploy-staging (read-only in CI), operational-failure observer wiring.
- Prerequisites: VOC-032-T07 merged (deploy-staging exists); VOC-050/VOC-084/VOC-088
  merged (post-deploy gates); VOC-094 merged (concurrency queue — do not regress);
  VOC-110 merged (container runtime merge gate in pipeline.yml — separate from this
  package's push selection).

## File reconciliation and implementation sequence

### T00 — Add fail-closed push selection, selector tests, and doc updates

| File | Action | Notes |
|------|--------|-------|
| `specs/changes/VOC-111-.../t00-evidence.md` | update | Issue evidence, allowlist decision, validation results |
| `.github/workflows/deploy-staging.yml` | modify | Push-only path allowlist; update header comment |
| `scripts/foundation/voc111-deploy-staging-paths.test.mjs` | create | Positive/negative selector fixtures |
| `docs/operations/11-devops-and-ci-cd.md` | modify (if needed) | Accurate staging push selection description |
| Existing voc084/voc088/voc094 foundation tests | regression | Must remain green |

Ordered steps:

1. Record issue #920 run metadata and the pre-change no-filter behavior in evidence.
2. Add push-only allowlist selection to `deploy-staging.yml` per `VOC-111-D03`.
3. Preserve full deploy job behavior for selected pushes and unchanged
   `workflow_dispatch`.
4. Add deterministic selector tests covering negative docs/specs/evidence cases and
   positive runtime/deploy cases, including root manifests and shared packages.
5. Remove stale near-no-op documentation from workflow comments and DevOps docs.
6. Run applicable validation; record command results in `t00-evidence.md`.

### T01 — Record live verification

| File | Action | Notes |
|------|--------|-------|
| `specs/changes/VOC-111-.../t01-evidence.md` | create/update | Operator absence proof metadata |
| `.karsift/live-evidence/VOC-111-T01.yaml` | **not used** | Negative case; see `VOC-111-D06` |

Ordered steps:

1. After T00 merges to `develop`, merge a governed docs/evidence-only change.
2. Operator queries Actions metadata and records zero matching `deploy-staging` push
   runs for the verified SHA in `t01-evidence.md`.
3. Independent verifier confirms metadata-only evidence satisfies AC-01 without
   fabricated deployment claims.

## Validation and independent verification

Deterministic (T00):

```bash
node --test scripts/foundation/voc111-deploy-staging-paths.test.mjs
node --test scripts/foundation/voc084-deploy-staging-oauth.test.mjs
node --test scripts/foundation/voc088-deploy-staging-allowlist.test.mjs
node --test scripts/foundation/voc094-deploy-concurrency.test.mjs
bash scripts/governance/validate-governance.sh   # when required for changed paths
git diff --check
```

Independent verifier (exact reviewed task PR SHA):

- Confirm push allowlist matches `VOC-111-D03` and tests cover every declared class.
- Confirm `workflow_dispatch` and selected-push deploy steps/concurrency unchanged.
- Confirm documentation no longer claims docs-only near-no-op deploys.
- For T01, confirm operator absence metadata is allowlisted and bound to the exact
  reviewed docs/evidence-only push SHA.

## Deployment and rollback

- **Staging effect:** After T00, non-runtime pushes stop scheduling deploy-staging;
  selected pushes behave as today.
- **Production effect:** None intentional. Production deploy filtering is out of scope.
- **Rollback trigger:** Selected runtime changes fail to deploy, or non-runtime pushes
  still schedule full deploys after T00.
- **Rollback mechanism:** Revert the task PR(s); `workflow_dispatch` remains available
  for manual staging redeploy while revert is prepared.
- **Last-known-good reference:** Pre-change `deploy-staging.yml` on `develop` immediately
  before T00 merge.
