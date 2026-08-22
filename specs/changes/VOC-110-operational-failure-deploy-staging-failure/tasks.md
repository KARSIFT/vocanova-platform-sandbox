# VOC-110 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01**.

## VOC-110-T00 — Diagnose and fix deploy-staging failure from run 32566405628

- Requirement source: issue #911; `VOC-110-D00`–`D06`
- Acceptance criteria: `VOC-110-AC-00`, `VOC-110-AC-01`, `VOC-110-AC-02`
- Tests: `VOC-110-TEST-00` through `VOC-110-TEST-05`, applicable deploy-staging
  regression tests from VOC-084, VOC-088, and VOC-095
- Evidence: `VOC-110-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record in `t00-evidence.md` the confirmed failure chain and only bounded metadata:
   `Poll staging.vocanova.site/` failed after API health passed; the staging web
   returned 502; Next.js 16.3.1 standalone omitted an `@swc/helpers` ESM runtime
   module on Node 24. Do not copy secrets, SSH transcripts, cookies, OAuth state,
   tokens, full logs, or personal data.
2. Upgrade `next` and `@next/eslint-plugin-next` together to stable 16.3.2 and refresh
   `pnpm-lock.yaml`. Preserve the other PR #859 updates. Use 16.3.0 only as rollback
   if 16.3.2 cannot pass the real image-runtime proof.
3. Add a path-aware job to `.github/workflows/pipeline.yml` that, for changes capable
   of affecting the web runtime (root `package.json`, `pnpm-lock.yaml`,
   `pnpm-workspace.yaml`, `apps/web/**`, or any `packages/**` path), builds
   `apps/web/Dockerfile`, starts the image,
   verifies the container stays running, and requires HTTP 2xx. It must always clean
   up and must not use secrets or publish the smoke image.
4. Make `merge-gate` wait for that job and fail closed when it fails. Keep a cheap
   success/no-op path for irrelevant plan/docs-only diffs to avoid unnecessary image
   builds without creating a bypass for root manifests or the lockfile.
5. Add deterministic `voc110` workflow tests for path selection, build/run/HTTP and
   cleanup commands, and merge-gate dependency. Update the development workflow doc.
6. Run applicable commands and record results in `t00-evidence.md`:
   - VOC-110 foundation tests;
   - a local or exact-SHA CI Docker build/start/HTTP proof using the real Dockerfile;
   - deploy-staging regression foundation tests touched by the diff;
   - `pnpm validate` if `apps/web/` or `packages/` changed;
   - `bash scripts/governance/validate-governance.sh` when required for changed paths;
   - `git diff --check`.

### Explicitly out of scope for this task

- Live post-merge `deploy-staging` proof (T01).
- Changing `operational-failure-monitoring.yml`, `deploy-staging.yml`, or the
  benign-cancel classifier (VOC-094).
- Weakening deploy concurrency, health checks, OAuth guards, or core-loop ordering.
- Production deploy or scheduled-synthetics changes unless T00 proves a shared
  regression requiring explicit scope expansion (default: out of scope).

## VOC-110-T01 — Record live verification that deploy-staging succeeds on develop

- Requirement source: issue #911 remediation outcome; `VOC-110-D00`
- Acceptance criteria: `VOC-110-AC-03`, `VOC-110-AC-04`
- Tests: `VOC-110-TEST-06`
- Evidence: `VOC-110-EV-01` (`t01-evidence.md`)
- Live-evidence contract: `.karsift/live-evidence/VOC-110-T01.yaml` (operator-owned;
  governed reconcile per `docs/operations/live-evidence.md`)
- Status: pending — depends on `VOC-110-T00`
- Automation ownership: operator

### Required work

1. After T00 merges to `develop`, enter the operator-owned waiting state. Through
   repository-controlled observe/reconcile (not implementer Actions access), confirm
   a `deploy-staging` run triggered by push to `develop` for a HEAD SHA that contains
   the T00 fix completes with conclusion `success` and job `deploy to staging`
   success. Record only allowlisted run metadata (run URL/ID, conclusion, duration,
   head SHA) — no secrets.
2. Confirm no **new** open issue exists with marker
   `<!-- operational-failure:deploy-staging:failure -->` beyond issue #911.
3. Note whether issue #911 can close under normal roster closure after verification.

### Explicitly out of scope for this task

- Code changes (T00 owns all fixes).
- Implementer-owned Actions dispatch or log inspection.
- Manual issue closure outside the governed roster path.

## Task ordering notes

- T00 blocks T01: live proof requires the fix on `develop` before staging deploy can
  succeed.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
