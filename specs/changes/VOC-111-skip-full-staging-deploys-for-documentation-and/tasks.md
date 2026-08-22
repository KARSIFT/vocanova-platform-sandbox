# VOC-111 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01 → T02**.

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
   `on.push.paths` allowlist. Select every repository-root file with a root-only
   pattern so a newly introduced root build/runtime input cannot silently bypass
   staging deployment.
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
   - `apps/web/**`, `apps/api/**`, `packages/**`, `infra/**`, representative current
     root manifests/lockfiles and an otherwise-unlisted future root filename,
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

- Governed docs-only fixture merge (T01) and live post-merge absence proof (T02).
- Production deploy filtering or release-promotion changes.
- Weakening health checks, OAuth guards, core-loop ordering, or concurrency posture.
- Changing operational-failure observer behavior except as required by unchanged
  failure semantics on **selected** pushes.

## VOC-111-T01 — Merge a governed docs-only selector fixture

- Requirement source: issue #920 required outcome item 7; `VOC-111-D01`, `VOC-111-D06`
- Acceptance criteria: `VOC-111-AC-01`
- Tests: `VOC-111-TEST-10`
- Evidence: `VOC-111-EV-01` (`t01-evidence.md`)
- Status: pending — depends on `VOC-111-T00`
### Required work

1. After T00 merges to `develop`, update only this package's `t01-evidence.md` into
   the deterministic docs-only fixture carrier. Do not change application,
   infrastructure, workflow, root, shared-package, or staging-e2e paths.
2. Record the fixture PR number and its exact pre-merge head SHA. The future
   integration merge SHA is deliberately left for T02 because it does not exist
   until this task PR merges.
3. Run governance validation and `git diff --check`; independently verify that the
   entire task diff is under `specs/changes/VOC-111-.../` and therefore outside the
   deploy allowlist.
4. Merge through the normal exact-SHA task lifecycle. That merge creates the bounded
   `develop` push which T02 observes; T01 must not claim the post-merge absence result.

### Explicitly out of scope for this task

- Code changes or selector changes (T00 owns them).
- Actions queries or claims that no run exists (T02 owns post-merge observation).
- Production deploy behavior.

## VOC-111-T02 — Record live absence proof for the T01 fixture push

- Requirement source: issue #920 required outcome item 7; `VOC-111-D01`, `VOC-111-D06`
- Acceptance criteria: `VOC-111-AC-01`
- Tests: `VOC-111-TEST-07`
- Evidence: `VOC-111-EV-02` (`t02-evidence.md`)
- Status: pending — depends on `VOC-111-T01`
- Automation ownership: operator

### Required work

1. After T01 merges, resolve its exact `develop` integration push SHA and merge
   timestamp using operator-owned repository metadata only.
2. Confirm T01's merged changed-file set is limited to the declared docs/specs
   fixture surface.
3. Query Actions metadata and record in `t02-evidence.md` that exactly zero
   `deploy-staging` runs have `event=push`, `head_branch=develop`, and `head_sha`
   equal to the T01 integration SHA. Record only the SHA, timestamp, PR/run count,
   and scrubbed outcome; never logs, credentials, or personal data.
4. Keep the T02 carrier draft until that negative evidence is complete. Do not use
   governed success-run reconciliation: the accepted outcome intentionally has no
   workflow run to observe.

### Explicitly out of scope for this task

- Repository code or workflow changes.
- Implementer-owned Actions access or dispatch.
- Production deploy behavior.

## Task ordering notes

- T00 blocks T01: push selection must be on `develop` before the fixture merge.
- T01 blocks T02: only T01's completed merge supplies a non-circular integration SHA
  whose absence of deploy runs can be observed honestly.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
