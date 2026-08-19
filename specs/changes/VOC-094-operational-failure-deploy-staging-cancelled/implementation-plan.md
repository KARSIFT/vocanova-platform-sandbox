# VOC-094 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `.github/workflows/deploy-staging.yml`,
  `.github/workflows/deploy-production.yml`,
  `.github/workflows/operational-failure-monitoring.yml`,
  `infra/scripts/open-failure-issue.sh` (and any new bounded classifier helper).
- Prerequisites: VOC-032-T07 merged (deploy-staging exists); VOC-088-T02 merged
  (operational-failure observer opened issue #781).

## File reconciliation and implementation sequence

### T00 — Fix deploy concurrency queue and benign-cancel observer filter

| File | Action | Notes |
|------|--------|-------|
| `.github/workflows/deploy-staging.yml` | modify | Add `queue: max`; document in header |
| `.github/workflows/deploy-production.yml` | modify | Add `queue: max` for parity |
| `.github/workflows/operational-failure-monitoring.yml` | modify | Gate issue creation on benign-cancel classifier |
| `infra/scripts/open-failure-issue.sh` and/or new classifier helper | modify/create | Bounded metadata skip path |
| `scripts/foundation/voc094-deploy-concurrency.test.mjs` | create | Queue + classifier deterministic tests |
| `scripts/foundation/voc088-failure-to-issue.test.mjs` | modify (minimal) | New skip fixtures; preserve existing tests |
| `specs/changes/VOC-094-.../t00-evidence.md` | create | Root-cause metadata from run 32290409156 |

Ordered steps:

1. Confirm run 32290409156 metadata (annotation, zero jobs, duration); record in evidence.
2. Add `queue: max` to both deploy workflow concurrency blocks.
3. Implement classifier + observer wiring; keep fail-closed default.
4. Add/extend deterministic tests.
5. Run VOC-094 tests, VOC-088 failure-to-issue tests, and applicable deploy regression tests.

### T01 — Record live verification

| File | Action | Notes |
|------|--------|-------|
| `specs/changes/VOC-094-.../t01-evidence.md` | create | Green deploy run + supersession hygiene proof |

Ordered steps:

1. After T00 merges to `develop`, confirm latest `deploy-staging` success.
2. Execute controlled supersession proof or document approved fallback per open question 3.
3. Record scrubbed run URLs; confirm issue fingerprint hygiene for #781.

## Validation and independent verification

Deterministic (T00):

```bash
node --test scripts/foundation/voc094-deploy-concurrency.test.mjs
node --test scripts/foundation/voc088-failure-to-issue.test.mjs
node --test scripts/foundation/voc084-deploy-staging-oauth.test.mjs
node --test scripts/foundation/voc088-deploy-staging-allowlist.test.mjs
bash scripts/governance/validate-governance.sh   # if invoked by CI on changed paths
git diff --check
```

Live (T01):

- Inspect latest `deploy-staging` run on GitHub Actions after T00 merge
- Optional controlled triple-push supersession on `develop` (operator-coordinated)

Independent verifier binds to exact task PR SHA; confirms scope, tests, and
evidence under active A-004.

## Deployment and rollback

- **Authorization:** Merging task PRs to `develop` updates workflow behavior on
  the next push or dispatch; no separate production deploy step for this package.
- **Rollout:** Immediate effect on the next `deploy-staging` / `deploy-production`
  concurrency scheduling and on subsequent operational-failure observations.
- **Rollback trigger:** Classifier false negatives hide real deploy cancellations;
  deep queue causes unacceptable deploy latency during merge bursts.
- **Rollback mechanism:** Revert T00 commits; remove `queue: max` and classifier gate.
- **Owner:** unassigned (set at adoption).
- **Last-known-good:** `develop` HEAD before T00 merge.
