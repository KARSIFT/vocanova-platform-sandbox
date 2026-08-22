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
| `specs/changes/VOC-110-.../t00-evidence.md` | update | Confirmed run/root-cause metadata and validation |
| `apps/web/package.json` | modify | Paired `next` / `@next/eslint-plugin-next` 16.3.2 repair |
| `pnpm-lock.yaml` | modify | Frozen lockfile for the paired repair |
| `.github/workflows/pipeline.yml` | modify | Path-aware, merge-gating production Docker boot/HTTP job |
| `scripts/foundation/voc110-web-container-runtime.test.mjs` | create | Workflow path/command/merge-gate contract |
| `docs/operations/10-development-workflow.md` | modify | Document shipped-artifact gate |
| Existing voc084/voc088/voc095 foundation tests | regression | Must remain green |

Ordered steps:

1. Record the confirmed `Poll staging.vocanova.site/` failure and sanitized runtime
   diagnosis in evidence.
2. Upgrade the paired Next.js packages to stable 16.3.2 and refresh the frozen lock.
3. Add a pipeline job that detects relevant paths, builds the real web Dockerfile,
   starts the image, asserts it remains running, requires HTTP 2xx, and always cleans
   up. Make merge-gate depend on this job so failure cannot merge.
4. Add deterministic workflow tests for relevant/irrelevant paths, runtime commands,
   cleanup, and merge-gate dependency; document the gate.
5. Run applicable validation and a real Docker build/run proof; record results in
   `t00-evidence.md`.

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
docker build -f apps/web/Dockerfile -t vocanova-web:voc110 .
# start the image, require it remains running and serves HTTP 2xx, then remove it
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
- **Rollback mechanism:** pin the paired Next.js packages back to 16.3.0 through a
  governed PR if 16.3.2 cannot satisfy the image-runtime gate; re-run deploy-staging.
- **Owner:** unassigned (set at adoption).
- **Last-known-good:** `develop` HEAD immediately before T00 merge (record exact SHA
  at task completion).
