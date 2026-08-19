# VOC-090 — Fix scheduled-synthetics cancellation from staging core-journey timeout: Specification

## Objective and requirement source

Remediate the operational failure recorded in
[GitHub issue #759](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/759):
the hourly `scheduled-synthetics` workflow cancelled because
`synthetic.staging.authenticated-core-journey` exceeded its GitHub Actions job
timeout.

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004 plan-review / adopt path.

Primary evidence (issue #759 + public run page for run 32271016931):

| Item | Value |
|------|-------|
| Workflow | `scheduled-synthetics` |
| Run | [#22](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32271016931), schedule-triggered 2026-08-19 15:35 UTC |
| Conclusion | `cancelled` (workflow wall clock ~30m 19s) |
| Failing job | `synthetic.staging.authenticated-core-journey` |
| Public annotation | exceeded maximum execution time **30m0s** |
| Other jobs | staging OAuth, production OAuth, production journey, production route sweep — no timeout error on public page |
| Issue origin | VOC-088-T02 `operational-failure-monitoring.yml` sanitized issue body |

Drafting-time repo read of the failing job's steps:

1. Checkout
2. SSH refresh of reserved synthetic review state (`seed-synthetic-smoke-user.sh`)
3. Mint staging session
4. Setup Node + enable pnpm
5. **`pnpm install --frozen-lockfile`** (no cache)
6. **`playwright install --with-deps chromium`** (no cache)
7. Run `core-loop.staging.spec.ts` via `playwright.staging.config.ts`
   (`timeout: 240_000` ms, `MAX_REVIEW_CARDS = 8`)

The job declares `timeout-minutes: 30`. `deploy-staging.yml` runs the same
dependency install + journey pattern with **`timeout-minutes: 40`** for the
overall deploy job.

`synthetic.staging.authenticated-core-journey` in `synthetics.yaml` declares
`timeout_seconds: 600` (10 minutes), which understates the end-to-end CI job
budget once setup is included.

## Scope and non-goals

In scope:

1. Reduce cold-start overhead for the scheduled staging core-journey job by
   caching pnpm store / workspace install artifacts and Playwright browser
   binaries (exact mechanism is an implementer shape choice; must be
   deterministic and testable from committed workflow structure).
2. Align the GitHub Actions job `timeout-minutes` for
   `staging-authenticated-core-journey` with the corrected end-to-end budget
   (setup + SSH seed + mint + journey). The corrected value must be **at least**
   as generous as `deploy-staging.yml`'s proven post-deploy gate for the same
   journey unless caching reduces setup below 30 minutes with measured margin.
3. Update `infra/monitoring/synthetics.yaml`
   `synthetic.staging.authenticated-core-journey.timeout_seconds` to reflect the
   corrected job budget (convert minutes to seconds; document that this field
   represents the workflow job wall clock, not Playwright's per-test timeout).
4. Extend deterministic tests (`scripts/foundation/voc090-*.test.mjs` or extend
   `voc086-scheduled-synthetics.test.mjs`) to assert:
   - caching steps exist for pnpm and Playwright in the staging core-journey job;
   - job timeout is ≥ registry `timeout_seconds` / 60 (rounded consistently);
   - job timeout is ≥ `playwright.staging.config.ts` journey timeout plus a
     documented minimum setup reserve (SSH + mint + install).
5. Live verification: a post-merge `workflow_dispatch` of
   `scheduled-synthetics.yml` (full suite or staging core-journey only) completes
   success within the declared budget; record scrubbed run URL in `t01-evidence.md`.

Non-goals / explicitly excluded:

- Weakening `core-loop.staging.spec.ts` assertions, reducing `MAX_REVIEW_CARDS`,
  adding Playwright retries, or otherwise masking a real staging regression.
- Changing production synthetic behavior, OAuth policy, Kuma inventory, or
  `error-monitoring.yml`.
- Modifying `operational-failure-monitoring.yml` (the observer worked as designed).
- Application code or schema changes unless a measured staging performance defect
  outside CI setup is proven in T00 evidence (open question — see below).
- Self-adoption / self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R3** (`.github/workflows/`, `infra/monitoring/`).
- **Measured path floor at drafting:** **R3**. Not R4 unless a task touches
  `scripts/governance/*`.
- Protected areas: `.github/workflows/scheduled-synthetics.yml`,
  `infra/monitoring/synthetics.yaml`, `apps/web/tests/staging-e2e/`,
  `apps/web/playwright.staging.config.ts`.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate.

## Decisions

`VOC-090-D00`: Run 32271016931 cancelled because job
`synthetic.staging.authenticated-core-journey` hit GitHub Actions'
`timeout-minutes: 30` wall clock. The remediation target is CI budget and setup
efficiency, not reclassifying cancellation as success.

`VOC-090-D01`: The fix must preserve the functional contract of
`synthetic.staging.authenticated-core-journey`: SSH seed of the reserved synthetic
account, session mint, and the full authenticated staging core-loop Playwright
journey against real staging.

`VOC-090-D02`: Primary mitigation is **dependency caching** for pnpm and Playwright
on scheduled runs, plus **timeout alignment** between the workflow job,
`synthetics.yaml`, and the known journey timeout. Increasing timeout alone without
caching is insufficient if cold installs continue to consume most of the budget on
every hourly tick.

`VOC-090-D03`: If implementer evidence from the failing run's job log (read at
implementation time, not copied into the issue) shows the Playwright journey itself
exceeded `playwright.staging.config.ts`'s 240s timeout, a separate application or
test-performance fix may be required in addition to CI caching. That path is out of
scope for T00 unless T00 evidence proves it; it would become a follow-up package
rather than silent scope expansion here.

`VOC-090-D04`: Shared setup refactors that touch other jobs in
`scheduled-synthetics.yml` are allowed only when they preserve existing production
job behavior and pass all VOC-086 scheduled-synthetics deterministic tests.

## Open questions for the reviewing human

1. Accept proposed **R3**, or raise in writing if the adopting human treats
   first correction of hourly behavioral synthetics as R4 operational risk.
2. Confirm acceptable cache mechanism: `actions/setup-node` with pnpm cache,
   explicit `actions/cache` for Playwright browsers, or a small reusable composite
   shared with `deploy-staging.yml` (would widen T00 file touch list — record in
   the task PR if chosen).
3. If job logs show journey slowness rather than install slowness (`VOC-090-D03`),
   decide whether to extend this package or open a separate staging-performance
   investigation.

## Data, migrations, analytics, and accessibility

- No application schema migration.
- No database mutation.
- No product UI change — evidence-backed non-applicability.
- No analytics change — evidence-backed non-applicability.
- No accessibility change — evidence-backed non-applicability.
