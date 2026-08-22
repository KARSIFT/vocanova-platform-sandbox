# VOC-110 — Restore staging web and gate the shipped container runtime: Specification

## Objective and requirement source

Remediate the operational failure recorded in
[GitHub issue #911](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/911):
`deploy-staging` ended `failure` after merge of Dependabot PR #859 to `develop`.

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004 plan-review / adopt path.

Primary evidence (issue #911, Actions metadata/logs, public probes, and read-only
container diagnosis for run 32566405628):

| Item | Value |
|------|-------|
| Workflow | `deploy-staging` |
| Run | [#364](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32566405628), push-triggered 2026-08-22 09:59 UTC |
| Head branch / SHA | `develop` @ `f25e4ccf5fc28dcc5b14a438fbdc4f93e5c53a46` |
| Conclusion | `failure` (workflow wall clock ~8m 28s) |
| Job | `deploy to staging` (~8m 23s, exit code 1) |
| Failing step | `Poll staging.vocanova.site/` after API health passed |
| Live impact | Staging web HTTP 502; staging API and production healthy/isolated |
| Runtime cause | Next.js 16.3.1 standalone artifact omitted an `@swc/helpers` ESM module on Node 24 |
| Trigger commit | Merge PR **#859** — Dependabot `minor-and-patch` npm group (eight updates) |
| Public artifacts | Two Docker build attestations produced during the run |
| Public annotations | Process exit code 1; Playwright report/test-results paths missing on failure upload; benign Go cache restore warning |
| Issue origin | VOC-088-T02 `operational-failure-monitoring.yml` sanitized issue body |

Confirmed root-cause bounds:

- **Not VOC-094:** job ran ~8m with conclusion `failure`, not `cancelled` with
  zero jobs and a concurrency-supersession annotation.
- **Not VOC-095:** total duration is well below the 40-minute job timeout; public
  signals point past image build/push rather than an unbounded Playwright install stall.
- **Post-build runtime failure:** both images built and the staging API health poll
  passed. The public web poll failed while `vocanova-web` restarted.
- **Dependency cause:** PR #859 moved Next.js 16.3.0 to 16.3.1. The runtime error
  matches upstream issue #97358 exactly. Stable 16.3.2 contains the 16.3 backport
  from upstream #97372/#97453.
- **Preventive-control gap:** repository validation ran `next build` but never
  started the standalone production image, so all pre-merge checks could pass while
  the deployable artifact was unable to boot.

## Scope and non-goals

In scope:

1. Record the confirmed failing step and sanitized failure class in
   `t00-evidence.md` (no secrets, SSH transcript, session cookies, OAuth state,
   tokens, personal data, or complete application logs).
2. Upgrade `next` and `@next/eslint-plugin-next` together from 16.3.1 to stable
   16.3.2. Pin 16.3.0 only as the documented rollback if 16.3.2 cannot satisfy the
   container-runtime proof.
3. Add a merge-gating CI job for relevant web/dependency changes that builds the
   real `apps/web/Dockerfile`, starts the image, verifies the container remains
   running, and requires HTTP 2xx from its root route. A source `next build` or
   string-only fixture is not sufficient evidence.
4. Preserve fail-closed deploy semantics: no `continue-on-error`, no skipped health
   checks, no weakened OAuth or core-loop ordering (VOC-050, VOC-084, VOC-088).
5. Add deterministic workflow tests proving the runtime job runs for relevant paths,
   safely no-ops for irrelevant plan/docs-only diffs, cleans up its container, and
   is a dependency of merge-gate.
6. Live verification (T01): after T00 merges to `develop`, record a `deploy-staging`
   run for a revision containing the fix that reaches conclusion `success`.

Non-goals / explicitly excluded:

- Weakening or removing the staging core-loop gate, OAuth-start check, health polls,
  or deploy concurrency posture (VOC-094 `queue: max` remains).
- Changing operational-failure observer behavior (VOC-088 worked as designed by opening
  issue #911).
- Reverting Dependabot routing to `main` or bypassing governed integration for future
  dependency updates.
- Reverting the other seven package updates from PR #859.
- Modifying production deploy semantics, signup policy, Kuma inventory, or migrations
  unless T00 proves a separate defect requiring its own package.
- Self-adoption / self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R3** (`.github/workflows/pipeline.yml`, with R1/R2
  application manifest and lockfile changes).
- **Measured path floor at drafting:** **R3**. Not R4 unless a task touches
  `scripts/governance/*`.
- Protected areas: `.github/workflows/deploy-staging.yml`, staging SSH deploy path,
  `apps/web/tests/staging-e2e/core-loop.staging.spec.ts`, repository secrets consumed
  by deploy-staging (read-only use in CI only).
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate.

## Decisions

`VOC-110-D00`: Run 32566405628 is an **actionable deploy-staging failure** (job
started, exit code 1, ~8m duration), not a benign concurrency supersession or a
Playwright-install timeout class failure.

`VOC-110-D01`: The failing step is `Poll staging.vocanova.site/`; the product image
failed before OAuth/core-loop execution because Next.js 16.3.1 standalone output
omitted an SWC helper required on Node 24.

`VOC-110-D02`: Evidence and fixes stay within VOC-088 sanitization boundaries: bounded
metadata in issues and package evidence; job logs inform T00 but are not copied into
issue bodies or committed evidence files.

`VOC-110-D03`: The fix must preserve VOC-050's post-deploy core-loop gate unless T00
proves the gate assertion is wrong — in that case adjust tests or product behavior,
do not skip the gate.

`VOC-110-D04`: The repair target is Next.js 16.3.2, which contains the upstream
16.3 backport. Next.js 16.3.0 is the rollback only if the repaired production image
cannot boot and serve HTTP.

`VOC-110-D05`: Dependabot PR #859 is context, not automatic blame. The task PR must
record whether the fix is a direct consequence of a listed package bump, a latent
defect exposed by timing, or an unrelated staging host issue.

`VOC-110-D06`: A deployable-artifact regression requires deployable-artifact
verification. The CI gate must exercise the real Dockerfile/runtime boundary and
must block merge-gate when that check fails.

## Open questions for the reviewing human

1. Accept proposed **R3** for the `.github/workflows/pipeline.yml` merge-gate change,
   or raise the exact task risk during adoption if semantic escalation applies.
2. Confirm the CI cost-control path matcher remains fail-closed for root manifests,
   `pnpm-lock.yaml`, `apps/web/**`, and shared package changes that can affect the
   standalone artifact.

## Data, migrations, analytics, and accessibility

- No application schema migration unless T00 proves an unrelated data defect requiring
  a separate governed package.
- No intentional staging database mutation beyond existing deploy seed/migrate steps.
- Possible product UI fix if core-loop failure is a rendering or routing regression.
- No analytics change — evidence-backed non-applicability unless a route fix
  incidentally touches analytics (record in task PR if so).
- Accessibility impact follows any UI fix; no standalone a11y scope unless required
  by the chosen fix (record in task PR).
