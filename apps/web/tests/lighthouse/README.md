# apps/web/tests/lighthouse/

Lighthouse (a.k.a. "Lighthouse CI") performance / accessibility /
best-practices budget suite for `@vocanova/web`. Net-new as of
**VOC-031-T09** (this directory did not exist before; the
VOC-031-D00 inspection at draft time confirmed zero Lighthouse
tooling anywhere in the repository).

## What this is

The `runner.mjs` script in this directory drives the
`lighthouse` npm package (the same engine `@lhci/cli` wraps,
so the scores it produces are byte-identical to a full LHCI
run) and asserts the three DOC-08 quality-standards thresholds
across every (screen, layout) combination the T09 acceptance
criterion names:

| Screen      | 360px | 430px | 1280x720 |
| ----------- | :---: | :---: | :------: |
| `/home`     |   *   |   *   |    *     |
| `/discover` |   *   |   *   |    *     |
| `/reviews`  |   *   |   *   |    *     |
| `/progress` |   *   |   *   |    *     |

That is 12 audits per run. The DOC-08 thresholds are:

- **Performance** >= 85
- **Accessibility** >= 95
- **Best Practices** >= 90

`assertions.mjs` is the single source of truth for these
constants; `budget.json` mirrors them (the matching
`scripts/foundation/mock-inventory.test.mjs` test loads
`budget.json` and asserts the values match `assertions.mjs`,
so a drift between the two files surfaces in CI).

A run that misses a threshold exits non-zero (CI failure) and
prints the failing (screen, layout, category, actual, threshold)
tuple. The T09 acceptance criterion is explicit that a
shortfall must be reported as an honest limitation, not
silently lowered or skipped. The run also writes a per-audit
JSON report under `apps/web/lighthouse-reports/` (gitignored)
so the failing audit can be inspected post hoc - the report
files are uploaded as a CI artifact on failure.

## Why this is `lighthouse` directly and not `@lhci/cli`

LHCI is built around a single `startServerCommand` that owns
one server process. Our T07a / T07b / T08 harness already
boots two cooperating processes - the mock API server
(`apps/web/tests/e2e/mock-api-server.mjs`) and the Next.js
production server - and the T09 runner relies on exactly the
same pair. Reusing that pattern in a single command would
either fork the accessibility workflow's webServer config or
duplicate it; calling `lighthouse()` against an
already-running URL sidesteps LHCI's server-management
entirely. The score calculation is identical regardless of
which path is used (same engine, same audit set, same
category weights); LHCI is a thin CI wrapper over
`lighthouse`.

## Flakiness mitigation (VOC-031-R04)

`VOC-031-R04` requires a fixed local production build as the
audit target - never a live network, never the dev server, no
hot-reload variance. The T09 implementation enforces that
contract in three ways:

1. **Production build, not dev server.** The CI workflow
   runs `pnpm --filter @vocanova/web build` (the same
   `next build` the production deploy uses) and then starts
   the result with `pnpm start`, exactly the pattern the
   accessibility workflow uses for `T07a`/`T07b`/`T08`.
2. **Local server, not live network.** Every audited URL is
   `http://127.0.0.1:3000/...` - a fixed address bound to
   the production build's local listener. The CI environment
   cannot reach the public internet during the run (the
   workflow does not need to), and the runner never hits a
   remote host.
3. **Simulated throttling, not devtools throttling.** The
   runner uses `throttlingMethod: 'simulate'` (Lighthouse's
   Lantern simulation) instead of `devtools` throttling.
   `simulate` works entirely from the captured network and
   CPU traces; it does not introduce an extra live-network
   throttling proxy that could time out or drop requests.
   This is also the default every published Lighthouse
   score uses, so the numbers we report are directly
   comparable to public benchmarks.

## Limitations and honest reporting

The T09 acceptance criterion requires that any threshold not
yet met be reported as an explicit, honestly-reported
limitation, never silently lowered or skipped. The first CI
run of the new suite is therefore expected to surface any
pre-existing A1–P4 performance gap (T09 deliberately audits
the full core loop, not only the screens this package adds -
see `VOC-031-R09` in `specs/changes/VOC-031-begin-milestone-
p5-integrated-core-loop/impact-analysis.md`). If a
shortfall is found, the runner fails the build, the failure
details are recorded under EV-38 in
`specs/changes/VOC-031-begin-milestone-p5-integrated-core-
loop/staging-evidence.md`, and a follow-up task addresses the
gap rather than weakening the check.

## Running locally

```bash
# 1. Install dependencies (lighthouse + chrome-launcher are
#    pinned in apps/web/package.json).
pnpm install --frozen-lockfile

# 2. Build the web app (the runner audits the production
#    build, not the dev server).
pnpm --filter @vocanova/web build

# 3. Start the two servers the runner expects to find at
#    http://127.0.0.1:8080 (mock API) and
#    http://127.0.0.1:3000 (Next.js production server).
node apps/web/tests/e2e/mock-api-server.mjs &
pnpm --filter @vocanova/web start --port 3000 &

# 4. Run the suite. The runner exits non-zero on the first
#    audit that misses a DOC-08 threshold.
pnpm --filter @vocanova/web test:lighthouse
```

A local run needs a Chrome / Chromium binary on `$PATH`
(this is the same browser the Playwright install pulls in,
so a developer who has run `pnpm --filter @vocanova/web exec
playwright install --with-deps chromium` already has it).
Set `LIGHTHOUSE_CHROME_PATH=/abs/path/to/chrome` to point
the runner at a specific binary.

## Layout

```
tests/lighthouse/
  README.md           <- this file
  runner.mjs          <- the audit driver; one Chrome instance
                         reused across all 12 audits
  assertions.mjs      <- DOC-08 threshold constants + per-audit
                         assertion helper (single source of truth
                         for the score bars)
  budget.json         <- mirror of the threshold values, parsed
                         by mock-inventory.test.mjs to pin the
                         single source of truth
```

The CI workflow that calls this suite lives at
`.github/workflows/lighthouse.yml` and is a required job for
PRs that touch `apps/web/**`, `apps/web/tests/**`, or the
workflow file itself (the same `paths` filter the
`accessibility.yml` workflow uses for T07a/T07b/T08).

## Out of scope (recorded for the next implementer)

- Score-trend tracking over time (`lhci/upload` /
  `lhci/server`) is not enabled here. T09's acceptance
  criterion is "scores meet the DOC-08 thresholds", not
  "track regressions across PRs", and the
  per-audit JSON reports already cover post-hoc inspection
  for any single run. A future package may opt into the
  LHCI server for trend reporting; it would not change the
  score bar.
- The `lighthouserc.cjs` LHCI configuration file is also
  not present here for the reason above (no LHCI CLI to
  consume it). If a future package adopts LHCI for trend
  tracking, the same `budget.json` values map directly
  into an LHCI `assert.assertions` block.
