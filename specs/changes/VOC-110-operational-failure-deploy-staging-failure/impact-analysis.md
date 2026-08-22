# VOC-110 — Impact Analysis

## Security and privacy

Remediation is driven by privacy-safe Actions metadata and read-only runtime diagnosis.
Evidence files and issues contain bounded metadata only — no SSH transcripts, session
cookies, OAuth codes, repository secrets, or user identifiers.

If the fix touches authentication, cookies, or OAuth verification paths, preserve
existing fail-closed guards from VOC-084 and VOC-088. No new secrets are introduced
by this package.

## Protected and operational surfaces

- **Primary write surface:** `.github/workflows/pipeline.yml` gains a path-aware
  Docker build/start/HTTP job and merge-gate dependency. The smoke image is local to
  the runner, is not published, receives no repository secrets, and is always cleaned
  up with a run-unique container name.
- **Preserve unchanged:** `.github/workflows/deploy-staging.yml`, its SSH deployment
  path and secret consumers, and
  `apps/web/tests/staging-e2e/core-loop.staging.spec.ts`. Health, OAuth-start, and
  core-loop steps stay fail-closed and ordered after deployment.
- **Isolation:** no server topology, network, database, deploy-user, Kuma, or secret
  boundary changes are authorized.

## Data and migrations

Default: no schema migration. Deploy-staging may apply existing migrations on the
staging host as today; this package does not authorize new migration files unless
T00 discovers an unrelated defect requiring a separate governed package.

Staging database mutations remain limited to existing idempotent seed steps invoked
by deploy-staging.

## Analytics and accessibility

No analytics changes unless a UI fix incidentally touches analytics instrumentation
(record in task PR if so).

Accessibility impact follows any route or component fix required for the core-loop
journey; no standalone accessibility scope unless the chosen fix demands it.

## Risks, dependencies, and evidence

- `VOC-110-R00`: **Staging web outage** — run 32566405628 deployed a web container
  that restarts and leaves the public staging web host at HTTP 502. Staging API and
  production remain healthy and isolated. Mitigation: T00 image/runtime fix + T01
  green deploy proof.
- `VOC-110-R01`: **Mis-scoped dependency revert** — reverting all eight Dependabot
  bumps without log evidence could hide a smaller fix. Mitigation: `VOC-110-D05` and
  AC-01/TEST-01 require explicit package-level traceability.
- `VOC-110-R02`: **False fix on environmental staging fault** — SSH/host issues
  unrelated to PR #859 could recur. Mitigation: T00 must record causal link; T01 live
  proof on the next push-triggered deploy.
- `VOC-110-R03`: **False-negative path bypass** — an incomplete matcher could skip
  runtime validation for a deployable-artifact change. Mitigation: root manifests,
  `pnpm-lock.yaml`, `apps/web/**`, and relevant shared packages are explicitly tested;
  ambiguous runtime-affecting paths select the Docker job rather than the no-op path.
- `VOC-110-R04`: **False-positive merge block or wasted CI** — building the image for
  unrelated plan/docs changes would waste runner time, while flaky polling could block
  safe merges. Mitigation: deterministic negative path tests, bounded startup polling,
  process-state checks, and unconditional cleanup; actual build/start failure remains
  fail-closed.
- `VOC-110-DEP-00`: Issue #911, run 32566405628, and sanitized read-only runtime
  diagnosis (resolved: web health poll failed due Next.js 16.3.1 standalone output).
- `VOC-110-DEP-01`: VOC-032/VOC-050/VOC-095 deploy-staging building blocks (resolved
  at drafting as predecessors).
- `VOC-110-EV-00`: `t00-evidence.md` — root cause, failing step, fix commands/results.
- `VOC-110-EV-01`: `t01-evidence.md` — post-fix green `deploy-staging` run metadata.
