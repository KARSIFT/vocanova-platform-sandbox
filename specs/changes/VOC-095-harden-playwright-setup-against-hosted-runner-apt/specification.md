# VOC-095 — Harden Playwright setup against hosted-runner apt mirror stalls: Specification

## Objective and requirement source

Remediate the CI reliability failure recorded in
[GitHub issue #792](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/792):
hosted GitHub Actions runners can stall indefinitely inside
`playwright install --with-deps chromium` when the Ubuntu package mirror is slow or
unavailable, causing workflow timeouts after otherwise-successful deploy convergence.

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004 plan-review / adopt path.

Primary evidence (issue #792 + public run metadata for run 32299315180):

| Item | Value |
|------|-------|
| Workflow | `deploy-staging` |
| Run | [32299315180](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32299315180), attempts 1 and 2 |
| Head SHA | `fdde12daeccf521a5a8be171294eba79a138717b` |
| Failure phase | `playwright install --with-deps chromium` waiting on apt |
| Job outcome | Cancelled at workflow timeout after application convergence succeeded |
| Related transient | Accessibility workflow hit the same mirror stall; exact-SHA rerun succeeded |

Drafting-time repo read:

- Inline install command in four workflows:
  `pnpm --filter @vocanova/web exec playwright install --with-deps chromium`
- Only `scheduled-synthetics.yml` restores `~/.cache/ms-playwright` via `actions/cache`
  before install (key `playwright-chromium-${{ runner.os }}-${{ hashFiles('pnpm-lock.yaml') }}`).
- `infra/monitoring/scheduled-synthetics.mjs` enforces cache-before-install ordering for
  the staging authenticated core-journey job.
- Lighthouse reuses the Playwright Chromium binary via shell-expanded
  `LIGHTHOUSE_CHROME_PATH` (see `.karsift/lessons.md` — must remain in `run:` body).

This is distinct from VOC-094 (deploy concurrency queue supersession) and from true
application or deploy defects surfaced by core-loop test failures.

## Scope and non-goals

In scope:

1. Add `infra/scripts/install-playwright-chromium.sh` implementing a bounded,
   retry-limited install contract: browser download (CDN) plus system dependencies
   (apt) with explicit per-attempt timeout, limited retries, and binary verification.
2. Replace inline `playwright install --with-deps chromium` in:
   - `.github/workflows/accessibility.yml`
   - `.github/workflows/lighthouse.yml`
   - `.github/workflows/deploy-staging.yml` (staging core-loop job steps)
   - `.github/workflows/scheduled-synthetics.yml`
3. Add Playwright browser cache restore to workflows that lack it, using the same cache
   path and key scheme as `scheduled-synthetics.yml`, **before** the install script runs.
4. Update `infra/monitoring/scheduled-synthetics.mjs` (and its foundation tests) so
   staging core-loop validation asserts the shared script + cache ordering.
5. Add `scripts/foundation/voc095-playwright-install.test.mjs` locking script structure,
   workflow wiring, timeout/retry constants, and fail-closed semantics.
6. Update `apps/web/tests/e2e/README.md` install documentation to reference the shared
   contract (local operators may still run the script or equivalent commands).
7. Live verification (T02): after T01 merges to `develop`, record a `deploy-staging` run
   for the latest integration commit that reaches `success` including the staging
   core-loop Playwright gate.

Non-goals / explicitly excluded:

- Skipping accessibility, Lighthouse, staging core-loop, or scheduled core-loop execution
  on install failure (`continue-on-error`, optional jobs, or mirror-bypass flags that
  disable browser tests).
- Weakening workflow `timeout-minutes` values (deploy-staging 40, accessibility/lighthouse
  30, scheduled synthetics budgets unchanged).
- Changing Playwright test suites, application code, migrations, signup policy, secrets,
  staging/production isolation, deploy ordering, or OAuth controls.
- Modifying operational-failure observer classification (VOC-088/VOC-094).
- Self-adoption / self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R3** (`.github/workflows/`, `infra/scripts/`).
- **Measured path floor at drafting:** **R3**. Not R4 unless a task touches
  `scripts/governance/*`.
- Protected areas: all four workflows listed above; `infra/scripts/`; staging core-loop
  gate inside `deploy-staging.yml`.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate.

## Decisions

`VOC-095-D00`: Run 32299315180 timed out because the unbounded `--with-deps` apt path
blocked the job after deploy convergence; remediation targets install boundedness and
cache reuse, not application or SSH deploy logic.

`VOC-095-D01`: One shared shell script (`infra/scripts/install-playwright-chromium.sh`)
is the repository-managed install contract. Workflows invoke it instead of inline
`playwright install --with-deps`.

`VOC-095-D02`: Browser binaries cache under `~/.cache/ms-playwright` with key
`playwright-chromium-${{ runner.os }}-${{ hashFiles('pnpm-lock.yaml') }}` and restore
**before** install in all four workflows.

`VOC-095-D03`: System dependency installation uses explicit bounded per-attempt timeout
and a small fixed retry count (draft proposes 3 attempts, 120 s per-attempt apt/deps
timeout, 30 s inter-attempt sleep — implementer may tune within AC-01 bounds). After
retries exhaust, the script exits non-zero; workflows remain fail-closed.

`VOC-095-D04`: On cache hit, the script still verifies the Chromium binary exists and
runs the deps install path (deps are not cached separately at drafting time). Browser
download may no-op when cache is warm, reducing total install time.

`VOC-095-D05`: Workflow timeouts, deploy step ordering, health checks, session minting,
and core-loop env vars remain unchanged except replacing the install step body and adding
cache restore where missing.

`VOC-095-D06`: Lighthouse `LIGHTHOUSE_CHROME_PATH` resolution stays in the workflow
`run:` step body with shell expansion; the install script must leave the Playwright
default cache layout intact so existing glob resolution continues to work.

`VOC-095-D07`: The install script must not echo secrets, session cookies, or OAuth
state. CI logs may record attempt counts and exit codes only.

## Open questions for the reviewing human

1. Accept proposed **R3**, or raise in writing if the adopting human treats CI browser
   install hardening on deploy-staging as R4 operational risk.
2. Confirm per-attempt deps timeout budget: is **120 s × 3 attempts** acceptable, or
   should deploy-staging allow a higher ceiling (still well below job timeout)?
3. Should T00 add an optional `actions/cache` entry for apt archives (`/var/cache/apt`)
   in addition to browser cache, or defer apt caching to a follow-up if mirror stalls
   persist after browser cache rollout?
4. If T02 cannot capture a mirror-stall reproduction on demand, is evidence of (a) a green
   `deploy-staging` core-loop run on latest `develop` after T01 plus (b) deterministic
   fixture proof of bounded failure behavior sufficient for closure?

## Data, migrations, analytics, and accessibility

- No application schema migration.
- No database mutation.
- No product UI change — evidence-backed non-applicability.
- No analytics change — evidence-backed non-applicability.
- Accessibility **test execution** is in scope only as a CI workflow consumer of the
  hardened install contract; no change to axe scan thresholds or spec coverage.
