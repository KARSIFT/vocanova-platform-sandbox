# VOC-110 — Implementation Plan

## Preconditions and protected areas

- Do not begin until the package is adopted and implementation-authorized.
- Protected areas: `.github/workflows/deploy-staging.yml`, staging SSH deploy semantics,
  `apps/web/tests/staging-e2e/core-loop.staging.spec.ts`, repository secrets used by
  deploy-staging (read-only in CI).
- Prerequisites: VOC-032-T07 merged (deploy-staging exists); VOC-050-T02 merged
  (staging core-loop gate); VOC-088-T02 merged (operational-failure observer opened
  issue #911); VOC-094 merged (concurrency queue — do not regress); VOC-095 merged
  (bounded Playwright install).

## File reconciliation and implementation sequence

### T00 — Diagnose and fix deploy-staging failure

| File | Action | Notes |
|------|--------|-------|
| `specs/changes/VOC-110-.../t00-evidence.md` | create/update | Run metadata + log-derived failing step |
| `apps/web/`, `packages/`, `pnpm-lock.yaml` | modify (if proven) | Dependency or UI regression fix |
| `apps/web/tests/staging-e2e/` | modify (if proven) | Stale assertion updates |
| `infra/scripts/` | modify (if proven) | Harness/OAuth/Playwright helper fix |
| `.github/workflows/deploy-staging.yml` | modify (if proven) | Minimal miswiring only |
| `scripts/foundation/voc110-*.test.mjs` | create (if warranted) | Regression fixture |
| Existing voc084/voc088/voc095 foundation tests | regression | Must remain green |

Ordered steps:

1. Pull run 32566405628 job metadata and logs; record failing step in evidence.
2. Choose remediation surface per evidence (application, test, harness, or workflow).
3. Implement smallest correct fix; preserve fail-closed deploy semantics.
4. Add/extend deterministic tests.
5. Run applicable validation commands; record in `t00-evidence.md`.

### T01 — Record live verification

| File | Action | Notes |
|------|--------|-------|
| `specs/changes/VOC-110-.../t01-evidence.md` | create | Green deploy-staging run metadata |
| `.karsift/live-evidence/VOC-110-T01.yaml` | present | Operator-owned contract (drafted in plan PR) |

Ordered steps:

1. After T00 merges to `develop`, wait for push-triggered `deploy-staging` or use
   governed reconcile per live-evidence contract.
2. Record scrubbed success run metadata in `t01-evidence.md`.
3. Confirm issue #911 fingerprint hygiene.

## Validation and independent verification

Deterministic (T00):

```bash
node --test scripts/foundation/voc110-*.test.mjs          # if added
node --test scripts/foundation/voc084-deploy-staging-oauth.test.mjs
node --test scripts/foundation/voc088-deploy-staging-allowlist.test.mjs
node --test scripts/foundation/voc095-playwright-install.test.mjs
# plus any other deploy-staging wiring tests named in the task PR
pnpm validate                                              # if apps/web/ or packages/ changed
bash scripts/governance/validate-governance.sh             # if invoked by CI on changed paths
bash scripts/governance/classify-change-risk.sh            # confirm task path floor
git diff --check
```

Live (T01):

- Repository-controlled reconcile of `.karsift/live-evidence/VOC-110-T01.yaml`
- Inspect qualifying `deploy-staging` run on GitHub Actions (metadata only)

Independent verifier binds to exact task PR SHA; confirms scope, tests, and
evidence under active A-004.

## Deployment and rollback

- **Authorization:** Merging T00 to `develop` triggers `deploy-staging` on the next
  push; T01 verifies that path succeeds. No separate production deploy authorization
  from this package alone.
- **Rollout:** Fix takes effect on staging at the first successful post-merge deploy.
  Production receives the fix only via normal develop → main promotion after roster
  completion.
- **Rollback trigger:** Fix introduces staging regression, breaks core-loop gate
  falsely, or reintroduces deploy-staging failures.
- **Rollback mechanism:** Revert T00 commits on `develop`; re-run deploy-staging.
- **Owner:** unassigned (set at adoption).
- **Last-known-good:** `develop` HEAD immediately before T00 merge (record exact SHA
  at task completion).
