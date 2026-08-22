# VOC-111 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01**.

## VOC-111-T00 — Add fail-closed push selection, selector tests, and doc updates

- Requirement source: issue #920; `VOC-111-D00`–`D06`
- Acceptance criteria: `VOC-111-AC-00` through `VOC-111-AC-04`
- Tests: `VOC-111-TEST-00` through `VOC-111-TEST-06`, `VOC-111-TEST-08`, `VOC-111-TEST-09`,
  applicable deploy-staging regression tests from VOC-084, VOC-088, and VOC-094
- Evidence: `VOC-111-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record in `t00-evidence.md` the issue #920 evidence table (runs `32568473144`,
   `32568622178`, `32572863842`; example SHAs; no path filter today). Do not copy
   secrets, SSH transcripts, cookies, OAuth state, tokens, or personal data.
2. Implement **push-only** path selection on `.github/workflows/deploy-staging.yml`
   using the explicit runtime/deploy allowlist in `VOC-111-D03`. Prefer GitHub's native
   `on.push.paths` allowlist when it can express the required patterns; if root-file
   detection requires a companion check, document the approach in evidence and keep
   selection fail-closed.
3. Ensure merges that touch **only** non-allowlisted paths (for example `docs/**`,
   `specs/**` plan/roster/evidence carriers, `.karsift/**` lesson files) do **not**
   schedule the deploy workflow on push.
4. Ensure allowlisted paths continue to trigger the **full** existing deploy job
   unchanged when selected, including build/push, SSH deploy, health/OAuth checks,
   core-loop gate, concurrency (`staging-deploy`, `queue: max`), and observer-visible
   failure semantics.
5. Leave `workflow_dispatch` unchanged (full manual retry/redeploy; not gated by push
   path selection).
6. Add `scripts/foundation/voc111-deploy-staging-paths.test.mjs` with deterministic
   positive and negative fixtures, including:
   - docs-only and specs/evidence-only negative cases;
   - `apps/web/**`, `apps/api/**`, `packages/**`, `infra/**`, root manifest/lockfile,
     `tests/staging-e2e/**`, and `.github/workflows/deploy-staging.yml` positive cases;
   - a regression case proving the selector test file itself is allowlisted.
7. Update the stale deploy-staging header comment and, if needed,
   `docs/operations/11-devops-and-ci-cd.md` so docs/package/evidence-only pushes are
   described as **skipped**, not near-no-op cached deploys.
8. Run applicable commands and record results in `t00-evidence.md`:
   - `node --test scripts/foundation/voc111-deploy-staging-paths.test.mjs`;
   - applicable deploy-staging foundation tests (`voc084`, `voc088`, `voc094`, etc.);
   - `bash scripts/governance/validate-governance.sh` when required for changed paths;
   - `git diff --check`.

### Explicitly out of scope for this task

- Live post-merge absence proof (T01).
- Production deploy filtering or release-promotion changes.
- Weakening health checks, OAuth guards, core-loop ordering, or concurrency posture.
- Changing operational-failure observer behavior except as required by unchanged
  failure semantics on **selected** pushes.

## VOC-111-T01 — Record live verification that docs/evidence-only pushes skip deploy-staging

- Requirement source: issue #920 required outcome item 7; `VOC-111-D01`, `VOC-111-D06`
- Acceptance criteria: `VOC-111-AC-01`
- Tests: `VOC-111-TEST-07`
- Evidence: `VOC-111-EV-01` (`t01-evidence.md`)
- Status: pending — depends on `VOC-111-T00`
- Automation ownership: operator

### Required work

1. After T00 merges to `develop`, verify push selection with operator-owned metadata
   only (not implementer Actions access). Use a merge to `develop` whose changed-file
   set includes **only** documentation, governed package material under `specs/**`,
   and/or evidence carriers — for example this package's own `t01-evidence.md` update
   or another governed docs/evidence-only task PR.
2. Record in `t01-evidence.md` the integration push SHA, merge timestamp, and an
   allowlisted Actions query result showing **zero** `deploy-staging` workflow runs
   whose `head_sha` equals that push and whose `event` is `push` on `develop`.
3. Confirm the verification does not weaken or fabricate deployment evidence. Do not
   paste logs, secrets, or personal data.
4. Note that governed live-evidence reconcile (VOC-097) is **not** used for this
   negative case because no success run exists to observe; operator metadata in
   `t01-evidence.md` is the acceptance carrier.

### Explicitly out of scope for this task

- Code changes (T00 owns all workflow/test/doc edits).
- Implementer-owned Actions dispatch or log inspection.
- Proving production deploy behavior.

## Task ordering notes

- T00 blocks T01: push selection must be on `develop` before live absence proof.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
