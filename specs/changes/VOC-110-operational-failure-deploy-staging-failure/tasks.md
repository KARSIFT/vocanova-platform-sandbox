# VOC-110 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01**.

## VOC-110-T00 — Diagnose and fix deploy-staging failure from run 32566405628

- Requirement source: issue #911; `VOC-110-D00`–`D05`
- Acceptance criteria: `VOC-110-AC-00`, `VOC-110-AC-01`, `VOC-110-AC-02`
- Tests: `VOC-110-TEST-00` through `VOC-110-TEST-05`, applicable deploy-staging
  regression tests from VOC-084, VOC-088, and VOC-095
- Evidence: `VOC-110-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Read run 32566405628 job logs for `deploy to staging` at implementation time
   (`gh run view`, Actions REST jobs API, or equivalent authorized access). Record
   in `t00-evidence.md`:
   - failing step name and conclusion;
   - sanitized failure class (e.g. core-loop assertion, OAuth check, health poll,
     SSH/migration, image build, Playwright install);
   - whether PR #859's dependency diff plausibly caused the failure.
   Do not copy secrets, SSH output, session cookies, OAuth state, tokens, or personal
   data into evidence.
2. Compare public metadata already recorded at drafting time (duration, Docker build
   artifacts, Playwright artifact-upload warnings) with the log-identified failing step.
3. Implement the smallest correct fix per `VOC-110-D01`:
   - **Application path:** `apps/web/` and/or `packages/` when staging runtime/UI
     regressed after the Dependabot merge.
   - **Test path:** `apps/web/tests/staging-e2e/` when the product is correct but
     the journey assertion is stale.
   - **Harness path:** `infra/scripts/` when a verify/install/mint helper is wrong.
   - **Workflow path:** minimal `deploy-staging.yml` change only when miswiring is
     proven unrelated to spurious dependency blame.
4. If the fix is reverting or pinning a specific Dependabot bump, document which
   package and why in the task PR (do not silently revert the entire PR #859 group).
5. Extend deterministic tests (`scripts/foundation/voc110-*.test.mjs` and/or
   existing voc084/voc088/voc095 deploy-staging suites) to lock the fix.
6. Run applicable commands and record results in `t00-evidence.md`:
   - new/extended VOC-110 foundation tests;
   - deploy-staging regression foundation tests touched by the diff;
   - `pnpm validate` if `apps/web/` or `packages/` changed;
   - `bash scripts/governance/validate-governance.sh` when required for changed paths;
   - `git diff --check`.

### Explicitly out of scope for this task

- Live post-merge `deploy-staging` proof (T01).
- Changing operational-failure-monitoring.yml or benign-cancel classifier (VOC-094).
- Weakening deploy concurrency, health checks, OAuth guards, or core-loop ordering.
- Production deploy or scheduled-synthetics changes unless T00 proves a shared
  regression requiring explicit scope expansion (default: out of scope).

## VOC-110-T01 — Record live verification that deploy-staging succeeds on develop

- Requirement source: issue #911; `VOC-110-D04`
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
