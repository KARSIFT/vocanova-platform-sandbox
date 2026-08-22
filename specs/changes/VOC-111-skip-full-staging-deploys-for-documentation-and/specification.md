# VOC-111 — Skip full staging deploys for documentation and evidence-only pushes: Specification

## Objective and requirement source

Stop unnecessary full `deploy-staging` runs on `develop` when merged changes cannot
affect the staging runtime image or deploy bundle, while preserving fail-closed
selection for every runtime/deployment-relevant path.

**Requirement source:** [GitHub issue #920](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/920).

**This draft package does not adopt or authorize itself**; adoption remains a separate
A-004 plan-review / adopt path.

### Confirmed problem evidence (issue #920)

| Item | Value |
|------|-------|
| Workflow | `deploy-staging` |
| Trigger | Every `push` to `develop` (no path filter today) |
| Plan-only merge | `86df6779` → run `32568473144` |
| Roster-only merge | `60822aa5` → run `32568622178` |
| Evidence-only merge | PR #917 → run `32572863842` (~3m39s; unchanged runtime inputs) |
| Stale comment | Header claims docs-only deploy is a near-no-op (cached layers) |
| Side effect | Unrelated pre-VOC-110 staging failures replayed on non-runtime merges |

## Scope and non-goals

### In scope

1. Add **push-only** path selection to `.github/workflows/deploy-staging.yml` so
   documentation, governed change-package material (`specs/**`), and package evidence
   carriers do not start the full deploy job.
2. Preserve **fail-closed** selection for at minimum:
   - Application code (`apps/**`)
   - Shared packages (`packages/**`)
   - Infrastructure/deploy assets (`infra/**`)
   - Every repository-root file (`*`), including workspace manifests, lockfiles,
     Docker/build configuration, and future root-level runtime inputs
   - Staging post-deploy gate inputs (`tests/staging-e2e/**`)
   - The deploy workflow and its selector tests (`.github/workflows/deploy-staging.yml`,
     `scripts/foundation/voc111-deploy-staging-paths.test.mjs`)
3. Leave **`workflow_dispatch`** behavior unchanged (always eligible to run the full
   workflow for manual retry/redeploy).
4. Leave staging host layout, secrets sync, database, Docker networks, deploy-user
   isolation, shared-edge nginx, health/OAuth/core-loop gates, concurrency posture
   (VOC-094 `queue: max`), and operational-failure observer semantics unchanged.
5. Add deterministic positive and negative selector tests, including root-file and
   shared-package cases, so a runtime-affecting path cannot silently bypass deployment.
6. Update stale documentation/comments that describe docs-only deploys as a near-no-op.
7. A governed docs-only fixture task (T01), followed by operator-owned live
   verification (T02) that its known integration push SHA produces **no**
   `deploy-staging` workflow run. The fixture and observation are separate so the
   evidence never claims facts about its own future merge.

### Non-goals / explicitly excluded

- Production deploy filtering (`.github/workflows/deploy-production.yml`).
- Release-promotion policy or develop→main automation changes.
- Manual server mutation on the staging host.
- Weakening health checks, OAuth guards, core-loop ordering, concurrency queueing, or
  operational-failure observer behavior.
- Changing merge-gating container runtime checks in `pipeline.yml` (VOC-110) except
  where a shared path-pattern helper is intentionally reused without weakening either
  gate.
- Self-adoption / self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R3** (`.github/workflows/deploy-staging.yml`).
- **Measured path floor at drafting:** **R3**. Not R4 unless a task touches
  `scripts/governance/*`.
- Protected areas: staging SSH deploy path, repository secrets consumed by
  deploy-staging, `tests/staging-e2e/core-loop.staging.spec.ts`, operational-failure
  observer wiring.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate.

## Decisions

`VOC-111-D00`: Push-triggered `deploy-staging` on `develop` must use an explicit
**runtime/deploy allowlist** (fail-closed). A denylist-only approach is insufficient
because new runtime paths could silently skip deployment.

`VOC-111-D01`: Documentation (`docs/**`), governed package material (`specs/**`),
and other evidence-only carriers under this repository's change-package conventions
must **not** trigger push deploys. A merge that changes **only** paths outside the
allowlist must produce **no** `deploy-staging` workflow run.

`VOC-111-D02`: `workflow_dispatch` remains a full manual retry/redeploy path and is
**not** subject to push path selection.

`VOC-111-D03`: Recommended allowlist (implementer may refine naming but not scope):

| Class | Paths / patterns |
|-------|------------------|
| Repository-root inputs | `*` (root files only), so current and future root-level build/runtime inputs fail closed; this intentionally also selects rare root documentation edits |
| Application | `apps/**` |
| Shared packages | `packages/**` |
| Deploy bundle / host assets | `infra/**` |
| Staging e2e gate | `tests/staging-e2e/**` |
| Workflow + selector tests | `.github/workflows/deploy-staging.yml`, `scripts/foundation/voc111-deploy-staging-paths.test.mjs` |

`VOC-111-D04`: When a push matches **no** allowlisted path, GitHub must not schedule
the deploy job at all (preferred over a no-op job that still consumes runner time and
can interact with concurrency queueing).

`VOC-111-D05`: Deterministic tests in T00 prove both:

- **Negative:** representative docs/specs/evidence-only file lists do **not** select deploy.
- **Positive:** representative `apps/**`, `packages/**`, `infra/**`, current and
  future repository-root filenames, and workflow/test edits **do** select deploy.

`VOC-111-D06`: T01 is a separately reviewed docs-only fixture merge. Live
verification (T02) uses operator-owned metadata only after that merge supplies a
stable integration SHA. Absence of a workflow run is recorded in `t02-evidence.md`;
governed live-evidence reconcile (VOC-097) expects success runs and therefore does
**not** apply to the negative case.

## Data, migrations, analytics, and accessibility

None. CI/workflow scheduling only; no database, product analytics, or UI accessibility
surface changes.

## Security, privacy, and authorization

No new secrets. No change to which secrets are written on deploy. Evidence and tasks
must not include logs, credentials, session values, OAuth data, cohort values, or
personal data.

## Open questions

None at drafting time. Every root file is deliberately selected. If implementation
discovers a nested path that affects the deploy bundle but is outside the drafted
allowlist, T00 must extend the allowlist and tests in the same task PR rather than
deferring silently.
