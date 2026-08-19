# VOC-095 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `.github/workflows/accessibility.yml`,
  `.github/workflows/lighthouse.yml`, `.github/workflows/deploy-staging.yml`,
  `.github/workflows/scheduled-synthetics.yml`, `infra/scripts/`,
  `infra/monitoring/scheduled-synthetics.mjs`.
- Prerequisites: VOC-031 Playwright harness merged; VOC-086 scheduled-synthetics browser
  cache pattern exists; issue #792 opened from deploy-staging timeout evidence.

## File reconciliation and implementation sequence

### T00 — Bounded install script and foundation tests

| File | Action | Notes |
|------|--------|-------|
| `infra/scripts/install-playwright-chromium.sh` | create | Bounded browser + deps install contract |
| `scripts/foundation/voc095-playwright-install.test.mjs` | create | Structure, constants, fail-closed denylist |
| `specs/changes/VOC-095-.../t00-evidence.md` | create | Run 32299315180 root-cause metadata |

Ordered steps:

1. Confirm run 32299315180 metadata; record failure phase in evidence.
2. Implement install script with separated browser/deps steps, timeout/retry, binary verify.
3. Add foundation tests; run locally in task PR.
4. Run governance validation if required for `infra/scripts/` changes.

### T01 — Workflow wiring and validator alignment

| File | Action | Notes |
|------|--------|-------|
| `.github/workflows/accessibility.yml` | modify | Add cache + script |
| `.github/workflows/lighthouse.yml` | modify | Add cache + script; preserve LIGHTHOUSE_CHROME_PATH |
| `.github/workflows/deploy-staging.yml` | modify | Add cache + script on core-loop install step |
| `.github/workflows/scheduled-synthetics.yml` | modify | Replace inline install with script |
| `infra/monitoring/scheduled-synthetics.mjs` | modify | Validator asserts script + cache order |
| `scripts/foundation/voc095-playwright-install.test.mjs` | modify | Four-workflow wiring assertions |
| `apps/web/tests/e2e/README.md` | modify | Document shared contract |
| `specs/changes/VOC-095-.../t01-evidence.md` | create | Test output + workflow summary |

Ordered steps:

1. Add browser cache restore to accessibility, lighthouse, deploy-staging (scheduled
   already has cache — keep key scheme identical).
2. Replace inline install with script invocation in all four workflows.
3. Update scheduled-synthetics.mjs and extend foundation tests.
4. Update e2e README install section.
5. Run VOC-095 tests, VOC-086 scheduled-synthetics tests, and applicable deploy regression tests.

### T02 — Live staging core-loop verification

| File | Action | Notes |
|------|--------|-------|
| `specs/changes/VOC-095-.../t02-evidence.md` | create | Green deploy-staging run URL + core-loop confirmation |

Ordered steps:

1. After T01 merges to `develop`, wait for or dispatch `deploy-staging` on latest SHA.
2. Record scrubbed success evidence; confirm install step completed within job budget.
3. Document fallback proof if mirror stall cannot be reproduced live (open question 4).

## Validation and independent verification

Deterministic (T00):

```bash
node --test scripts/foundation/voc095-playwright-install.test.mjs
bash scripts/governance/validate-governance.sh   # if invoked by CI on changed paths
git diff --check
```

Deterministic (T01):

```bash
node --test scripts/foundation/voc095-playwright-install.test.mjs
node --test scripts/foundation/voc086-scheduled-synthetics.test.mjs
node --test scripts/foundation/voc084-deploy-staging-oauth.test.mjs
node --test scripts/foundation/voc088-deploy-staging-allowlist.test.mjs
bash scripts/governance/validate-governance.sh
git diff --check
```

Live (T02):

- Inspect latest `deploy-staging` run on GitHub Actions after T01 merge
- Confirm staging core-loop Playwright step success

Independent verifier binds to exact task PR SHA; confirms scope, tests, and
evidence under active A-004.

## Deployment and rollback

- **Authorization:** Merging task PRs to `develop` updates CI install behavior immediately
  on the next workflow run; no separate production application deploy for this package.
- **Rollout:** Next accessibility, lighthouse, deploy-staging, and scheduled-synthetics
  runs use the shared contract.
- **Rollback trigger:** Install script false failures block legitimate deploys/tests;
  timeout too aggressive causes widespread CI flakes.
- **Rollback mechanism:** Revert T00/T01 commits; restore inline `--with-deps` one-liner.
- **Owner:** unassigned (set at adoption).
- **Last-known-good:** `develop` HEAD before T00 merge.
